package fake

import (
	"math/rand"
	"time"

	"github.com/jschaf/b2/pkg/api"
)

// EventFaker is a non-thread-safe, deterministic pseudo-RNG. Deterministic
// means the same sequence of function calls on newly created EventFaker
// instances produces the same output.
type EventFaker struct {
	intSrc          rand.Source
	jsIntSrc        rand.Source
	epochMillisSrc  rand.Source
	randWordsOffset int
}

func NewEventFaker() *EventFaker {
	// Use custom sources to ensure deterministic output.
	return &EventFaker{
		intSrc:          rand.NewSource(1),
		jsIntSrc:        rand.NewSource(2),
		epochMillisSrc:  rand.NewSource(3),
		randWordsOffset: 0,
	}
}

func (f *EventFaker) WriteInts(p []int) {
	for i := range p {
		p[i] = int(f.intSrc.Int63())
	}
}

// WriteJSInts fills p with integers in [0, 2^53).
// https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Number/MAX_SAFE_INTEGER
func (f *EventFaker) WriteJSInts(p []int) {
	for i := range p {
		p[i] = int(f.intSrc.Int63() & ((1 << 53) - 1))
	}
}

// WriteEpochMillis fills p with ints that represent the epoch milliseconds.
// The range is roughly 2012 to 2020.
func (f *EventFaker) WriteEpochMillis(p []int) {
	// Start relative to Unix epoch, 1970-01-01.
	const start = 42 * 365 * 24 * (time.Hour / time.Millisecond)
	for i := range p {
		// 2^38 milliseconds is 8.7 years. Use a power of 2 to use masking to
		// reduce range.
		n := f.intSrc.Int63() & ((1 << 38) - 1)
		p[i] = int(n) + int(start)
	}
}

// WriteWords fills p with random words.
func (f *EventFaker) WriteWords(p []string) {
	for i := range p {
		word := randWords[(f.randWordsOffset+i)%len(randWords)]
		p[i] = word
	}
	f.randWordsOffset += len(p)
}

func (f *EventFaker) WriteEvents(p []api.Event) error {
	ints := make([]int, len(p))

	f.WriteJSInts(ints)
	for i := range p {
		p[i].EventID = api.EventID(ints[i])
	}

	f.WriteJSInts(ints)
	for i := range p {
		p[i].UserID = api.UserID(ints[i])
	}

	f.WriteJSInts(ints)
	for i := range p {
		p[i].SessionID = api.SessionID(ints[i])
	}

	f.WriteEpochMillis(ints)
	for i := range p {
		p[i].Time = api.EpochMillis(ints[i])
	}

	strings := make([]string, len(p))

	type data struct {
		Path         string
		Hash         string
		Hierarchy    string
		Price        int
		CheckoutTime int
	}
	datas := make([]data, len(p))

	f.WriteWords(strings)
	for i := range strings {
		datas[i].Path = strings[i]
	}

	f.WriteWords(strings)
	for i := range strings {
		datas[i].Hash = strings[i]
	}

	f.WriteWords(strings)
	for i := range strings {
		datas[i].Hierarchy = strings[i]
	}

	f.WriteInts(ints)
	for i := range ints {
		datas[i].Price = ints[i]
	}

	f.WriteEpochMillis(ints)
	for i := range ints {
		datas[i].CheckoutTime = ints[i]
	}

	enc := NewUnsafeJSONEncoder(EncoderConfig{NumObjects: len(p), EstObjSize: 32})
	for i, data := range datas {
		enc.StartObject()
		enc.WriteStringEntry("path", data.Path)
		enc.WriteStringEntry("hash", data.Hash)
		enc.WriteStringEntry("hierarchy", data.Hierarchy)
		enc.WriteIntEntry("price", data.Price)
		enc.WriteIntEntry("checkout_time", data.CheckoutTime)
		p[i].Data = enc.EndObject()
	}
	return nil
}
