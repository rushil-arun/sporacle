package rediscoord

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

const (
	GameServersHash = "game_servers"
	ServerLoadZSet  = "server_load"
)

func NewClient(addr string) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{Addr: addr})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		rdb.Close()
		return nil, fmt.Errorf("redis ping %s: %w", addr, err)
	}
	return rdb, nil
}
