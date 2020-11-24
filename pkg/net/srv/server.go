package srv

import (
	"context"
	stdflag "flag"
	"fmt"
	"github.com/jschaf/b2/pkg/errs"
	"github.com/jschaf/b2/pkg/log"
	"go.uber.org/zap"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
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
	logLevel      zap.AtomicLevel
	l             *zap.Logger
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

func NewServer(cfg ServerCfg, logLvl zap.AtomicLevel, l *zap.Logger) *Server {
	flag := stdflag.CommandLine
	flag.Usage = makeFlagUsage(flag)

	srvCfg := cfg
	// Set defaults.
	if srvCfg.ShutdownGracePeriod == 0 {
		srvCfg.ShutdownGracePeriod = defaultShutdownGracePeriod
	}
	logger := l.With(zap.String("name", string(srvCfg.Name)), zap.String("version", string(srvCfg.Version)))
	return &Server{
		Cfg:      srvCfg,
		flag:     flag,
		stopCh:   make(chan string),
		logLevel: logLvl,
		l:        logger,
	}
}

func (s *Server) Init() error {
	s.l.Info("initialize server")
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
			if e, ok := err.(*net.OpError); ok && e.Err.Error() == "use of closed network connection" {
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
		s.l.Info("shutdown requested before server start", zap.String("reason", reason))
	default: // continue
	}

	doneCh := make(chan error)
	numServers := 0
	fieldServers := zap.Intp("servers_remaining", &numServers)

	// Start the non-debug HTTP server with customized support for HTTP/2.
	ln, err := net.Listen("tcp", s.Cfg.HTTPAddr)
	if err != nil {
		return fmt.Errorf("listen on HTTP address %s: %w", s.Cfg.HTTPAddr, err)
	}
	defer errs.Capturing(&mErr, newLnCloser(ln), "close HTTP listener")
	http2Server := &http2.Server{}
	httpServer := &http.Server{
		Handler:  h2c.NewHandler(s.httpHandler, http2Server),
		ErrorLog: zap.NewStdLog(s.l.Named("http")),
	}

	s.l.Info("listening http", zap.String("server", "http"), zap.String("addr", ln.Addr().String()))
	numServers++
	go func() {
		// Ignore http.ErrServerClosed because it's returned on graceful shutdown.
		if err := httpServer.Serve(ln); err != nil && err != http.ErrServerClosed {
			doneCh <- fmt.Errorf("http serve: %w", err)
		} else {
			doneCh <- nil
		}
	}()

	// Once any server requests a stop or stops serving due to an error, start to
	// shutdown everything.
	select {
	case reason := <-s.stopCh:
		s.l.Info("shutdown requested", zap.String("reason", reason), fieldServers)
	case doneErr := <-doneCh:
		numServers--
		s.l.Error("server unexpectedly errored", zap.Error(doneErr), fieldServers)
	}
	shutdownCtx, cancel := context.WithTimeout(context.Background(), s.Cfg.ShutdownGracePeriod)
	defer cancel()
	go func() {
		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			s.l.Error("http server shutdown", zap.Error(err))
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
			s.l.Info("server exited during shutdown", zap.Error(err), fieldServers)
		}
	}
	s.l.Info("all servers exited")
	return nil
}

func (s *Server) ListenAndServe() error {
	// Stop server on SIGINT or SIGTERM.
	go func() {
		sigCh := make(chan os.Signal)
		signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)
		sig := <-sigCh
		name := sig.String()
		s.l.Info("got signal", zap.String("signal", name))
		signal.Stop(sigCh)
		close(sigCh)
		s.stopCh <- name
	}()

	termMsg := []byte("clean shutdown")
	termLog := "/tmp/server-termination"
	srvErr := s.listenAndServe()
	if srvErr != nil {
		s.l.Error("server error", zap.Error(srvErr))
	}
	// TODO: Export Event like Datadog.Event instead of write to file.
	if err := ioutil.WriteFile(termLog, termMsg, 0666); err != nil {
		s.l.Info("write termination log", zap.String("path", termLog), zap.Error(err))
	}

	log.Flush(s.l)
	return srvErr
}
