package livereload

import "github.com/gorilla/websocket"

// connPub publishes messages to all attached LiveReload websocket connections.
type connPub struct {
	// All connections registered on this connPub.
	conns map[*conn]struct{}
	// Messages to publish to all livereload client connections.
	publish chan interface{}
	// LiveReload client connections to attach.
	attach chan *conn
	// LiveReload client connections to detach.
	detach chan *conn
}

func newConnPub() *connPub {
	return &connPub{
		conns:   make(map[*conn]struct{}),
		publish: make(chan interface{}),
		attach:  make(chan *conn),
		detach:  make(chan *conn),
	}
}

func (h *connPub) start() {
	for {
		select {
		case c := <-h.attach:
			h.conns[c] = struct{}{}

		case c := <-h.detach:
			delete(h.conns, c)
			c.closeCode(websocket.CloseNormalClosure)

		case m := <-h.publish:
			for c := range h.conns {
				select {
				case c.send <- m:
				default:
					// If the connection is not accepting data either it's closed or
					// congested. Force the connection to reconnect if it's still alive.
					delete(h.conns, c)
					c.closeCode(websocket.CloseTryAgainLater)
				}
			}
		}
	}
}
