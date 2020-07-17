package errs

import (
	"bytes"
	"fmt"
	"io"
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

// CapturingClose runs closer.Close() and assigns the error, if any, to err.
// Preserves the original err by wrapping in a MultiError if err is non-nil.
//
// - If closer.Close() does not error, do nothing.
// - If closer.Close() errors and *err == nil, replace *err with the Close()
//   error.
// - If closer.Close() errors and *err != nil, create a MultiError containing
//   *err and the Close() err, then replace *err with the MultiError.
func CapturingClose(err *error, closer io.Closer, msg string) {
	cErr := closer.Close()
	if cErr == nil {
		return
	}

	// Wrap if we have a msg.
	wErr := cErr
	if msg != "" {
		wErr = fmt.Errorf(msg+": %w", cErr)
	}

	if *err == nil {
		// Only 1 error from Close() so replace the error pointed at by err
		*err = wErr
		return
	}

	// Both *err and Close() error are non-nil.
	m := NewMultiError()
	m.Append(*err)
	m.Append(wErr)
	*err = m
}

// CloseWithTestError runs closer.Close() and calls t.Error if Close() returned
// an error.
func CloseWithTestError(t *testing.T, closer io.Closer) {
	t.Helper()
	err := closer.Close()
	if err != nil {
		t.Errorf("close in test: %s", err.Error())
	}
}
