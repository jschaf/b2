package api

import "encoding/json"

type (
	EventID     int
	UserID      int
	SessionID   int
	EpochMillis int
)

type Event struct {
	EventID   EventID
	UserID    UserID
	Time      EpochMillis
	SessionID SessionID
	Data      json.RawMessage
}
