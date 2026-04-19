package gameinit_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	gameinit "server/game-init"
	rediscoord "server/redis"
	"server/state"
	test "server/tst"

	"github.com/gorilla/websocket"
)

func TestConnect_LoadTracking(t *testing.T) {
	saved := state.TriviaBasePath
	state.TriviaBasePath = "../../../trivia"
	defer func() { state.TriviaBasePath = saved }()

	_, rdb := newTestRedis(t)
	ctx := context.Background()
	const selfAddr = "localhost:8080"
	rediscoord.RegisterServer(ctx, rdb, selfAddr)

	gs := state.NewGlobalState()
	m := gs.Create("US Capitals", test.LOBBY_TIME, test.GAME_TIME)
	if m == nil {
		t.Fatal("Create failed")
	}
	go func() {
		defer gs.RemoveGame(m.Code)
		m.Run()
	}()

	mux := http.NewServeMux()
	gameinit.RegisterRoutes(mux, gs, rdb, selfAddr)
	server := httptest.NewServer(mux)
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws?game=" + m.Code + "&user=alice"
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}

	// Read the success message.
	var msg map[string]string
	if err := conn.ReadJSON(&msg); err != nil {
		t.Fatalf("read: %v", err)
	}
	if msg["type"] != "success" {
		t.Fatalf("expected success, got %+v", msg)
	}

	// Score should be 1 after connect.
	score, err := rdb.ZScore(ctx, rediscoord.ServerLoadZSet, selfAddr).Result()
	if err != nil {
		t.Fatalf("zscore after connect: %v", err)
	}
	if score != 1 {
		t.Errorf("expected score 1 after connect, got %f", score)
	}

	// Close the connection and wait briefly for the goroutine to decrement.
	conn.Close()
	time.Sleep(100 * time.Millisecond)

	score, err = rdb.ZScore(ctx, rediscoord.ServerLoadZSet, selfAddr).Result()
	if err != nil {
		t.Fatalf("zscore after disconnect: %v", err)
	}
	if score != 0 {
		t.Errorf("expected score 0 after disconnect, got %f", score)
	}
}
