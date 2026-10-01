package game

import (
	"crypto/rand"
	"encoding/hex"
	"sync"

	"github.com/gorilla/websocket"
)

type Player struct {
	Username         string          `json:"username"` // identifies the player
	Connection       *websocket.Conn `json:"-"`        // WebSocket connection to the server (e.g. *websocket.Conn)
	Color            string          `json:"color"`    // hex color, unique within the game
	Code             string          `json:"code"`     // game code this player belongs to
	OutboundRequests chan GameEvent  `json:"-"`
	Token            string          `json:"-"` // secret handed to the client so it can reclaim this player after a disconnect
	connClosed       chan struct{}   // closes when Read() terminates, so Write() knows to terminate
	connMu           sync.Mutex      // guards Connection and connClosed, which change on Reattach
}

type PlayerMetaData struct {
	Username string
	Color    string
}

// ConnClosed returns a channel that is closed when the player's Read loop exits.
func (p *Player) ConnClosed() <-chan struct{} {
	p.connMu.Lock()
	defer p.connMu.Unlock()
	return p.connClosed
}

// CloseConnection closes the player's current WebSocket, if any.
func (p *Player) CloseConnection() {
	p.connMu.Lock()
	conn := p.Connection
	p.connMu.Unlock()
	if conn != nil {
		conn.Close()
	}
}

// Reattach swaps in a new WebSocket for a player whose client reconnected,
// closes the previous one, and restarts the Read/Write goroutines. The player's
// identity, color, score and outbound buffer are preserved.
func (p *Player) Reattach(m *Manager, conn *websocket.Conn) {
	p.connMu.Lock()
	old := p.Connection
	p.Connection = conn
	p.connClosed = make(chan struct{})
	p.connMu.Unlock()
	if old != nil {
		old.Close()
	}
	go p.Read(m)
	go p.Write()
}

func newToken() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func NewPlayer(username string, connection *websocket.Conn, color string, code string) *Player {
	return &Player{
		Username:         username,
		Connection:       connection,
		Color:            color,
		Code:             code,
		Token:            newToken(),
		OutboundRequests: make(chan GameEvent, 64),
		connClosed:       make(chan struct{}),
	}
}

func (p *Player) Write() {
	p.connMu.Lock()
	conn, closed := p.Connection, p.connClosed
	p.connMu.Unlock()
	defer conn.Close()
	for {
		select {
		case event, ok := <-p.OutboundRequests:
			if !ok {
				return
			}
			if err := conn.WriteJSON(event); err != nil {
				return
			}
		case <-closed:
			return
		}
	}
}

func (p *Player) Read(m *Manager) {
	p.connMu.Lock()
	conn, closed := p.Connection, p.connClosed
	p.connMu.Unlock()
	defer conn.Close()
	defer close(closed)
	for {
		var req PlayerRequest
		if err := conn.ReadJSON(&req); err != nil {
			return
		}

		if req.Username == "" || req.Code == "" || (req.Item == "" && req.Message == "") {
			continue
		}

		_, playerExists := m.Players[req.Username]
		if (req.Code != m.Code) || !playerExists {
			continue
		}

		select {
		case m.InboundRequests <- req:
		default: // don't block the channel
		}
	}
}
