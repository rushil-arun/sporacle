# Sporacle
A multi-player trivia game where you prove to your friends that you are an oracle.

## About

This is a multi-player version of the popular trivia game Sporcle. In this version, any number of players can compete on the same board to claim squares. After a period of time has passed, the player with the most squares wins. 

Games can be played in numerous categories, such as sports, music, or entertainment. 

## Running Locally

The app has three pieces that need to be running at the same time: Redis, the Go backend, and the React frontend.

### 1. Redis

The backend uses Redis to track which server is hosting which game (even when running a single backend instance). Install and start it locally:

```bash
# macOS
brew install redis
brew services start redis

# Debian/Ubuntu
sudo apt install redis-server
sudo systemctl start redis-server
```

Verify it's up: `redis-cli ping` should return `PONG`. By default it listens on `localhost:6379`.

### 2. Backend

```bash
cd server
cp .env.example .env   # defaults already point at localhost:6379, port :8080
make gen                # regenerate shared constants (only needed after editing shared/constants.json)
go run main.go
```

Key `.env` values:

- `SERVER_BASE_URL` — bind address for the HTTP/WS server (default `:8080`)
- `SERVER_ADDR` — publicly advertised address, used for Redis routing and building WebSocket URLs (default `localhost:8080`)
- `REDIS_ADDR` — Redis connection address (default `localhost:6379`)
- `WS_SCHEME` — scheme used in WebSocket URLs returned to clients; `ws` for local dev, `wss` when deployed behind TLS
- `LOBBY_TIME` — default lobby countdown in seconds

The server logs `Listening on :8080` once it's up.

### 3. Frontend

```bash
cd client
npm install
npm run dev
```

Create `client/.env` (gitignored) pointing at the backend, e.g.:

```
VITE_SERVER_BASE_URL=http://localhost:8080
```

> **Note:** if `VITE_SERVER_URLS` is also set (a comma-separated list, used for load-balancing across multiple backend instances), it takes priority over `VITE_SERVER_BASE_URL` — see `client/src/lib/serverPool.ts`. `pickRandomServer()` picks randomly among *every* URL in that list for each API call, so if it includes a port with nothing listening on it, requests will intermittently fail (e.g. trivia categories failing to load). For a single local backend, leave `VITE_SERVER_URLS` unset or make sure every URL in it is actually running.

Open the Vite dev server URL it prints (typically `http://localhost:5173`) to play.
