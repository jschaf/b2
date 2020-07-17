package atomics

import (
	"math"
	"sync"
	"testing"
)

func TestNewBool_startValue(t *testing.T) {
	t.Parallel()
	b1 := NewBool(true)
	assertTrue(t, b1)

	b2 := NewBool(false)
	assertFalse(t, b2)
}

func TestBool_Set(t *testing.T) {
	t.Parallel()
	b1 := NewBool(true)
	b1.Set(true)
	assertTrue(t, b1)
	b1.Set(false)
	assertFalse(t, b1)
	b1.Set(true)
	assertTrue(t, b1)
	b1.Set(false)
	assertFalse(t, b1)
}

func TestBool_Toggle(t *testing.T) {
	t.Parallel()
	b1 := NewBool(true)
	b1.Toggle()
	assertFalse(t, b1)
	b1.Toggle()
	assertTrue(t, b1)
	b1.Toggle()
	assertFalse(t, b1)
}

func TestBool_Toggle_overflow(t *testing.T) {
	t.Parallel()
	m := int32(math.MaxInt32)
	b1 := (*Bool)(&m)
	assertTrue(t, b1)
	b1.Toggle()
	assertFalse(t, b1)
	b1.Toggle()
	assertTrue(t, b1)
	b1.Toggle()
	assertFalse(t, b1)
}

func TestBool_race(t *testing.T) {
	t.Parallel()
	repeat := 10000
	wg := sync.WaitGroup{}
	wg.Add(repeat * 4)
	ab := NewBool(false)

	go func() {
		for i := 0; i < repeat; i++ {
			ab.Set(true)
			wg.Done()
		}
	}()

	go func() {
		for i := 0; i < repeat; i++ {
			ab.Get()
			wg.Done()
		}
	}()

	go func() {
		for i := 0; i < repeat; i++ {
			ab.Set(false)
			wg.Done()
		}
	}()

	go func() {
		for i := 0; i < repeat; i++ {
			ab.Toggle()
			wg.Done()
		}
	}()

	wg.Wait()
}

func assertTrue(t *testing.T, b *Bool) {
	t.Helper()
	if !b.Get() {
		t.Fatalf("expected true; got false")
	}
}

func assertFalse(t *testing.T, b *Bool) {
	t.Helper()
	if b.Get() {
		t.Fatalf("expected false; got true")
	}
}
