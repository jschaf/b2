package livereload

import (
	"bytes"
	"encoding/json"
	"github.com/go-test/deep"
	"github.com/gorilla/websocket"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestServeJSHandler(t *testing.T) {
	req := httptest.NewRequest("GET", "http://example.com/livereload.js", nil)
	w := httptest.NewRecorder()
	ServeJSHandler(w, req)

	resp := w.Result()
	body, _ := ioutil.ReadAll(resp.Body)

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
			req := httptest.NewRequest("GET", "http://example.com", nil)
			w := httptest.NewRecorder()
			injector := NewHTMLInjector(newTag, tt.handler)
			injector(w, req)

			resp := w.Result()
			body, err := ioutil.ReadAll(resp.Body)
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
			w.Write([]byte(h))
		}
	}
}

func TestLiveReload_WebSocketHandler_ImmediateClose(t *testing.T) {
	lr := NewWebsocketServer()
	server := httptest.NewServer(http.HandlerFunc(lr.WebSocketHandler))
	defer server.Close()

	go lr.Start()

	req, _ := http.NewRequest("GET", server.URL, strings.NewReader(""))
	req.Header.Set("Connection", "upgrade")
	req.Header.Set("UpgradeOnSIGHUP", "websocket")
	req.Header.Set("Sec-WebSocket-Version", "13")
	req.Header.Set("Sec-WebSocket-Key", "unused")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Do() returned error: %v", err)
	}
	defer resp.Body.Close()

	go func() {
		<-time.After(time.Millisecond * 5)
		lr.Shutdown()
	}()

	body, _ := ioutil.ReadAll(resp.Body)

	expected := bytes.Join([][]byte{
		websocketTextMsg(newHelloMsg()),
		websocketCloseMsg(websocket.CloseNormalClosure),
	}, []byte{})
	if diff := deep.Equal(body, expected); diff != nil {
		t.Errorf("body doesn't match\nexpected:\n%s\ngot:\n%s\n%s",
			expected,
			body,
			strings.Join(diff, "\n"))
	}
}

func TestLiveReload_WebSocketHandler_ClientShouldGetHello(t *testing.T) {
	server, _ := newLiveReloadServer()
	defer server.Close()

	conn, resp := newWebSocketClient(t, server)
	expected := "101 Switching Protocols"
	if resp.Status != expected {
		t.Fatalf("expected websocket status code to be %s, got %s", expected, resp.Status)
	}
	assertReadsHelloMsg(t, conn)
}

func TestLiveReload_WebSocketHandler_BadHandshake(t *testing.T) {
	server, _ := newLiveReloadServer()
	defer server.Close()

	conn, _ := newWebSocketClient(t, server)
	assertReadsHelloMsg(t, conn)

	randomMsg := struct {
		Command string `json:"command"`
	}{"foo"}
	writeClientJSON(t, conn, randomMsg)

	_, _, err := conn.NextReader()
	if _, ok := err.(*websocket.CloseError); !ok {
		t.Fatalf("expected CloseError after bad client handshake; got %s", err)
	}
}

func TestLiveReload_WebSocketHandler_UnknownClientMessage(t *testing.T) {
	server, _ := newLiveReloadServer()
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
	server, lr := newLiveReloadServer()
	defer server.Close()
	conn, _ := newWebSocketClient(t, server)
	assertReadsHelloMsg(t, conn)
	writeClientJSON(t, conn, newHelloMsg())

	lr.ReloadFile("foo_bar")

	actual := new(reloadMsg)
	readClientJSON(t, conn, actual)
	expected := newReloadMsg("foo_bar")
	if diff := deep.Equal(expected, *actual); diff != nil {
		t.Fatalf("expected reload response from server:\n%v\ngot:\n%v\n%s",
			expected, actual, strings.Join(diff, "\n"))
	}
}

func TestLiveReload_Alert(t *testing.T) {
	server, lr := newLiveReloadServer()
	defer server.Close()
	conn, _ := newWebSocketClient(t, server)
	assertReadsHelloMsg(t, conn)
	writeClientJSON(t, conn, newHelloMsg())

	for i := 0; i < 4; i++ {
		lr.Alert("alert!")

		actual := new(alertMsg)
		readClientJSON(t, conn, actual)
		expected := newAlertResponse("alert!")
		if diff := deep.Equal(expected, *actual); diff != nil {
			t.Fatalf("expected alert response from server:\n%v\ngot:\n%v\n%s",
				expected, actual, strings.Join(diff, "\n"))
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

func newLiveReloadServer() (*httptest.Server, *LiveReload) {
	lr := NewWebsocketServer()
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

const (
	websocketText = 129
)

func websocketTextMsg(v interface{}) []byte {
	b := &bytes.Buffer{}
	b.Write([]byte{websocketText, byte(0)})
	err := json.NewEncoder(b).Encode(v)
	if err != nil {
		panic(err)
	}
	bs := b.Bytes()
	bs[1] = byte(len(bs) - 2)
	return bs
}

func assertReadsHelloMsg(t *testing.T, conn *websocket.Conn) {
	t.Helper()
	hello := new(helloMsg)
	err := conn.ReadJSON(hello)
	if err != nil {
		t.Fatalf("failed to read server hello request: %s", err)
	}
	expectedResp := newHelloMsg()
	if diff := deep.Equal(*hello, expectedResp); diff != nil {
		t.Fatalf("didn't receive hello request from server; expected:\n%s\ngot:\n%s\n%s",
			expectedResp, *hello, strings.Join(diff, "\n"))
	}
}

func websocketCloseMsg(code int) []byte {
	err := &websocket.CloseError{Code: code}
	bs := websocket.FormatCloseMessage(code, err.Error())
	finalBit := byte(1 << 7)
	b := bytes.Buffer{}
	b0 := byte(websocket.CloseMessage) | finalBit
	b1 := byte(len(bs))
	b.Write([]byte{b0, b1})
	b.Write(bs)
	return b.Bytes()
}
