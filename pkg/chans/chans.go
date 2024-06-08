package chans

import (
	"errors"
	"time"
)

var TimeoutErr = errors.New("timeout waiting for channel")

// Wait waits for ch to close for up to duration d. If ch doesn't close with
// duration d, Wait returns a TimeoutError.
func Wait(ch <-chan struct{}, d time.Duration) error {
	select {
	case <-ch:
		return nil
	case <-time.After(d):
		return TimeoutErr
	}
}
