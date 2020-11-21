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
	serverPIDFileTmpl = "/var/run/user/%d/b2_server.pid"
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
	s.l = l.Sugar()
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
		PIDFile:        fmt.Sprintf(serverPIDFileTmpl, os.Getuid()),
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
		// We might get multiple upgrade requests.
		for range upgrade {
			server.l.Infof("upgrading because got signal %s", syscall.SIGHUP)
			if err := upg.Upgrade(); err != nil {
				server.l.Errorf("upgrade server: %s", err)
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

	// CompileIndex in case content changed since last run.
	if err := sites.Rebuild(pubDir, server.l.Desugar()); err != nil {
		return fmt.Errorf("rebuild site: %w", err)
	}

	<-upg.Exit()
	server.l.Infof("upgrader exited")

	// Set a deadline on exiting the process after upg.Exit() is closed. No new
	// upgrades can be performed if the parent doesn't exit in Shutdown() below.
	time.AfterFunc(serverShutdownTimeout, func() {
		server.l.Errorf("graceful shutdown timed out")
		os.Exit(1)
	})

	// Wait for connections to drain.
	if err := httpServer.Shutdown(context.Background()); err != nil {
		server.l.Errorf("server shutdown: %w", err)
	}

	return nil
}

func main() {
	l, err := logs.NewShortDevLogger(zapcore.InfoLevel)
	if err != nil {
		log.Fatalf("create logger: %s", err)
	}
	pid := os.Getpid()
	l = l.With(zap.Int("pid", pid))
	if err := run(l); err != nil {
		l.Error(err.Error())
	}
	l.Info("server shutdown")
	logs.Flush(l)
}
