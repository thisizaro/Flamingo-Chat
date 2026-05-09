package matchmaking

import (
	"time"
)

type QueueStatus string

const (
	StatusWaiting QueueStatus = "waiting"
	StatusMatched QueueStatus = "matched"
	StatusLeft    QueueStatus = "left"
)

type QueueEntry struct {
	UserID     string
	Gender     string
	Preference string
	JoinedAt   time.Time
	Status     QueueStatus
}

type Queue struct {
	entries []*QueueEntry
}

func NewQueue() *Queue {
	return &Queue{
		entries: []*QueueEntry{},
	}
}


func (q *Queue) AddUser(userID, gender, preference string) *QueueEntry {
	entry := &QueueEntry{
		UserID:     userID,
		Gender:     gender,
		Preference: preference,
		JoinedAt:   time.Now(),
		Status:     StatusWaiting,
	}

	q.entries = append(q.entries, entry)
	return entry
}


func (q *Queue) RemoveUser(userID string) {
	for i, e := range q.entries {
		if e.UserID == userID {
			q.entries = append(q.entries[:i], q.entries[i+1:]...)
			return
		}
	}
}

func (q *Queue) RemoveUsers(userIDs ...string) {
	newEntries := []*QueueEntry{}

	for _, e := range q.entries {
		remove := false
		for _, id := range userIDs {
			if e.UserID == id {
				remove = true
				break
			}
		}

		if !remove {
			newEntries = append(newEntries, e)
		}
	}

	q.entries = newEntries
}