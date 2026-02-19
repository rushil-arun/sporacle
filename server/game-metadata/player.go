package game

import (
	"github.com/gorilla/websocket"
)

type Player struct {
	Username        string          // identifies the player
	Connection      *websocket.Conn // WebSocket connection to the server (e.g. *websocket.Conn)
	Color           string          // hex color, unique within the game
	Code            string          // game code this player belongs to
	InboundRequests chan PlayerRequest
}

func NewPlayer(username string, connection *websocket.Conn, color string, code string) *Player {
	return &Player{
		Username:   username,
		Connection: connection,
		Color:      color,
		Code:       code,
	}
}
