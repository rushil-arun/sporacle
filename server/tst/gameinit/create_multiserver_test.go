package gameinit_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	gameinit "server/game-init"
	rediscoord "server/redis"
	"server/state"
	test "server/tst"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func newTestRedis(t *testing.T) (*miniredis.Miniredis, *redis.Client) {
	t.Helper()
	mr := miniredis.RunT(t)
	rdb, err := rediscoord.NewClient(mr.Addr())
	if err != nil {
		t.Fatalf("redis client: %v", err)
	}
	t.Cleanup(func() { rdb.Close() })
	return mr, rdb
}

func TestCreateHandler_MultiServer_SelfAssigned(t *testing.T) {
	saved := state.TriviaBasePath
	state.TriviaBasePath = "../../../trivia"
	defer func() { state.TriviaBasePath = saved }()

	_, rdb := newTestRedis(t)
	ctx := context.Background()
	rediscoord.RegisterServer(ctx, rdb, "localhost:8080")

	gs := state.NewGlobalState()
	body, _ := json.Marshal(gameinit.CreateRequest{Title: "US Capitals", LobbyTime: test.LOBBY_TIME, GameTime: test.GAME_TIME})
	req := httptest.NewRequest(http.MethodPost, "/create-game", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	gameinit.CreateHandler(gs, rdb, "localhost:8080", rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var resp gameinit.CreateResponse
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp.Code == "" {
		t.Error("expected non-empty code")
	}
	if resp.ServerAddr != "localhost:8080" {
		t.Errorf("expected serverAddr localhost:8080, got %s", resp.ServerAddr)
	}
	if gs.GetGame(resp.Code) == nil {
		t.Error("expected game to exist in local state")
	}
}

func TestCreateHandler_MultiServer_ForwardToOther(t *testing.T) {
	saved := state.TriviaBasePath
	state.TriviaBasePath = "../../../trivia"
	defer func() { state.TriviaBasePath = saved }()

	_, rdb := newTestRedis(t)
	ctx := context.Background()

	// Stand up a fake "other" server that handles /internal/create-game.
	otherGs := state.NewGlobalState()
	otherMux := http.NewServeMux()
	otherServer := httptest.NewServer(otherMux)
	defer otherServer.Close()
	otherAddr := otherServer.Listener.Addr().String()

	otherMux.HandleFunc("/internal/create-game", func(w http.ResponseWriter, r *http.Request) {
		gameinit.InternalCreateHandler(otherGs, otherAddr, w, r)
	})

	// Register self with high load and other with zero load so other is chosen.
	rediscoord.RegisterServer(ctx, rdb, "localhost:8080")
	rediscoord.RegisterServer(ctx, rdb, otherAddr)
	rdb.ZIncrBy(ctx, rediscoord.ServerLoadZSet, 10, "localhost:8080")

	gs := state.NewGlobalState()
	body, _ := json.Marshal(gameinit.CreateRequest{Title: "US Capitals", LobbyTime: test.LOBBY_TIME, GameTime: test.GAME_TIME})
	req := httptest.NewRequest(http.MethodPost, "/create-game", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	gameinit.CreateHandler(gs, rdb, "localhost:8080", rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var resp gameinit.CreateResponse
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp.Code == "" {
		t.Error("expected non-empty code in forwarded response")
	}
	if resp.ServerAddr != otherAddr {
		t.Errorf("expected serverAddr %s, got %s", otherAddr, resp.ServerAddr)
	}
	// Game should exist on the other server, not on self.
	if gs.GetGame(resp.Code) != nil {
		t.Error("game should NOT be in self's state when forwarded")
	}
	if otherGs.GetGame(resp.Code) == nil {
		t.Error("game should exist in other server's state")
	}
}
