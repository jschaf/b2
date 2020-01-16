package livereload

import (
	"encoding/json"
	"errors"
	"github.com/gorilla/websocket"
	"log"
	"sync"
	"time"
)

// conn is a websocket connection to a LiveReload client.
type conn struct {
	handshake bool
	ws        *websocket.Conn
	send      chan interface{}
	closer    sync.Once
}

func newConn(ws *websocket.Conn) *conn {
	return &conn{
		handshake: false,
		ws:        ws,
		send:      make(chan interface{}, 5),
		closer:    sync.Once{},
	}
}

func (c *conn) start() {
	if err := c.ws.WriteJSON(helloResponse); err != nil {
		c.closeCode(websocket.CloseInternalServerErr)
	}

	go c.receive()
	c.transmit()
}

func (c *conn) receive() {
	for {
		msgType, reader, err := c.ws.NextReader()
		if err != nil {
			c.close(err)
			return
		}

		if msgType == websocket.BinaryMessage {
			c.closeCode(websocket.CloseUnsupportedData)
			return
		}

		helloReq := new(helloRequest)
		err = json.NewDecoder(reader).Decode(helloReq)
		if err != nil {
			c.closeCode(websocket.ClosePolicyViolation)
			return
		}

		if validateHelloRequest(helloReq) {
			c.handshake = true
		} else {
			c.close(websocket.ErrBadHandshake)
			return
		}
	}
}

func (c *conn) transmit() {
	for m := range c.send {
		if !c.handshake {
			c.close(errors.New("handshake not established"))
			return
		}

		if err := c.ws.WriteJSON(m); err != nil {
			c.close(err)
			return
		}
	}
}

func (c *conn) closeCode(code int) {
	err := &websocket.CloseError{Code: code}
	c.close(err)
}

func (c *conn) close(err error) {
	closeCode := websocket.CloseNoStatusReceived
	if closeErr, ok := err.(*websocket.CloseError); ok {
		closeCode = closeErr.Code
	}

	closeMsg := websocket.FormatCloseMessage(closeCode, err.Error())
	deadline := time.Now().Add(time.Second)
	err = c.ws.WriteControl(websocket.CloseMessage, closeMsg, deadline)
	log.Printf("failed to write websocket control: %s", err)

	c.closer.Do(func() {
		close(c.send)
	})
}
