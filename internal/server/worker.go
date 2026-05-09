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

			fmt.Println("[WORKER] Running cleanup worker")

			s.cleanupInactiveSessions()
		}
	}()
}

func (s *Server) cleanupInactiveSessions() {

	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()

	for sessionID, session := range s.state.Sessions {

		// already inactive
		if !session.IsActive {
			continue
		}

		// timeout check
		if now.Sub(session.LastPingAt) > 30*time.Second {

			fmt.Println("[WORKER] Session timeout")
			fmt.Println("  SessionID:", sessionID)

			session.IsActive = false

			u, ok := s.state.Users[session.UserID]
			if ok {

				u.Status = user.StatusOffline

				// remove from queue
				s.queue.RemoveUsers(u.ID)

				fmt.Println("[WORKER] User removed from active state")
				fmt.Println("  UserID:", u.ID)
			}
		}
	}
}