package main

import (
	"context"
	"fmt"
	"github.com/jschaf/b2/pkg/dirs"
	"github.com/jschaf/b2/pkg/errs"
	"github.com/jschaf/b2/pkg/git"
	"github.com/jschaf/b2/pkg/logs"
	"github.com/jschaf/b2/pkg/sites"
	"go.uber.org/zap/zapcore"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/cloudflare/tableflip"
	"github.com/jschaf/b2/pkg/livereload"
	"go.uber.org/zap"
)

const (

	// Time after which an upgrade is considered a failure.
	upgradeTimeout = 30 * time.Second
	// How long to allow the server to handle existing connections.

	serverShutdownTimeout = 25 * time.Second
)

type server struct {
	*http.ServeMux
	port string
	once sync.Once
	l    *zap.SugaredLogger
}

func newServer(port string, l *zap.Logger) *server {
	s := new(server)
	s.ServeMux = http.NewServeMux()
	s.port = port
	s.once = sync.Once{}
	pid := os.Getpid()
	s.l = l.Sugar().With(zap.Int("pid", pid))
	return s
}

func run(l *zap.Logger) (mErr error) {
	port := "8080"
	server := newServer(port, l)
	pubDir := dirs.PublicMemfs

	if err := dirs.CleanDir(pubDir); err != nil {
		return fmt.Errorf("clean public dir: %w", err)
	}

	lrJSPath := "/dev/livereload.js"
	lrPath := "/dev/livereload"
	lr := livereload.NewServer(server.l.Named("livereload"))
	server.HandleFunc(lrJSPath, lr.ServeJSHandler)
	server.HandleFunc(lrPath, lr.WebSocketHandler)
	go lr.Start()

	pubDirHandler := http.FileServer(http.Dir(pubDir))

	lrScript := strings.Join([]string{
		fmt.Sprintf("<script defer src=%s?port=%s&path=%s type='application/javascript'>",
			lrJSPath, port, strings.TrimLeft(lrPath, "/")),
		"</script>",
	}, "")
	server.Handle("/", lr.NewHTMLInjector(lrScript, pubDirHandler))

	watcher := NewFSWatcher(pubDir, lr, server.l)
	root := git.MustFindRootDir()
	if err := watcher.watchDirs(
		filepath.Join(root, dirs.Cmd),
		filepath.Join(root, dirs.Pkg),
		filepath.Join(root, dirs.Posts),
		filepath.Join(root, dirs.Static),
		filepath.Join(root, dirs.Style),
		filepath.Join(root, dirs.TIL),
	); err != nil {
		return fmt.Errorf("watch dirs: %w", err)
	}

	httpServer := http.Server{
		Handler: server.ServeMux,
	}

	upg, err := tableflip.New(tableflip.Options{
		UpgradeTimeout: upgradeTimeout,
		PIDFile:        pidFilePath(),
	})
	if err != nil {
		return fmt.Errorf("failed to create upgrader")
	}
	defer upg.Stop()

	// Upgrade on SIGHUP. The SIGHUP is sent by the watcher when any Go file
	// changes. On a Go file change, the watcher recompiles the server binary.
	// After the binary is compiled, the watcher sends SIGHUP to the current
	// process which tells the upgrader to start shutting down this process.
	go func() {
		upgrade := make(chan os.Signal, 1)
		signal.Notify(upgrade, syscall.SIGHUP)
		// We might get multiple upgrade requests. If an upgrade fails, like when
		// the replacement go server fails to compile, we want to keep trying to
		// upgrade future requests.
		for {
			select {
			// If the upgrader is done, stop listening for notifications.
			case <-upg.Exit():
				server.l.Infof("stopped listening for upgrade notifications")
				signal.Stop(upgrade)
				return
			case <-upgrade:
				server.l.Infof("upgrading because got signal %s", syscall.SIGHUP)
				if err := upg.Upgrade(); err != nil {
					server.l.Errorf("upgrade server: %s", err)
				}
			}
		}
	}()

	// upgrader.listen must be called before ready.
	ln, err := upg.Listen("tcp", "localhost:"+port)
	if err != nil {
		return fmt.Errorf("server listen: %w", err)
	}
	defer errs.CapturingErr(&mErr, func() error {
		if err := ln.Close(); err != nil {
			// Ignore closed network connection errors. That's expected because the
			// upgrader ceded control to the replacement process.
			if err != http.ErrServerClosed {
				return nil
			}
			if e, ok := err.(*net.OpError); ok {
				// Ignore certain types of network connection errors that don't matter
				// because we're closing the connection.
				if e.Temporary() || e.Timeout() || e.Err.Error() == "use of closed network connection" {
					return nil
				}
			}
			return err
		}
		return nil
	}, "close server upgrader listener")

	go func() {
		if err := httpServer.Serve(ln); err != nil {
			if err != http.ErrServerClosed {
				server.l.Infof("serve error: %s", err)
			}
			upg.Stop()
		}
	}()

	if err := upg.Ready(); err != nil {
		return fmt.Errorf("upgrader not ready: %w", err)
	}

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
		<-c
		server.l.Info("received quit signal")
		upg.Stop()
	}()

	go func() {
		if err := watcher.Start(); err != nil {
			server.l.Infof("server watcher start error: %s", err)
			upg.Stop()
		}
	}()

	server.l.Infof("serving at http://localhost:%s", port)

	// Rebuild in case content changed since last run.
	if err := sites.Rebuild(pubDir, server.l.Desugar()); err != nil {
		return fmt.Errorf("rebuild site: %w", err)
	}

	<-upg.Exit()
	upg.Stop()
	server.l.Infof("upgrader exited")

	// Set a deadline on server.Shutdown after upg.Exit() is closed. No new
	// upgrades can be performed if server has not Shutdown below.
	shutdownC := make(chan struct{})
	go func() {
		select {
		case <-time.After(serverShutdownTimeout):
			server.l.Errorf("graceful shutdown timed out")
			os.Exit(1)
		case <-shutdownC:
			// If we get here, the server successfully shutdown.
			return
		}
	}()

	// Wait for connections to drain.
	if err := httpServer.Shutdown(context.Background()); err != nil {
		server.l.Errorf("server shutdown: %w", err)
	}
	close(shutdownC)
	watcher.Stop()
	// Defer so it shutdown log entry appears after other defers.
	defer server.l.Info("server shutdown")
	return nil
}

func pidFilePath() string {
	return fmt.Sprintf("/var/run/user/%d/b2_server.pid", os.Getuid())
}

func isFirstServer() (first bool, cleanup func() error, mErr error) {
	pid := os.Getpid()
	pgid, err := syscall.Getpgid(pid)
	nop := func() error { return nil }
	if err != nil {
		return false, nop, fmt.Errorf("register pgid, get process group ID: %w", err)
	}
	lockfile := fmt.Sprintf("/dev/shm/b2_serve.%d.pgid", pgid)
	if _, err := os.Stat(lockfile); err == nil {
		// Path exists.
		return false, nop, nil
	} else if os.IsNotExist(err) {
		// Path doesn't exist.
		if err := ioutil.WriteFile(lockfile, nil, 0777); err != nil {
			return false, nop, fmt.Errorf("write pgid lockfile: %w", err)
		}
		cleanup := func() error {
			if err := os.Remove(lockfile); err != nil {
				return fmt.Errorf("clean up lockfile: %w", err)
			}
			return nil
		}
		return true, cleanup, nil
	} else {
		// Unknown file error.
		return false, nop, fmt.Errorf("check pgid lockfile existence: %w", err)
	}
}

// forwardSignals forwards signals to all processes in the same process group
// as this process.
func forwardSignals(log *zap.Logger) (mErr error) {
	pid := os.Getpid()
	l := log.Sugar().With("pid", pid)

	isFirst, cleanup, err := isFirstServer()
	if err != nil {
		return err
	}
	defer errs.CapturingErr(&mErr, cleanup, "cleanup lockfile")
	if !isFirst {
		l.Infof("not forwarding signals because not first server")
		return nil
	}

	l.Infof("forwarding quit signals because is first server")
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGHUP, syscall.SIGTERM)
	for {
		sig := <-c
		l.Infof("forwarding signal to process group: %s", sig)
		// The negative PID kills all processes within the process group started by the PID.
		// https://stackoverflow.com/a/11000554/30900
		switch sig {
		case syscall.SIGHUP:
			// Ignore. The active server handles SIGHUP. Ignore explicitly because
			// the default Go action is to kill this process.
		case syscall.SIGINT, syscall.SIGTERM:
			signal.Stop(c)
			if err := syscall.Kill(-pid, syscall.SIGINT); err != nil {
				return fmt.Errorf("forward %s from PID %d: %w", sig, pid, err)
			}
			return nil
		default:
			panic("unhandled forwarded signal " + sig.String())
		}
	}
	return nil
}

func main() {
	l, err := logs.NewShortDevLogger(zapcore.InfoLevel)
	if err != nil {
		log.Fatalf("create logger: %s", err)
	}
	if err := run(l); err != nil {
		l.Error(err.Error())
	}
	logs.Flush(l)
	if err := forwardSignals(l); err != nil {
		l.Error(err.Error())
	}
	logs.Flush(l)
}
