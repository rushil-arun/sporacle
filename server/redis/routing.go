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

// assignGameScript atomically picks the least-loaded server,
// and stores code → server in game_servers. Returns the chosen server address.
var assignGameScript = redis.NewScript(`
local servers = redis.call('ZRANGE', KEYS[1], 0, 0)
if #servers == 0 then return redis.error_reply('no servers registered') end
local addr = servers[1]
redis.call('HSET', KEYS[2], ARGV[1], addr)
return addr
`)

// AssignGame picks the least-loaded server, records the code→server mapping in
// game_servers, and increments that server's load score. Returns the chosen address.
func AssignGame(ctx context.Context, rdb *redis.Client, code string) (string, error) {
	res, err := assignGameScript.Run(ctx, rdb,
		[]string{ServerLoadZSet, GameServersHash},
		code,
	).Text()
	if err != nil {
		return "", err
	}
	return res, nil
}

// LookupGame returns the server address hosting the given game code,
// or ("", nil) if the code is not found.
func LookupGame(ctx context.Context, rdb *redis.Client, code string) (string, error) {
	addr, err := rdb.HGet(ctx, GameServersHash, code).Result()
	if err == redis.Nil {
		return "", nil
	}
	return addr, err
}

// RemoveGame deletes the code→server mapping from game_servers.
func RemoveGame(ctx context.Context, rdb *redis.Client, code string) error {
	return rdb.HDel(ctx, GameServersHash, code).Err()
}

// IncrLoad increments the player-count score for serverAddr by 1.
func IncrLoad(ctx context.Context, rdb *redis.Client, serverAddr string) error {
	return rdb.ZIncrBy(ctx, ServerLoadZSet, 1, serverAddr).Err()
}

// DecrLoad decrements the player-count score for serverAddr by 1, floored at 0.
func DecrLoad(ctx context.Context, rdb *redis.Client, serverAddr string) error {
	score, err := rdb.ZScore(ctx, ServerLoadZSet, serverAddr).Result()
	if err == redis.Nil {
		return nil
	}
	if err != nil {
		return err
	}
	if score <= 0 {
		return nil
	}
	return rdb.ZIncrBy(ctx, ServerLoadZSet, -1, serverAddr).Err()
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
