package gameinit

// CreateRequest is the JSON body for /create-game.
type CreateRequest struct {
	Title string `json:"title"`
}

type CreateResponse struct {
	Code string `json:"code"`
}

// JoinRequest is the JSON body for /join-game.
type JoinRequest struct {
	Username string `json:"username"`
	Code     string `json:"code"`
}

// WSURLResponse is the JSON response with the WebSocket URL.
type WSURLResponse struct {
	URL string `json:"url"`
}

// ErrorResponse is the JSON response for errors.
type ErrorResponse struct {
	Error string `json:"error"`
}
