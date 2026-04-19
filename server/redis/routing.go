package rediscoord

import (
	"context"

	"github.com/redis/go-redis/v9"
)

// RegisterServer adds serverAddr to the server_load sorted set with score 0.
// Uses NX so an already-registered server's score is not reset.
func RegisterServer(ctx context.Context, rdb *redis.Client, serverAddr string) error {
	return rdb.ZAddNX(ctx, ServerLoadZSet, redis.Z{Score: 0, Member: serverAddr}).Err()
}

// DeregisterServer removes serverAddr from the server_load sorted set and deletes
// all game_servers entries that point to it.
func DeregisterServer(ctx context.Context, rdb *redis.Client, serverAddr string) error {
	if err := rdb.ZRem(ctx, ServerLoadZSet, serverAddr).Err(); err != nil {
		return err
	}
	// Scan game_servers hash and delete entries pointing to this server.
	games, err := rdb.HGetAll(ctx, GameServersHash).Result()
	if err != nil {
		return err
	}
	for code, addr := range games {
		if addr == serverAddr {
			if err := rdb.HDel(ctx, GameServersHash, code).Err(); err != nil {
				return err
			}
		}
	}
	return nil
}
