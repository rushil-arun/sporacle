package gameinit_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	gameinit "server/game-init"
	rediscoord "server/redis"
	"server/state"
	test "server/tst"
)

func TestLobbiesHandler_SingleServer_ListsUnstartedGame(t *testing.T) {
	saved := state.TriviaBasePath
	state.TriviaBasePath = "../../../trivia"
	defer func() { state.TriviaBasePath = saved }()

	gs := state.NewGlobalState()
	body, _ := json.Marshal(gameinit.CreateRequest{Title: "US Capitals", GameTime: test.GAME_TIME})
	createReq := httptest.NewRequest(http.MethodPost, "/create-game", bytes.NewReader(body))
	createRec := httptest.NewRecorder()
	gameinit.CreateHandler(gs, nil, "", createRec, createReq)

	var createResp gameinit.CreateResponse
	if err := json.NewDecoder(createRec.Body).Decode(&createResp); err != nil {
		t.Fatalf("decode create response: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/lobbies", nil)
	rec := httptest.NewRecorder()
	gameinit.LobbiesHandler(gs, nil, rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var resp gameinit.LobbiesResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode lobbies response: %v", err)
	}
	if len(resp.Lobbies) != 1 {
		t.Fatalf("expected 1 lobby, got %d", len(resp.Lobbies))
	}
	got := resp.Lobbies[0]
	if got.Code != createResp.Code {
		t.Errorf("code = %q, want %q", got.Code, createResp.Code)
	}
	if got.Title != "US Capitals" {
		t.Errorf("title = %q, want %q", got.Title, "US Capitals")
	}
	if got.Creator != "" {
		t.Errorf("creator = %q, want empty before anyone joins", got.Creator)
	}
}

func TestLobbiesHandler_MultiServer_ListsAcrossCluster(t *testing.T) {
	saved := state.TriviaBasePath
	state.TriviaBasePath = "../../../trivia"
	defer func() { state.TriviaBasePath = saved }()

	_, rdb := newTestRedis(t)
	rediscoord.RegisterServer(t.Context(), rdb, "localhost:8080")

	gs := state.NewGlobalState()
	body, _ := json.Marshal(gameinit.CreateRequest{Title: "US Capitals", GameTime: test.GAME_TIME})
	createReq := httptest.NewRequest(http.MethodPost, "/create-game", bytes.NewReader(body))
	createRec := httptest.NewRecorder()
	gameinit.CreateHandler(gs, rdb, "localhost:8080", createRec, createReq)

	var createResp gameinit.CreateResponse
	if err := json.NewDecoder(createRec.Body).Decode(&createResp); err != nil {
		t.Fatalf("decode create response: %v", err)
	}

	// A different server instance in the cluster, with only Redis in common,
	// should still be able to see the lobby.
	otherGs := state.NewGlobalState()
	req := httptest.NewRequest(http.MethodGet, "/lobbies", nil)
	rec := httptest.NewRecorder()
	gameinit.LobbiesHandler(otherGs, rdb, rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var resp gameinit.LobbiesResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode lobbies response: %v", err)
	}
	if len(resp.Lobbies) != 1 || resp.Lobbies[0].Code != createResp.Code {
		t.Fatalf("expected lobby %q to be listed, got %+v", createResp.Code, resp.Lobbies)
	}

	// Once the game starts, it stops being an open lobby.
	m := gs.GetGame(createResp.Code)
	if m == nil {
		t.Fatalf("expected game %q to exist locally", createResp.Code)
	}
	m.Lock()
	m.GameStarted = true
	m.Unlock()
	if m.OnGameStart != nil {
		m.OnGameStart()
	}

	rec2 := httptest.NewRecorder()
	gameinit.LobbiesHandler(otherGs, rdb, rec2, req)
	var resp2 gameinit.LobbiesResponse
	if err := json.NewDecoder(rec2.Body).Decode(&resp2); err != nil {
		t.Fatalf("decode lobbies response: %v", err)
	}
	if len(resp2.Lobbies) != 0 {
		t.Fatalf("expected 0 lobbies after start, got %+v", resp2.Lobbies)
	}
}
