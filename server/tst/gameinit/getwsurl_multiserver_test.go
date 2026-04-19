package gameinit_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	gameinit "server/game-init"
	rediscoord "server/redis"
	"server/state"
)

func TestGetWSURLHandler_MultiServer_GameFound(t *testing.T) {
	_, rdb := newTestRedis(t)
	ctx := context.Background()

	// Pre-populate Redis as if a game was created on localhost:8081.
	rdb.HSet(ctx, rediscoord.GameServersHash, "GAME01", "localhost:8081")

	gs := state.NewGlobalState()
	req := httptest.NewRequest(http.MethodGet, "/get-ws-url?code=GAME01&username=alice", nil)
	rec := httptest.NewRecorder()
	gameinit.GetWSURLHandler(gs, rdb, rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var resp gameinit.WSURLResponse
	json.NewDecoder(rec.Body).Decode(&resp)
	expected := "ws://localhost:8081/ws?game=GAME01&user=alice"
	if resp.URL != expected {
		t.Errorf("expected URL %q, got %q", expected, resp.URL)
	}
}

func TestGetWSURLHandler_MultiServer_GameNotFound(t *testing.T) {
	_, rdb := newTestRedis(t)

	gs := state.NewGlobalState()
	req := httptest.NewRequest(http.MethodGet, "/get-ws-url?code=NOSUCH&username=alice", nil)
	rec := httptest.NewRecorder()
	gameinit.GetWSURLHandler(gs, rdb, rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}
