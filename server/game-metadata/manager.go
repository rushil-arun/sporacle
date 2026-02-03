package game

import "sync"

// Predefined hex colors for players; first unused one is assigned.
var PlayerColors = []string{
	"#E63946", "#F4A261", "#2A9D8F", "#264653", "#E9C46A",
	"#9B5DE5", "#00F5D4", "#F15BB5", "#00BBF9", "#Fee440",
}

type Manager struct {
	Title   string              // name of the game; key into trivia/*.json
	Code    string              // unique game code, 6 uppercase letters/numbers
	Players map[string]*Player  // maps player usernames to player objects
	Board   map[string]*Player  // category item -> player who claimed it (nil if unclaimed)
	Colors  map[string]struct{} // set of assigned colors
	Time    int                 // seconds remaining (60 until start, then 180)
	mu      sync.RWMutex
}

// NewManager creates a Manager with the given title and code. Time is set to 60,
// board and colors are initialized empty.
func NewManager(title, code string) *Manager {
	return &Manager{
		Title:   title,
		Code:    code,
		Players: make(map[string]*Player),
		Board:   make(map[string]*Player),
		Colors:  make(map[string]struct{}),
		Time:    60,
	}
}

func (m *Manager) SetBoardValue(item string, player *Player) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Board[item] = player
}

func (m *Manager) GetBoardValue(item string) *Player {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.Board[item]
}

// HasPlayer returns whether the username is already a player in this game.
func (m *Manager) HasPlayer(username string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, ok := m.Players[username]
	return ok
}

// AddPlayer adds the player to this game.
func (m *Manager) AddPlayer(username string, p *Player) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if p == nil {
		return
	}
	m.Players[username] = p
	m.Colors[p.Color] = struct{}{}
}

// AssignColor returns a hex color not yet used in this game. Caller should add it when adding the player.
func (m *Manager) AssignColor() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, c := range PlayerColors {
		if _, used := m.Colors[c]; !used {
			return c
		}
	}
	return "#888888"
}
