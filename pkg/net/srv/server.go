package srv

import (
	"context"
	"errors"
	stdflag "flag"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/jschaf/b2/pkg/errs"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

const (
	defaultShutdownGracePeriod = 30 * time.Second
)

type Server struct {
	Cfg           ServerCfg
	flag          *stdflag.FlagSet
	httpHandler   http.Handler // handler for non-debug HTTP endpoints
	drainHandlers []func()
	stopCh        chan string // handle server stop requests
}

type ServerCfg struct {
	Name     Name
	Version  Version
	HTTPAddr string
	// How long to wait for connections to drain before exiting. Defaults to
	// defaultShutdownGracePeriod.
	ShutdownGracePeriod time.Duration
}

func makeFlagUsage(flag *stdflag.FlagSet) func() {
	return func() {
		_, _ = fmt.Fprintf(os.Stderr, "server: usage\n")
		flag.PrintDefaults()
	}
}

func NewServer(cfg ServerCfg) *Server {
	flag := stdflag.CommandLine
	flag.Usage = makeFlagUsage(flag)

	srvCfg := cfg
	// Set defaults.
	if srvCfg.ShutdownGracePeriod == 0 {
		srvCfg.ShutdownGracePeriod = defaultShutdownGracePeriod
	}
	return &Server{
		Cfg:    srvCfg,
		flag:   flag,
		stopCh: make(chan string),
	}
}

func (s *Server) Init() error {
	slog.Info("initialize server")
	if err := s.initFlags(); err != nil {
		return err
	}
	if s.Cfg.Name == "" {
		return fmt.Errorf("no server name set")
	}
	return nil
}

func (s *Server) initFlags() error {
	if err := s.flag.Parse(os.Args[1:]); err != nil {
		return fmt.Errorf("parse server flags: %w", err)
	}
	if s.flag.NArg() != 0 {
		return fmt.Errorf("unparsed server flags: %s", strings.Join(s.flag.Args(), ", "))
	}
	return nil
}

// SetHTTPHandler registers an HTTP handler to serve all non-debug HTTP
// requests.  You may only register a single handler; to serve multiple URLs,
// use an http.ServeMux.
func (s *Server) SetHTTPHandler(h http.Handler) {
	if s.httpHandler != nil {
		panic("attempt to set HTTP handler more than once")
	}
	s.httpHandler = h
}

// AddDrainHandler registers a function to be called when the server begins
// draining. It is not safe to call while ListenAndServe is running. If your
// function blocks, it will interfere with a clean shutdown, so don't block.
//
// To cancel select statements, share a channel between the drain handler and
// your loop:
//
//	drainCh := make(chan struct{})
//	server.AddDrainHandler(func() { close(drainCh) })
//	for {
//		select {
//			case <-drainCh:
//			// draining
//			case <- whatever:
//			// whatever
//		}
//	}
func (s *Server) AddDrainHandler(f func()) {
	s.drainHandlers = append(s.drainHandlers, f)
}

// Stop stops the server. If the server hasn't yet started, it won't be started.
func (s *Server) Stop() {
	s.stopCh <- "called server.Stop()"
}

// newLnCloser closes the listener, ignoring already closed errors.
func newLnCloser(ln net.Listener) func() error {
	return func() error {
		if err := ln.Close(); err != nil {
			var e *net.OpError
			if errors.As(err, &e) && e.Err.Error() == "use of closed network connection" {
				// Ignore closed network connections, we already closed them.
				return nil
			}
			return err
		}
		return nil
	}
}

func (s *Server) listenAndServe() (mErr error) {
	if s.httpHandler == nil {
		return fmt.Errorf("no http hander set; use server.SetHTTPHandler()")
	}

	// Check if Stop() was already called.
	select {
	case reason := <-s.stopCh:
		slog.Info("shutdown requested before server start", "reason", reason)
	default: // continue
	}

	doneCh := make(chan error)
	numServers := 0

	// Start the non-debug HTTP server with customized support for HTTP/2.
	ln, err := net.Listen("tcp", s.Cfg.HTTPAddr)
	if err != nil {
		return fmt.Errorf("listen on HTTP address %s: %w", s.Cfg.HTTPAddr, err)
	}
	defer errs.Capture(&mErr, newLnCloser(ln), "close HTTP listener")
	http2Server := &http2.Server{}
	httpServer := &http.Server{
		Handler: h2c.NewHandler(s.httpHandler, http2Server),
	}

	slog.Info("listening http", "addr", ln.Addr().String())
	numServers++
	go func() {
		// Ignore http.ErrServerClosed because it's returned on graceful shutdown.
		if err := httpServer.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
			doneCh <- fmt.Errorf("http serve: %w", err)
		} else {
			doneCh <- nil
		}
	}()

	// Once any server requests a stop or stops serving due to an error, start to
	// shutdown everything.
	select {
	case reason := <-s.stopCh:
		slog.Info("shutdown requested", "reason", reason, "servers_remaining", numServers)
	case doneErr := <-doneCh:
		numServers--
		slog.Error("server unexpectedly errored", "error", doneErr, "servers_remaining", numServers)
	}
	shutdownCtx, cancel := context.WithTimeout(context.Background(), s.Cfg.ShutdownGracePeriod)
	defer cancel()
	go func() {
		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			slog.Error("http server shutdown", "error", err)
		}
	}()

	// Shutdown remaining servers.
	for numServers > 0 {
		select {
		case <-shutdownCtx.Done():
			return fmt.Errorf("server shutdown exceeded deadline, servers_remaining=%d: %w", numServers, shutdownCtx.Err())
		case err := <-doneCh:
			numServers--
			// err is nil if shutdown occurred cleanly.
			slog.Info("server exited during shutdown", "error", err, "servers_remaining", numServers)
		}
	}
	slog.Info("all servers exited")
	return nil
}

func (s *Server) ListenAndServe() error {
	// Stop server on SIGINT or SIGTERM.
	go func() {
		sigCh := make(chan os.Signal)
		signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)
		sig := <-sigCh
		name := sig.String()
		slog.Info("got signal", "signal", name)
		signal.Stop(sigCh)
		close(sigCh)
		s.stopCh <- name
	}()

	termMsg := []byte("clean shutdown")
	termLog := "/tmp/server-termination"
	srvErr := s.listenAndServe()
	if srvErr != nil {
		slog.Error("server error", "error", srvErr)
	}
	if err := os.WriteFile(termLog, termMsg, 0o666); err != nil {
		slog.Info("write termination log", "path", termLog, "error", err)
	}

	return srvErr
}
