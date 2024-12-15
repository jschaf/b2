package difftest

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

// Diff diffs want and got.
func Diff[T any](want, got T, opts ...cmp.Option) string {
	allOpts := append([]cmp.Option{
		cmpopts.EquateEmpty(), // useful so nil is the same as 0-sized slice
	}, opts...)
	d := cmp.Diff(want, got, allOpts...)
	if d == "" {
		return ""
	}
	return "mismatch (-want +got)\n" + d
}

// AssertSame asserts that there is no diff between want and got.
func AssertSame[T any](t *testing.T, want, got T, opts ...cmp.Option) {
	t.Helper()
	d := Diff(want, got, opts...)
	if d == "" {
		return
	}
	t.Error("mismatch (-want +got)\n" + d)
}
