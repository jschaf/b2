package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/cloudflare/tableflip"
	"github.com/jschaf/b2/pkg/git"
	"github.com/jschaf/b2/pkg/livereload"
	"github.com/jschaf/b2/pkg/markdown"
	"github.com/jschaf/b2/pkg/markdown/compiler"
	"go.uber.org/zap"
)

type server struct {
	*http.ServeMux
	port     string
	stopC    chan struct{}
	once     sync.Once
	upgrader *tableflip.Upgrader
	logger   *zap.SugaredLogger
}

func newServer(port string) (*server, error) {
	s := new(server)
	s.ServeMux = http.NewServeMux()
	s.port = port
	s.once = sync.Once{}
	s.stopC = make(chan struct{})

	if l, err := newShortDevLogger(); err != nil {
		return nil, fmt.Errorf("failed to create logger: %w", err)
	} else {
		pid := os.Getpid()
		s.logger = l.Sugar().With("pid", pid)
	}

	return s, nil
}

func (s *server) Serve() error {
	srv := http.Server{
		Handler: s.ServeMux,
	}
	upg, err := tableflip.New(tableflip.Options{
		UpgradeTimeout: time.Second * 5,
	})
	if err != nil {
		return fmt.Errorf("failed to create upgrader")
	}
	s.upgrader = upg
	if err := s.upgrader.Ready(); err != nil {
		return fmt.Errorf("upgrader not ready: %w", err)
	}

	ln, err := s.upgrader.Listen("tcp", "localhost:"+s.port)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}
	defer ln.Close()
	return srv.Serve(ln)
}

func (s *server) UpgradeOnSIGHUP() {
	upgrade := make(chan os.Signal, 1)
	signal.Notify(upgrade, syscall.SIGHUP)
	// We might get multiple upgrade requests.
	for range upgrade {
		s.logger.Info("upgrading because SIGHUP")
		if err := s.upgrader.Upgrade(); err != nil {
			s.logger.Errorf("failed to upgrade: %s", err)
			continue
		}
	}
}

func (s *server) Stop() {
	s.once.Do(func() {
		s.upgrader.Stop()
		close(s.stopC)
	})
	s.logger.Info("server stopped")
	_ = s.logger.Sync()
}

func main() {
	port := "8080"
	server, err := newServer(port)
	if err != nil {
		log.Fatalf("failed to create server: %s", err)
	}
	defer server.Stop()

	lrJSPath := "/dev/livereload.js"
	lrPath := "/dev/livereload"
	lr := livereload.NewWebsocketServer(server.logger.Named("livereload"))
	server.HandleFunc(lrJSPath, livereload.ServeJSHandler)
	server.HandleFunc(lrPath, lr.WebSocketHandler)
	go lr.Start()

	root, err := git.FindRootDir()
	if err != nil {
		server.logger.Errorf("failed to find root dir: %s", err)
		return
	}
	pubDir := filepath.Join(root, "public")
	pubDirHandler := http.FileServer(http.Dir(pubDir))

	lrScript := strings.Join([]string{
		fmt.Sprintf("<script defer src=%s?port=%s&path=%s type='application/javascript'>",
			lrJSPath, port, strings.TrimLeft(lrPath, "/")),
		"</script>",
	}, "")
	server.Handle("/", livereload.NewHTMLInjector(lrScript, pubDirHandler))

	watcher := NewFSWatcher(lr, server.logger)
	if err := watcher.watchDirs(
		filepath.Join(root, "public"),
		filepath.Join(root, "style"),
		filepath.Join(root, "posts"),
		filepath.Join(root, "cmd"),
		filepath.Join(root, "pkg"),
	); err != nil {
		server.logger.Error(err)
		return
	}

	go server.UpgradeOnSIGHUP()

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
		<-c
		server.logger.Info("received quit signal")
		server.Stop()
	}()

	go func() {
		if err := watcher.Start(); err != nil {
			server.logger.Infof("stopping server because watcher error: %s", err)
			server.Stop()
		}
	}()

	go func() {
		if err := server.Serve(); err != nil {
			server.Stop()
		}
	}()

	server.logger.Infof("Serving at http://localhost:%s", port)

	// Compile because it might have changed since last run.
	md := markdown.New()
	c := compiler.New(md)
	server.logger.Debug("compiling all markdown files")
	if err := compiler.CompileEverything(c); err != nil {
		server.logger.Error(err)
		return
	}

	select {
	case <-server.upgrader.Exit():
		server.logger.Debug("upgrader exiting")
	case <-server.stopC:
		server.logger.Debug("server stopping")
	}
}
