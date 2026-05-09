package chat

import "time"

type ChatStatus string

const (
	StatusActive ChatStatus = "active"
	StatusEnded  ChatStatus = "ended"
)

type Chat struct {
	ID        string
	UserIDs   []string
	Messages  []*Message
	CreatedAt time.Time
	Status    ChatStatus
}

func (c *Chat) GetOtherUserID(
	userID string,
) string {

	for _, id := range c.UserIDs {

		if id != userID {
			return id
		}
	}

	return ""
}
