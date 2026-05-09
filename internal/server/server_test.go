package server

import (
	"testing"

	"flamingo_chat/internal/user"
)

// --- InitUser ---

func TestInitUser_NewUser(t *testing.T) {
	s := NewServer()
	u := s.InitUser("")

	if u == nil {
		t.Fatal("expected a user, got nil")
	}
	if u.ID == "" {
		t.Error("expected a generated ID, got empty string")
	}
	if u.Status != user.StatusIdle {
		t.Errorf("expected StatusIdle, got %s", u.Status)
	}
}

func TestInitUser_ReturningUser(t *testing.T) {
	s := NewServer()

	// create user first
	first := s.InitUser("")
	id := first.ID

	// now "return" with that ID
	second := s.InitUser(id)

	if second.ID != id {
		t.Errorf("expected same ID %s, got %s", id, second.ID)
	}
}

func TestInitUser_UnknownID_CreatesNew(t *testing.T) {
	s := NewServer()
	u := s.InitUser("totally-fake-id")

	if u == nil {
		t.Fatal("expected a new user, got nil")
	}
	if u.ID == "totally-fake-id" {
		t.Error("should have generated a new ID, not reused the fake one")
	}
}

func TestInitUser_UniqueIDs(t *testing.T) {
	s := NewServer()
	ids := map[string]bool{}

	for i := 0; i < 100; i++ {
		u := s.InitUser("")
		if ids[u.ID] {
			t.Fatalf("duplicate ID generated: %s", u.ID)
		}
		ids[u.ID] = true
	}
}

// --- JoinQueue / TryMatch ---

func TestJoinQueue_StatusChanges(t *testing.T) {
	s := NewServer()

	u := s.InitUser("")
	s.SetPreferences(u.ID, "male", "female")
	s.JoinQueue(u.ID)

	s.mu.Lock()
	status := s.state.Users[u.ID].Status
	s.mu.Unlock()

	// If no match found, user stays in queue
	if status != user.StatusInQueue {
		t.Errorf("expected StatusInQueue, got %s", status)
	}
}

func TestTryMatch_TwoCompatibleUsers(t *testing.T) {
	s := NewServer()

	matched := false
	s.OnMatchFound = func(userAID, userBID, chatID string) {
		matched = true
	}

	u1 := s.InitUser("")
	u2 := s.InitUser("")

	s.SetPreferences(u1.ID, "male", "female")
	s.SetPreferences(u2.ID, "female", "male")

	s.JoinQueue(u1.ID)
	s.JoinQueue(u2.ID)

	if !matched {
		t.Error("expected OnMatchFound to be called")
	}

	s.mu.Lock()
	status1 := s.state.Users[u1.ID].Status
	status2 := s.state.Users[u2.ID].Status
	s.mu.Unlock()

	if status1 != user.StatusInChat {
		t.Errorf("user1: expected StatusInChat, got %s", status1)
	}
	if status2 != user.StatusInChat {
		t.Errorf("user2: expected StatusInChat, got %s", status2)
	}
}

func TestTryMatch_IncompatibleUsers_NoMatch(t *testing.T) {
	s := NewServer()

	matched := false
	s.OnMatchFound = func(a, b, c string) { matched = true }

	u1 := s.InitUser("")
	u2 := s.InitUser("")

	s.SetPreferences(u1.ID, "male", "female")
	s.SetPreferences(u2.ID, "male", "female")

	s.JoinQueue(u1.ID)
	s.JoinQueue(u2.ID)

	if matched {
		t.Error("incompatible users should not be matched")
	}
}

// --- LeaveChat ---

func TestLeaveChat_NotifiesPartner(t *testing.T) {
	s := NewServer()

	notifiedUser := ""
	s.OnChatEnded = func(userID string, data []byte) {
		notifiedUser = userID
	}
	s.OnMatchFound = func(a, b, chatID string) {}

	u1 := s.InitUser("")
	u2 := s.InitUser("")
	s.SetPreferences(u1.ID, "male", "female")
	s.SetPreferences(u2.ID, "female", "male")

	var chatID string
	s.OnMatchFound = func(a, b, cID string) { chatID = cID }

	s.JoinQueue(u1.ID)
	s.JoinQueue(u2.ID)

	if chatID == "" {
		t.Fatal("no chat was created")
	}

	s.LeaveChat(u1.ID, chatID)

	if notifiedUser != u2.ID {
		t.Errorf("expected partner %s to be notified, got %s", u2.ID, notifiedUser)
	}
}

func TestLeaveChat_StatusResetToIdle(t *testing.T) {
	s := NewServer()

	var chatID string
	s.OnMatchFound = func(a, b, cID string) { chatID = cID }
	s.OnChatEnded = func(string, []byte) {}

	u1 := s.InitUser("")
	u2 := s.InitUser("")
	s.SetPreferences(u1.ID, "male", "female")
	s.SetPreferences(u2.ID, "female", "male")

	s.JoinQueue(u1.ID)
	s.JoinQueue(u2.ID)

	s.LeaveChat(u1.ID, chatID)

	s.mu.Lock()
	s1 := s.state.Users[u1.ID].Status
	s2 := s.state.Users[u2.ID].Status
	s.mu.Unlock()

	if s1 != user.StatusIdle {
		t.Errorf("u1: expected StatusIdle after leaving, got %s", s1)
	}
	if s2 != user.StatusIdle {
		t.Errorf("u2: expected StatusIdle after chat ended, got %s", s2)
	}
}

// --- SendMessage ---

func TestSendMessage_DeliveredToPartner(t *testing.T) {
	s := NewServer()

	var receiverID string
	s.OnMessage = func(rID string, data []byte) { receiverID = rID }

	var chatID string
	s.OnMatchFound = func(a, b, cID string) { chatID = cID }
	s.OnChatEnded = func(string, []byte) {}

	u1 := s.InitUser("")
	u2 := s.InitUser("")
	s.SetPreferences(u1.ID, "male", "female")
	s.SetPreferences(u2.ID, "female", "male")

	s.JoinQueue(u1.ID)
	s.JoinQueue(u2.ID)

	s.SendMessage(u1.ID, chatID, "hello", "text")

	if receiverID != u2.ID {
		t.Errorf("expected message delivered to %s, got %s", u2.ID, receiverID)
	}
}
