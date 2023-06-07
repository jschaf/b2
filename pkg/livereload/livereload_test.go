package livereload

import (
	"github.com/google/go-cmp/cmp"
	"github.com/gorilla/websocket"
	"github.com/jschaf/b2/pkg/errs"
	"go.uber.org/zap/zaptest"
	"io"

	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestServeJSHandler(t *testing.T) {
	server, lr := newLiveReloadServer(t)
	defer server.Close()
	req := httptest.NewRequest("GET", "http://example.com/livereload.js", nil)
	w := httptest.NewRecorder()
	lr.ServeJSHandler(w, req)

	resp := w.Result()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	defer errs.CapturingT(t, resp.Body.Close, "")

	if !strings.Contains(string(body), "var LiveReload") {
		t.Error("expected LiveReload JS to contain 'var LiveReload'")
	}
}

func TestLiveReload_NewHTMLInjector(t *testing.T) {
	newTag := "<meta foo=qux>"
	replaced := "  " + newTag + "\n</head>"
	tests := []struct {
		name    string
		handler http.Handler
		header  http.Header
		want    string
	}{
		{"only head tag", writeHTMLBody("</head>"), headers(), replaced},
		{"2 tags with head",
			writeHTMLBody("<html><head></head></html>"), headers(),
			"<html><head>" + replaced + "</html>"},
		{"split head tag </he ad>",
			writeHTMLBody("</he", "ad>"), headers(), replaced},
		{"split head tag with prefix",
			writeHTMLBody("<html></he", "ad>"), headers(), "<html>" + replaced},
		{"split head tag multiple writes",
			writeHTMLBody("<html>", "</he", "ad>"), headers(), "<html>" + replaced},
		{"headers preserved",
			writeHTML(headers("FOO", "bar"), "<html>", "</he", "ad>"), headers("FOO", "bar"),
			"<html>" + replaced},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server, lr := newLiveReloadServer(t)
			defer server.Close()
			req := httptest.NewRequest("GET", "http://example.com", nil)
			w := httptest.NewRecorder()
			injector := lr.NewHTMLInjector(newTag, tt.handler)
			injector(w, req)

			resp := w.Result()
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatal(err)
			}

			for k, vs := range tt.header {
				if resp.Header.Get(k) != vs[0] {
					t.Errorf("expected header '%s: %s' but not found", k, vs[0])
				}
			}
			if string(body) != tt.want {
				t.Errorf("mismatch in injected HTML, expected:\n%s\ngot:\n%s",
					tt.want, string(body))
			}
		})
	}
}

func headers(hs ...string) http.Header {
	if len(hs)%2 != 0 {
		panic("hs must be divisible by two")
	}
	header := make(http.Header)
	for i := 0; i < len(hs); i += 2 {
		j := i + 1
		header.Set(hs[i], hs[j])
	}
	return header
}

func writeHTMLBody(hs ...string) http.HandlerFunc {
	return writeHTML(nil, hs...)
}

func writeHTML(h http.Header, hs ...string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		for k, vs := range h {
			for _, v := range vs {
				w.Header().Add(k, v)
			}
		}
		for _, h := range hs {
			_, _ = w.Write([]byte(h))
		}
	}
}

func TestLiveReload_WebSocketHandler_ClientShouldGetHello(t *testing.T) {
	server, _ := newLiveReloadServer(t)
	defer server.Close()

	conn, resp := newWebSocketClient(t, server)
	expected := "101 Switching Protocols"
	if resp.Status != expected {
		t.Fatalf("expected websocket status code to be %s, got %s", expected, resp.Status)
	}
	assertReadsHelloMsg(t, conn)
}

func TestLiveReload_WebSocketHandler_UnknownClientMessage(t *testing.T) {
	server, _ := newLiveReloadServer(t)
	defer server.Close()
	conn, _ := newWebSocketClient(t, server)
	assertReadsHelloMsg(t, conn)
	writeClientJSON(t, conn, newHelloMsg())

	randomMsg := struct {
		Command string `json:"command"`
	}{"foo"}
	writeClientJSON(t, conn, randomMsg)

	writeClientJSON(t, conn, newHelloMsg())
	hello := new(helloMsg)
	err := conn.ReadJSON(hello)
	if err == nil {
		t.Fatalf("expected server to be closed")
	}
}

func TestLiveReload_ReloadFile(t *testing.T) {
	server, lr := newLiveReloadServer(t)
	defer server.Close()
	conn, _ := newWebSocketClient(t, server)
	assertReadsHelloMsg(t, conn)
	writeClientJSON(t, conn, newHelloMsg())

	lr.ReloadFile("foo_bar")

	actual := new(reloadMsg)
	readClientJSON(t, conn, actual)
	expected := newReloadMsg("foo_bar")
	if diff := cmp.Diff(expected, *actual); diff != "" {
		t.Fatalf("ReloadFile() mismatch (-want +got):\n%s", diff)
	}
}

func TestLiveReload_Alert(t *testing.T) {
	server, lr := newLiveReloadServer(t)
	defer server.Close()
	conn, _ := newWebSocketClient(t, server)
	assertReadsHelloMsg(t, conn)
	writeClientJSON(t, conn, newHelloMsg())

	for i := 0; i < 4; i++ {
		lr.Alert("alert!")

		actual := new(alertMsg)
		readClientJSON(t, conn, actual)
		expected := newAlertResponse("alert!")
		if diff := cmp.Diff(expected, *actual); diff != "" {
			t.Fatalf("Alert() mismatch (-want +got):\n%s", diff)
		}
	}
}

func writeClientJSON(t *testing.T, conn *websocket.Conn, value interface{}) {
	t.Helper()
	if err := conn.WriteJSON(value); err != nil {
		t.Fatal(err)
	}
	// Give a bit of time after writing to let the server process the result.
	<-time.After(time.Millisecond)
}

func readClientJSON(t *testing.T, conn *websocket.Conn, value interface{}) {
	t.Helper()
	if err := conn.ReadJSON(value); err != nil {
		t.Fatal(err)
	}
}

func newLiveReloadServer(t *testing.T) (*httptest.Server, *LiveReload) {
	lr := NewServer(zaptest.NewLogger(t).Sugar())
	go lr.Start()
	return httptest.NewServer(http.HandlerFunc(lr.WebSocketHandler)), lr
}

func newWebSocketClient(t *testing.T, server *httptest.Server) (*websocket.Conn, *http.Response) {
	dialer := new(websocket.Dialer)
	wsURL := strings.Replace(server.URL, "http://", "ws://", 1)
	conn, resp, err := dialer.Dial(wsURL, http.Header{})
	if err != nil {
		t.Fatal(err)
		return nil, nil
	}
	return conn, resp
}

func assertReadsHelloMsg(t *testing.T, conn *websocket.Conn) {
	t.Helper()
	hello := new(helloMsg)
	err := conn.ReadJSON(hello)
	if err != nil {
		t.Fatalf("failed to read server hello request: %s", err)
	}
	expectedResp := newHelloMsg()
	if diff := cmp.Diff(expectedResp, *hello); diff != "" {
		t.Fatalf("didn't receive hello request from server; ReadJSON() mismatch (-want +got):\n%s", diff)
	}
}
