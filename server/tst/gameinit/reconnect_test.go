package gameinit_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	gameinit "server/game-init"
	state "server/state"
	test "server/tst"

	"github.com/gorilla/websocket"
)

func dialWS(t *testing.T, base, code, user, token string) (*websocket.Conn, map[string]interface{}) {
	t.Helper()
	url := "ws" + strings.TrimPrefix(base, "http") + "/ws?game=" + code + "&user=" + user
	if token != "" {
		url += "&token=" + token
	}
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	var msg map[string]interface{}
	if err := conn.ReadJSON(&msg); err != nil {
		t.Fatalf("read handshake: %v", err)
	}
	return conn, msg
}

func TestConnect_ReconnectWithToken(t *testing.T) {
	saved := state.TriviaBasePath
	state.TriviaBasePath = "../../../trivia"
	defer func() { state.TriviaBasePath = saved }()

	globalState := state.NewGlobalState()
	m := globalState.Create("US Capitals", test.LOBBY_TIME, test.GAME_TIME)
	if m == nil {
		t.Fatal("Create failed")
	}
	mux := http.NewServeMux()
	gameinit.RegisterRoutes(mux, globalState, nil, "", "ws")
	server := httptest.NewServer(mux)
	defer server.Close()

	first, msg := dialWS(t, server.URL, m.Code, "alice", "")
	defer first.Close()
	token, _ := msg["token"].(string)
	if msg["type"] != "success" || token == "" {
		t.Fatalf("first connect: got %+v, want success with token", msg)
	}
	colorBefore := m.Players["alice"].Color

	// Same user, no token: still rejected.
	c, msg := dialWS(t, server.URL, m.Code, "alice", "")
	c.Close()
	if msg["type"] != "error" {
		t.Errorf("no token: got %+v, want error", msg)
	}

	// Same user, wrong token: rejected.
	c, msg = dialWS(t, server.URL, m.Code, "alice", "bogus")
	c.Close()
	if msg["type"] != "error" {
		t.Errorf("wrong token: got %+v, want error", msg)
	}

	// Same user, right token: reclaims the player (old socket is replaced).
	second, msg := dialWS(t, server.URL, m.Code, "alice", token)
	defer second.Close()
	if msg["type"] != "success" || msg["started"] == true {
		t.Fatalf("reconnect: got %+v, want success and not started", msg)
	}
	if got := m.Players["alice"].Color; got != colorBefore {
		t.Errorf("color changed on reconnect: %q -> %q", colorBefore, got)
	}
	if len(m.Players) != 1 {
		t.Errorf("players = %d, want 1", len(m.Players))
	}
}
