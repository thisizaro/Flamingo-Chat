package events

type EventType string

const (

	// incoming events (client → server)
	EventInit        EventType = "init"
	EventJoinQueue   EventType = "join_queue"
	EventSendMessage EventType = "send_message"
	EventPing        EventType = "ping"
	EventLeaveChat   EventType = "leave_chat"
	EventNextPartner EventType = "next_partner"

	// outgoing events (server → client)
	EventReady           EventType = "ready"
	EventQueueJoined     EventType = "queue_joined"
	EventMatchFound      EventType = "match_found"
	EventMessageReceived EventType = "message_received"
	EventPresenceUpdate  EventType = "presence_update"
	EventError           EventType = "error"
	EventChatEnded       EventType = "chat_ended"
)

type Event struct {
	Type    EventType `json:"type"`
	Payload any       `json:"payload"`
}

// --- incoming payloads ---

// InitPayload is the first thing a client sends.
// UserID is empty string if they are new.
// UserID is their stored ID if they are returning.
type InitPayload struct {
	UserID string `json:"user_id"`
}

type JoinQueuePayload struct {
	Gender     string `json:"gender"`
	Preference string `json:"preference"`
}

type SendMessagePayload struct {
	ChatID  string `json:"chat_id"`
	Content string `json:"content"`
}

type NextPartnerPayload struct {
	ChatID string `json:"chat_id"`
}

// --- outgoing payloads ---

// ReadyPayload is what the server sends back after init.
// The client must store this UserID locally (localStorage etc).
type ReadyPayload struct {
	UserID string `json:"user_id"`
}

type MatchFoundPayload struct {
	ChatID    string `json:"chat_id"`
	PartnerID string `json:"partner_id"`
}

type MessageReceivedPayload struct {
	ChatID      string `json:"chat_id"`
	SenderID    string `json:"sender_id"`
	Content     string `json:"content"`
	MessageType string `json:"message_type"`
}

// PresenceUpdatePayload is broadcast to all connected users
// on a regular interval so the frontend can show live counts.
type PresenceUpdatePayload struct {
	TotalOnline  int `json:"total_online"`
	MaleOnline   int `json:"male_online"`
	FemaleOnline int `json:"female_online"`
}

type ErrorPayload struct {
	Message string `json:"message"`
}

type ChatEndedPayload struct {
	Reason string `json:"reason"`
}
