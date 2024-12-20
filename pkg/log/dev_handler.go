package log

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/jschaf/b2/pkg/texts/humanize"
	"github.com/jschaf/b2/pkg/tty"
)

type DevHandler struct {
	w    io.Writer
	opts slog.HandlerOptions
}

func NewDevHandler(w io.Writer, opts *slog.HandlerOptions) *DevHandler {
	if opts == nil {
		opts = &slog.HandlerOptions{}
	}
	return &DevHandler{w: w, opts: *opts}
}

func (h *DevHandler) Enabled(_ context.Context, l slog.Level) bool {
	minLevel := slog.LevelInfo
	if h.opts.Level != nil {
		minLevel = h.opts.Level.Level()
	}
	return l >= minLevel
}

const (
	align       = 40 // make logs easier to scan by aligning the first attr
	alignStr    = "                                        "
	readyPrefix = "ready: "
)

func (h *DevHandler) Handle(_ context.Context, r slog.Record) error {
	buf := newBuffer()
	defer buf.free()

	// Time
	appendTime(buf, r.Time)

	// Level
	buf.appendByte('\t')
	appendLevel(buf, r)

	// Message
	r.Message = strings.TrimPrefix(r.Message, readyPrefix)
	buf.appendByte('\t')
	buf.appendString(r.Message)

	// Attrs
	if r.NumAttrs() > 0 {
		padCount := max(align-len(r.Message), 2)
		pad := alignStr[:padCount]
		buf.appendString(pad)
		r.Attrs(func(attr slog.Attr) bool {
			appendAttr(buf, attr)
			return true
		})
	}

	// Newline
	buf.appendByte('\n')

	_, err := h.w.Write(*buf)
	if err != nil {
		return fmt.Errorf("write record: %w", err)
	}
	return nil
}

func (h *DevHandler) WithAttrs(_ []slog.Attr) slog.Handler {
	panic("not implemented")
}

func (h *DevHandler) WithGroup(_ string) slog.Handler {
	panic("not implemented")
}

func appendTime(buf *buffer, t time.Time) {
	h, m, s := t.Clock()

	// Hours
	buf.appendByte('0' + byte(h/10))
	buf.appendByte('0' + byte(h%10))

	// Minutes
	buf.appendByte(':')
	buf.appendByte('0' + byte(m/10))
	buf.appendByte('0' + byte(m%10))

	// Seconds
	buf.appendByte(':')
	buf.appendByte('0' + byte(s/10))
	buf.appendByte('0' + byte(s%10))

	// Milliseconds
	buf.appendByte('.')
	ms := t.Nanosecond() / 1e6
	lo := ms % 10
	ms /= 10
	mid := ms % 10
	ms /= 10
	hi := ms
	buf.appendByte('0' + byte(hi))
	buf.appendByte('0' + byte(mid))
	buf.appendByte('0' + byte(lo))
}

func appendLevel(buf *buffer, r slog.Record) {
	switch {
	case r.Level < slog.LevelInfo:
		buf.appendString(tty.Magenta.Code())
		buf.appendString("debug")
		buf.appendString(tty.Reset.Code())
	case r.Level < slog.LevelWarn:
		if strings.HasPrefix(r.Message, readyPrefix) {
			buf.appendString(tty.Green.Code())
			buf.appendString("ready")
		} else {
			buf.appendString(tty.Blue.Code())
			buf.appendString("info")
		}
		buf.appendString(tty.Reset.Code())
	case r.Level < slog.LevelError:
		buf.appendString(tty.Yellow.Code())
		buf.appendString("warn")
		buf.appendString(tty.Reset.Code())
	default:
		buf.appendString(tty.Red.Code())
		buf.appendString("error")
		buf.appendString(tty.Reset.Code())
	}
}

func appendAttr(buf *buffer, attr slog.Attr) {
	buf.appendByte(' ')
	switch attr.Key {
	case "url":
		appendValue(buf, attr.Value)
	default:
		buf.appendString(attr.Key)
		buf.appendByte('=')
		appendValue(buf, attr.Value)
	}
}

func appendValue(buf *buffer, v slog.Value) {
	switch v.Kind() {
	case slog.KindString:
		buf.appendString(v.String())
	case slog.KindInt64:
		*buf = strconv.AppendInt(*buf, v.Int64(), 10)
	case slog.KindUint64:
		*buf = strconv.AppendUint(*buf, v.Uint64(), 10)
	case slog.KindFloat64:
		*buf = strconv.AppendFloat(*buf, v.Float64(), 'g', -1, 64)
	case slog.KindBool:
		*buf = strconv.AppendBool(*buf, v.Bool())
	case slog.KindDuration:
		buf.appendString(humanize.Duration(v.Duration()))
	case slog.KindTime:
		*buf = appendRFC3339Millis(*buf, v.Time())
	case slog.KindAny:
		a := v.Any()
		switch a := a.(type) {
		case error:
			buf.appendString(a.Error())
		default:
			_, _ = fmt.Fprint(buf, a)
		}
	default:
		panic(fmt.Sprintf("bad kind: %s", v.Kind()))
	}
}

func appendRFC3339Millis(b []byte, t time.Time) []byte {
	// Format according to time.RFC3339Nano since it is highly optimized,
	// but truncate it to use millisecond resolution.
	// Unfortunately, that format trims trailing 0s, so add 1/10 millisecond
	// to guarantee that there are exactly 4 digits after the period.
	const prefixLen = len("2006-01-02T15:04:05.000")
	n := len(b)
	t = t.Truncate(time.Millisecond).Add(time.Millisecond / 10)
	b = t.AppendFormat(b, time.RFC3339Nano)
	b = append(b[:n+prefixLen], b[n+prefixLen+1:]...) // drop the 4th digit
	return b
}
