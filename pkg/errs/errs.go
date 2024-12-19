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
func Capture(errPtr *error, errFunc func() error, msg string) {
	err := errFunc()
	if err == nil {
		return
	}
	*errPtr = errors.Join(*errPtr, fmt.Errorf("%s: %w", msg, err))
}

// testingTB is a subset of *testing.T and *testing.B methods.
type testingTB interface {
	Helper()
	Errorf(format string, args ...any)
}

// CaptureT call t.Error if errF returns an error with an optional message.
func CaptureT(t testingTB, errFunc func() error, msg string) {
	t.Helper()
	if err := errFunc(); err != nil {
		t.Errorf("%s: %s", msg, err)
	}
}
