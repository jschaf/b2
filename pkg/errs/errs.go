package errs

import (
	"errors"
	"fmt"
)

// Capture runs errF and assigns the error, if any, to *err. Preserves the
// original error by wrapping with errors.Join if err is non-nil. If msg is not
// empty, wrap the error returned by closer with the msg.
//
//   - If errF returns nil, do nothing.
//   - If errF returns an error and *err == nil, replace *err with the error.
//   - If errF returns an error and *err != nil, replace *err with a errors.Join
//     containing *err and the errF err.
func Capture(err *error, errF func() error, msg string) {
	fErr := errF()
	if fErr == nil {
		return
	}
	wErr := fErr
	if msg != "" {
		wErr = fmt.Errorf(msg+": %w", fErr)
	}
	if *err == nil {
		// Only 1 error so avoid a MultiError and replace the err pointer.
		*err = wErr
		return
	}
	*err = errors.Join(*err, wErr)
}

type testingTB interface {
	Helper()
	Error(args ...interface{})
	Errorf(format string, args ...interface{})
}

// CaptureT call t.Error if errF returns an error with an optional message.
func CaptureT(t testingTB, errF func() error, msg string) {
	t.Helper()
	if err := errF(); err != nil {
		if msg == "" {
			t.Error(err)
		} else {
			t.Errorf(msg+": %s", err)
		}
	}
}
