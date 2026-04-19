package gameinit

import (
	"net/http"

	"github.com/redis/go-redis/v9"

	state "server/state"
)

// RegisterRoutes registers all public and internal routes.
// Pass nil for rdb and "" for serverAddr to run in single-server mode
// (existing behaviour, no Redis lookups performed).
func RegisterRoutes(mux *http.ServeMux, globalState *state.GlobalState, rdb *redis.Client, serverAddr string) {
	mux.HandleFunc("/create-game", func(w http.ResponseWriter, r *http.Request) {
		CreateHandler(globalState, rdb, serverAddr, w, r)
	})
	mux.HandleFunc("/get-ws-url", func(w http.ResponseWriter, r *http.Request) {
		GetWSURLHandler(globalState, w, r)
	})
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		Connect(globalState, w, r)
	})
	mux.HandleFunc("/internal/create-game", func(w http.ResponseWriter, r *http.Request) {
		InternalCreateHandler(globalState, serverAddr, w, r)
	})
}
