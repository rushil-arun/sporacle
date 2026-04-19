package gameinit_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	gameinit "server/game-init"
)

func TestForwardCreate_Success(t *testing.T) {
	want := gameinit.CreateResponse{Code: "ABCD12", ServerAddr: "localhost:8081"}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/internal/create-game" || r.Method != http.MethodPost {
			http.Error(w, "unexpected", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(want)
	}))
	defer ts.Close()

	// ts.Listener.Addr().String() gives "127.0.0.1:<port>"; strip the scheme.
	addr := ts.Listener.Addr().String()
	req := gameinit.CreateRequest{Title: "US Capitals", LobbyTime: 10, GameTime: 10}
	got, err := gameinit.ForwardCreate(context.Background(), addr, req)
	if err != nil {
		t.Fatalf("ForwardCreate: %v", err)
	}
	if got.Code != want.Code || got.ServerAddr != want.ServerAddr {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

func TestForwardCreate_ServerError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "internal error", http.StatusInternalServerError)
	}))
	defer ts.Close()

	_, err := gameinit.ForwardCreate(context.Background(), ts.Listener.Addr().String(),
		gameinit.CreateRequest{Title: "US Capitals", LobbyTime: 10, GameTime: 10})
	if err == nil {
		t.Fatal("expected error when target returns 500, got nil")
	}
}

func TestForwardCreate_Timeout(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(500 * time.Millisecond)
	}))
	defer ts.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err := gameinit.ForwardCreate(ctx, ts.Listener.Addr().String(),
		gameinit.CreateRequest{Title: "US Capitals", LobbyTime: 10, GameTime: 10})
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
}
