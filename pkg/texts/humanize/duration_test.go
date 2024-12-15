package humanize

import (
	"fmt"
	"testing"
	"time"

	"github.com/jschaf/b2/pkg/testing/difftest"
)

func TestDuration(t *testing.T) {
	tests := []struct {
		d    time.Duration
		want string
	}{
		{0, "0"},
		{999 * time.Nanosecond, "999 ns"},
		{-999 * time.Nanosecond, "-999 ns"}, // nonsensical, but safe
		{1000 * time.Nanosecond, "1.0 µs"},
		{1023 * time.Nanosecond, "1.0 µs"},
		{1051 * time.Nanosecond, "1.0 µs"},
		{1100 * time.Microsecond, "1.1 ms"},
		{1460 * time.Microsecond, "1.4 ms"},
		{1510 * time.Microsecond, "1.5 ms"},
		{1600 * time.Microsecond, "1.6 ms"},
		{88 * time.Millisecond, "88 ms"},
		{2000 * time.Millisecond, "2.0 s"},
		{2400 * time.Millisecond, "2.4 s"},
		{24 * time.Second, "24 s"},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%d == %s", tt.d, tt.want), func(t *testing.T) {
			got := Duration(tt.d)
			difftest.AssertSame(t, tt.want, got)
		})
	}
}
