package fake

import (
	"math"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/jschaf/b2/pkg/api"
)

func TestEventFaker_WriteEvents_deterministic(t *testing.T) {
	faker1 := NewEventFaker()
	faker2 := NewEventFaker()
	evs1 := make([]api.Event, 10)
	evs2 := make([]api.Event, 10)

	if err := faker1.WriteEvents(evs1); err != nil {
		t.Error(err)
	}

	if err := faker2.WriteEvents(evs2); err != nil {
		t.Error(err)
	}

	if diff := cmp.Diff(evs1, evs2); diff != "" {
		t.Errorf("WriteEvents() non-deterministic output - diffs:\n%s", diff)
	}
}

func TestEventFaker_WriteEvents_noNullBytes(t *testing.T) {
	faker := NewEventFaker()
	evs := make([]api.Event, 1<<14)
	if err := faker.WriteEvents(evs); err != nil {
		t.Error(err)
	}

	for _, ev := range evs {
		for _, ch := range ev.Data {
			if ch == 0 {
				t.Fatalf("got null byte in JSON")
			}
		}
	}
}

func TestEventFaker_WriteEvents_overwrites(t *testing.T) {
	faker := NewEventFaker()
	evs := make([]api.Event, 10)

	// Fill with JS ints which are always less than 2^53 - 1.
	evs[4].EventID = math.MaxInt64
	evs[4].UserID = math.MaxInt64
	evs[4].Time = math.MaxInt64
	evs[4].SessionID = math.MaxInt64
	preData := []byte("foobar")
	evs[4].Data = preData

	if err := faker.WriteEvents(evs); err != nil {
		t.Error(err)
	}

	if evs[4].EventID == math.MaxInt64 {
		t.Errorf("expected event ID to be overwritten")
	}

	if evs[4].UserID == math.MaxInt64 {
		t.Errorf("expected user ID to be overwritten")
	}

	if evs[4].Time == math.MaxInt64 {
		t.Errorf("expected time to be overwritten")
	}

	if evs[4].SessionID == math.MaxInt64 {
		t.Errorf("expected session ID to be overwritten")
	}

	if string(evs[4].Data) == string(preData) {
		t.Errorf("expected data to be overwritten")
	}
}

func BenchmarkEventFaker_WriteEvents(b *testing.B) {
	evs := make([]api.Event, 8192)

	for i := 0; i < b.N; i++ {
		f := NewEventFaker()
		if err := f.WriteEvents(evs); err != nil {
			b.Error(err)
		}
	}
}
