package app

import (
	"encoding/json"
	"flamingo_chat/internal/chat"
	"flamingo_chat/internal/events"
	"flamingo_chat/internal/server"
	ws "flamingo_chat/internal/transport/websocket"
	"fmt"
)

// App is the glue between transport and server.
// It owns no state of its own — it just routes.
type App struct {
	server *server.Server
	hub    *ws.Hub
}

func New(srv *server.Server, hub *ws.Hub) *App {
	return &App{
		server: srv,
		hub:    hub,
	}
}

// Wire sets up all server callbacks and the hub disconnect handler.
// Call this once before starting the HTTP server.
func (a *App) Wire() {

	// When two users are matched, send each of them a match_found event.
	a.server.OnMatchFound = func(userAID, userBID, chatID string) {

		dataA, _ := events.BuildEvent(events.EventMatchFound, events.MatchFoundPayload{
			ChatID:    chatID,
			PartnerID: userBID,
		})

		dataB, _ := events.BuildEvent(events.EventMatchFound, events.MatchFoundPayload{
			ChatID:    chatID,
			PartnerID: userAID,
		})

		a.hub.SendToUser(userAID, dataA)
		a.hub.SendToUser(userBID, dataB)
	}

	// When a message is sent, forward it to the receiver.
	a.server.OnMessage = func(receiverID string, data []byte) {
		a.hub.SendToUser(receiverID, data)
	}

	// When a chat ends (partner left), notify the remaining user.
	a.server.OnChatEnded = func(userID string, data []byte) {
		a.hub.SendToUser(userID, data)
	}

	// When a user disconnects from WebSocket, clean up their server state.
	a.hub.SetOnDisconnect(func(userID string) {
		fmt.Println("[APP] User disconnected:", userID)
		// For now: just mark them offline via a session disconnect.
		// In a fuller implementation you'd look up their session ID.
		// The cleanup worker will handle timed-out sessions automatically.
	})
}

// HandleMessage is called by the transport layer for every incoming message.
// client is the WebSocket client, data is raw JSON bytes.
func (a *App) HandleMessage(client *ws.Client, data []byte) {
	event, err := events.ParseEvent(data)
	if err != nil {
		fmt.Println("[APP] Bad event from", client.GetUserID(), ":", err)
		a.sendError(client, "invalid message format")
		return
	}

	fmt.Println("[APP] Event received:", event.Type, "from:", client.GetUserID())

	switch event.Type {

	case events.EventInit:
		a.handleInit(client, event)

	case events.EventJoinQueue:
		a.handleJoinQueue(client, event)

	case events.EventSendMessage:
		a.handleSendMessage(client, event)

	case events.EventPing:
		a.handlePing(client)

	case events.EventLeaveChat:
		a.handleLeaveChat(client, event)

	default:
		fmt.Println("[APP] Unknown event type:", event.Type)
		a.sendError(client, "unknown event type")
	}
}

// --- handlers ---

func (a *App) handleInit(client *ws.Client, event *events.Event) {
	payload, err := unmarshalPayload[events.InitPayload](event.Payload)
	if err != nil {
		a.sendError(client, "invalid init payload")
		return
	}

	// InitUser handles both new users (empty ID) and returning users.
	// If the ID is unknown it creates a new user and returns that.
	u := a.server.InitUser(payload.UserID)

	// Now we know who this client is — tell the hub.
	client.SetUserID(u.ID)
	a.hub.RegisterUser(u.ID, client)

	// Send ready event back to the client with their user ID.
	readyData, _ := events.BuildEvent(events.EventReady, events.ReadyPayload{
		UserID: u.ID,
	})
	client.Send(readyData)

	fmt.Println("[APP] Init complete for user:", u.ID)
}

func (a *App) handleJoinQueue(client *ws.Client, event *events.Event) {
	userID := client.GetUserID()
	if userID == "" {
		a.sendError(client, "not initialized — send init first")
		return
	}

	payload, err := unmarshalPayload[events.JoinQueuePayload](event.Payload)
	if err != nil {
		a.sendError(client, "invalid join_queue payload")
		return
	}

	a.server.SetPreferences(userID, payload.Gender, payload.Preference)
	a.server.JoinQueue(userID)

	// Acknowledge that they joined the queue.
	ack, _ := events.BuildEvent(events.EventQueueJoined, nil)
	client.Send(ack)
}

func (a *App) handleSendMessage(client *ws.Client, event *events.Event) {
	userID := client.GetUserID()
	if userID == "" {
		a.sendError(client, "not initialized")
		return
	}

	payload, err := unmarshalPayload[events.SendMessagePayload](event.Payload)
	if err != nil {
		a.sendError(client, "invalid send_message payload")
		return
	}

	a.server.SendMessage(userID, payload.ChatID, payload.Content, chat.MessageTypeText)
}

func (a *App) handlePing(client *ws.Client) {
	// The WebSocket layer handles low-level ping/pong automatically.
	// This is an app-level ping to keep the session alive on the server.
	// We look up the session by userID here.
	// For now: just update LastActiveAt via SetOnline as a proxy.
	userID := client.GetUserID()
	if userID != "" {
		a.server.SetOnline(userID)
	}
}

func (a *App) handleLeaveChat(client *ws.Client, event *events.Event) {
	userID := client.GetUserID()
	if userID == "" {
		a.sendError(client, "not initialized")
		return
	}

	payload, err := unmarshalPayload[events.NextPartnerPayload](event.Payload)
	if err != nil {
		a.sendError(client, "invalid leave_chat payload")
		return
	}

	a.server.LeaveChat(userID, payload.ChatID)
}

// --- helpers ---

func (a *App) sendError(client *ws.Client, message string) {
	data, _ := events.BuildEvent(events.EventError, events.ErrorPayload{
		Message: message,
	})
	client.Send(data)
}

// unmarshalPayload is a generic helper that re-encodes the payload
// (which comes in as map[string]any after JSON parsing) into the
// correct struct type.
func unmarshalPayload[T any](payload any) (T, error) {
	var result T
	b, err := json.Marshal(payload)
	if err != nil {
		return result, err
	}
	err = json.Unmarshal(b, &result)
	return result, err
}
