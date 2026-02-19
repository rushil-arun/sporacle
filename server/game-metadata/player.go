package game

import (
	"github.com/gorilla/websocket"
)

type Player struct {
	Username         string          // identifies the player
	Connection       *websocket.Conn // WebSocket connection to the server (e.g. *websocket.Conn)
	Color            string          // hex color, unique within the game
	Code             string          // game code this player belongs to
	OutboundRequests chan GameEvent
	connClosed       chan struct{} // closes when Read() terminates, so Write() knows to terminate
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
		select {
		case m.InboundRequests <- req:
		default: // don't block the channel
		}
	}
}
