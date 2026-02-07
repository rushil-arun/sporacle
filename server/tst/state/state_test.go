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
	if state.NewGlobalState().GetGame("any") != nil {
		t.Error("new state should not contain any game")
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

func TestCreate(t *testing.T) {
	saved := state.TriviaBasePath
	state.TriviaBasePath = "../../../trivia"
	defer func() { state.TriviaBasePath = saved }()

	s := state.NewGlobalState()
	title := "US Capitals"

	m := s.Create(title)
	if m == nil {
		t.Fatal("Create with valid title expected non-nil Manager")
	}
	if m.Title != title {
		t.Errorf("Create Manager Title = %q, want %q", m.Title, title)
	}
	if m.Code == "" || len(m.Code) != 6 {
		t.Errorf("Create Manager Code should be 6 chars, got %q", m.Code)
	}
	if len(m.Board) == 0 {
		t.Error("Create should populate Board from trivia")
	}

	// Second Create returns a different game with a different code
	m2 := s.Create("NBA Teams")
	if m2 == nil {
		t.Fatal("Create second game expected non-nil Manager")
	}
	if m2.Code == m.Code {
		t.Error("Create should generate distinct codes for different games")
	}

	// Invalid title should return nil
	m3 := s.Create("NonExistentTitleXYZ")
	if m3 != nil {
		t.Error("Create with invalid title expected nil")
	}
}

func TestCanJoin(t *testing.T) {
	saved := state.TriviaBasePath
	state.TriviaBasePath = "../../../trivia"
	defer func() { state.TriviaBasePath = saved }()

	s := state.NewGlobalState()
	m := s.Create("US Capitals")
	if m == nil {
		t.Fatal("Create failed")
	}
	code := m.Code

	gameExists1, usernameFree1 := s.CanJoin(code, "player1")
	gameExists2, usernameFree2 := s.CanJoin(code, "player2")
	gameExists3, usernameFree3 := s.CanJoin("BADCODE", "player3")
	// Valid code and new username
	if !gameExists1 || !usernameFree1 {
		t.Error("CanJoin(code, player1) expected true, true")
	}
	// Same game, different username
	if !gameExists2 || !usernameFree2 {
		t.Error("CanJoin(code, player2) expected true, true")
	}
	// Invalid code
	if gameExists3 || usernameFree3 {
		t.Error("CanJoin with invalid code expected false, false")
	}
}
