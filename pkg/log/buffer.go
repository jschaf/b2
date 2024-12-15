package log

import "sync"

// buffer is a byte buffer.
//
// This implementation is adapted from the unexported type buffer
// in go/src/fmt/print.go.
type buffer []byte

// Having an initial size gives a dramatic speedup.
var bufPool = sync.Pool{
	New: func() any {
		b := make([]byte, 0, 1024)
		return (*buffer)(&b)
	},
}

func newBuffer() *buffer {
	return bufPool.Get().(*buffer)
}

func (b *buffer) Write(p []byte) (n int, err error) {
	*b = append(*b, p...)
	return len(p), nil
}

func (b *buffer) free() {
	// To reduce peak allocation, return only smaller buffers to the pool.
	const maxBufferSize = 16 << 10
	if cap(*b) <= maxBufferSize {
		*b = (*b)[:0]
		bufPool.Put(b)
	}
}

func (b *buffer) appendString(s string) {
	*b = append(*b, s...)
}

func (b *buffer) appendByte(c byte) {
	*b = append(*b, c)
}
