package livereload

import (
	"fmt"
	"github.com/gorilla/websocket"
)

// connPub publishes messages to all attached LiveReload websocket connections.
type connPub struct {
	// All connections registered on this connPub.
	conns map[*conn]struct{}
	// Messages to publish to all LiveReload client connections.
	publish chan interface{}
	// LiveReload client connections to attach.
	attach chan *conn
	// LiveReload client connections to detach.
	detach chan *conn
	stop   chan struct{}
}

func newConnPub() *connPub {
	return &connPub{
		conns:   make(map[*conn]struct{}),
		publish: make(chan interface{}),
		attach:  make(chan *conn),
		detach:  make(chan *conn),
		stop:    make(chan struct{}),
	}
}

func (p *connPub) start() {
	fmt.Println("starting conn publisher")
	for {
		select {
		case <-p.stop:
			return

		case c := <-p.attach:
			fmt.Println("attaching connection")
			p.conns[c] = struct{}{}

		case c := <-p.detach:
			fmt.Println("detaching connection")
			delete(p.conns, c)
			c.closeWithCode(websocket.CloseNormalClosure,
				"detaching connection")

		case m := <-p.publish:
			fmt.Println("publishing to all connections")
			for c := range p.conns {
				select {
				case c.send <- m:
				default:
					// If the connection is not accepting data either it's closed or
					// congested. Force the connection to reconnect if it's still alive.
					delete(p.conns, c)
					c.closeWithCode(websocket.CloseTryAgainLater,
						"connection is congested")
				}
			}
		}
	}
}

func (p *connPub) shutdown() {
	close(p.stop)
	for c := range p.conns {
		delete(p.conns, c)
		c.closeWithCode(websocket.CloseNormalClosure,
			"shutting down server")
	}
}
