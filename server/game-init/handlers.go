package gameinit

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	game "server/game"
	rediscoord "server/redis"
	state "server/state"

	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// CreateHandler handles POST /create-game.
// When rdb is nil it runs in single-server mode (original behaviour).
// When rdb is non-nil it uses Redis to route the game to the least-loaded server.
func CreateHandler(globalState *state.GlobalState, rdb *redis.Client, serverAddr string, w http.ResponseWriter, r *http.Request) {
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
	fmt.Println(req)
	if req.LobbyTime < 10 || req.GameTime < 10 {
		writeError(w, http.StatusBadRequest, "Must have at least 10s for lobby/game")
		return
	}

	if rdb == nil {
		// Single-server mode: original behaviour.
		m := globalState.Create(req.Title, req.LobbyTime, req.GameTime)
		if m == nil {
			writeError(w, http.StatusBadRequest, "Invalid title")
			return
		}
		go func() {
			defer globalState.RemoveGame(m.Code)
			m.Run()
		}()
		writeJSON(w, http.StatusOK, CreateResponse{Code: m.Code, ServerAddr: r.Host})
		return
	}

	// Multi-server mode: generate a code, ask Redis which server should host it,
	// then either create locally or forward to the chosen server.
	ctx := r.Context()
	code := globalState.GenerateCode()

	chosenServer, err := rediscoord.AssignGame(ctx, rdb, code)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "routing unavailable")
		return
	}

	if chosenServer == serverAddr {
		m := globalState.CreateWithCode(req.Title, code, req.LobbyTime, req.GameTime)
		if m == nil {
			rediscoord.RemoveGame(context.Background(), rdb, code)
			writeError(w, http.StatusBadRequest, "Invalid title")
			return
		}
		go func() {
			defer func() {
				globalState.RemoveGame(m.Code)
				rediscoord.RemoveGame(context.Background(), rdb, m.Code)
			}()
			m.Run()
		}()
		writeJSON(w, http.StatusOK, CreateResponse{Code: code, ServerAddr: serverAddr})
		return
	}

	// Forward to the chosen server, passing the pre-assigned code.
	req.Code = code
	resp, err := ForwardCreate(ctx, chosenServer, req)
	if err != nil {
		rediscoord.RemoveGame(context.Background(), rdb, code)
		writeError(w, http.StatusBadGateway, "failed to reach target server")
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

// GetWSURLHandler returns a WS URL for a client trying to join a game.
// When rdb is non-nil it looks up which server hosts the game and returns a URL
// pointing to that server. When rdb is nil it falls back to single-server mode.
func GetWSURLHandler(globalState *state.GlobalState, rdb *redis.Client, w http.ResponseWriter, r *http.Request) {
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

	if rdb == nil {
		// Single-server mode: build URL using the current request's host.
		url := buildWSURL(r, req.Code, req.Username)
		writeJSON(w, http.StatusOK, WSURLResponse{URL: url})
		return
	}

	// Multi-server mode: look up which server hosts this game.
	serverAddr, err := rediscoord.LookupGame(r.Context(), rdb, req.Code)
	if err != nil || serverAddr == "" {
		writeError(w, http.StatusNotFound, "game not found")
		return
	}
	url := buildWSURLForAddr(serverAddr, req.Code, req.Username)
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

	if m.GameStarted {
		conn.WriteJSON(map[string]string{
			"type":    "error",
			"message": "This game has already started",
		})
		conn.Close()
		return
	}

	color := m.AssignColorLocked()
	player := game.NewPlayer(username, conn, color, code)
	// this will start routines for the player
	m.AddPlayerLocked(username, player)
	conn.WriteJSON(map[string]string{
		"type":    "success",
		"message": m.Title,
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

// buildWSURL returns the wss:// or ws:// URL with game and user query params,
// using the host from the incoming request.
func buildWSURL(r *http.Request, code, username string) string {
	scheme := "ws"
	if r.TLS != nil {
		scheme = "wss"
	}
	return scheme + "://" + r.Host + "/ws?game=" + code + "&user=" + username
}

// buildWSURLForAddr builds a ws:// URL pointing at a specific server address.
func buildWSURLForAddr(serverAddr, code, username string) string {
	return "ws://" + serverAddr + "/ws?game=" + code + "&user=" + username
}
