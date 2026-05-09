package events

type EventType string

const (

	// incoming events
	EventJoinQueue   EventType = "join_queue"
	EventSendMessage EventType = "send_message"
	EventPing        EventType = "ping"
	EventLeaveChat   EventType = "leave_chat"
	EventNextPartner EventType = "next_partner"

	// outgoing events
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

type JoinQueuePayload struct {
	Gender     string `json:"gender"`
	Preference string `json:"preference"`
}

type SendMessagePayload struct {
	ChatID  string `json:"chat_id"`
	Content string `json:"content"`
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

type ErrorPayload struct {
	Message string `json:"message"`
}

type NextPartnerPayload struct {
	ChatID string `json:"chat_id"`
}

type ChatEndedPayload struct {
	Reason string `json:"reason"`
}
