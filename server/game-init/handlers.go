package gameinit

import (
	"encoding/json"
	"net/http"

	game "server/game-metadata"
	state "server/state"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// Create handles POST /create-game: creates a game and returns wss URL.
func CreateHandler(globalState *state.GlobalState, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var req CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Username == "" || req.Code == "" || req.Title == "" {
		writeError(w, http.StatusBadRequest, "username, code, and title required")
		return
	}
	if globalState.HasUsername(req.Username) {
		writeError(w, http.StatusConflict, "username already in use")
		return
	}
	m := globalState.Create(req.Code, req.Title)
	if m == nil {
		writeError(w, http.StatusBadRequest, "code already in use or invalid title")
		return
	}
	globalState.AddUsername(req.Username)

	url := buildWSURL(r, req.Code, req.Username)
	writeJSON(w, http.StatusOK, WSURLResponse{URL: url})
}

// Join handles POST /join-game: validates join and returns wss URL.
func JoinHandler(globalState *state.GlobalState, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var req JoinRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Username == "" || req.Code == "" {
		writeError(w, http.StatusBadRequest, "username and code required")
		return
	}
	if globalState.HasUsername(req.Username) {
		writeError(w, http.StatusConflict, "username already in use")
		return
	}
	if !globalState.Join(req.Code, req.Username) {
		writeError(w, http.StatusBadRequest, "invalid code or username already in game")
		return
	}
	globalState.AddUsername(req.Username)

	url := buildWSURL(r, req.Code, req.Username)
	writeJSON(w, http.StatusOK, WSURLResponse{URL: url})
}

// Connect handles GET /ws: upgrades to WebSocket and adds the player to the game.
func Connect(globalState *state.GlobalState, w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("game")
	username := r.URL.Query().Get("user")
	if code == "" || username == "" {
		writeError(w, http.StatusBadRequest, "game and user query params required")
		return
	}
	m := globalState.GetGame(code)
	if m == nil {
		writeError(w, http.StatusNotFound, "game not found")
		return
	}
	if m.HasPlayer(username) {
		writeError(w, http.StatusConflict, "username already connected to this game")
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	color := m.AssignColor()
	player := game.NewPlayer(username, conn, color, code)
	m.AddPlayer(username, player)
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, ErrorResponse{Error: msg})
}

// buildWSURL returns the wss:// or ws:// URL with game and user query params.
func buildWSURL(r *http.Request, code, username string) string {
	scheme := "ws"
	if r.TLS != nil {
		scheme = "wss"
	}
	return scheme + "://" + r.Host + "/ws?game=" + code + "&user=" + username
}
