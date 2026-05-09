package matchmaking

import "testing"

func TestFindMatch_BasicPair(t *testing.T) {
	q := NewQueue()
	q.AddUser("user1", "male", "female")
	q.AddUser("user2", "female", "male")

	a, b := q.FindMatch()

	if a == nil || b == nil {
		t.Fatal("expected a match, got nil")
	}

	ids := []string{a.UserID, b.UserID}
	if !contains(ids, "user1") || !contains(ids, "user2") {
		t.Errorf("matched wrong users: %v", ids)
	}
}

func TestFindMatch_NoMatchWrongPreference(t *testing.T) {
	q := NewQueue()
	q.AddUser("user1", "male", "female")
	q.AddUser("user2", "male", "female") // both want female, neither qualifies

	a, b := q.FindMatch()

	if a != nil || b != nil {
		t.Error("expected no match, got one")
	}
}

func TestFindMatch_EmptyQueue(t *testing.T) {
	q := NewQueue()
	a, b := q.FindMatch()
	if a != nil || b != nil {
		t.Error("expected no match on empty queue")
	}
}

func TestFindMatch_OneUser(t *testing.T) {
	q := NewQueue()
	q.AddUser("user1", "male", "female")
	a, b := q.FindMatch()
	if a != nil || b != nil {
		t.Error("expected no match with only one user")
	}
}

func TestFindMatch_PicksFirstCompatiblePair(t *testing.T) {
	q := NewQueue()
	q.AddUser("user1", "male", "female")
	q.AddUser("user2", "male", "female") // incompatible with user1
	q.AddUser("user3", "female", "male") // compatible with user1

	a, b := q.FindMatch()
	if a == nil || b == nil {
		t.Fatal("expected a match")
	}

	ids := []string{a.UserID, b.UserID}
	if !contains(ids, "user1") || !contains(ids, "user3") {
		t.Errorf("expected user1+user3, got %v", ids)
	}
}

func TestRemoveUser(t *testing.T) {
	q := NewQueue()
	q.AddUser("user1", "male", "female")
	q.AddUser("user2", "female", "male")
	q.RemoveUser("user1")

	a, b := q.FindMatch()
	if a != nil || b != nil {
		t.Error("expected no match after removing user1")
	}
}

func TestRemoveUsers_Multiple(t *testing.T) {
	q := NewQueue()
	q.AddUser("user1", "male", "female")
	q.AddUser("user2", "female", "male")
	q.RemoveUsers("user1", "user2")

	a, b := q.FindMatch()
	if a != nil || b != nil {
		t.Error("expected no match after removing both users")
	}
}

func contains(slice []string, val string) bool {
	for _, s := range slice {
		if s == val {
			return true
		}
	}
	return false
}
