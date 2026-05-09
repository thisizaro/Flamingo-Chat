package server

import (
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

	mu sync.Mutex

	OnMatchFound func(
		userAID string,
		userBID string,
		chatID string,
	)

	OnMessage func(
		receiverID string,
		data []byte,
	)

	OnChatEnded func(
		userID string,
		data []byte,
	)
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

// Create a new user
func (s *Server) CreateUser(id string) *user.User {
	s.mu.Lock()
	defer s.mu.Unlock()

	u := &user.User{
		ID:     id,
		Status: user.StatusIdle,
	}

	s.state.Users[id] = u
	return u
}

// Set preferences
func (s *Server) SetPreferences(id, gender, preference string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if u, ok := s.state.Users[id]; ok {
		u.Gender = gender
		u.Preference = preference
		u.LastActiveAt = time.Now()
	}
}

// Set user online
func (s *Server) SetOnline(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if u, ok := s.state.Users[id]; ok {
		u.Status = user.StatusIdle
		u.LastActiveAt = time.Now()
	}
}

// Join Queue function
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

	// try matching AFTER unlock (important)
	s.TryMatch()
}

func (s *Server) TryMatch() {
	s.mu.Lock()
	defer s.mu.Unlock()

	a, b := s.queue.FindMatch()
	if a == nil || b == nil {
		return
	}

	// remove from queue
	s.queue.RemoveUsers(a.UserID, b.UserID)

	// update user states
	userA := s.state.Users[a.UserID]
	userB := s.state.Users[b.UserID]

	userA.Status = user.StatusInChat
	userB.Status = user.StatusInChat

	// TEMP: just print match
	fmt.Println("Matched:", userA.ID, "with", userB.ID)

	// create chat
	chatID := a.UserID + "_" + b.UserID // simple for now
	chat := s.chats.CreateChat(chatID, []string{userA.ID, userB.ID})

	fmt.Println("Matched:", userA.ID, "with", userB.ID)
	fmt.Println("Chat created:", chat.ID)

	// trigger match found callback
	if s.OnMatchFound != nil {

		s.OnMatchFound(
			userA.ID,
			userB.ID,
			chat.ID,
		)
	}
}

// send message.
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

	fmt.Println("[SERVER] Processing message from:", userID)

	// send message to chat manager
	s.chats.SendMessage(
		chatID,
		userID,
		content,
		messageType,
	)

	// get chat room to find receiver
	chatRoom, ok := s.chats.GetChat(chatID)

	if !ok {
		return
	}

	// find the other user in the chat
	receiverID := chatRoom.GetOtherUserID(userID)

	// build event data
	eventData, _ := events.BuildEvent(
		events.EventMessageReceived,
		events.MessageReceivedPayload{
			ChatID:      chatID,
			SenderID:    userID,
			Content:     content,
			MessageType: string(messageType),
		},
	)

	// trigger message callback
	if s.OnMessage != nil {

		s.OnMessage(
			receiverID,
			eventData,
		)
	}
}

// create session
func (s *Server) CreateSession(
	sessionID,
	userID,
	token string,
) *user.Session {

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

	fmt.Println("[SESSION] Session created")
	fmt.Println("  SessionID:", sessionID)
	fmt.Println("  UserID:", userID)

	return session
}

func (s *Server) PingSession(sessionID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	session, ok := s.state.Sessions[sessionID]
	if !ok {
		fmt.Println("[SESSION] Session not found:", sessionID)
		return
	}

	session.LastPingAt = time.Now()

	fmt.Println("[SESSION] Ping updated")
	fmt.Println("  SessionID:", sessionID)
}

func (s *Server) DisconnectSession(sessionID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	session, ok := s.state.Sessions[sessionID]
	if !ok {
		fmt.Println("[SESSION] Session not found:", sessionID)
		return
	}

	session.IsActive = false

	u, exists := s.state.Users[session.UserID]
	if exists {
		u.Status = user.StatusOffline
		u.LastActiveAt = time.Now()

		// remove from queue if needed
		s.queue.RemoveUsers(u.ID)

		fmt.Println("[SESSION] User marked offline")
		fmt.Println("  UserID:", u.ID)
	}

	fmt.Println("[SESSION] Session disconnected")
	fmt.Println("  SessionID:", sessionID)
}

func (s *Server) LeaveChat(
	userID string,
	chatID string,
) {

	s.mu.Lock()
	defer s.mu.Unlock()

	chatRoom, ok := s.chats.GetChat(chatID)

	if !ok {

		fmt.Println("[CHAT] Chat not found")
		return
	}

	partnerID := chatRoom.GetOtherUserID(userID)

	// end chat
	s.chats.EndChat(chatID)

	// remove chat
	s.chats.RemoveChat(chatID)

	// reset users
	if u, ok := s.state.Users[userID]; ok {

		u.Status = user.StatusOnline
	}

	if u, ok := s.state.Users[partnerID]; ok {

		u.Status = user.StatusOnline
	}

	fmt.Println("[CHAT] Chat ended")
	fmt.Println("  ChatID:", chatID)

	// notify partner
	if s.OnChatEnded != nil {

		data, _ := events.BuildEvent(
			events.EventChatEnded,
			events.ChatEndedPayload{
				Reason: "partner_left",
			},
		)

		s.OnChatEnded(partnerID, data)
	}
}
