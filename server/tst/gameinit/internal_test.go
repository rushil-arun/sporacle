package gameinit_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	gameinit "server/game-init"
	"server/state"
	test "server/tst"
)

const testServerAddr = "localhost:8080"

func TestInternalCreateHandler_Success(t *testing.T) {
	saved := state.TriviaBasePath
	state.TriviaBasePath = "../../../trivia"
	defer func() { state.TriviaBasePath = saved }()

	globalState := state.NewGlobalState()
	body, _ := json.Marshal(gameinit.CreateRequest{
		Title:     "US Capitals",
		LobbyTime: test.LOBBY_TIME,
		GameTime:  test.GAME_TIME,
	})
	req := httptest.NewRequest(http.MethodPost, "/internal/create-game", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	gameinit.InternalCreateHandler(globalState, testServerAddr, rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp gameinit.CreateResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Code == "" {
		t.Error("expected non-empty code in response")
	}
	if resp.ServerAddr != testServerAddr {
		t.Errorf("expected serverAddr %s, got %s", testServerAddr, resp.ServerAddr)
	}
}

func TestInternalCreateHandler_MethodNotAllowed(t *testing.T) {
	globalState := state.NewGlobalState()
	req := httptest.NewRequest(http.MethodGet, "/internal/create-game", nil)
	rec := httptest.NewRecorder()
	gameinit.InternalCreateHandler(globalState, testServerAddr, rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rec.Code)
	}
}

func TestInternalCreateHandler_InvalidTitle(t *testing.T) {
	saved := state.TriviaBasePath
	state.TriviaBasePath = "../../../trivia"
	defer func() { state.TriviaBasePath = saved }()

	globalState := state.NewGlobalState()
	body, _ := json.Marshal(gameinit.CreateRequest{Title: "NoSuchTitle", LobbyTime: test.LOBBY_TIME, GameTime: test.GAME_TIME})
	req := httptest.NewRequest(http.MethodPost, "/internal/create-game", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	gameinit.InternalCreateHandler(globalState, testServerAddr, rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}
