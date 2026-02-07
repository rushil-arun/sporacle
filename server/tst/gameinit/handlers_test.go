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

func TestCreateHandler_MissingTitle(t *testing.T) {
	globalState := state.NewGlobalState()
	body, _ := json.Marshal(map[string]string{}) // missing title
	req := httptest.NewRequest(http.MethodPost, "/create-game", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	gameinit.CreateHandler(globalState, rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("CreateHandler missing title: status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestCreateHandler_InvalidTitle(t *testing.T) {
	saved := state.TriviaBasePath
	state.TriviaBasePath = "../../../trivia"
	defer func() { state.TriviaBasePath = saved }()

	globalState := state.NewGlobalState()
	body, _ := json.Marshal(gameinit.CreateRequest{Title: "NoSuchTitle"})
	req := httptest.NewRequest(http.MethodPost, "/create-game", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	gameinit.CreateHandler(globalState, rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("CreateHandler invalid title: status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestCreateHandler_Success(t *testing.T) {
	saved := state.TriviaBasePath
	state.TriviaBasePath = "../../../trivia"
	defer func() { state.TriviaBasePath = saved }()

	globalState := state.NewGlobalState()
	body, _ := json.Marshal(gameinit.CreateRequest{Title: "US Capitals"})
	req := httptest.NewRequest(http.MethodPost, "/create-game", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	gameinit.CreateHandler(globalState, rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("CreateHandler success: status = %d, want 200", rec.Code)
	}
	var resp gameinit.CreateResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Code == "" {
		t.Error("CreateHandler success: Code should be non-empty")
	}
	if len(resp.Code) != 6 {
		t.Errorf("CreateHandler success: Code length = %d, want 6", len(resp.Code))
	}
}

// GetWSURLHandler returns a WS URL for the given code/username without validating
// that the game exists or the username is free; that is checked in Connect().

func TestGetWSURLHandler_MethodNotAllowed(t *testing.T) {
	globalState := state.NewGlobalState()
	body, _ := json.Marshal(gameinit.JoinRequest{Username: "bob", Code: "ABC123"})
	req := httptest.NewRequest(http.MethodPost, "/get-ws-url", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	gameinit.GetWSURLHandler(globalState, rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("GetWSURLHandler POST: status = %d, want %d", rec.Code, http.StatusMethodNotAllowed)
	}
}

func TestGetWSURLHandler_InvalidBody(t *testing.T) {
	globalState := state.NewGlobalState()
	req := httptest.NewRequest(http.MethodGet, "/get-ws-url", bytes.NewReader([]byte("not json")))
	rec := httptest.NewRecorder()
	gameinit.GetWSURLHandler(globalState, rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("GetWSURLHandler invalid body: status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestGetWSURLHandler_MissingFields(t *testing.T) {
	globalState := state.NewGlobalState()
	body, _ := json.Marshal(map[string]string{"username": "LeBron"}) // missing code
	req := httptest.NewRequest(http.MethodGet, "/get-ws-url", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	gameinit.GetWSURLHandler(globalState, rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("GetWSURLHandler missing code: status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
	body2, _ := json.Marshal(map[string]string{"code": "ABC123"}) // missing username
	req2 := httptest.NewRequest(http.MethodGet, "/get-ws-url", bytes.NewReader(body2))
	rec2 := httptest.NewRecorder()
	gameinit.GetWSURLHandler(globalState, rec2, req2)
	if rec2.Code != http.StatusBadRequest {
		t.Errorf("GetWSURLHandler missing username: status = %d, want %d", rec2.Code, http.StatusBadRequest)
	}
}

func TestGetWSURLHandler_Success(t *testing.T) {
	globalState := state.NewGlobalState()
	body, _ := json.Marshal(gameinit.JoinRequest{Username: "bob", Code: "JOIN3"})
	req := httptest.NewRequest(http.MethodGet, "/get-ws-url", bytes.NewReader(body))
	req.Host = "test.local"
	rec := httptest.NewRecorder()
	gameinit.GetWSURLHandler(globalState, rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("GetWSURLHandler success: status = %d, want 200", rec.Code)
	}
	var resp gameinit.WSURLResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	expected := "ws://test.local/ws?game=JOIN3&user=bob"
	if resp.URL != expected {
		t.Errorf("GetWSURLHandler success: URL = %q, want %s", resp.URL, expected)
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
	m := globalState.Create("US Capitals")
	if m == nil {
		t.Fatal("Create failed")
	}
	code := m.Code
	// Username-already-in-game is checked in Connect(), not in GetWSURL. Simulate player
	// already in game so Connect returns 409 before Upgrade.
	fakePlayer := game.NewPlayer("LeBron", nil, m.AssignColor(), code)
	m.AddPlayer("LeBron", fakePlayer)

	req := httptest.NewRequest(http.MethodGet, "/ws?game="+code+"&user=LeBron", nil)
	rec := httptest.NewRecorder()
	gameinit.Connect(globalState, rec, req)

	if rec.Code != http.StatusConflict {
		t.Errorf("Connect username already connected: status = %d, want %d", rec.Code, http.StatusConflict)
	}
}

func TestConnect_FirstConnection(t *testing.T) {
	saved := state.TriviaBasePath
	state.TriviaBasePath = "../../../trivia"
	defer func() { state.TriviaBasePath = saved }()

	globalState := state.NewGlobalState()
	m := globalState.Create("US Capitals")
	if m == nil {
		t.Fatal("Create failed")
	}
	code := m.Code

	mux := http.NewServeMux()
	gameinit.RegisterRoutes(mux, globalState)
	server := httptest.NewServer(mux)
	defer server.Close()

	user := "LeBron"
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") +
		"/ws?game=" + code + "&user=" + user
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("WebSocket dial: %v", err)
	}
	defer conn.Close()

	gameM := globalState.GetGame(code)
	if gameM == nil || !gameM.HasPlayer(user) {
		t.Errorf("expected %s to be added to game after Connect", user)
	}
}

func TestConnect_DuplicateUsernameRejected(t *testing.T) {
	saved := state.TriviaBasePath
	state.TriviaBasePath = "../../../trivia"
	defer func() { state.TriviaBasePath = saved }()

	globalState := state.NewGlobalState()
	m := globalState.Create("US Capitals")
	if m == nil {
		t.Fatal("Create failed")
	}
	code := m.Code

	mux := http.NewServeMux()
	gameinit.RegisterRoutes(mux, globalState)
	server := httptest.NewServer(mux)
	defer server.Close()

	user := "LeBron"
	conn, _, err := websocket.DefaultDialer.Dial("ws"+
		strings.TrimPrefix(server.URL, "http")+"/ws?game="+
		code+"&user="+user, nil)
	if err != nil {
		t.Fatalf("WebSocket dial: %v", err)
	}
	defer conn.Close()

	// Second connection with same username for same game must be rejected (409).
	dupReq := httptest.NewRequest(http.MethodGet, "/ws?game="+code+"&user="+user, nil)
	dupRec := httptest.NewRecorder()
	mux.ServeHTTP(dupRec, dupReq)
	if dupRec.Code != http.StatusConflict {
		t.Errorf("second Connect with same user: status = %d, want %d"+
			"(username already connected)", dupRec.Code, http.StatusConflict)
	}
}

func TestConnect_TwoDifferentUsers(t *testing.T) {
	saved := state.TriviaBasePath
	state.TriviaBasePath = "../../../trivia"
	defer func() { state.TriviaBasePath = saved }()

	globalState := state.NewGlobalState()
	m := globalState.Create("US Capitals")
	if m == nil {
		t.Fatal("Create failed")
	}
	code := m.Code

	mux := http.NewServeMux()
	gameinit.RegisterRoutes(mux, globalState)
	server := httptest.NewServer(mux)
	defer server.Close()

	user1 := "LeBron"
	conn1, _, err := websocket.DefaultDialer.Dial("ws"+
		strings.TrimPrefix(server.URL, "http")+"/ws?game="+
		code+"&user="+user1, nil)
	if err != nil {
		t.Fatalf("WebSocket dial first user: %v", err)
	}
	defer conn1.Close()

	user2 := "Steph"
	conn2, _, err := websocket.DefaultDialer.Dial("ws"+
		strings.TrimPrefix(server.URL, "http")+"/ws?game="+
		code+"&user="+user2, nil)
	if err != nil {
		t.Fatalf("WebSocket dial second user: %v", err)
	}
	defer conn2.Close()

	gameM := globalState.GetGame(code)
	if gameM == nil || !gameM.HasPlayer(user1) || !gameM.HasPlayer(user2) {
		t.Errorf("expected both %s and %s to be in the game", user1, user2)
	}
}

func TestRegisterRoutes(t *testing.T) {
	globalState := state.NewGlobalState()
	mux := http.NewServeMux()
	gameinit.RegisterRoutes(mux, globalState)
	// Verify routes respond: create-game with GET returns 405
	req := httptest.NewRequest(http.MethodGet, "/create-game", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("RegisterRoutes /create-game: status = %d, want 405", rec.Code)
	}
	// get-ws-url with empty body returns 400 (invalid request body)
	req2 := httptest.NewRequest(http.MethodGet, "/get-ws-url", nil)
	rec2 := httptest.NewRecorder()
	mux.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusBadRequest {
		t.Errorf("RegisterRoutes /get-ws-url: status = %d, want 400", rec2.Code)
	}
	// /ws with no query params returns 400
	req3 := httptest.NewRequest(http.MethodGet, "/ws", nil)
	rec3 := httptest.NewRecorder()
	mux.ServeHTTP(rec3, req3)
	if rec3.Code != http.StatusBadRequest {
		t.Errorf("RegisterRoutes /ws: status = %d, want 400", rec3.Code)
	}
}
