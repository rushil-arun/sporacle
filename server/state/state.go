package state

import (
	"encoding/json"
	"math/rand"
	"os"
	"path/filepath"
	"sync"

	game "server/game-metadata"
)

const codeChars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// GlobalState holds games and usernames. Use getters/setters for concurrent access.
type GlobalState struct {
	games map[string]*game.Manager
	mu    sync.RWMutex
}

// NewGlobalState returns an initialized GlobalState.
func NewGlobalState() *GlobalState {
	return &GlobalState{
		games: make(map[string]*game.Manager),
	}
}

// generateCode returns a random code of 6 capitalized letters/numbers
// that is not already a key in games.
// Assumes caller has the lock for the global state.
func (s *GlobalState) generateCode() string {
	for {
		b := make([]byte, 6)
		for i := range b {
			b[i] = codeChars[rand.Intn(len(codeChars))]
		}
		code := string(b)
		if s.games[code] == nil {
			return code
		}
	}
}

// GetGame returns the Manager for the given code, or nil. Caller holds read lock.
func (s *GlobalState) GetGame(code string) *game.Manager {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.games[code]
}

// SetGame stores the Manager for the given code.
func (s *GlobalState) SetGame(code string, m *game.Manager) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.games[code] = m
}

// Create checks code and title, then creates a new Manager with board keys from trivia.
// Returns nil if code already exists or title is not found in trivia.
func (state *GlobalState) Create(title string) *game.Manager {
	state.mu.Lock()
	items := loadTriviaItems(title)
	if items == nil {
		state.mu.Unlock()
		return nil
	}
	code := state.generateCode()
	m := game.NewManager(title, code)
	for _, item := range items {
		m.Board[item] = nil
	}
	state.games[code] = m
	state.mu.Unlock()
	return m
}

/*
	 CanJoin returns a tuple representing:
		1. False, False if there is no game matching the code.
		2. True, (x) if there is a game, where x is True if the game
			doesn't have a user with a matching username as the input.
*/
func (state *GlobalState) CanJoin(code, username string) (bool, bool) {
	m := state.GetGame(code)
	if m == nil {
		return false, false
	}
	return true, !m.HasPlayer(username)
}

// TriviaBasePath is the path to the trivia directory (relative to server when run from server/).
var TriviaBasePath = "../trivia"

// loadTriviaItems finds title in any trivia/*.json and returns the list of items, or nil.
func loadTriviaItems(title string) []string {
	entries, err := os.ReadDir(TriviaBasePath)
	if err != nil {
		return nil
	}
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".json" {
			continue
		}
		path := filepath.Join(TriviaBasePath, e.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		var obj map[string][]string
		if json.Unmarshal(data, &obj) != nil {
			continue
		}
		if items, ok := obj[title]; ok {
			return items
		}
	}
	return nil
}
