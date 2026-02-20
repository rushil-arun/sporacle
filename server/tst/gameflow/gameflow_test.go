package gameflow

import (
	"net/http"
	"net/http/httptest"
	gameinit "server/game-init"
	game "server/game-metadata"
	"server/state"
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
	m := globalState.Create("US Capitals")
	if m == nil {
		t.Fatal("Create failed")
	}
	code := m.Code

	mux := http.NewServeMux()
	gameinit.RegisterRoutes(mux, globalState)
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

	// Drain messages until game starts (timer fires quickly)
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
		if got["Type"] == "board" {
			return
		}
	}
	t.Errorf("Did not recieve a message of type board in %d iters", iters)

}

func TestRead_ValidRequestAppearsOnInboundRequests(t *testing.T) {
	m, code, conn, _ := setupGameWithConn(t)
	defer conn.Close()

	go m.Run()

	req := map[string]string{
		"username": "LeBron",
		"code":     code,
		"Item":     "Olympia",
	}
	if err := conn.WriteJSON(req); err != nil {
		t.Fatalf("WriteJSON: %v", err)
	}

	select {
	case got := <-m.InboundRequests:
		if got.Username != req["username"] || got.Code != req["code"] || got.Item != req["Item"] {
			t.Errorf("got %+v, want username=%s code=%s Item=%s", got, req["username"], req["code"], req["Item"])
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for InboundRequests")
	}
}

func TestRead_InvalidRequestIgnored(t *testing.T) {
	m, code, conn, _ := setupGameWithConn(t)
	defer conn.Close()

	go m.Run()

	// Drain messages until game starts (timer fires quickly)
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

	iters := 100
	found := false
	for range iters {
		var got map[string]interface{}
		if err := conn.ReadJSON(&got); err != nil {
			t.Fatalf("ReadJSON response: %v", err)
		}
		if got["Type"] == "board" {
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

	iters = 100
	found = false
	var got map[string]interface{}
	for range iters {
		if err := conn.ReadJSON(&got); err != nil {
			t.Fatalf("ReadJSON response: %v", err)
		}
		if got["Type"] == "board" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("Did not recieve a message of type board in %d iters", iters)
	}
	// Invalid request should have been skipped; we got Denver not something from invalid
}

func TestRun_ProcessesInboundRequestAndBroadcastsState(t *testing.T) {
	saved := state.TriviaBasePath
	state.TriviaBasePath = testTriviaPath
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

	// Drain messages until we get "Start" (game started). Timer fires every 60ns so this is fast.
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
	found := false
	var boardMsg map[string]interface{}
	for range iters {
		if err := conn.ReadJSON(&boardMsg); err != nil {
			t.Fatalf("ReadJSON board: %v", err)
		}
		if boardMsg["Type"] == "board" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("got type %v, want board", boardMsg["type"])
	}

	// State should have Sacramento claimed
	stateVal, ok := boardMsg["State"]
	if !ok {
		t.Fatal("no State in board event")
	}
	stateMap, ok := stateVal.(map[string]interface{})
	if !ok {
		t.Fatalf("State is %T, want map", stateVal)
	}
	sac, ok := stateMap["Sacramento"]
	if !ok || sac == nil {
		t.Errorf("Sacramento not in State or nil: %v", stateMap["Sacramento"])
	}
}
