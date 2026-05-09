package server

import (
	"fmt"
	"time"

	"flamingo_chat/internal/user"
)

func (s *Server) StartCleanupWorker() {
	go func() {
		for {
			time.Sleep(10 * time.Second)
			s.cleanupInactiveSessions()
		}
	}()
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
