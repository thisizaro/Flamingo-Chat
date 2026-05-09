package chat

import "time"

type MessageType string

const (
	MessageTypeText   MessageType = "text"
	MessageTypeSystem MessageType = "system"
)

type Message struct {
	ID          string
	ChatID      string
	SenderID    string
	Content     string
	MessageType MessageType
	CreatedAt   time.Time
}