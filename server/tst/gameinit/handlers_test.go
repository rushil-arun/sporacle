package gameinit_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	gameinit "server/game-init"
	game "server/game-metadata"
	"server/state"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
)

func TestCreateHandler_MethodNotAllowed(t *testing.T) {
	globalState := state.NewGlobalState()
	req := httptest.NewRequest(http.MethodGet, "/create-game", nil)
	rec := httptest.NewRecorder()
	gameinit.CreateHandler(globalState, rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("CreateHandler GET: status = %d, want %d", rec.Code, http.StatusMethodNotAllowed)
	}
}

func TestCreateHandler_InvalidBody(t *testing.T) {
	globalState := state.NewGlobalState()
	req := httptest.NewRequest(http.MethodPost, "/create-game", bytes.NewReader([]byte("not json")))
	rec := httptest.NewRecorder()
	gameinit.CreateHandler(globalState, rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("CreateHandler invalid body: status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestCreateHandler_MissingFields(t *testing.T) {
	globalState := state.NewGlobalState()
	body, _ := json.Marshal(map[string]string{"username": "LeBron", "code": "ABC123"}) // missing title
	req := httptest.NewRequest(http.MethodPost, "/create-game", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	gameinit.CreateHandler(globalState, rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("CreateHandler missing title: status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestCreateHandler_UsernameConflict(t *testing.T) {
	saved := state.TriviaBasePath
	state.TriviaBasePath = "../../../trivia"
	defer func() { state.TriviaBasePath = saved }()

	globalState := state.NewGlobalState()
	globalState.AddUsername("LeBron")
	body, _ := json.Marshal(gameinit.CreateRequest{Username: "LeBron", Code: "CODE1", Title: "US Capitals"})
	req := httptest.NewRequest(http.MethodPost, "/create-game", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	gameinit.CreateHandler(globalState, rec, req)
	if rec.Code != http.StatusConflict {
		t.Errorf("CreateHandler username conflict: status = %d, want %d", rec.Code, http.StatusConflict)
	}
}

func TestCreateHandler_CodeExistsOrInvalidTitle(t *testing.T) {
	saved := state.TriviaBasePath
	state.TriviaBasePath = "../../../trivia"
	defer func() { state.TriviaBasePath = saved }()

	globalState := state.NewGlobalState()
	body, _ := json.Marshal(gameinit.CreateRequest{Username: "LeBron", Code: "CODE1", Title: "US Capitals"})
	req := httptest.NewRequest(http.MethodPost, "/create-game", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	gameinit.CreateHandler(globalState, rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("first CreateHandler: status = %d, want 200", rec.Code)
	}
	// Same code again
	body2, _ := json.Marshal(gameinit.CreateRequest{Username: "bob", Code: "CODE1", Title: "NBA Teams"})
	req2 := httptest.NewRequest(http.MethodPost, "/create-game", bytes.NewReader(body2))
	rec2 := httptest.NewRecorder()
	gameinit.CreateHandler(globalState, rec2, req2)
	if rec2.Code != http.StatusBadRequest {
		t.Errorf("CreateHandler duplicate code: status = %d, want %d", rec2.Code, http.StatusBadRequest)
	}
	// Invalid title
	body3, _ := json.Marshal(gameinit.CreateRequest{Username: "bob", Code: "CODE2", Title: "NoSuchTitle"})
	req3 := httptest.NewRequest(http.MethodPost, "/create-game", bytes.NewReader(body3))
	rec3 := httptest.NewRecorder()
	gameinit.CreateHandler(globalState, rec3, req3)
	if rec3.Code != http.StatusBadRequest {
		t.Errorf("CreateHandler invalid title: status = %d, want %d", rec3.Code, http.StatusBadRequest)
	}
}

func TestCreateHandler_Success(t *testing.T) {
	saved := state.TriviaBasePath
	state.TriviaBasePath = "../../../trivia"
	defer func() { state.TriviaBasePath = saved }()

	globalState := state.NewGlobalState()
	body, _ := json.Marshal(gameinit.CreateRequest{Username: "LeBron", Code: "NEW01", Title: "US Capitals"})
	req := httptest.NewRequest(http.MethodPost, "/create-game", bytes.NewReader(body))
	req.Host = "example.com"
	rec := httptest.NewRecorder()
	gameinit.CreateHandler(globalState, rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("CreateHandler success: status = %d, want 200", rec.Code)
	}
	var resp gameinit.WSURLResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.URL == "" {
		t.Error("CreateHandler success: URL should be non-empty")
	}
	expected := "ws://example.com/ws?game=NEW01&user=LeBron"
	if resp.URL != expected {
		t.Errorf("CreateHandler success: URL = %q, want %q", resp.URL, expected)
	}
}

func TestJoinHandler_MethodNotAllowed(t *testing.T) {
	globalState := state.NewGlobalState()
	req := httptest.NewRequest(http.MethodGet, "/join-game", nil)
	rec := httptest.NewRecorder()
	gameinit.JoinHandler(globalState, rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("JoinHandler GET: status = %d, want %d", rec.Code, http.StatusMethodNotAllowed)
	}
}

func TestJoinHandler_InvalidBody(t *testing.T) {
	globalState := state.NewGlobalState()
	req := httptest.NewRequest(http.MethodPost, "/join-game", bytes.NewReader([]byte("not json")))
	rec := httptest.NewRecorder()
	gameinit.JoinHandler(globalState, rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("JoinHandler invalid body: status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestJoinHandler_MissingFields(t *testing.T) {
	globalState := state.NewGlobalState()
	body, _ := json.Marshal(map[string]string{"username": "LeBron"}) // missing code
	req := httptest.NewRequest(http.MethodPost, "/join-game", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	gameinit.JoinHandler(globalState, rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("JoinHandler missing code: status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestJoinHandler_UsernameConflict(t *testing.T) {
	saved := state.TriviaBasePath
	state.TriviaBasePath = "../../../trivia"
	defer func() { state.TriviaBasePath = saved }()

	globalState := state.NewGlobalState()
	globalState.Create("JOIN1", "US Capitals")
	globalState.AddUsername("LeBron")
	body, _ := json.Marshal(gameinit.JoinRequest{Username: "LeBron", Code: "JOIN1"})
	req := httptest.NewRequest(http.MethodPost, "/join-game", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	gameinit.JoinHandler(globalState, rec, req)
	if rec.Code != http.StatusConflict {
		t.Errorf("JoinHandler username conflict: status = %d, want %d", rec.Code, http.StatusConflict)
	}
}

func TestJoinHandler_InvalidCodeOrAlreadyInGame(t *testing.T) {
	saved := state.TriviaBasePath
	state.TriviaBasePath = "../../../trivia"
	defer func() { state.TriviaBasePath = saved }()

	globalState := state.NewGlobalState()
	globalState.Create("JOIN2", "US Capitals")
	// Invalid code
	body, _ := json.Marshal(gameinit.JoinRequest{Username: "player2", Code: "BADCODE"})
	req := httptest.NewRequest(http.MethodPost, "/join-game", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	gameinit.JoinHandler(globalState, rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("JoinHandler invalid code: status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
	// Username already in game (player added to game but not via JoinHandler, so no global username)
	m := globalState.GetGame("JOIN2")
	if m == nil {
		t.Fatal("game JOIN2 not found")
	}
	body1, _ := json.Marshal(gameinit.JoinRequest{Username: "player2", Code: "JOIN2"})
	req1 := httptest.NewRequest(http.MethodPost, "/join-game", bytes.NewReader(body1))
	rec1 := httptest.NewRecorder()
	gameinit.JoinHandler(globalState, rec1, req1)
	body2, _ := json.Marshal(gameinit.JoinRequest{Username: "player2", Code: "JOIN2"})
	req2 := httptest.NewRequest(http.MethodPost, "/join-game", bytes.NewReader(body2))
	rec2 := httptest.NewRecorder()
	gameinit.JoinHandler(globalState, rec2, req2)
	if rec2.Code != http.StatusConflict {
		t.Errorf("JoinHandler username already in game: status = %d, want %d", rec2.Code, http.StatusBadRequest)
	}
}

func TestJoinHandler_Success(t *testing.T) {
	saved := state.TriviaBasePath
	state.TriviaBasePath = "../../../trivia"
	defer func() { state.TriviaBasePath = saved }()

	globalState := state.NewGlobalState()
	globalState.Create("JOIN3", "US Capitals")
	body, _ := json.Marshal(gameinit.JoinRequest{Username: "bob", Code: "JOIN3"})
	req := httptest.NewRequest(http.MethodPost, "/join-game", bytes.NewReader(body))
	req.Host = "test.local"
	rec := httptest.NewRecorder()
	gameinit.JoinHandler(globalState, rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("JoinHandler success: status = %d, want 200", rec.Code)
	}
	var resp gameinit.WSURLResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.URL != "ws://test.local/ws?game=JOIN3&user=bob" {
		t.Errorf("JoinHandler success: URL = %q, want ws://test.local/ws?game=JOIN3&user=bob", resp.URL)
	}
}

func TestConnect_MissingParams(t *testing.T) {
	globalState := state.NewGlobalState()
	req := httptest.NewRequest(http.MethodGet, "/ws", nil)
	rec := httptest.NewRecorder()
	gameinit.Connect(globalState, rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("Connect no params: status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
	req2 := httptest.NewRequest(http.MethodGet, "/ws?game=ABC", nil)
	rec2 := httptest.NewRecorder()
	gameinit.Connect(globalState, rec2, req2)
	if rec2.Code != http.StatusBadRequest {
		t.Errorf("Connect missing user: status = %d, want %d", rec2.Code, http.StatusBadRequest)
	}
}

func TestConnect_GameNotFound(t *testing.T) {
	globalState := state.NewGlobalState()
	req := httptest.NewRequest(http.MethodGet, "/ws?game=NOSUCH&user=alice", nil)
	rec := httptest.NewRecorder()
	gameinit.Connect(globalState, rec, req)
	if rec.Code != http.StatusNotFound {
		t.Errorf("Connect game not found: status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestConnect_UsernameAlreadyConnected(t *testing.T) {
	saved := state.TriviaBasePath
	state.TriviaBasePath = "../../../trivia"
	defer func() { state.TriviaBasePath = saved }()

	globalState := state.NewGlobalState()
	globalState.Create("WS1", "US Capitals")
	m := globalState.GetGame("WS1")
	if m == nil {
		t.Fatal("game not found")
	}
	// Simulate player already in game (no real WebSocket) so Connect returns 409 before Upgrade.
	fakePlayer := game.NewPlayer("LeBron", nil, m.AssignColor(), "WS1")
	m.AddPlayer("LeBron", fakePlayer)

	req := httptest.NewRequest(http.MethodGet, "/ws?game=WS1&user=LeBron", nil)
	rec := httptest.NewRecorder()
	gameinit.Connect(globalState, rec, req)

	if rec.Code != http.StatusConflict {
		t.Errorf("Connect username already connected: status = %d, want %d", rec.Code, http.StatusConflict)
	}
}

func TestConnect_SuccessWebSocket(t *testing.T) {
	saved := state.TriviaBasePath
	state.TriviaBasePath = "../../../trivia"
	defer func() { state.TriviaBasePath = saved }()

	globalState := state.NewGlobalState()
	globalState.Create("WS2", "US Capitals")
	if globalState.GetGame("WS2") == nil {
		t.Fatal("game not found")
	}

	mux := http.NewServeMux()
	gameinit.RegisterRoutes(mux, globalState)
	server := httptest.NewServer(mux)
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws?game=WS2&user=alice"
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("WebSocket dial: %v", err)
	}
	defer conn.Close()

	// Optionally verify the player was added
	m := globalState.GetGame("WS2")
	if m == nil || !m.HasPlayer("alice") {
		t.Error("expected alice to be added to game after Connect")
	}
}

func TestRegisterRoutes(t *testing.T) {
	globalState := state.NewGlobalState()
	mux := http.NewServeMux()
	gameinit.RegisterRoutes(mux, globalState)
	// Verify routes respond (create-game with GET returns 405)
	req := httptest.NewRequest(http.MethodGet, "/create-game", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("RegisterRoutes /create-game: status = %d, want 405", rec.Code)
	}
	req2 := httptest.NewRequest(http.MethodGet, "/join-game", nil)
	rec2 := httptest.NewRecorder()
	mux.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusMethodNotAllowed {
		t.Errorf("RegisterRoutes /join-game: status = %d, want 405", rec2.Code)
	}
	req3 := httptest.NewRequest(http.MethodGet, "/ws", nil)
	rec3 := httptest.NewRecorder()
	mux.ServeHTTP(rec3, req3)
	if rec3.Code != http.StatusBadRequest {
		t.Errorf("RegisterRoutes /ws: status = %d, want 400", rec3.Code)
	}
}
