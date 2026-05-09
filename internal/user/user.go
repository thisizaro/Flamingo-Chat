package user

import "time"

type Status string

const (
	StatusOffline Status = "offline"  // not connected
	StatusIdle    Status = "idle"     // connected, not in queue or chat
	StatusInQueue Status = "in_queue" // waiting for a match
	StatusInChat  Status = "in_chat"  // currently in a chat
)

type User struct {
	ID           string
	Gender       string
	Preference   string
	Status       Status
	CreatedAt    time.Time
	LastActiveAt time.Time
	IsBanned     bool
}
