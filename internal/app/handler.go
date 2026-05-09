package app

import (
	"encoding/json"
	"fmt"

	"flamingo_chat/internal/events"
	"flamingo_chat/internal/server"
)

type Sender interface {
	SendToClient(userID string, data []byte)
	Broadcast(data []byte)
}

type Handler struct {
	server *server.Server
	sender Sender
}

func NewHandler(s *server.Server, sender Sender) *Handler {
	return &Handler{
		server: s,
		sender: sender,
	}
}

func (h *Handler) Handle(clientID string, event *events.Event) {

	switch event.Type {

	case events.EventJoinQueue:
		h.handleJoinQueue(clientID, event.Payload)

	case events.EventSendMessage:
		h.handleSendMessage(clientID, event.Payload)

	case events.EventNextPartner:
		h.handleNextPartner(clientID, event.Payload)

	default:
		fmt.Println("[APP] Unknown event:", event.Type)
	}
}

func (h *Handler) handleJoinQueue(clientID string, payload any) {

	payloadBytes, _ := json.Marshal(payload)

	var joinPayload events.JoinQueuePayload

	err := json.Unmarshal(payloadBytes, &joinPayload)
	if err != nil {
		fmt.Println("[APP] Invalid join_queue payload")
		return
	}

	h.server.SetPreferences(
		clientID,
		joinPayload.Gender,
		joinPayload.Preference,
	)

	h.server.JoinQueue(clientID)
	h.broadcastPresence()

	response, _ := events.BuildEvent(
		events.EventQueueJoined,
		map[string]string{"status": "joined"},
	)

	h.sender.SendToClient(clientID, response)
}

func (h *Handler) handleSendMessage(clientID string, payload any) {

	payloadBytes, _ := json.Marshal(payload)

	var msgPayload events.SendMessagePayload

	err := json.Unmarshal(payloadBytes, &msgPayload)
	if err != nil {
		fmt.Println("[APP] Invalid message payload")
		return
	}

	h.server.SendMessage(
		clientID,
		msgPayload.ChatID,
		msgPayload.Content,
		"text",
	)
}

func (h *Handler) handleNextPartner(clientID string, payload any) {

	payloadBytes, _ := json.Marshal(payload)

	var nextPayload events.NextPartnerPayload

	err := json.Unmarshal(payloadBytes, &nextPayload)
	if err != nil {
		fmt.Println("[APP] Invalid next_partner payload")
		return
	}

	h.server.LeaveChat(clientID, nextPayload.ChatID)

	h.server.JoinQueue(clientID)
	h.broadcastPresence()
}

func (h *Handler) broadcastPresence() {

	stats := h.server.GetPresenceStats()

	data, err := events.BuildEvent(
		events.EventPresenceUpdate,
		stats,
	)

	if err != nil {
		fmt.Println("[APP] Failed to build presence event")
		return
	}

	h.sender.Broadcast(data)
}
