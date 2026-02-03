package gameinit

import (
	"net/http"

	state "server/state"
)

// RegisterRoutes registers /create-game, /join-game, and /ws on mux with the given state.
func RegisterRoutes(mux *http.ServeMux, globalState *state.GlobalState) {
	mux.HandleFunc("/create-game", func(w http.ResponseWriter, r *http.Request) {
		CreateHandler(globalState, w, r)
	})
	mux.HandleFunc("/join-game", func(w http.ResponseWriter, r *http.Request) {
		JoinHandler(globalState, w, r)
	})
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		Connect(globalState, w, r)
	})
}
