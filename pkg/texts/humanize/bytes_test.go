package humanize

import (
	"math"
	"testing"

	"github.com/jschaf/jsc/pkg/testing/difftest"
)

func TestBytes(t *testing.T) {
	rnd := func(n float64) int64 { return int64(math.Round(n)) }
	tests := []struct {
		n    int64
		want string
	}{
		{0, "0 B"},
		{999, "999 B"},
		{-999, "-999 B"}, // nonsensical, but safe
		{1000, "1000 B"},
		{1023, "1023 B"},
		{1024, "1.0 KiB"},
		{1025, "1.0 KiB"},
		{6143, "5.9 KiB"},
		{rnd(1.1 * Kibibyte), "1.1 KiB"},
		{rnd(1.46 * Kibibyte), "1.4 KiB"},
		{rnd(1.51 * Kibibyte), "1.5 KiB"},
		{rnd(1.6 * Kibibyte), "1.6 KiB"},
		{rnd(2.0 * Mebibyte), "2.0 MiB"},
		{rnd(2.4 * Mebibyte), "2.4 MiB"},
		{rnd(32.5 * Mebibyte), "32 MiB"},
		{rnd(1.0 * Tebibyte), "1.0 TiB"},
		{rnd(2 * Exbibyte), "2.0 EiB"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := Bytes(tt.n)
			difftest.AssertSame(t, tt.want, got)
		})
	}
}
