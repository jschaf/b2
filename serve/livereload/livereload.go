// Package livereload provides the handlers necessary to run a LiveReload
// server.
//
// The protocol follows http://livereload.com/api/protocol/.
package livereload

import (
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

type LiveReload struct {
	upgrader      websocket.Upgrader
	connPublisher *connPub
}

func New() *LiveReload {
	return &LiveReload{
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		connPublisher: newConnPub(),
	}
}

// Handler is a http.HandlerFunc to handle LiveReload websocket interaction.
// The goroutine running the handler is left open.
func (lr *LiveReload) WebSocketHandler(w http.ResponseWriter, r *http.Request) {
	ws, err := lr.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("failed to upgrade HTTP to websocket: %s", err)
		return
	}
	c := newConn(ws)
	lr.connPublisher.attach <- c
	defer func() { lr.connPublisher.detach <- c }()
	log.Printf("Start LiveReload connection")
	c.start()
	log.Printf("Finish LiveReload connection")
}

// ServeJS is a http.HandlerFunc to serve the livereload.js script.
func (lr *LiveReload) ServeJSHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/javascript")
	w.Write(liveReloadJS)
}

func (lr *LiveReload) Start() {
	lr.connPublisher.start()
}

func (lr *LiveReload) Shutdown() {
	log.Printf("Shutting down livereload")
	for c := range lr.connPublisher.conns {
		log.Printf("Shutting down livereload connection")
		c.closeWithCode(websocket.CloseNormalClosure)
	}
}

// ReloadFile instructs all registered LiveReload clients to reload path.
func (lr *LiveReload) ReloadFile(path string) {
	lr.connPublisher.publish <- newReloadResponse(path)
}

// Alert instructs all registered LiveReload clients to display an alert
// with msg.
func (lr *LiveReload) Alert(msg string) {
	lr.connPublisher.publish <- newAlertResponse(msg)
}
