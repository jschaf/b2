package log

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/jschaf/jsc/pkg/chans"
)

var TriggerNotFoundErr = errors.New("trigger not found")

type triggerState int

const (
	triggerStateUnknown triggerState = iota
	triggerStateNotFound
	triggerStateFound
)

// TriggerWriter is a writer that looks for a trigger string in newline
// delimited text. Once the trigger string is found, the Wait function returns.
// If the trigger is not found after input is exhausted, Wait returns
// TriggerNotFoundErr.
type TriggerWriter struct {
	pipeW  *io.PipeWriter
	doneCh chan struct{}
	state  triggerState
	once   sync.Once
}

func (tw *TriggerWriter) Close() error {
	return tw.pipeW.Close()
}

func NewTriggerWriter(trigger string) *TriggerWriter {
	pipeR, pipeW := io.Pipe()
	tw := &TriggerWriter{
		pipeW:  pipeW,
		doneCh: make(chan struct{}),
		state:  triggerStateUnknown,
	}

	// Start the goroutine here to avoid blocking any upstream writers.
	go func() {
		scanner := bufio.NewScanner(pipeR)
		for scanner.Scan() {
			text := scanner.Text()
			if strings.Contains(text, trigger) {
				tw.state = triggerStateFound
				break
			}
		}
		if tw.state == triggerStateUnknown {
			tw.state = triggerStateNotFound
		}
		close(tw.doneCh)
		// Make sure we don't block other writers before Write closes the pipe.
		_, _ = io.Copy(io.Discard, pipeR)
		_ = pipeR.Close()
	}()

	return tw
}

func (tw *TriggerWriter) Write(p []byte) (n int, err error) {
	// Racy check; okay because this is purely an optimization.
	if tw.state != triggerStateUnknown {
		// Safe to close pipeW because state will not change.
		tw.once.Do(func() { _ = tw.pipeW.Close() })
		// Do a no-op write as an optimization since we don't need to search for
		// the trigger anymore.
		return len(p), nil
	}
	return tw.pipeW.Write(p)
}

// Wait blocks until the trigger string is found. If the string is not found
// before Close, returns TriggerNotFoundErr.
func (tw *TriggerWriter) Wait(deadline time.Duration) error {
	if err := chans.Wait(tw.doneCh, deadline); err != nil {
		return fmt.Errorf("exceeded wait deadline in trigger writer wait: %w", err)
	}
	if tw.state == triggerStateNotFound {
		return TriggerNotFoundErr
	}
	return nil
}
