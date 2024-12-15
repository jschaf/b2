package tty

import (
	"os"
)

var noColor = os.Getenv("NO_COLOR") != "" || os.Getenv("TERM") == "dumb"

// Foreground colors.
//
//goland:noinspection GoUnusedConst
const (
	Reset   Color = 0
	Black   Color = 30
	Red     Color = 31
	Green   Color = 32
	Yellow  Color = 33
	Blue    Color = 34
	Magenta Color = 35
	Cyan    Color = 36
	White   Color = 37
)

var codes = [8]string{
	"\x1b[30m",
	"\x1b[31m",
	"\x1b[32m",
	"\x1b[33m",
	"\x1b[34m",
	"\x1b[35m",
	"\x1b[36m",
	"\x1b[37m",
}

// Color represents a text color.
type Color uint8

func (c Color) Code() string {
	if noColor {
		return ""
	}
	if c < 30 || c > 37 {
		return "\x1b[0m"
	}
	return codes[c-30]
}

// Add adds the coloring to the given string.
func (c Color) Add(s string) string {
	if noColor || c == 0 {
		return s
	}
	return c.Code() + s + Reset.Code()
}
