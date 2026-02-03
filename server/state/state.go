package state

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"

	game "server/game-metadata"
)

// GlobalState holds games and usernames. Use getters/setters for concurrent access.
type GlobalState struct {
	games     map[string]*game.Manager
	usernames map[string]struct{}
	mu        sync.RWMutex
}

// NewGlobalState returns an initialized GlobalState.
func NewGlobalState() *GlobalState {
	return &GlobalState{
		games:     make(map[string]*game.Manager),
		usernames: make(map[string]struct{}),
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

// HasUsername returns whether the username is already in use.
func (s *GlobalState) HasUsername(username string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, ok := s.usernames[username]
	return ok
}

// AddUsername marks the username as in use. Returns false if already in use.
func (s *GlobalState) AddUsername(username string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.usernames[username]; ok {
		return false
	}
	s.usernames[username] = struct{}{}
	return true
}

// RemoveUsername removes the username from the set (e.g. when player leaves).
func (s *GlobalState) RemoveUsername(username string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.usernames, username)
}

// Create checks code and title, then creates a new Manager with board keys from trivia.
// Returns nil if code already exists or title is not found in trivia.
func (state *GlobalState) Create(code, title string) *game.Manager {
	state.mu.Lock()
	if _, exists := state.games[code]; exists {
		state.mu.Unlock()
		return nil
	}
	items := loadTriviaItems(title)
	if items == nil {
		state.mu.Unlock()
		return nil
	}

	m := game.NewManager(title, code)
	for _, item := range items {
		m.Board[item] = nil
	}
	state.SetGame(code, m)
	state.mu.Unlock()
	return m
}

// Join returns true if the code exists and the username is not already in that game.
func (state *GlobalState) Join(code, username string) bool {
	m := state.GetGame(code)
	if m == nil {
		return false
	}
	return !m.HasPlayer(username)
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
