package rediscoord_test

import (
	"testing"

	rediscoord "server/redis"

	"github.com/alicebob/miniredis/v2"
)

func TestNewClient_Success(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb, err := rediscoord.NewClient(mr.Addr())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	defer rdb.Close()
}

func TestNewClient_Unreachable(t *testing.T) {
	_, err := rediscoord.NewClient("localhost:19999")
	if err == nil {
		t.Fatal("expected error for unreachable address, got nil")
	}
}
