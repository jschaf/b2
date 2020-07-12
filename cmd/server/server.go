package main

import (
	"fmt"
	"github.com/jschaf/b2/pkg/dirs"
	"github.com/jschaf/b2/pkg/errs"
	"github.com/jschaf/b2/pkg/logs"
	"github.com/jschaf/b2/pkg/sites"
	"go.uber.org/zap/zapcore"
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
	"github.com/jschaf/b2/pkg/css"
	"github.com/jschaf/b2/pkg/git"
	"github.com/jschaf/b2/pkg/livereload"
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

func newServer(port string, l *zap.Logger) *server {
	s := new(server)
	s.ServeMux = http.NewServeMux()
	s.port = port
	s.once = sync.Once{}
	s.stopC = make(chan struct{})
	pid := os.Getpid()
	s.logger = l.Sugar().With("pid", pid)
	return s
}

func (s *server) Serve() (mErr error) {
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
	defer errs.CloseWithErrCapture(&mErr, ln, "close server upgrader")
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

func run(l *zap.Logger) error {
	port := "8080"
	server := newServer(port, l)
	defer server.Stop()
	root, err := git.FindRootDir()
	if err != nil {
		return fmt.Errorf("failed to find root dir: %s", err)
	}
	pubDir := filepath.Join(root, dirs.Public)

	if err := dirs.CleanPubDir(); err != nil {
		return fmt.Errorf("failed to clean public dir: %w", err)
	}

	if err := os.MkdirAll(pubDir, 0755); err != nil {
		return fmt.Errorf("failed to make public dir: %w", err)
	}

	lrJSPath := "/dev/livereload.js"
	lrPath := "/dev/livereload"
	lr := livereload.NewWebsocketServer(server.logger.Named("livereload"))
	server.HandleFunc(lrJSPath, livereload.ServeJSHandler)
	server.HandleFunc(lrPath, lr.WebSocketHandler)
	go lr.Start()

	pubDirHandler := http.FileServer(http.Dir(pubDir))

	lrScript := strings.Join([]string{
		fmt.Sprintf("<script defer src=%s?port=%s&path=%s type='application/javascript'>",
			lrJSPath, port, strings.TrimLeft(lrPath, "/")),
		"</script>",
	}, "")
	server.Handle("/", livereload.NewHTMLInjector(lrScript, pubDirHandler))

	watcher := NewFSWatcher(lr, server.logger)
	if err := watcher.watchDirs(
		filepath.Join(root, dirs.Public),
		filepath.Join(root, dirs.Style),
		filepath.Join(root, dirs.Posts),
		filepath.Join(root, dirs.Cmd),
		filepath.Join(root, dirs.Pkg),
		filepath.Join(root, dirs.Static),
		filepath.Join(root, dirs.Scripts),
	); err != nil {
		return fmt.Errorf("watch dirs: %w", err)
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
			server.logger.Infof("server watcher start error: %s", err)
			server.Stop()
		}
	}()

	go func() {
		if err := server.Serve(); err != nil {
			server.logger.Infof("server serve error: %s", err)
			server.Stop()
		}
	}()

	server.logger.Infof("Serving at http://localhost:%s", port)

	if _, err := css.WriteMainCSS(root); err != nil {
		return fmt.Errorf("write main.css: %w", err)
	}

	// Compile in case content changed since last run.
	if err := sites.Rebuild(server.logger.Desugar()); err != nil {
		return fmt.Errorf("rebuild site: %w", err)
	}

	select {
	case <-server.upgrader.Exit():
		server.logger.Debug("upgrader exiting")
	case <-server.stopC:
		server.logger.Debug("server stopping")
	}
	return nil
}

func main() {
	l, err := logs.NewShortDevLogger(zapcore.InfoLevel)
	if err != nil {
		log.Fatalf("failed to create logger: %s", err)
	}
	if err := run(l); err != nil {
		l.Error(err.Error())
	}
}
