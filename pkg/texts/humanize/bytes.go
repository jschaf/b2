package humanize

import (
	"strconv"
)

const (
	Byte = 1 << (iota * 10)
	Kibibyte
	Mebibyte
	Gibibyte
	Tebibyte
	Pebibyte
	Exbibyte
)

// Bytes formats a number of bytes into a human-readable string using power-of-2
// units defined by the IEC like KiB for Kibibyte. Truncates (not rounds) to 1
// decimal place if the units are less than 10.
func Bytes(n int64) string {
	switch {
	case n < Kibibyte:
		return formatTruncatedInt(n, Byte, "B")
	case n < Mebibyte:
		return formatTruncatedInt(n, Kibibyte, "KiB")
	case n < Gibibyte:
		return formatTruncatedInt(n, Mebibyte, "MiB")
	case n < Tebibyte:
		return formatTruncatedInt(n, Gibibyte, "GiB")
	case n < Pebibyte:
		return formatTruncatedInt(n, Tebibyte, "TiB")
	case n < Exbibyte:
		return formatTruncatedInt(n, Pebibyte, "PiB")
	default:
		return formatTruncatedInt(n, Exbibyte, "EiB")
	}
}

func formatTruncatedInt(n int64, base int64, suffix string) string {
	units := n / base
	b := make([]byte, 0, 8)
	b = strconv.AppendInt(b, units, 10)
	if units < 10 && base > Byte {
		rem := n % base
		msd := min(rem/(base/10), 9)
		b = append(b, '.')
		b = append(b, '0'+byte(msd))
	}
	b = append(b, ' ')
	b = append(b, suffix...)
	return string(b)
}
