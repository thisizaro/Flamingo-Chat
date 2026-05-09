package server

import (
	"fmt"

	"flamingo_chat/internal/user"
)

type PresenceStats struct {
	TotalOnline  int
	MaleOnline   int
	FemaleOnline int
}

func (s *Server) GetPresenceStats() PresenceStats {
	s.mu.Lock()
	defer s.mu.Unlock()

	stats := PresenceStats{}

	for _, u := range s.state.Users {

		if u.Status == user.StatusOffline {
			continue
		}

		stats.TotalOnline++

		switch u.Gender {
		case "male":
			stats.MaleOnline++

		case "female":
			stats.FemaleOnline++
		}
	}

	fmt.Println("[PRESENCE] Stats calculated")
	fmt.Println("  Total:", stats.TotalOnline)
	fmt.Println("  Male:", stats.MaleOnline)
	fmt.Println("  Female:", stats.FemaleOnline)

	return stats
}