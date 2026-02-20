package game

import (
	"github.com/gorilla/websocket"
)

type Player struct {
	Username         string          `json:"username"` // identifies the player
	Connection       *websocket.Conn `json:"-"`        // WebSocket connection to the server (e.g. *websocket.Conn)
	Color            string          `json:"color"`    // hex color, unique within the game
	Code             string          `json:"code"`     // game code this player belongs to
	OutboundRequests chan GameEvent  `json:"-"`
	connClosed       chan struct{}   // closes when Read() terminates, so Write() knows to terminate
}

type PlayerMetaData struct {
	Username string
	Color    string
}

func NewPlayer(username string, connection *websocket.Conn, color string, code string) *Player {
	return &Player{
		Username:         username,
		Connection:       connection,
		Color:            color,
		Code:             code,
		OutboundRequests: make(chan GameEvent, 64),
		connClosed:       make(chan struct{}),
	}
}

func (p *Player) Write() {
	defer p.Connection.Close()
	for {
		select {
		case event, ok := <-p.OutboundRequests:
			if !ok {
				return
			}
			if err := p.Connection.WriteJSON(event); err != nil {
				return
			}
		case <-p.connClosed:
			return
		}
	}
}

func (p *Player) Read(m *Manager) {
	defer p.Connection.Close()
	defer close(p.connClosed)
	for {
		var req PlayerRequest
		if err := p.Connection.ReadJSON(&req); err != nil {
			return
		}

		if req.Username == "" || req.Code == "" || req.Item == "" {
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
