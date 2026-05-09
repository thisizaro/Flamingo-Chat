package user

import "time"

type Status string

const (
	StatusIdle    Status = "idle"
	StatusInQueue Status = "in_queue"
	StatusInChat  Status = "in_chat"
	StatusOffline Status = "offline"
	StatusOnline  Status = "online"
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
