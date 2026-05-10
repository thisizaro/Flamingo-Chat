package server

import (
	"fmt"
	"time"

	"flamingo_chat/internal/user"
)

func (s *Server) StartCleanupWorker() {
	go s.runCleanup()
	go s.runPresenceBroadcast()
}

// runCleanup ticks every 10 seconds and removes inactive sessions.
func (s *Server) runCleanup() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		s.cleanupInactiveSessions()
	}
}

// runPresenceBroadcast ticks every 5 seconds and pushes
// live online counts to all connected clients.
func (s *Server) runPresenceBroadcast() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		s.BroadcastPresence()
	}
}

func (s *Server) cleanupInactiveSessions() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()

	for sessionID, session := range s.state.Sessions {
		if !session.IsActive {
			continue
		}

		if now.Sub(session.LastPingAt) > 30*time.Second {
			session.IsActive = false

			u, ok := s.state.Users[session.UserID]
			if ok {
				u.Status = user.StatusOffline
				s.queue.RemoveUsers(u.ID)
				fmt.Println("[WORKER] Session timed out, user offline:", u.ID)
			} else {
				fmt.Println("[WORKER] Session timed out:", sessionID)
			}
		}
	}
}
