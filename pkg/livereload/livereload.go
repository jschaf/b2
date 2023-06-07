// Package livereload provides the handlers necessary to run a LiveReload
// server.
//
// The protocol follows http://livereload.com/api/protocol/.
package livereload

import (
	"bytes"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"

	"net/http"
	"net/http/httptest"
	"strconv"
)

type LiveReload struct {
	upgrader      websocket.Upgrader
	connPublisher *connPub
	logger        *zap.SugaredLogger
}

func NewServer(l *zap.SugaredLogger) *LiveReload {
	return &LiveReload{
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		connPublisher: newConnPub(l.Desugar()),
		logger:        l,
	}
}

// WebSocketHandler is a http.HandlerFunc to handle LiveReload websocket interaction.
// The goroutine running the handler is left open.
func (lr *LiveReload) WebSocketHandler(w http.ResponseWriter, r *http.Request) {
	ws, err := lr.upgrader.Upgrade(w, r, nil)
	if err != nil {
		lr.logger.Infof("failed to upgrade HTTP to websocket: %s", err)
		return
	}
	lr.logger.Debugf("received LiveReload websocket connection")
	lr.connPublisher.runConn(ws)
}

// NewHTMLInjector intercepts all output from the next handler and injects
// a script tag to load the LiveReload script. The script is injected before
// the </head> tag.
func (lr *LiveReload) NewHTMLInjector(scriptTag string, next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		recorder := httptest.NewRecorder()
		next.ServeHTTP(recorder, r)

		headTag := []byte("</head>")
		replacement := []byte("  " + scriptTag + "\n</head>")
		s := bytes.Replace(recorder.Body.Bytes(), headTag, replacement, 1)
		if len(s) == len(recorder.Body.Bytes()) {
			// lr.l.Warnf("did not inject livereload script")
		}
		for k, vs := range recorder.Header() {
			for _, v := range vs {
				w.Header().Add(k, v)
			}
		}
		w.Header().Set("Content-Length", strconv.Itoa(len(s)))
		w.WriteHeader(recorder.Code)
		_, _ = w.Write(s)
	}
}

// ServeJS is a http.HandlerFunc to serve the livereload.js script.
func (lr *LiveReload) ServeJSHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/javascript")
	_, _ = w.Write(liveReloadJS)
}

func (lr *LiveReload) Start() {
	lr.logger.Debug("starting LiveReload")
	lr.connPublisher.start()
}

func (lr *LiveReload) Shutdown() {
	lr.logger.Debug("shutting down livereload")
	lr.connPublisher.shutdown()
}

// ReloadFile instructs all registered LiveReload clients to reload path.
// The path should be absolute if possible.
func (lr *LiveReload) ReloadFile(path string) {
	lr.connPublisher.publish <- newReloadMsg(path)
}

// Alert instructs all registered LiveReload clients to display an alert
// with msg.
func (lr *LiveReload) Alert(msg string) {
	lr.connPublisher.publish <- newAlertResponse(msg)
}
