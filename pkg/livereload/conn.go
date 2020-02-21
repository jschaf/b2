package livereload

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"io/ioutil"
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
	if err := c.ws.WriteJSON(newHelloMsg()); err != nil {
		c.closeWithCode(websocket.CloseInternalServerErr,
			"failed to write hello message")
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
			c.closeWithCode(websocket.CloseUnsupportedData,
				"expected text data, got binary")
			return
		}

		bs, err := ioutil.ReadAll(reader)
		cmd := new(baseCmd)
		if err := json.NewDecoder(bytes.NewReader(bs)).Decode(cmd); err != nil {
			c.close(errors.New("unable to decode JSON with {command}"))
			return
		}

		switch cmd.Command {
		case helloCmd:
			hello := new(helloMsg)
			err = json.NewDecoder(bytes.NewReader(bs)).Decode(hello)
			if err != nil {
				c.closeWithCode(websocket.ClosePolicyViolation,
					"failed to decode client hello message")
				return
			}
			if !validateHelloMsg(hello) {
				c.close(websocket.ErrBadHandshake)
				return
			}
			c.handshake = true

		case infoCmd:
			info := new(infoMsg)
			err = json.NewDecoder(bytes.NewReader(bs)).Decode(info)
			if err != nil {
				c.closeWithCode(websocket.ClosePolicyViolation,
					"failed to decode info message")
				return
			}
			log.Printf("LiveReload client info: url=%s, plugins=%s",
				info.URL, formatInfoMsg(info))

		default:
			log.Printf("unsupported command received from websocket client: %s",
				cmd.Command)
			c.close(fmt.Errorf("unexpected command: %s", cmd.Command))
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

func (c *conn) closeWithCode(code int, message string) {
	err := &websocket.CloseError{Code: code, Text: message}
	c.close(err)
}

func (c *conn) close(err error) {
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
