package livereload

import (
	"context"
	"log/slog"

	"github.com/gorilla/websocket"
	"go.uber.org/atomic"
)

type closeReq struct {
	conn *conn
	err  error
}

func newCloseReq(conn *conn, err error) closeReq {
	return closeReq{conn, err}
}

func newCloseError(code int, msg string) *websocket.CloseError {
	return &websocket.CloseError{Code: code, Text: msg}
}

// connPub publishes messages to all attached LiveReload websocket connections.
type connPub struct {
	conns   map[*conn]struct{} // all connections registered on this connPub
	publish chan any           // messages to publish to all LiveReload client connections
	attach  chan *conn         // LiveReload client connections to attach
	detach  chan closeReq      // LiveReload client connections to detach
	stop    chan struct{}
	connSeq *atomic.Int32 // the next ID to use for a connection
}

func newConnPub() *connPub {
	return &connPub{
		conns:   make(map[*conn]struct{}),
		publish: make(chan any),
		attach:  make(chan *conn),
		detach:  make(chan closeReq),
		stop:    make(chan struct{}),
		connSeq: atomic.NewInt32(770),
	}
}

func (p *connPub) start(ctx context.Context) {
	slog.Debug("starting connPub")
	// detachConn unregisters the conn from receiving new messages and closes the
	// websocket connection. Not thread-safe.
	detachConn := func(req closeReq) {
		if _, ok := p.conns[req.conn]; !ok {
			slog.Debug("conn already deleted", "error", req.err)
			return
		}
		delete(p.conns, req.conn)
		slog.Debug("detach conn", "error", req.err)
		req.conn.close(req.err)
	}

main:
	for {
		select {
		case <-ctx.Done():
			break main

		case <-p.stop:
			break main

		case c := <-p.attach:
			p.conns[c] = struct{}{}

		case closeReq := <-p.detach:
			detachConn(closeReq)

		case m := <-p.publish:
			for c := range p.conns {
				select {
				case c.send <- m:
				default:
					// If the connection is not accepting data either it's closed or
					// congested. Force the connection to reconnect if it's still alive.
					detachConn(
						newCloseReq(c, newCloseError(websocket.CloseTryAgainLater, "congested connection")))
				}
			}
		}
	}

	for c := range p.conns {
		detachConn(newCloseReq(
			c, newCloseError(websocket.CloseNormalClosure, "shutting down server")))
	}
}

// runConn runs a LiveReload websocket connection and blocks until the
// connection closes.
func (p *connPub) runConn(ws *websocket.Conn) {
	c := newConn(ws, p.detach)
	if err := c.start(); err != nil {
		slog.Error("start livereload connection", "error", err.Error())
		return
	}
	p.attach <- c
}

func (p *connPub) shutdown() {
	slog.Debug("shutting down conn pub")
	close(p.stop)
}
