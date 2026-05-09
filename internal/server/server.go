package server

import (
	"crypto/rand"
	"encoding/hex"
	"flamingo_chat/internal/chat"
	"flamingo_chat/internal/events"
	"flamingo_chat/internal/matchmaking"
	"flamingo_chat/internal/user"
	"fmt"
	"sync"
	"time"
)

type Server struct {
	state *State
	queue *matchmaking.Queue
	chats *chat.Manager
	mu    sync.Mutex

	// Callbacks — set by the app layer, not main.
	// The server calls these when something happens that
	// the outside world (transport) needs to know about.

	// OnMatchFound: tell both users they have been matched
	OnMatchFound func(userAID, userBID, chatID string)

	// OnMessage: deliver a message to a specific user
	OnMessage func(receiverID string, data []byte)

	// OnChatEnded: tell a user their chat has ended
	OnChatEnded func(userID string, data []byte)

	// OnReady: send the ready event back to a specific user
	OnReady func(userID string, data []byte)
}

func NewServer() *Server {
	return &Server{
		state: &State{
			Users:    make(map[string]*user.User),
			Sessions: make(map[string]*user.Session),
		},
		queue: matchmaking.NewQueue(),
		chats: chat.NewManager(),
	}
}

// generateUserID creates a short random hex ID, e.g. "a3f9c2b1e4d70c88"
func generateUserID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// --- Init ---

// InitUser handles a client connecting for the first time or returning.
// If userID is empty or not found, a new user is created.
// Returns the user (new or existing) so the caller can send EventReady.
func (s *Server) InitUser(userID string) *user.User {
	s.mu.Lock()
	defer s.mu.Unlock()

	// returning user — found in state
	if userID != "" {
		if u, ok := s.state.Users[userID]; ok {
			u.Status = user.StatusIdle
			u.LastActiveAt = time.Now()
			fmt.Println("[SERVER] Returning user:", u.ID)
			return u
		}
		// ID provided but not found — fall through and make a new one
		fmt.Println("[SERVER] Unknown user ID, creating new user:", userID)
	}

	// new user
	u := &user.User{
		ID:           generateUserID(),
		Status:       user.StatusIdle,
		CreatedAt:    time.Now(),
		LastActiveAt: time.Now(),
	}
	s.state.Users[u.ID] = u
	fmt.Println("[SERVER] New user created:", u.ID)
	return u
}

// --- User management ---

func (s *Server) SetPreferences(id, gender, preference string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if u, ok := s.state.Users[id]; ok {
		u.Gender = gender
		u.Preference = preference
		u.LastActiveAt = time.Now()
	}
}

func (s *Server) SetOnline(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if u, ok := s.state.Users[id]; ok {
		u.Status = user.StatusIdle
		u.LastActiveAt = time.Now()
	}
}

// --- Queue ---

func (s *Server) JoinQueue(userID string) {
	s.mu.Lock()

	u, ok := s.state.Users[userID]
	if !ok {
		s.mu.Unlock()
		return
	}

	u.Status = user.StatusInQueue
	u.LastActiveAt = time.Now()
	s.queue.AddUser(u.ID, u.Gender, u.Preference)

	s.mu.Unlock()

	// TryMatch is called WITHOUT the lock held.
	// TryMatch acquires its own lock internally.
	// Do not call TryMatch while holding the lock — deadlock.
	s.TryMatch()
}

func (s *Server) TryMatch() {
	s.mu.Lock()
	defer s.mu.Unlock()

	a, b := s.queue.FindMatch()
	if a == nil || b == nil {
		return
	}

	s.queue.RemoveUsers(a.UserID, b.UserID)

	userA := s.state.Users[a.UserID]
	userB := s.state.Users[b.UserID]

	userA.Status = user.StatusInChat
	userB.Status = user.StatusInChat

	chatID := generateUserID() // random chat ID, no collision risk
	newChat := s.chats.CreateChat(chatID, []string{userA.ID, userB.ID})

	fmt.Println("[SERVER] Matched:", userA.ID, "with", userB.ID)
	fmt.Println("[SERVER] Chat created:", newChat.ID)

	if s.OnMatchFound != nil {
		s.OnMatchFound(userA.ID, userB.ID, newChat.ID)
	}
}

// --- Messaging ---

func (s *Server) SendMessage(
	userID,
	chatID,
	content string,
	messageType chat.MessageType,
) {
	s.mu.Lock()
	defer s.mu.Unlock()

	u, ok := s.state.Users[userID]
	if !ok {
		fmt.Println("[SERVER] User not found:", userID)
		return
	}

	if u.Status != user.StatusInChat {
		fmt.Println("[SERVER] User not in chat:", userID)
		return
	}

	s.chats.SendMessage(chatID, userID, content, messageType)

	chatRoom, ok := s.chats.GetChat(chatID)
	if !ok {
		return
	}

	receiverID := chatRoom.GetOtherUserID(userID)

	eventData, _ := events.BuildEvent(
		events.EventMessageReceived,
		events.MessageReceivedPayload{
			ChatID:      chatID,
			SenderID:    userID,
			Content:     content,
			MessageType: string(messageType),
		},
	)

	if s.OnMessage != nil {
		s.OnMessage(receiverID, eventData)
	}
}

// --- Sessions ---

func (s *Server) CreateSession(sessionID, userID, token string) *user.Session {
	s.mu.Lock()
	defer s.mu.Unlock()

	session := &user.Session{
		ID:           sessionID,
		UserID:       userID,
		SessionToken: token,
		StartedAt:    time.Now(),
		LastPingAt:   time.Now(),
		IsActive:     true,
	}

	s.state.Sessions[sessionID] = session

	fmt.Println("[SESSION] Created — SessionID:", sessionID, "UserID:", userID)
	return session
}

func (s *Server) PingSession(sessionID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	session, ok := s.state.Sessions[sessionID]
	if !ok {
		return
	}

	session.LastPingAt = time.Now()
}

func (s *Server) DisconnectSession(sessionID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	session, ok := s.state.Sessions[sessionID]
	if !ok {
		return
	}

	session.IsActive = false

	u, exists := s.state.Users[session.UserID]
	if exists {
		u.Status = user.StatusOffline
		u.LastActiveAt = time.Now()
		s.queue.RemoveUsers(u.ID)
		fmt.Println("[SESSION] User offline:", u.ID)
	}

	fmt.Println("[SESSION] Disconnected:", sessionID)
}

// --- Chat ---

func (s *Server) LeaveChat(userID, chatID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	chatRoom, ok := s.chats.GetChat(chatID)
	if !ok {
		fmt.Println("[CHAT] Chat not found:", chatID)
		return
	}

	partnerID := chatRoom.GetOtherUserID(userID)

	s.chats.EndChat(chatID)
	s.chats.RemoveChat(chatID)

	// both users go back to idle, ready to queue again
	if u, ok := s.state.Users[userID]; ok {
		u.Status = user.StatusIdle
	}
	if u, ok := s.state.Users[partnerID]; ok {
		u.Status = user.StatusIdle
	}

	fmt.Println("[CHAT] Ended:", chatID)

	if s.OnChatEnded != nil {
		data, _ := events.BuildEvent(
			events.EventChatEnded,
			events.ChatEndedPayload{Reason: "partner_left"},
		)
		s.OnChatEnded(partnerID, data)
	}
}

// UserExists checks if a user ID is known to the server.
func (s *Server) UserExists(userID string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, ok := s.state.Users[userID]
	return ok
}
