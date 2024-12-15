package srv

import (
	"errors"
	"net"
)

// NewListenerCloser closes the listener, ignoring already closed errors.
func NewListenerCloser(ln net.Listener) func() error {
	return func() error {
		// Ignore closed network connections, we already closed them.
		if err := ln.Close(); err != nil && !errors.Is(err, net.ErrClosed) {
			return err
		}
		return nil
	}
}
