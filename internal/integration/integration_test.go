package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"flamingo_chat/internal/app"
	"flamingo_chat/internal/events"
	"flamingo_chat/internal/server"
	ws "flamingo_chat/internal/transport/websocket"

	"github.com/gorilla/websocket"
)

// startTestServer spins up a real HTTP+WebSocket test server.
// httptest.NewServer gives us a real port, no mocking.
func startTestServer(t *testing.T) *httptest.Server {
	t.Helper()

	srv := server.NewServer()
	hub := ws.NewHub()
	application := app.New(srv, hub)
	application.Wire()

	srv.StartCleanupWorker()
	go hub.Run()

	mux := http.NewServeMux()
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		ws.ServeWS(hub, application.HandleMessage, w, r)
	})

	return httptest.NewServer(mux)
}

// connect opens a WebSocket connection to the test server.
func connect(t *testing.T, serverURL string) *websocket.Conn {
	t.Helper()
	wsURL := "ws" + strings.TrimPrefix(serverURL, "http") + "/ws"
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	return conn
}

// send sends a JSON event over the connection.
func send(t *testing.T, conn *websocket.Conn, eventType events.EventType, payload any) {
	t.Helper()
	data, _ := events.BuildEvent(eventType, payload)
	if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
		t.Fatalf("failed to send event %s: %v", eventType, err)
	}
}

// receive reads one event from the connection, with a timeout.
func receive(t *testing.T, conn *websocket.Conn) *events.Event {
	t.Helper()
	conn.SetReadDeadline(time.Now().Add(3 * time.Second))
	_, data, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("failed to receive: %v", err)
	}

	var raw struct {
		Type    events.EventType `json:"type"`
		Payload json.RawMessage  `json:"payload"`
	}
	json.Unmarshal(data, &raw)

	return &events.Event{
		Type:    raw.Type,
		Payload: raw.Payload,
	}
}

// assertEventType checks the event type and fails the test if wrong.
func assertEventType(t *testing.T, event *events.Event, expected events.EventType) {
	t.Helper()
	if event.Type != expected {
		t.Fatalf("expected event %s, got %s", expected, event.Type)
	}
}

// --- Tests ---

func TestIntegration_InitNewUser(t *testing.T) {
	ts := startTestServer(t)
	defer ts.Close()

	conn := connect(t, ts.URL)
	defer conn.Close()

	// Send init with no user ID (new user)
	send(t, conn, events.EventInit, events.InitPayload{UserID: ""})

	// Expect ready with a generated user ID
	evt := receive(t, conn)
	assertEventType(t, evt, events.EventReady)

	var payload events.ReadyPayload
	json.Unmarshal(evt.Payload.(json.RawMessage), &payload)

	if payload.UserID == "" {
		t.Error("expected a non-empty user ID in ready event")
	}
}

func TestIntegration_InitReturningUser(t *testing.T) {
	ts := startTestServer(t)
	defer ts.Close()

	// First connection — get an ID
	conn1 := connect(t, ts.URL)
	send(t, conn1, events.EventInit, events.InitPayload{UserID: ""})
	evt := receive(t, conn1)
	var ready events.ReadyPayload
	json.Unmarshal(evt.Payload.(json.RawMessage), &ready)
	conn1.Close()

	// Second connection — return with that ID
	conn2 := connect(t, ts.URL)
	defer conn2.Close()
	send(t, conn2, events.EventInit, events.InitPayload{UserID: ready.UserID})
	evt2 := receive(t, conn2)
	assertEventType(t, evt2, events.EventReady)

	var ready2 events.ReadyPayload
	json.Unmarshal(evt2.Payload.(json.RawMessage), &ready2)

	if ready2.UserID != ready.UserID {
		t.Errorf("returning user got different ID: want %s, got %s", ready.UserID, ready2.UserID)
	}
}

func TestIntegration_InvalidUserID_GetsNewID(t *testing.T) {
	ts := startTestServer(t)
	defer ts.Close()

	conn := connect(t, ts.URL)
	defer conn.Close()

	send(t, conn, events.EventInit, events.InitPayload{UserID: "not-a-real-id"})
	evt := receive(t, conn)
	assertEventType(t, evt, events.EventReady)

	var payload events.ReadyPayload
	json.Unmarshal(evt.Payload.(json.RawMessage), &payload)

	if payload.UserID == "not-a-real-id" {
		t.Error("server should have generated a new ID, not echoed the fake one")
	}
	if payload.UserID == "" {
		t.Error("expected a valid generated ID")
	}
}

func TestIntegration_FullMatchAndMessage(t *testing.T) {
	ts := startTestServer(t)
	defer ts.Close()

	// Connect two users
	conn1 := connect(t, ts.URL)
	conn2 := connect(t, ts.URL)
	defer conn1.Close()
	defer conn2.Close()

	// Init both
	send(t, conn1, events.EventInit, events.InitPayload{})
	evt1 := receive(t, conn1)
	var ready1 events.ReadyPayload
	json.Unmarshal(evt1.Payload.(json.RawMessage), &ready1)

	send(t, conn2, events.EventInit, events.InitPayload{})
	evt2 := receive(t, conn2)
	var ready2 events.ReadyPayload
	json.Unmarshal(evt2.Payload.(json.RawMessage), &ready2)

	// Both join queue with compatible preferences
	send(t, conn1, events.EventJoinQueue, events.JoinQueuePayload{Gender: "male", Preference: "female"})
	ack1 := receive(t, conn1) // queue_joined
	assertEventType(t, ack1, events.EventQueueJoined)

	send(t, conn2, events.EventJoinQueue, events.JoinQueuePayload{Gender: "female", Preference: "male"})
	ack2 := receive(t, conn2) // queue_joined
	assertEventType(t, ack2, events.EventQueueJoined)

	// Both should receive match_found
	match1 := receive(t, conn1)
	match2 := receive(t, conn2)
	assertEventType(t, match1, events.EventMatchFound)
	assertEventType(t, match2, events.EventMatchFound)

	// Get chatID from match event
	var matchPayload events.MatchFoundPayload
	json.Unmarshal(match1.Payload.(json.RawMessage), &matchPayload)
	chatID := matchPayload.ChatID

	// User1 sends a message
	send(t, conn1, events.EventSendMessage, events.SendMessagePayload{
		ChatID:  chatID,
		Content: "hello there",
	})

	// User2 should receive it
	msg := receive(t, conn2)
	assertEventType(t, msg, events.EventMessageReceived)

	var msgPayload events.MessageReceivedPayload
	json.Unmarshal(msg.Payload.(json.RawMessage), &msgPayload)

	if msgPayload.Content != "hello there" {
		t.Errorf("expected 'hello there', got '%s'", msgPayload.Content)
	}
	if msgPayload.SenderID != ready1.UserID {
		t.Errorf("wrong sender ID")
	}
}

func TestIntegration_LeaveChat_PartnerNotified(t *testing.T) {
	ts := startTestServer(t)
	defer ts.Close()

	conn1 := connect(t, ts.URL)
	conn2 := connect(t, ts.URL)
	defer conn1.Close()
	defer conn2.Close()

	// Init and match (same as above)
	send(t, conn1, events.EventInit, events.InitPayload{})
	receive(t, conn1)
	send(t, conn2, events.EventInit, events.InitPayload{})
	receive(t, conn2)

	send(t, conn1, events.EventJoinQueue, events.JoinQueuePayload{Gender: "male", Preference: "female"})
	receive(t, conn1)
	send(t, conn2, events.EventJoinQueue, events.JoinQueuePayload{Gender: "female", Preference: "male"})
	receive(t, conn2)

	match1 := receive(t, conn1)
	receive(t, conn2) // match for conn2

	var matchPayload events.MatchFoundPayload
	json.Unmarshal(match1.Payload.(json.RawMessage), &matchPayload)

	// User1 leaves
	send(t, conn1, events.EventLeaveChat, events.NextPartnerPayload{ChatID: matchPayload.ChatID})

	// User2 should get chat_ended
	ended := receive(t, conn2)
	assertEventType(t, ended, events.EventChatEnded)
}
