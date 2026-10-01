package gameflow

import (
	"net/http"
	"net/http/httptest"
	game "server/game"
	gameinit "server/game-init"
	"server/state"
	test "server/tst"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

const testTriviaPath = "../../../trivia"

// setupGameWithConn creates a game, HTTP server, and a connected WebSocket client.
// Returns the manager, game code, conn, and the player. Caller must defer conn.Close().
// The Connect handler sends {"type":"success"} first; consume it before testing Read/Write.
func setupGameWithConn(t *testing.T) (*game.Manager, string, *websocket.Conn, *game.Player) {
	t.Helper()
	saved := state.TriviaBasePath
	state.TriviaBasePath = testTriviaPath
	defer func() { state.TriviaBasePath = saved }()

	globalState := state.NewGlobalState()
	m := globalState.Create("US Capitals", test.GAME_TIME)
	if m == nil {
		t.Fatal("Create failed")
	}
	code := m.Code

	mux := http.NewServeMux()
	gameinit.RegisterRoutes(mux, globalState, nil, "", "ws")
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws?game=" + code + "&user=LeBron"
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("WebSocket dial: %v", err)
	}

	// Consume initial "success" message from Connect
	var successMsg map[string]string
	if err := conn.ReadJSON(&successMsg); err != nil {
		conn.Close()
		t.Fatalf("read success msg: %v", err)
	}
	if successMsg["type"] != "success" {
		conn.Close()
		t.Fatalf("expected type=success, got %v", successMsg)
	}

	player := m.Players["LeBron"]
	if player == nil {
		conn.Close()
		t.Fatal("player LeBron not in game")
	}
	return m, code, conn, player
}

func TestWrite_SendsEventsToWebSocket(t *testing.T) {
	m, code, conn, _ := setupGameWithConn(t)
	defer conn.Close()

	// Start Run() so Read/Write routines run (Run calls StartRoutines)
	go m.Run()

	sendStart(t, conn, "LeBron", code)
	// Drain messages until the game starts
	for {
		var msg map[string]interface{}
		if err := conn.ReadJSON(&msg); err != nil {
			t.Fatalf("ReadJSON: %v", err)
		}
		if msg["Type"] == "Start" {
			break
		}
		if msg["Type"] == "error" {
			t.Fatalf("unexpected error: %v", msg)
		}
	}
	// Write a message into the connection (as client would); Read() picks it up,
	// Run() processes it and BroadcastState, Write() sends the response back.
	req := map[string]string{
		"username": "LeBron",
		"code":     code,
		"Item":     "Olympia",
	}
	if err := conn.WriteJSON(req); err != nil {
		t.Fatalf("WriteJSON: %v", err)
	}

	// Verify we get the board update response back through the WebSocket
	iters := 100
	for range iters {
		var got map[string]interface{}
		if err := conn.ReadJSON(&got); err != nil {
			t.Fatalf("ReadJSON response: %v", err)
		}
		if got["Type"] == "Board" {
			return
		}
	}
	t.Errorf("Did not recieve a message of type board in %d iters", iters)

}

func TestRead_ValidRequestAppearsOnInboundRequests(t *testing.T) {
	m, code, conn, _ := setupGameWithConn(t)
	defer conn.Close()

	go m.Run()

	sendStart(t, conn, "LeBron", code)
	// Drain messages until the game starts
	for {
		var msg map[string]interface{}
		if err := conn.ReadJSON(&msg); err != nil {
			t.Fatalf("ReadJSON: %v", err)
		}
		if msg["Type"] == "Start" {
			break
		}
		if msg["Type"] == "error" {
			t.Fatalf("unexpected error: %v", msg)
		}
	}

	req := map[string]string{
		"username": "LeBron",
		"code":     code,
		"Item":     "Olympia",
	}
	if err := conn.WriteJSON(req); err != nil {
		t.Fatalf("WriteJSON: %v", err)
	}

	iters := 10
	for range iters {
		var got map[string]interface{}
		if err := conn.ReadJSON(&got); err != nil {
			t.Fatalf("ReadJSON response: %v", err)
		}
		if got["Type"] == "Board" {
			stateVal, ok := got["State"]
			if !ok {
				continue
			}
			stateMap, ok := stateVal.(map[string]interface{})
			if !ok {
				continue
			}
			player, ok := stateMap["Olympia"]
			if !ok || player == nil {
				continue
			}
			playerMap, ok := player.(map[string]interface{})
			if playerMap["username"] != "LeBron" {
				continue
			}
			return
		}
	}
	t.Errorf("Did not recieve a message of type board with LeBron: Olympia mapping in %d iters", iters)
}

func TestRead_InvalidRequestIgnored(t *testing.T) {
	m, code, conn, _ := setupGameWithConn(t)
	defer conn.Close()

	go m.Run()

	sendStart(t, conn, "LeBron", code)
	// Drain messages until the game starts
	for {
		var msg map[string]interface{}
		if err := conn.ReadJSON(&msg); err != nil {
			t.Fatalf("ReadJSON: %v", err)
		}
		if msg["Type"] == "Start" {
			break
		}
		if msg["Type"] == "error" {
			t.Fatalf("unexpected error: %v", msg)
		}
	}

	// Send valid request first so we know Read is processing
	validReq := map[string]string{"username": "LeBron", "code": code, "Item": "Olympia"}
	if err := conn.WriteJSON(validReq); err != nil {
		t.Fatalf("WriteJSON valid: %v", err)
	}

	iters := 10
	found := false
	for range iters {
		var got map[string]interface{}
		if err := conn.ReadJSON(&got); err != nil {
			t.Fatalf("ReadJSON response: %v", err)
		}
		if got["Type"] == "Board" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("Did not recieve a message of type board in %d iters", iters)
	}

	// Send invalid (empty username) - should be ignored
	invalidReq := map[string]string{"username": "", "code": code, "Item": "Denver"}
	if err := conn.WriteJSON(invalidReq); err != nil {
		t.Fatalf("WriteJSON invalid: %v", err)
	}

	// Send another valid one we can detect
	validReq2 := map[string]string{"username": "LeBron", "code": code, "Item": "Oklahoma City"}
	if err := conn.WriteJSON(validReq2); err != nil {
		t.Fatalf("WriteJSON valid2: %v", err)
	}

	iters = 10
	var got map[string]interface{}
	for range iters {
		if err := conn.ReadJSON(&got); err != nil {
			t.Fatalf("ReadJSON response: %v", err)
		}
		if got["Type"] == "Board" {
			// Invalid request should have been skipped; we got Oklahoma City not something from invalid
			stateVal, ok := got["State"]
			if !ok {
				continue
			}
			stateMap, ok := stateVal.(map[string]interface{})
			if !ok {
				continue
			}
			city, ok := stateMap["Oklahoma City"]
			if !ok || city == nil {
				continue
			}
			playerMap, ok := city.(map[string]interface{})
			if playerMap["username"] != "LeBron" {
				continue
			}
			return
		}
	}
	t.Fatalf("Did not recieve a message of type board with LeBron: Oklahoma City mapping in %d iters", iters)

}

func TestChat_BroadcastDuringLobby(t *testing.T) {
	m, code, conn, _ := setupGameWithConn(t)
	defer conn.Close()

	go m.Run()

	req := map[string]string{
		"username": "LeBron",
		"code":     code,
		"Message":  "hello lobby",
	}
	if err := conn.WriteJSON(req); err != nil {
		t.Fatalf("WriteJSON: %v", err)
	}

	iters := 10
	for range iters {
		var got map[string]interface{}
		if err := conn.ReadJSON(&got); err != nil {
			t.Fatalf("ReadJSON response: %v", err)
		}
		if got["Type"] != "Chat" {
			continue
		}
		chat, ok := got["Chat"].(map[string]interface{})
		if !ok {
			t.Fatalf("expected Chat payload, got %v", got["Chat"])
		}
		if chat["username"] != "LeBron" || chat["text"] != "hello lobby" {
			t.Fatalf("unexpected chat payload: %v", chat)
		}
		return
	}
	t.Errorf("Did not receive a Chat message in %d iters", iters)
}

func TestRun_ProcessesInboundRequestAndBroadcastsState(t *testing.T) {
	saved := state.TriviaBasePath
	state.TriviaBasePath = testTriviaPath
	defer func() { state.TriviaBasePath = saved }()

	globalState := state.NewGlobalState()
	m := globalState.Create("US Capitals", 2)
	if m == nil {
		t.Fatal("Create failed")
	}
	code := m.Code

	mux := http.NewServeMux()
	gameinit.RegisterRoutes(mux, globalState, nil, "", "ws")
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws?game=" + code + "&user=Steph"
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("WebSocket dial: %v", err)
	}
	defer conn.Close()

	// Consume "success"
	var successMsg map[string]string
	if err := conn.ReadJSON(&successMsg); err != nil {
		t.Fatalf("read success: %v", err)
	}

	// Start Run (player already connected, so StartRoutines will pick them up)
	go m.Run()

	sendStart(t, conn, "Steph", code)
	// Drain messages until we get "Start" (game started).
	for {
		var msg map[string]interface{}
		if err := conn.ReadJSON(&msg); err != nil {
			t.Fatalf("ReadJSON: %v", err)
		}
		if msg["Type"] == "Start" {
			break
		}
		// Also break on error
		if msg["Type"] == "error" {
			t.Fatalf("unexpected error: %v", msg)
		}
		// Avoid infinite loop if something is wrong
		if msg["Type"] == "Time" {
			if tl, ok := msg["TimeLeft"].(float64); ok && tl < 0 {
				t.Fatal("got TimeLeft < 0 before Start")
			}
		}
	}

	// Send a valid claim
	req := map[string]string{"username": "Steph", "code": code, "Item": "Sacramento"}
	if err := conn.WriteJSON(req); err != nil {
		t.Fatalf("WriteJSON: %v", err)
	}

	// Expect a board update
	iters := 10
	var boardMsg map[string]interface{}
	for range iters {
		if err := conn.ReadJSON(&boardMsg); err != nil {
			t.Fatalf("ReadJSON board: %v", err)
		}
		if boardMsg["Type"] == "Board" {
			stateVal, ok := boardMsg["State"]
			if !ok {
				continue
			}
			stateMap, ok := stateVal.(map[string]interface{})
			if !ok {
				continue
			}
			city, ok := stateMap["Sacramento"]
			if !ok || city == nil {
				continue
			}
			playerMap, ok := city.(map[string]interface{})
			if playerMap["username"] != "Steph" {
				continue
			}
			return
		}
	}
	t.Fatalf("Did not recieve a message of type board with Steph: Sacramento mapping in %d iters", iters)
}

// sendStart asks the manager to start the game as the given user.
func sendStart(t *testing.T, conn *websocket.Conn, username, code string) {
	t.Helper()
	if err := conn.WriteJSON(map[string]interface{}{"username": username, "code": code, "Start": true}); err != nil {
		t.Fatalf("WriteJSON start: %v", err)
	}
}

// dial connects a player to the game and consumes the handshake message.
func dial(t *testing.T, serverURL, code, user string) *websocket.Conn {
	t.Helper()
	wsURL := "ws" + strings.TrimPrefix(serverURL, "http") + "/ws?game=" + code + "&user=" + user
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("WebSocket dial: %v", err)
	}
	t.Cleanup(func() { conn.Close() })
	var success map[string]string
	if err := conn.ReadJSON(&success); err != nil || success["type"] != "success" {
		t.Fatalf("handshake: %v %v", err, success)
	}
	return conn
}

// watchStart reads conn in the background and closes the returned channel when
// a Start event arrives. (A read deadline can't be used instead: gorilla
// websocket connections are unusable after a read times out.)
func watchStart(conn *websocket.Conn) <-chan struct{} {
	started := make(chan struct{})
	go func() {
		for {
			var msg map[string]interface{}
			if err := conn.ReadJSON(&msg); err != nil {
				return
			}
			if msg["Type"] == "Start" {
				close(started)
				return
			}
		}
	}()
	return started
}

// waitForStart reports whether started is closed within d.
func waitForStart(started <-chan struct{}, d time.Duration) bool {
	select {
	case <-started:
		return true
	case <-time.After(d):
		return false
	}
}

func TestStart_OnlyHostStartsGameAndLobbyDoesNotAutoStart(t *testing.T) {
	saved := state.TriviaBasePath
	state.TriviaBasePath = testTriviaPath
	defer func() { state.TriviaBasePath = saved }()

	globalState := state.NewGlobalState()
	m := globalState.Create("US Capitals", test.GAME_TIME)
	if m == nil {
		t.Fatal("Create failed")
	}
	mux := http.NewServeMux()
	gameinit.RegisterRoutes(mux, globalState, nil, "", "ws")
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	host := dial(t, server.URL, m.Code, "Host")
	guest := dial(t, server.URL, m.Code, "Guest")
	hostStarted := watchStart(host)
	guestStarted := watchStart(guest)
	go m.Run()

	// A guest asking to start must be ignored, and nothing starts on its own.
	sendStart(t, guest, "Guest", m.Code)
	if waitForStart(hostStarted, 2500*time.Millisecond) {
		t.Fatal("game started without the host pressing Start")
	}

	sendStart(t, host, "Host", m.Code)
	if !waitForStart(hostStarted, 3*time.Second) {
		t.Fatal("host did not receive Start after pressing Start")
	}
	if !waitForStart(guestStarted, 3*time.Second) {
		t.Fatal("guest did not receive Start")
	}
}
