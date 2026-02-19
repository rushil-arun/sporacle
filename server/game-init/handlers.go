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
	if req.Title == "" {
		writeError(w, http.StatusBadRequest, "title required")
		return
	}
	m := globalState.Create(req.Title)
	if m == nil {
		writeError(w, http.StatusBadRequest, "Invalid title")
		return
	}

	go m.Run()

	writeJSON(w, http.StatusOK, CreateResponse{Code: m.Code})
}

/*
Returns a WS URL for a client trying to join a game.
*/
func GetWSURLHandler(globalState *state.GlobalState, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var req JoinRequest
	req.Code = r.URL.Query().Get("code")
	req.Username = r.URL.Query().Get("username")
	if req.Username == "" || req.Code == "" {
		writeError(w, http.StatusBadRequest, "username and code required")
		return
	}
	// note that this URL could be invalid (code might not match a game, or username
	// could be taken).
	// Since this needs to be checked in Connect(), it won't be checked here.
	url := buildWSURL(r, req.Code, req.Username)
	writeJSON(w, http.StatusOK, WSURLResponse{URL: url})
}

// Connect handles GET /ws: upgrades to WebSocket and adds the player to the game.
func Connect(globalState *state.GlobalState, w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("game")
	username := r.URL.Query().Get("user")
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	if code == "" || username == "" {
		conn.WriteJSON(map[string]string{
			"type":    "error",
			"message": "Need to enter a code and a username.",
		})
		conn.Close()
		return
	}
	m := globalState.GetGame(code)
	if m == nil {
		conn.WriteJSON(map[string]string{
			"type":    "error",
			"message": "No game with this code.",
		})
		conn.Close()
		return
	}

	m.Lock()
	defer m.Unlock()

	if m.HasPlayerLocked(username) {
		conn.WriteJSON(map[string]string{
			"type":    "error",
			"message": "Username taken in this lobby.",
		})
		conn.Close()
		return
	}

	color := m.AssignColorLocked()
	player := game.NewPlayer(username, conn, color, code)
	m.AddPlayerLocked(username, player)
	conn.WriteJSON(map[string]string{
		"type": "success",
	})
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
