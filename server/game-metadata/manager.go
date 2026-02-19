package game

import (
	"sync"
	"time"
)

// Predefined hex colors for players; first unused one is assigned.
var PlayerColors = []string{
	"#E63946", "#F4A261", "#2A9D8F", "#264653", "#E9C46A",
	"#9B5DE5", "#00F5D4", "#F15BB5", "#00BBF9", "#Fee440",
}

const LOBBY_TIME = 60
const GAME_TIME = 180

type Manager struct {
	Title    string              // name of the game; key into trivia/*.json
	Code     string              // unique game code, 6 uppercase letters/numbers
	Players  map[string]*Player  // maps player usernames to player objects
	Board    map[string]*Player  // category item -> player who claimed it (nil if unclaimed)
	Colors   map[string]struct{} // set of assigned colors
	Time     *time.Ticker        // seconds remaining (60 until start, then 180)
	Outbound chan GameEvent
	mu       sync.RWMutex
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
		Time:    time.NewTicker(1 * LOBBY_TIME),
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
	return m.AssignColorLocked()
}

// Lock acquires the Manager's mutex. Use with Unlock to wrap critical sections.
func (m *Manager) Lock() {
	m.mu.Lock()
}

// Unlock releases the Manager's mutex.
func (m *Manager) Unlock() {
	m.mu.Unlock()
}

// HasPlayerLocked returns whether the username is already a player. Caller must hold lock via Lock/Unlock.
func (m *Manager) HasPlayerLocked(username string) bool {
	_, ok := m.Players[username]
	return ok
}

// AssignColorLocked returns a hex color not yet used. Caller must hold lock.
func (m *Manager) AssignColorLocked() string {
	for _, c := range PlayerColors {
		if _, used := m.Colors[c]; !used {
			return c
		}
	}
	return "#888888"
}

// AddPlayerLocked adds the player. Caller must hold lock.
func (m *Manager) AddPlayerLocked(username string, p *Player) {
	if p == nil {
		return
	}
	m.Players[username] = p
	m.Colors[p.Color] = struct{}{}
}
