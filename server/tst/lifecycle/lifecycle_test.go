package lifecycle_test

import (
	"context"
	"testing"

	rediscoord "server/redis"

	"github.com/alicebob/miniredis/v2"
)

func TestServerLifecycle(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb, err := rediscoord.NewClient(mr.Addr())
	if err != nil {
		t.Fatalf("client: %v", err)
	}
	defer rdb.Close()
	ctx := context.Background()

	const addr = "localhost:8080"

	if err := rediscoord.RegisterServer(ctx, rdb, addr); err != nil {
		t.Fatalf("register: %v", err)
	}

	score, err := rdb.ZScore(ctx, rediscoord.ServerLoadZSet, addr).Result()
	if err != nil {
		t.Fatalf("zscore after register: %v", err)
	}
	if score != 0 {
		t.Errorf("expected score 0 after register, got %f", score)
	}

	if err := rediscoord.DeregisterServer(ctx, rdb, addr); err != nil {
		t.Fatalf("deregister: %v", err)
	}

	count, _ := rdb.ZCard(ctx, rediscoord.ServerLoadZSet).Result()
	if count != 0 {
		t.Errorf("expected sorted set to be empty after deregister, got %d entries", count)
	}
}
