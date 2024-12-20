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
		{"nil_err_msg", errP(nil), closer(e("cl")), "", ": cl"},
		{"err_err", errP(e("orig")), closer(e("cl")), "msg", "orig\nmsg: cl"},
		{"multiErr_err", errP(errors.Join(e("o1"), e("o2"))), closer(e("cl")), "", "o1\no2\n: cl"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Capture(tt.err, tt.closer.Close, tt.msg)
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
