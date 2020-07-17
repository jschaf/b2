package atomics

import "sync/atomic"

// Bool is an atomic boolean. All methods are atomic and safe to use from
// multiple goroutines.
type Bool int32

const (
	boolFalse = int32(0)
	boolTrue  = int32(1)
)

func NewBool(val bool) *Bool {
	ab := new(int32)
	if val {
		*ab = boolTrue
	}
	return (*Bool)(ab)
}

// Get returns the current boolean value.
func (ab *Bool) Get() bool {
	return atomic.LoadInt32((*int32)(ab))&1 == boolTrue
}

// Set sets the boolean value.
func (ab *Bool) Set(val bool) {
	if val {
		atomic.StoreInt32((*int32)(ab), boolTrue)
	} else {
		atomic.StoreInt32((*int32)(ab), boolFalse)
	}
}

// Toggle returns the current value and inverts the boolean value.
func (ab *Bool) Toggle() bool {
	return atomic.AddInt32((*int32)(ab), 1)&1 == boolFalse
}
