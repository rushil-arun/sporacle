package game

import (
	"sort"
	"strings"
	"sync"
	"time"
)

var PlayerColors = []string{
	"356 75% 57%",  // #E63946
	"27 87% 67%",   // #F4A261
	"174 58% 39%",  // #2A9D8F
	"198 37% 24%",  // #264653
	"43 74% 66%",   // #E9C46A
	"272 72% 57%",  // #9B5DE5
	"171 100% 48%", // #00F5D4
	"322 84% 64%",  // #F15BB5
	"201 100% 49%", // #00BBF9
	"51 99% 62%",   // #Fee440
}

type Manager struct {
	Title           string              // name of the game; key into trivia/*.json
	Code            string              // unique game code, 6 uppercase letters/numbers
	Players         map[string]*Player  // maps player usernames to player objects
	Board           map[string]*Player  // category item -> player who claimed it (nil if unclaimed)
	Colors          map[string]struct{} // set of assigned colors
	Correct         map[*Player]int     // maps players to number of correct items they've inputted
	Time            int                 // seconds remaining (60 until start, then 180)
	InboundRequests chan PlayerRequest
	GameStarted     bool
	SquaresTaken    int
	LobbyTime       int
	GameTime        int
	mu              sync.RWMutex
}

type LeaderboardEntry struct {
	Username string `json:"username"`
	Color    string `json:"color"`
	Count    int    `json:"correct"`
	Rank     int    `json:"rank"`
	IsTied   bool   `json:"isTied"`
}

// NewManager creates a Manager with the given title and code. Time is set to 60,
// board and colors are initialized empty.
func NewManager(title, code string, lobbyTime, gameTime int) *Manager {
	return &Manager{
		Title:           title,
		Code:            code,
		Players:         make(map[string]*Player),
		Board:           make(map[string]*Player),
		Colors:          make(map[string]struct{}),
		Correct:         make(map[*Player]int),
		Time:            lobbyTime,
		GameStarted:     false,
		InboundRequests: make(chan PlayerRequest, 256),
		SquaresTaken:    0,
		LobbyTime:       lobbyTime,
		GameTime:        gameTime,
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
			if len(m.Players) == 0 {
				// don't tick until someone has joined
				continue
			}
			m.Time--
			if m.Time < -10 { // arbitrary threshold to end game
				m.CloseConnections()
				return
			}
			if m.Time == 0 {
				if !m.GameStarted {
					if len(m.Players) == 0 {
						return
					}
					m.Time = m.GameTime
					timer.Stop()
					timer = time.NewTicker(1 * time.Second)
					m.GameStarted = true
					for _, p := range m.Players {
						m.Correct[p] = 0
					}
					m.BroadcastStartGame()

				} else {
					// TODO: Need to send a winner here
					m.BroadcastWinner()
				}

			}
			m.BroadcastTime()
			m.BroadcastState()
			if !m.GameStarted {
				m.BroadcastPlayers()
			}

		case event, ok := <-m.InboundRequests:
			if event.Item == "GAME_OVER" && m.Time <= 0 {
				m.CloseConnections()
				return
			}

			if !m.GameStarted || m.Time < 0 {

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

			var boardKey string
			var currPlayer *Player
			itemExists := false
			for k, v := range m.Board {
				if strings.EqualFold(k, item) {
					boardKey = k
					currPlayer = v
					itemExists = true
					break
				}
			}
			if !itemExists || currPlayer != nil {
				continue
			}
			m.Board[boardKey] = player
			m.Correct[player] += 1
			m.SquaresTaken += 1
			if m.SquaresTaken == len(m.Board) {
				m.BroadcastWinner()
				m.Time = 0
			}

			m.BroadcastState()
		}
	}
}

func (m *Manager) BroadcastState() {
	for _, p := range m.Players {
		select {
		case p.OutboundRequests <- GameEvent{Type: "Board", State: m.Board}:
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

func (m *Manager) BroadcastWinner() {
	lst := make([]LeaderboardEntry, 0, len(m.Correct))
	for k, v := range m.Correct {
		lst = append(lst, LeaderboardEntry{Username: k.Username, Color: k.Color, Count: v})
	}
	sort.Slice(lst, func(i, j int) bool {
		return lst[i].Count > lst[j].Count
	})
	// Include all tied for 1st; extend to 2nd then 3rd until at least 3 podium spots are filled.
	{
		result := make([]LeaderboardEntry, 0, len(lst))
		i := 0
		for rank := 1; rank <= 3 && i < len(lst); rank++ {
			scoreAtRank := lst[i].Count
			j := i
			for j < len(lst) && lst[j].Count == scoreAtRank {
				j++
			}
			result = append(result, lst[i:j]...)
			i = j
			if len(result) >= 3 {
				break
			}
		}
		lst = result
	}

	// Assign ranks: same score → same rank
	for i := range lst {
		if i == 0 {
			lst[i].Rank = 1
		} else if lst[i].Count == lst[i-1].Count {
			lst[i].Rank = lst[i-1].Rank
		} else {
			lst[i].Rank = i + 1
		}
	}

	// Mark entries that share a rank as tied
	rankCounts := make(map[int]int)
	for _, e := range lst {
		rankCounts[e.Rank]++
	}
	for i := range lst {
		if rankCounts[lst[i].Rank] > 1 {
			lst[i].IsTied = true
		}
	}

	for _, p := range m.Players {
		select {
		case p.OutboundRequests <- GameEvent{Type: "Leaderboard", Leaderboard: lst}:
		default:
		}
	}
}

func (m *Manager) CloseConnections() {
	for _, p := range m.Players {
		if p.Connection == nil {
			continue
		}
		p.Connection.Close()
	}
}
