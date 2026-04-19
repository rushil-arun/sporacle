package gameinit

import (
	"encoding/json"
	"net/http"

	state "server/state"
)

// InternalCreateHandler handles POST /internal/create-game.
// It creates the game on this server without consulting Redis for routing,
// so it is safe to call from another server's forwarding logic without looping.
func InternalCreateHandler(globalState *state.GlobalState, serverAddr string, w http.ResponseWriter, r *http.Request) {
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
	if req.LobbyTime < 10 || req.GameTime < 10 {
		writeError(w, http.StatusBadRequest, "Must have at least 10s for lobby/game")
		return
	}
	m := globalState.Create(req.Title, req.LobbyTime, req.GameTime)
	if m == nil {
		writeError(w, http.StatusBadRequest, "Invalid title")
		return
	}

	go func() {
		defer globalState.RemoveGame(m.Code)
		m.Run()
	}()

	writeJSON(w, http.StatusOK, CreateResponse{Code: m.Code, ServerAddr: serverAddr})
}
