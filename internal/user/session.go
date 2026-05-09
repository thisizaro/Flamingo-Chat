package user

import "time"

type Session struct {
	ID           string
	UserID       string
	SessionToken string
	StartedAt    time.Time
	LastPingAt   time.Time
	IsActive     bool
}