package game

/*
An event that we will send back to a player
*/
type GameEvent struct {
	Type     string
	State    map[string]*Player
	TimeLeft int
	Winner   *Player
	Players  map[string]*Player
}

/*
An incoming request from a player.
The "Item" represents the item that the player
wants to enter into the board.
*/
type PlayerRequest struct {
	Username string `json:"username"`
	Code     string `json:"code"`
	Item     string `json:"Item"`
}
