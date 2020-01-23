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
	if err := c.ws.WriteJSON(newHelloResponse()); err != nil {
		c.closeWithCode(websocket.CloseInternalServerErr)
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
			c.closeWithCode(websocket.CloseUnsupportedData)
			return
		}

		helloReq := new(helloRequest)
		err = json.NewDecoder(reader).Decode(helloReq)
		if err != nil {
			c.closeWithCode(websocket.ClosePolicyViolation)
			return
		}

		if !c.handshake {
			if validateHelloRequest(helloReq) {
				log.Println("validated hello request")
				c.handshake = true
			} else {
				c.close(websocket.ErrBadHandshake)
				return
			}
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

func (c *conn) closeWithCode(code int) {
	err := &websocket.CloseError{Code: code}
	c.close(err)
}

func (c *conn) close(err error) {
	log.Printf("closing connection: %s", err)
	closeCode := websocket.CloseInternalServerErr
	if closeErr, ok := err.(*websocket.CloseError); ok {
		closeCode = closeErr.Code
	}

	msg := err.Error()
	closeMsg := websocket.FormatCloseMessage(closeCode, msg)
	deadline := time.Now().Add(time.Second)

	c.closer.Do(func() {
		err = c.ws.WriteControl(websocket.CloseMessage, closeMsg, deadline)
		if err != nil {
			log.Printf("failed to write websocket control: %s", err)
		}
		err = c.ws.Close()
		if err != nil {
			log.Printf("failed to close websocket: :%s", err)
		}
		close(c.send)
	})
}