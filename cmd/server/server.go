package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/jschaf/b2/pkg/dirs"
	"github.com/jschaf/b2/pkg/errs"
	"github.com/jschaf/b2/pkg/git"
	"github.com/jschaf/b2/pkg/livereload"
	"github.com/jschaf/b2/pkg/log"
	"github.com/jschaf/b2/pkg/net/srv"
	"github.com/jschaf/b2/pkg/process"
	"github.com/jschaf/b2/pkg/sites"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

const port = "2222"

type Server struct {
	// Server lifecycle context
	// Calling serverCancel causes all background goroutines to stop. To
	// stop the HTTP server, call Shutdown.
	serverCtx    context.Context
	serverCancel context.CancelFunc
	// Servers
	httpSrv       *http.Server
	liveReloadSrv *livereload.LiveReload
	// Locks
	mu sync.Mutex
}

type ServerOpts struct {
	PubDir string
	Cancel context.CancelFunc
}

func InitServer(ctx context.Context, opts ServerOpts) (*Server, error) {
	if err := dirs.CleanDir(opts.PubDir); err != nil {
		return nil, fmt.Errorf("clean public dir: %w", err)
	}

	// Rebuild in case content changed since last run.
	if err := sites.Rebuild(opts.PubDir); err != nil {
		return nil, fmt.Errorf("rebuild site: %w", err)
	}

	// Live reload.
	lr := livereload.NewServer()
	go lr.Start(ctx)

	// File system watcher.
	watcher := NewFSWatcher(opts.PubDir, lr)
	root := git.RootDir()
	if err := watcher.watchDirs(
		filepath.Join(root, dirs.Book),
		filepath.Join(root, dirs.Cmd),
		filepath.Join(root, dirs.Pkg),
		filepath.Join(root, dirs.Posts),
		filepath.Join(root, dirs.Static),
		filepath.Join(root, dirs.Style),
		filepath.Join(root, dirs.TIL),
	); err != nil {
		return nil, fmt.Errorf("watch dirs: %w", err)
	}
	go func() {
		if err := watcher.Start(); err != nil {
			slog.Error("watcher error", "error", err)
			opts.Cancel()
		}
	}()

	// HTTP server.
	routeHandler := buildRoutes(buildRoutesOpts{
		pubDir: opts.PubDir,
		lr:     lr,
	})
	h2s := &http2.Server{}
	httpSrv := &http.Server{
		Handler: h2c.NewHandler(routeHandler, h2s),
	}

	return &Server{
		serverCtx:    ctx,
		serverCancel: opts.Cancel,
		// Servers
		httpSrv:       httpSrv,
		liveReloadSrv: lr,
		// Locks
		mu: sync.Mutex{},
	}, nil
}

func (s *Server) ListenAndServe() (mErr error) {
	if err := s.serverCtx.Err(); err != nil {
		return fmt.Errorf("server context error: %w", err)
	}

	ln, err := net.Listen("tcp", "0.0.0.0:"+port)
	if err != nil {
		return fmt.Errorf("listen to http port: %w", err)
	}
	defer errs.Capture(&mErr, srv.NewListenerCloser(ln), "close http listener")
	slog.Info("ready: dev server listening", slog.String("url", "http://localhost:"+port))

	err = s.httpSrv.Serve(ln)
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("prod http serve: %w", err)
	}
	slog.Info("dev server shut down")
	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	// First, gracefully shutdown the HTTP server.
	err := s.httpSrv.Shutdown(ctx)
	// Then cancel the server context, which should trigger everything else to
	// stop.
	s.serverCancel()
	return err
}

func main() {
	process.RunMain(runMain)
}

func runMain(ctx context.Context) (mErr error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	fset := flag.CommandLine
	logLevel := log.DefineFlags(fset)
	if err := fset.Parse(os.Args[1:]); err != nil {
		return fmt.Errorf("parse flags: %w", err)
	}

	slog.SetDefault(slog.New(log.NewDevHandler(os.Stderr, &slog.HandlerOptions{
		Level: logLevel,
	})))

	slog.Info("start dev server", "process.args", os.Args[1:])

	devSrv, err := InitServer(ctx, ServerOpts{
		PubDir: dirs.Public,
		Cancel: cancel,
	})
	if err != nil {
		return fmt.Errorf("init server: %w", err)
	}

	// If context is done, shutdown.
	go func() {
		<-ctx.Done()
		slog.Debug("stop dev server")
		shutdownCtx, shutdownCancel := context.WithTimeout(context.WithoutCancel(ctx), 30*time.Second)
		defer shutdownCancel()
		err := devSrv.Shutdown(shutdownCtx)
		if err != nil {
			slog.Error("shutdown dev server", slog.String("error", err.Error()))
		}
	}()

	if err := devSrv.ListenAndServe(); err != nil {
		return fmt.Errorf("listen and serve dev server: %w", err)
	}

	return nil
}
