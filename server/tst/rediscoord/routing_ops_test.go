package rediscoord_test

import (
	"context"
	"testing"

	rediscoord "server/redis"

	"github.com/alicebob/miniredis/v2"
)

func TestLookupGame_Found(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb, err := rediscoord.NewClient(mr.Addr())
	if err != nil {
		t.Fatalf("client: %v", err)
	}
	defer rdb.Close()
	ctx := context.Background()

	rdb.HSet(ctx, rediscoord.GameServersHash, "GAME01", "localhost:8080")

	addr, err := rediscoord.LookupGame(ctx, rdb, "GAME01")
	if err != nil {
		t.Fatalf("lookup: %v", err)
	}
	if addr != "localhost:8080" {
		t.Errorf("expected localhost:8080, got %s", addr)
	}
}

func TestLookupGame_NotFound(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb, err := rediscoord.NewClient(mr.Addr())
	if err != nil {
		t.Fatalf("client: %v", err)
	}
	defer rdb.Close()

	addr, err := rediscoord.LookupGame(context.Background(), rdb, "MISSING")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if addr != "" {
		t.Errorf("expected empty string for missing game, got %s", addr)
	}
}

func TestRemoveGame(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb, err := rediscoord.NewClient(mr.Addr())
	if err != nil {
		t.Fatalf("client: %v", err)
	}
	defer rdb.Close()
	ctx := context.Background()

	rdb.HSet(ctx, rediscoord.GameServersHash, "GAME01", "localhost:8080")
	if err := rediscoord.RemoveGame(ctx, rdb, "GAME01"); err != nil {
		t.Fatalf("remove: %v", err)
	}

	addr, _ := rediscoord.LookupGame(ctx, rdb, "GAME01")
	if addr != "" {
		t.Errorf("expected empty after removal, got %s", addr)
	}
}

func TestIncrLoad_DecrLoad(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb, err := rediscoord.NewClient(mr.Addr())
	if err != nil {
		t.Fatalf("client: %v", err)
	}
	defer rdb.Close()
	ctx := context.Background()

	rediscoord.RegisterServer(ctx, rdb, "localhost:8080")

	if err := rediscoord.IncrLoad(ctx, rdb, "localhost:8080"); err != nil {
		t.Fatalf("incr: %v", err)
	}
	score, _ := rdb.ZScore(ctx, rediscoord.ServerLoadZSet, "localhost:8080").Result()
	if score != 1 {
		t.Errorf("expected score 1 after incr, got %f", score)
	}

	if err := rediscoord.DecrLoad(ctx, rdb, "localhost:8080"); err != nil {
		t.Fatalf("decr: %v", err)
	}
	score, _ = rdb.ZScore(ctx, rediscoord.ServerLoadZSet, "localhost:8080").Result()
	if score != 0 {
		t.Errorf("expected score 0 after decr, got %f", score)
	}
}

func TestDecrLoad_FloorsAtZero(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb, err := rediscoord.NewClient(mr.Addr())
	if err != nil {
		t.Fatalf("client: %v", err)
	}
	defer rdb.Close()
	ctx := context.Background()

	rediscoord.RegisterServer(ctx, rdb, "localhost:8080")

	// Decrement when already at 0 should be a no-op.
	if err := rediscoord.DecrLoad(ctx, rdb, "localhost:8080"); err != nil {
		t.Fatalf("decr at zero: %v", err)
	}
	score, _ := rdb.ZScore(ctx, rediscoord.ServerLoadZSet, "localhost:8080").Result()
	if score != 0 {
		t.Errorf("expected score to stay 0, got %f", score)
	}
}
