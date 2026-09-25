package game

/*
An event that we will send back to a player
*/
type GameEvent struct {
	Type        string
	State       map[string]*Player
	TimeLeft    int
	Winner      *Player
	Players     map[string]*Player
	Leaderboard []LeaderboardEntry
	Chat        *ChatMessage
}

/*
An incoming request from a player.
The "Item" represents the item that the player
wants to enter into the board. "Message" is set instead
of "Item" for a lobby chat message.
*/
type PlayerRequest struct {
	Username string `json:"username"`
	Code     string `json:"code"`
	Item     string `json:"Item"`
	Message  string `json:"Message"`
}

// ChatMessage is a single lobby chat message, broadcast to all players
// as it arrives and held in-memory on the Manager for the life of the game.
type ChatMessage struct {
	Username string `json:"username"`
	Color    string `json:"color"`
	Text     string `json:"text"`
}
