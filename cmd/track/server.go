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
	"sync"
	"time"

	"github.com/jschaf/jsc/pkg/errs"
	"github.com/jschaf/jsc/pkg/log"
	"github.com/jschaf/jsc/pkg/net/srv"
	"github.com/jschaf/jsc/pkg/process"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

const port = "3355"

type Server struct {
	// Lifecycle context.
	// Calling serverCancel causes all background goroutines to stop. To stop the
	// HTTP server, call Shutdown.
	serverCtx    context.Context
	serverCancel context.CancelFunc
	// Servers
	httpSrv *http.Server
	// Locks
	mu sync.Mutex
}

type ServerOpts struct {
	Cancel context.CancelFunc
}

func InitServer(ctx context.Context, opts ServerOpts) (*Server, error) {
	routeHandler := buildRoutes()
	h2s := &http2.Server{}
	httpSrv := &http.Server{
		Handler: h2c.NewHandler(routeHandler, h2s),
	}

	return &Server{
		serverCtx:    ctx,
		serverCancel: opts.Cancel,
		// Servers
		httpSrv: httpSrv,
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

	url := "http://localhost:" + port
	slog.Info("ready: track server listening", slog.String("url", url))

	// Serve
	err = s.httpSrv.Serve(ln)
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("prod http serve: %w", err)
	}
	slog.Info("track server shut down")
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

	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		AddSource: true,
		Level:     logLevel,
	})))

	slog.Info("start track server", "process.args", os.Args[1:])

	devSrv, err := InitServer(ctx, ServerOpts{
		Cancel: cancel,
	})
	if err != nil {
		return fmt.Errorf("init server: %w", err)
	}

	// If context is done, shutdown.
	go func() {
		<-ctx.Done()
		slog.Debug("stop track server")
		shutdownCtx, shutdownCancel := context.WithTimeout(context.WithoutCancel(ctx), 30*time.Second)
		defer shutdownCancel()
		err := devSrv.Shutdown(shutdownCtx)
		if err != nil {
			slog.Error("shutdown track server", slog.String("error", err.Error()))
		}
	}()

	if err := devSrv.ListenAndServe(); err != nil {
		return fmt.Errorf("listen and serve track server: %w", err)
	}

	return nil
}
