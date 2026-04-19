package rediscoord_test

import (
	"context"
	"testing"

	rediscoord "server/redis"

	"github.com/alicebob/miniredis/v2"
)

func TestRegisterServer(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb, err := rediscoord.NewClient(mr.Addr())
	if err != nil {
		t.Fatalf("client: %v", err)
	}
	defer rdb.Close()
	ctx := context.Background()

	if err := rediscoord.RegisterServer(ctx, rdb, "localhost:8080"); err != nil {
		t.Fatalf("register: %v", err)
	}
	if err := rediscoord.RegisterServer(ctx, rdb, "localhost:8081"); err != nil {
		t.Fatalf("register: %v", err)
	}

	members, err := rdb.ZRangeWithScores(ctx, rediscoord.ServerLoadZSet, 0, -1).Result()
	if err != nil {
		t.Fatalf("zrange: %v", err)
	}
	if len(members) != 2 {
		t.Fatalf("expected 2 members, got %d", len(members))
	}
	for _, m := range members {
		if m.Score != 0 {
			t.Errorf("expected score 0 for %s, got %f", m.Member, m.Score)
		}
	}
}

func TestRegisterServer_Idempotent(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb, err := rediscoord.NewClient(mr.Addr())
	if err != nil {
		t.Fatalf("client: %v", err)
	}
	defer rdb.Close()
	ctx := context.Background()

	rediscoord.RegisterServer(ctx, rdb, "localhost:8080")
	// Manually bump the score to simulate active load.
	rdb.ZIncrBy(ctx, rediscoord.ServerLoadZSet, 5, "localhost:8080")
	// Re-registering should not reset the score.
	rediscoord.RegisterServer(ctx, rdb, "localhost:8080")

	score, err := rdb.ZScore(ctx, rediscoord.ServerLoadZSet, "localhost:8080").Result()
	if err != nil {
		t.Fatalf("zscore: %v", err)
	}
	if score != 5 {
		t.Errorf("expected score 5 after idempotent re-register, got %f", score)
	}
}

func TestDeregisterServer(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb, err := rediscoord.NewClient(mr.Addr())
	if err != nil {
		t.Fatalf("client: %v", err)
	}
	defer rdb.Close()
	ctx := context.Background()

	rediscoord.RegisterServer(ctx, rdb, "localhost:8080")
	rediscoord.RegisterServer(ctx, rdb, "localhost:8081")
	// Assign two games to 8080, one to 8081.
	rdb.HSet(ctx, rediscoord.GameServersHash, "AAAA11", "localhost:8080")
	rdb.HSet(ctx, rediscoord.GameServersHash, "BBBB22", "localhost:8080")
	rdb.HSet(ctx, rediscoord.GameServersHash, "CCCC33", "localhost:8081")

	if err := rediscoord.DeregisterServer(ctx, rdb, "localhost:8080"); err != nil {
		t.Fatalf("deregister: %v", err)
	}

	// Server entry removed from sorted set.
	count, _ := rdb.ZCard(ctx, rediscoord.ServerLoadZSet).Result()
	if count != 1 {
		t.Errorf("expected 1 server remaining, got %d", count)
	}

	// Game entries for 8080 are cleaned up; 8081's game remains.
	games, _ := rdb.HGetAll(ctx, rediscoord.GameServersHash).Result()
	if _, ok := games["AAAA11"]; ok {
		t.Error("expected AAAA11 to be removed")
	}
	if _, ok := games["BBBB22"]; ok {
		t.Error("expected BBBB22 to be removed")
	}
	if games["CCCC33"] != "localhost:8081" {
		t.Error("expected CCCC33 to remain for localhost:8081")
	}
}
