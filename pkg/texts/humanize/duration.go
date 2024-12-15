package humanize

import "time"

// Duration formats a time.Duration into a human-readable string.
func Duration(d time.Duration) string {
	switch {
	case d == 0:
		return "0"
	case d < time.Microsecond:
		return formatTruncatedInt(int64(d), 1, "ns")
	case d < time.Millisecond:
		return formatTruncatedInt(int64(d), int64(time.Microsecond), "µs")
	case d < time.Second:
		return formatTruncatedInt(int64(d), int64(time.Millisecond), "ms")
	default:
		return formatTruncatedInt(int64(d), int64(time.Second), "s")
	}
}
