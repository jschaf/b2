package livereload

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"sync"
	"syscall"
	"time"

	"github.com/gorilla/websocket"
)

// conn is a websocket connection to a LiveReload client.
type conn struct {
	ws      *websocket.Conn
	send    chan any
	closer  sync.Once
	detachC chan<- closeReq // request the connPub to stop sending messages and close this conn
	stopC   chan struct{}
}

func newConn(ws *websocket.Conn, detachC chan<- closeReq) *conn {
	return &conn{
		ws:      ws,
		send:    make(chan any, 5),
		detachC: detachC,
		closer:  sync.Once{},
		stopC:   make(chan struct{}),
	}
}

// start blocks until the initial handshake completes and then runs
// in goroutines.
func (c *conn) start() error {
	if err := c.ws.WriteJSON(newHelloMsg()); err != nil {
		return newCloseError(
			websocket.CloseInternalServerErr,
			"failed to write hello message")
	}

	if err := c.receiveHandshake(); err != nil {
		return fmt.Errorf("expected handshake: %w", err)
	}

	go c.receive()
	go c.transmit()
	return nil
}

func (c *conn) requestDetach(err error) {
	c.detachC <- newCloseReq(c, err)
}

func (c *conn) readText() ([]byte, error) {
	msgType, reader, err := c.ws.NextReader()
	if err != nil {
		return nil, fmt.Errorf("read websocket: %w", err)
	}

	if msgType == websocket.BinaryMessage {
		return nil, newCloseError(websocket.CloseUnsupportedData, "expected text data, got binary")
	}
	bs, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("websocket read all: %w", err)
	}
	return bs, nil
}

func (c *conn) decodeCmd(bs []byte) (command, error) {
	cmd := new(baseCmd)
	if err := json.NewDecoder(bytes.NewReader(bs)).Decode(cmd); err != nil {
		return "", errors.New("unable to decode JSON with {command}")
	}
	return cmd.Command, nil
}

func (c *conn) readHelloCmd(bs []byte) error {
	hello := new(helloMsg)
	err := json.NewDecoder(bytes.NewReader(bs)).Decode(hello)
	if err != nil {
		return newCloseError(websocket.ClosePolicyViolation, "failed to decode client hello message")
	}
	if !validateHelloMsg(hello) {
		return websocket.ErrBadHandshake
	}
	return nil
}

func (c *conn) readInfoCmd(bs []byte) error {
	info := new(infoMsg)
	err := json.NewDecoder(bytes.NewReader(bs)).Decode(info)
	if err != nil {
		return newCloseError(websocket.ClosePolicyViolation, "failed to decode info message")
	}
	slog.Debug("LiveReload client info", "url", info.URL, "plugins", formatInfoMsg(info))
	return nil
}

func (c *conn) receiveHandshake() error {
	bs, err := c.readText()
	if err != nil {
		return err
	}

	cmd, err := c.decodeCmd(bs)
	if err != nil {
		return err
	}

	switch cmd {
	case helloCmd:
		if err := c.readHelloCmd(bs); err != nil {
			return err
		}

	default:
		c.requestDetach(fmt.Errorf("unexpected command: %s", cmd))
		return err
	}
	slog.Debug("received handshake")
	return nil
}

func (c *conn) receive() {
	slog.Debug("starting receive")
	for {
		bs, err := c.readText()
		if err != nil {
			c.requestDetach(err)
			return
		}

		cmd, err := c.decodeCmd(bs)
		if err != nil {
			c.requestDetach(err)
			return
		}

		switch cmd {
		case helloCmd:
			if err := c.readHelloCmd(bs); err != nil {
				c.requestDetach(err)
				return
			}

		case infoCmd:
			if err := c.readInfoCmd(bs); err != nil {
				c.requestDetach(err)
				return
			}

		default:
			c.requestDetach(fmt.Errorf("unexpected command: %s", cmd))
			return
		}
	}
}

func (c *conn) transmit() {
	slog.Debug("starting transmit")
	for m := range c.send {
		slog.Debug("sending LiveReload message", "message", m)
		if err := c.ws.WriteJSON(m); err != nil {
			c.requestDetach(err)
			return
		}
	}
}

// close closes the websocket connection. Must only be called by connPublisher.
func (c *conn) close(err error) {
	slog.Debug("close livereload socket", "error", err)
	c.closer.Do(func() {
		slog.Debug("actually closing livereload socket")
		closeCode := websocket.CloseInternalServerErr
		var closeErr *websocket.CloseError
		if errors.As(err, &closeErr) {
			closeCode = closeErr.Code
		}

		msg := err.Error()
		closeMsg := websocket.FormatCloseMessage(closeCode, msg)
		deadline := time.Now().Add(time.Second)

		writeErr := c.ws.WriteControl(websocket.CloseMessage, closeMsg, deadline)
		if writeErr != nil && !errors.Is(writeErr, websocket.ErrCloseSent) {
			slog.Debug("write websocket close message", "error", writeErr)
		}
		err := c.ws.Close()
		// Ignore broken pipe since that's it means the browser client disconnected.
		// Likely caused by reloading.
		if err != nil && !errors.Is(err, syscall.EPIPE) {
			slog.Error("close websocket", "error", err)
		}
		close(c.send)
	})
}
