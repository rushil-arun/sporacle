package state_test

import (
	"testing"

	game "server/game-metadata"
	state "server/state"
)

func TestNewGlobalState(t *testing.T) {
	s := state.NewGlobalState()
	if s == nil {
		t.Fatal("NewGlobalState returned nil")
	}
	// New state should have no games and no usernames
	if state.NewGlobalState().GetGame("any") != nil {
		t.Error("new state should not contain any game")
	}
	if state.NewGlobalState().HasUsername("LeBron") {
		t.Error("new state should not have any username")
	}
}

func TestSetGameAndGetGame(t *testing.T) {
	s := state.NewGlobalState()
	code := "ABC123"
	if g := s.GetGame(code); g != nil {
		t.Errorf("GetGame(%q) expected nil, got %v", code, g)
	}
	m := game.NewManager("US Capitals", code)
	s.SetGame(code, m)
	if g := s.GetGame(code); g != m {
		t.Errorf("GetGame(%q) expected same manager, got %v", code, g)
	}
	// Overwrite
	m2 := game.NewManager("NBA Teams", code)
	s.SetGame(code, m2)
	if g := s.GetGame(code); g != m2 {
		t.Errorf("GetGame after SetGame expected m2, got %v", g)
	}
}

func TestHasUsername(t *testing.T) {
	s := state.NewGlobalState()
	if s.HasUsername("alice") {
		t.Error("HasUsername(alice) expected false on empty state")
	}
	ok := s.AddUsername("alice")
	if !ok {
		t.Fatal("AddUsername(alice) expected true")
	}
	if !s.HasUsername("alice") {
		t.Error("HasUsername(alice) expected true after AddUsername")
	}
}

func TestAddUsername(t *testing.T) {
	s := state.NewGlobalState()
	ok := s.AddUsername("bob")
	if !ok {
		t.Error("AddUsername(bob) first time expected true")
	}
	ok2 := s.AddUsername("bob")
	if ok2 {
		t.Error("AddUsername(bob) second time expected false (already in use)")
	}
}

func TestRemoveUsername(t *testing.T) {
	s := state.NewGlobalState()
	s.AddUsername("carol")
	if !s.HasUsername("carol") {
		t.Fatal("expected carol to be present")
	}
	s.RemoveUsername("carol")
	if s.HasUsername("carol") {
		t.Error("RemoveUsername(carol) should remove the username")
	}
	// Removing again should be no-op
	s.RemoveUsername("carol")
	if s.HasUsername("carol") {
		t.Error("second RemoveUsername should still leave carol absent")
	}
}

func TestCreate(t *testing.T) {
	// Point to repo trivia from tst/state (server/tst/state -> ../../../trivia)
	saved := state.TriviaBasePath
	state.TriviaBasePath = "../../../trivia"
	defer func() { state.TriviaBasePath = saved }()

	s := state.NewGlobalState()
	code := "CODE01"
	title := "US Capitals"

	m := s.Create(code, title)
	if m == nil {
		t.Fatal("Create with valid code and title expected non-nil Manager")
	}
	if m.Title != title {
		t.Errorf("Create Manager Title = %q, want %q", m.Title, title)
	}
	if m.Code != code {
		t.Errorf("Create Manager Code = %q, want %q", m.Code, code)
	}
	if len(m.Board) == 0 {
		t.Error("Create should populate Board from trivia")
	}

	// Duplicate code should return nil
	m2 := s.Create(code, "NBA Teams")
	if m2 != nil {
		t.Error("Create with existing code expected nil")
	}

	// Invalid title should return nil
	m3 := s.Create("OTHER", "NonExistentTitleXYZ")
	if m3 != nil {
		t.Error("Create with invalid title expected nil")
	}
}

func TestCanJoin(t *testing.T) {
	saved := state.TriviaBasePath
	state.TriviaBasePath = "../../../trivia"
	defer func() { state.TriviaBasePath = saved }()

	s := state.NewGlobalState()
	code := "JOIN01"
	s.Create(code, "US Capitals")

	// Join with valid code and new username
	if !s.CanJoin(code, "player1") {
		t.Error("Join(code, player1) expected true")
	}
	// Different username in same game should be true
	if !s.CanJoin(code, "player2") {
		t.Error("Join(code, player2) expected true")
	}
	// Invalid code
	if s.CanJoin("BADCODE", "player3") {
		t.Error("Join with invalid code expected false")
	}
}
