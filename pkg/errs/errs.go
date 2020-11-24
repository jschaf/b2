package errs

import (
	"bytes"
	"fmt"
	"strconv"
	"testing"
)

// MultiError implements the error interface and contains many errors.
// MultiError is nil if no errors are added, so err == nil is still
// a valid construct.
type MultiError []error

func NewMultiError(errs ...error) MultiError {
	if len(errs) == 0 {
		// Return nil so err == nil guards work.
		return nil
	}

	m := MultiError{}
	for _, err := range errs {
		m.Append(err)
	}
	return m
}

// Error returns a string of all errors contained in the MultiError.
func (m MultiError) Error() string {
	var buf bytes.Buffer

	if len(m) > 1 {
		buf.WriteString(strconv.Itoa(len(m)))
		buf.WriteString(" errors: ")
	}

	for i, err := range m {
		if i != 0 {
			buf.WriteString("; ")
		}
		buf.WriteString(err.Error())
	}

	return buf.String()
}

// Append adds the error to the error list if it is not nil.
func (m *MultiError) Append(err error) {
	if err == nil {
		return
	}
	switch t := err.(type) {
	case MultiError:
		*m = append(*m, t...)
	default:
		*m = append(*m, err)
	}
}

// ErrorOrNil returns an error interface if this Error represents a list of
// errors, or returns nil if the list of errors is empty. This function is
// useful at the end of accumulation to make sure that the value returned
// represents the existence of errors.
func (m *MultiError) ErrorOrNil() error {
	if m == nil || len(*m) == 0 {
		return nil
	}
	return m
}

// Capturing runs errF and assigns the error, if any, to err.
// Preserves the original err by wrapping in a MultiError if err is non-nil.
//
// If msg is not empty, wrap the error returned by closer with the msg.
//
// - If errF does not error, do nothing.
// - If errF errors and *err == nil, replace *err with the error.
// - If errF errors and *err != nil, create a MultiError containing
//   *err and the errF err, then replace *err with the MultiError.
func Capturing(err *error, errF func() error, msg string) {
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

	*err = NewMultiError(*err, wErr)
}

// CapturingT call t.Error if errF returns an error with an optional message.
func CapturingT(t *testing.T, errF func() error, msg string) {
	t.Helper()
	if err := errF(); err != nil {
		if msg == "" {
			t.Error(err)
		} else {
			t.Errorf(msg+": %s", err)
		}
	}
}
