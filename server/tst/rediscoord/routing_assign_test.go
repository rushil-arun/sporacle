package rediscoord_test

import (
	"context"
	"sync"
	"testing"

	rediscoord "server/redis"

	"github.com/alicebob/miniredis/v2"
)

func TestAssignGame_PicksLeastLoaded(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb, err := rediscoord.NewClient(mr.Addr())
	if err != nil {
		t.Fatalf("client: %v", err)
	}
	defer rdb.Close()
	ctx := context.Background()

	rediscoord.RegisterServer(ctx, rdb, "localhost:8080")
	rediscoord.RegisterServer(ctx, rdb, "localhost:8081")
	// Manually set 8080 to have more load.
	rdb.ZIncrBy(ctx, rediscoord.ServerLoadZSet, 5, "localhost:8080")

	chosen, err := rediscoord.AssignGame(ctx, rdb, "GAME01")
	if err != nil {
		t.Fatalf("assign: %v", err)
	}
	if chosen != "localhost:8081" {
		t.Errorf("expected least-loaded server localhost:8081, got %s", chosen)
	}

	// Verify the mapping was stored.
	stored, _ := rdb.HGet(ctx, rediscoord.GameServersHash, "GAME01").Result()
	if stored != "localhost:8081" {
		t.Errorf("expected game_servers[GAME01]=localhost:8081, got %s", stored)
	}

	// Verify the score was incremented.
	score, _ := rdb.ZScore(ctx, rediscoord.ServerLoadZSet, "localhost:8081").Result()
	if score != 1 {
		t.Errorf("expected score 1 for localhost:8081, got %f", score)
	}
}

func TestAssignGame_NoServers(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb, err := rediscoord.NewClient(mr.Addr())
	if err != nil {
		t.Fatalf("client: %v", err)
	}
	defer rdb.Close()

	_, err = rediscoord.AssignGame(context.Background(), rdb, "GAME01")
	if err == nil {
		t.Fatal("expected error when no servers registered, got nil")
	}
}

func TestAssignGame_Concurrent(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb, err := rediscoord.NewClient(mr.Addr())
	if err != nil {
		t.Fatalf("client: %v", err)
	}
	defer rdb.Close()
	ctx := context.Background()

	rediscoord.RegisterServer(ctx, rdb, "localhost:8080")
	rediscoord.RegisterServer(ctx, rdb, "localhost:8081")

	const n = 100
	codes := make([]string, n)
	for i := range codes {
		codes[i] = "GAME" + string(rune('A'+i%26)) + string(rune('0'+i/26))
	}

	var wg sync.WaitGroup
	for _, code := range codes {
		wg.Add(1)
		go func(c string) {
			defer wg.Done()
			rediscoord.AssignGame(ctx, rdb, c)
		}(code)
	}
	wg.Wait()

	// Every code should have a mapping.
	games, _ := rdb.HGetAll(ctx, rediscoord.GameServersHash).Result()
	if len(games) != n {
		t.Errorf("expected %d game mappings, got %d", n, len(games))
	}
	// Total load across servers should equal n.
	members, _ := rdb.ZRangeWithScores(ctx, rediscoord.ServerLoadZSet, 0, -1).Result()
	var total float64
	for _, m := range members {
		total += m.Score
	}
	if total != n {
		t.Errorf("expected total load %d, got %f", n, total)
	}
}
