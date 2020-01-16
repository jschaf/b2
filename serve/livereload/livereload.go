// Package livereload provides the handlers necessary to run a LiveReload
// server.
//
// The protocol follows http://livereload.com/api/protocol/.
package livereload

import (
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/jschaf/b2/serve/paths"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
	"sync"
)

var (
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	// Global register of LiveReload connections.
	connPublisher = newConnPub()
)

// Handler is a http.HandlerFunc to handle LiveReload websocket interaction.
func Handler(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("failed to upgrade HTTP to websocket: %s", err)
		return
	}
	c := newConn(ws)
	connPublisher.attach <- c
	defer func() { connPublisher.detach <- c }()
	c.start()
}

var (
	liveReloadOnce  sync.Once
	liveReloadJS    []byte
	liveReloadJSErr error
)

// ServeJS is a http.HandlerFunc to serve the livereload.js script.
func ServeJS(w http.ResponseWriter, _ *http.Request) {
	liveReloadOnce.Do(func() {
		liveReloadJS, liveReloadJSErr = loadJS()
	})

	if liveReloadJSErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(liveReloadJSErr.Error()))
		return
	}

	w.Header().Set("Content-Type", "application/javascript")
	w.Write(liveReloadJS)
}

func loadJS() ([]byte, error) {
	root, err := paths.FindRootDir()
	if err != nil {
		return nil, err
	}
	jsPath := filepath.Join(root, "third_party", "livereload", "livereload.js")
	file, err := ioutil.ReadFile(jsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open livereload.js: %w", err)
	}
	return file, nil
}

// ReloadFile instructs all registered LiveReload clients to reload path.
func ReloadFile(path string) {
	connPublisher.publish <- newReloadResponse(path)
}

// Alert instructs all registered LiveReload clients to display an alert
// with msg.
func Alert(msg string) {
	connPublisher.publish <- newAlertResponse(msg)
}
