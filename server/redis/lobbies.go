package rediscoord

import (
	"context"
	"encoding/json"

	"github.com/redis/go-redis/v9"
)

// OpenLobbiesHash holds code -> JSON-encoded LobbyInfo for every game that is
// still in its lobby phase (i.e. hasn't started yet), across all servers.
const OpenLobbiesHash = "open_lobbies"

// LobbyInfo is the metadata stored per open lobby so any server can serve a
// cluster-wide "available lobbies" listing without querying other servers directly.
type LobbyInfo struct {
	Code        string `json:"code"`
	Title       string `json:"title"`
	Creator     string `json:"creator"`
	ServerAddr  string `json:"serverAddr"`
	LobbyEndsAt int64  `json:"lobbyEndsAt"` // unix seconds; when the lobby phase is expected to end
}

// SetLobbyInfo upserts the lobby's metadata in the open_lobbies hash.
func SetLobbyInfo(ctx context.Context, rdb *redis.Client, info LobbyInfo) error {
	data, err := json.Marshal(info)
	if err != nil {
		return err
	}
	return rdb.HSet(ctx, OpenLobbiesHash, info.Code, data).Err()
}

// SetLobbyCreator updates the creator field of an existing open lobby entry.
// It is a no-op (returns nil) if the lobby is no longer listed, e.g. because
// it already started.
func SetLobbyCreator(ctx context.Context, rdb *redis.Client, code, creator string) error {
	raw, err := rdb.HGet(ctx, OpenLobbiesHash, code).Result()
	if err == redis.Nil {
		return nil
	}
	if err != nil {
		return err
	}
	var info LobbyInfo
	if err := json.Unmarshal([]byte(raw), &info); err != nil {
		return err
	}
	info.Creator = creator
	return SetLobbyInfo(ctx, rdb, info)
}

// RemoveLobbyInfo deletes the lobby's entry from open_lobbies, e.g. once it starts.
func RemoveLobbyInfo(ctx context.Context, rdb *redis.Client, code string) error {
	return rdb.HDel(ctx, OpenLobbiesHash, code).Err()
}

// ListLobbies returns the metadata for every open lobby across the cluster.
// Entries that fail to unmarshal are skipped rather than failing the whole call.
func ListLobbies(ctx context.Context, rdb *redis.Client) ([]LobbyInfo, error) {
	raw, err := rdb.HGetAll(ctx, OpenLobbiesHash).Result()
	if err != nil {
		return nil, err
	}
	lobbies := make([]LobbyInfo, 0, len(raw))
	for _, v := range raw {
		var info LobbyInfo
		if err := json.Unmarshal([]byte(v), &info); err != nil {
			continue
		}
		lobbies = append(lobbies, info)
	}
	return lobbies, nil
}
