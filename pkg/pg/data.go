package pg

import (
	"errors"
	"github.com/jschaf/b2/pkg/api"
)

// EventSource provides events to Postgres in a one-time manner similar to
// io.Reader. Once an event is sourced, it's not retained.
type EventSource struct {
	evs []api.Event
	idx int
}

// NewEventSource creates a new source of events. The provided events must not
// be modified.
func NewEventSource(evs []api.Event) *EventSource {
	return &EventSource{
		evs: evs,
		idx: 0,
	}
}

func (s *EventSource) Rows() []string {
	return []string{"event_id", "user_id", "time", "session_id", "data"}
}

func (s *EventSource) Next() bool {
	return s.idx < len(s.evs)
}

func (s *EventSource) Values() ([]interface{}, error) {
	if s.idx >= len(s.evs) {
		return nil, errors.New("index out of bounds - api.Event pgx.CopyFromSource")
	}
	row := []interface{}{
		s.evs[s.idx].EventID,
		s.evs[s.idx].UserID,
		s.evs[s.idx].Time,
		s.evs[s.idx].SessionID,
		s.evs[s.idx].Data,
	}
	s.idx++
	return row, nil
}

func (s *EventSource) Err() error {
	return nil
}
