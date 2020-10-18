package fake

import (
	"github.com/jschaf/b2/pkg/texts"
)

// UnsafeJSONEncoder is a barebones, streaming JSON encoder that doesn't
// attempt any escaping. If any key or value contains illegal identifier
// characters like double quotes or literal newlines, this encoder will
// produce invalid JSON.
//
// This encoder only encodes one object at a time with explicit calls to write
// object entries. The object must only contain either string or number values.
// This encoder is generally safe because we control the input data. This
// encoder is about 4x faster than json.NewEncoder and has the benefit that
// we can incrementally encode JSON.
type UnsafeJSONEncoder struct {
	buf           []byte
	objStart      int    // offset of the current object
	off           int    // next write at buf[off]
	sizeEst       int    // estimated size of an object
	objsRemaining int    // number of objects remaining
	sizes         [5]int // ring-buffer of last seen object sizes
	sizesIdx      int    // index into sizes ring-buffer
}

const (
	jsonNumDelimSize = 5  // 2 quotes + colon + trailing comma + closing curly bracket
	jsonStrDelimSize = 7  // 4 quotes + colon + trailing comma + closing curly bracket
	maxNumByteLen    = 18 // 2^64 in decimal is 18 chars long
	minObjSize       = 32 // the smallest object size estimate
)

type EncoderConfig struct {
	NumObjects int // number of objects to encode
	EstObjSize int // estimated object sized
}

func NewUnsafeJSONEncoder(conf EncoderConfig) *UnsafeJSONEncoder {
	enc := &UnsafeJSONEncoder{
		sizeEst:       conf.EstObjSize,
		objsRemaining: conf.NumObjects,
	}
	if enc.sizeEst < minObjSize {
		enc.sizeEst = minObjSize
	}
	if enc.objsRemaining < 16 {
		enc.objsRemaining = 16
	}
	enc.buf = enc.newBuffer()
	return enc
}

// newBuffer allocates a new byte array to use for appending JSON.
// The size is based on the 80th percentile of the last 5 objects or the
// original estimated size.
func (e *UnsafeJSONEncoder) newBuffer() []byte {
	const minBufferSize = 8 * (1 << 10)  // min backing buffer to allocate
	const maxBufferSize = 16 * (1 << 20) // max backing buffer to allocate
	maxSize := 0
	p90Size := 0 // 2nd biggest element in 5 elems is the 80th percentile
	for _, size := range e.sizes {
		if size > maxSize {
			maxSize = size
		} else if size > p90Size {
			p90Size = size
		}
	}
	if p90Size == 0 {
		p90Size = e.sizeEst
	}
	if p90Size < minObjSize {
		p90Size = minObjSize
	}
	numBytes := p90Size * e.objsRemaining
	if numBytes > maxBufferSize {
		numBytes = maxBufferSize
	} else if numBytes < minBufferSize {
		numBytes = minBufferSize
	}

	return make([]byte, numBytes)
}

func (e *UnsafeJSONEncoder) grow() {
	cur := e.buf[e.objStart:e.off]
	e.buf = e.newBuffer()
	copy(e.buf, cur)
	e.objStart = 0
	e.off = len(cur)
}

func (e *UnsafeJSONEncoder) remainingBytes() int {
	return len(e.buf) - e.off
}

func (e *UnsafeJSONEncoder) writeByte(b byte) {
	e.buf[e.off] = b
	e.off++
}

func (e *UnsafeJSONEncoder) writeString(s string) {
	bs := texts.ReadOnlyBytes(s)
	copy(e.buf[e.off:], bs)
	e.off += len(bs)
}

func (e *UnsafeJSONEncoder) StartObject() {
	e.objStart = e.off
	if e.remainingBytes() < e.sizeEst {
		e.grow()
	}
	e.objStart = e.off
	e.writeByte('{')
}

func (e *UnsafeJSONEncoder) WriteIntEntry(key string, val int) {
	if e.remainingBytes() < jsonNumDelimSize+len(key)+maxNumByteLen {
		e.grow()
	}

	// key
	e.writeByte('"')
	e.writeString(key)
	e.writeByte('"')

	// delimiter
	e.writeByte(':')

	// value
	e.writeInt(val)
	e.writeByte(',')
}

func (e *UnsafeJSONEncoder) WriteStringEntry(key, val string) {
	if e.remainingBytes() < jsonStrDelimSize+len(key)+len(val) {
		e.grow()
	}

	// key
	e.writeByte('"')
	e.writeString(key)
	e.writeByte('"')

	// delimiter
	e.writeByte(':')

	// value
	e.writeByte('"')
	e.writeString(val)
	e.writeByte('"')
	e.writeByte(',')
}

func (e *UnsafeJSONEncoder) EndObject() []byte {
	// No need to check for size because WriteStringEntry, WriteIntEntry, and StartObject
	// guarantee enough room for the closing bracket.
	if e.buf[e.off-1] == ',' {
		e.off--
	}
	e.writeByte('}')
	e.objsRemaining--
	e.sizes[e.sizesIdx] = e.off - e.objStart
	e.sizesIdx = (e.sizesIdx + 1) % len(e.sizes)
	return e.buf[e.objStart:e.off]
}

var smallsInts = []byte(
	"00010203040506070809" +
		"10111213141516171819" +
		"20212223242526272829" +
		"30313233343536373839" +
		"40414243444546474849" +
		"50515253545556575859" +
		"60616263646566676869" +
		"70717273747576777879" +
		"80818283848586878889" +
		"90919293949596979899")

// writeInt writes a single integer into the buffer.
// This is an optimized version of:
//
//   strconv.AppendInt(e.buf[e.off:e.off], int64(val), 10)
//   e.off += len(intBuf)
//
// This version is 7% faster than the above code and doesn't allocate 64 bytes.
func (e *UnsafeJSONEncoder) writeInt(n int) {
	if n < 0 {
		e.writeByte('-')
		n *= -1
	}

	// Compute log10 manually.
	length := 1
	x := n
	if x >= 10_000_000_000_000_000 {
		length += 16
		x /= 10_000_000_000_000_000
	}
	if x >= 100_000_000 {
		length += 8
		x /= 100_000_000
	}
	if x >= 10_000 {
		length += 4
		x /= 10_000
	}
	if x >= 100 {
		length += 2
		x /= 100
	}
	if x >= 10 {
		length += 1
	}

	i := e.off + length
	for n >= 100 {
		is := n % 100 * 2
		n /= 100
		i -= 2
		e.buf[i+1] = smallsInts[is+1]
		e.buf[i+0] = smallsInts[is+0]
	}

	// n < 100
	is := n * 2
	i--
	e.buf[i] = smallsInts[is+1]
	if n >= 10 {
		i--
		e.buf[i] = smallsInts[is]
	}

	e.off += length
}
