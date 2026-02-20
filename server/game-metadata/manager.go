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

const LOBBY_TIME = 100
const GAME_TIME = 2

type Manager struct {
	Title           string              // name of the game; key into trivia/*.json
	Code            string              // unique game code, 6 uppercase letters/numbers
	Players         map[string]*Player  // maps player usernames to player objects
	Board           map[string]*Player  // category item -> player who claimed it (nil if unclaimed)
	Colors          map[string]struct{} // set of assigned colors
	Time            int                 // seconds remaining (60 until start, then 180)
	InboundRequests chan PlayerRequest
	GameStarted     bool
	mu              sync.RWMutex
}

// NewManager creates a Manager with the given title and code. Time is set to 60,
// board and colors are initialized empty.
func NewManager(title, code string) *Manager {
	return &Manager{
		Title:           title,
		Code:            code,
		Players:         make(map[string]*Player),
		Board:           make(map[string]*Player),
		Colors:          make(map[string]struct{}),
		Time:            LOBBY_TIME,
		GameStarted:     false,
		InboundRequests: make(chan PlayerRequest, 256),
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
	go p.Read(m)
	go p.Write()
}

func (m *Manager) Run() {
	timer := time.NewTicker(1 * time.Second)
	for {
		select {
		case <-timer.C:
			m.Time--
			if m.Time <= 0 {
				if !m.GameStarted {
					m.Time = GAME_TIME
					timer.Stop()
					timer = time.NewTicker(1 * time.Second)
					m.GameStarted = true
					m.BroadcastStartGame()
				} else {
					// TODO: Need to send a winner here
					m.CloseConnections()
					return
				}

			}
			m.BroadcastTime()
			if !m.GameStarted {
				m.BroadcastPlayers()
			}

		case event, ok := <-m.InboundRequests:
			if !m.GameStarted {
				continue
			}
			if !ok || event.Code != m.Code {
				m.CloseConnections()
				return
			}
			player, playerExists := m.Players[event.Username]
			if !playerExists {
				continue
			}
			item := event.Item
			currPlayer, itemExists := m.Board[item]
			if !itemExists || currPlayer != nil {
				continue
			}
			m.Board[item] = player
			m.BroadcastState()
		}
	}
}

func (m *Manager) BroadcastState() {
	for _, p := range m.Players {
		select {
		case p.OutboundRequests <- GameEvent{Type: "board", State: m.Board}:
		default:
		}
	}
}

func (m *Manager) BroadcastTime() {
	for _, p := range m.Players {
		select {
		case p.OutboundRequests <- GameEvent{Type: "Time", TimeLeft: m.Time}:
		default:
		}
	}
}

func (m *Manager) BroadcastStartGame() {
	for _, p := range m.Players {
		select {
		case p.OutboundRequests <- GameEvent{Type: "Start"}:
		default:
		}
	}
}

func (m *Manager) BroadcastPlayers() {
	for _, p := range m.Players {
		select {
		case p.OutboundRequests <- GameEvent{Type: "Players", Players: m.Players}:
		default:
		}
	}
}

func (m *Manager) CloseConnections() {
	for _, p := range m.Players {
		p.Connection.Close()
	}
}
