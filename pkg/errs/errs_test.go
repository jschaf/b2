package errs

import (
	"errors"
	"io"
	"testing"

	"github.com/google/go-cmp/cmp"
)

type errCloser struct {
	err error
}

func (e errCloser) Close() error {
	return e.err
}

func closer(err error) io.Closer {
	return errCloser{err: err}
}

func errP(err error) *error {
	return &err
}

func TestCloseWithErrCapture(t *testing.T) {
	e := errors.New
	me := NewMultiError

	tests := []struct {
		name   string
		err    *error
		closer io.Closer
		msg    string
		want   string
	}{
		{"nil_nil", errP(nil), closer(nil), "msg", "<nil error>"},
		{"err_nil", errP(e("orig")), closer(nil), "msg", "orig"},
		{"nil_err", errP(nil), closer(e("cl")), "msg", "msg: cl"},
		{"nil_err_msg", errP(nil), closer(e("cl")), "", "cl"},
		{"err_err", errP(e("orig")), closer(e("cl")), "msg", "2 errors: orig; msg: cl"},
		{"multiErr_err", errP(me(e("o1"), e("o2"))), closer(e("cl")), "", "3 errors: o1; o2; cl"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Capturing(tt.err, tt.closer.Close, tt.msg)
			got := "<nil error>"
			if *tt.err != nil {
				got = (*tt.err).Error()
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("CapturingClose() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
