package errs

import (
	"bytes"
	"fmt"
	"io"
	"strconv"
)

// MultiError implements the error interface and contains many errors.
type MultiError []error

func NewMultiError(errs ...error) MultiError {
	m := MultiError{}
	for _, err := range errs {
		m.Add(err)
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

// Add adds the error to the error list if it is not nil.
func (m *MultiError) Add(err error) {
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

// Err returns the error list as an error or nil if it is empty.
func (m MultiError) Err() error {
	if len(m) == 0 {
		return nil
	}
	return m
}

// CloseWithErrCapture runs closer.Close() and assigns the error, if any, to
// err while preserving the original err in a MultiError if necessary.
//
// - If closer.Close() does not error, do nothing.
// - If closer.Close() errors and *err == nil, replace *err with the Close()
//   error.
// - If closer.Close() errors and *err != nil, create a MultiError containing
//   *err and the Close() err, then replace *err with the MultiError.
func CloseWithErrCapture(err *error, closer io.Closer, msg string) {
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
	m.Add(*err)
	m.Add(wErr)
	*err = m
}
