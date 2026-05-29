# Sporcle API Reference

This document covers every HTTP endpoint and WebSocket message exchanged between the frontend and backend. Use it as the source of truth when adding new routes or modifying existing ones.

---

## Base URL

All HTTP requests target a server address selected from the pool configured in `VITE_SERVER_BASE_URL` (frontend `.env`). The server listens on the address set by `SERVER_BASE_URL` (backend `.env`).

```
http://<server-addr>
ws://<server-addr>
```

---

## Error Format

All error responses share the same shape:

```json
{ "error": "<human-readable message>" }
```

| HTTP status | Meaning |
|---|---|
| `400` | Bad request — missing or invalid fields |
| `404` | Resource not found |
| `405` | Wrong HTTP method |
| `500` / `502` | Server-side or routing failure |

---

## Public Endpoints

### `POST /create-game`

Creates a new game. In multi-server mode the request is routed by Redis to the least-loaded server; the response always names the server that will host the game.

**Request body** (`Content-Type: application/json`)

| Field | Type | Required | Description |
|---|---|---|---|
| `title` | string | yes | Trivia category / game name |
| `lobbyTime` | number | yes | Lobby countdown in seconds (minimum 10) |
| `gameTime` | number | yes | Game duration in seconds (minimum 10) |

```json
{
  "title": "World Capitals",
  "lobbyTime": 60,
  "gameTime": 180
}
```

**Response `200 OK`**

| Field | Type | Description |
|---|---|---|
| `code` | string | 6-character alphanumeric game code (e.g. `"A3BX9Z"`) |
| `serverAddr` | string | Address of the server hosting this game |

```json
{
  "code": "A3BX9Z",
  "serverAddr": "localhost:8080"
}
```

---

### `GET /get-ws-url`

Resolves the WebSocket URL a client should connect to for a given game. In multi-server mode this performs a Redis lookup so the URL always points to the correct host.

**Query parameters**

| Param | Required | Description |
|---|---|---|
| `code` | yes | Game code returned by `/create-game` |
| `username` | yes | Display name the player wants to use |

**Response `200 OK`**

| Field | Type | Description |
|---|---|---|
| `url` | string | Full WebSocket URL including query params |

```json
{ "url": "ws://localhost:8080/ws?game=A3BX9Z&user=alice" }
```

---

### `GET /ws`

Upgrades the HTTP connection to a WebSocket and adds the player to the specified game. After this point all communication uses the WebSocket protocol described in the [WebSocket Messages](#websocket-messages) section below.

**Query parameters**

| Param | Required | Description |
|---|---|---|
| `game` | yes | Game code |
| `user` | yes | Player username |

**Connection handshake — server sends one of:**

```json
{ "type": "success", "message": "<game title>" }
```

```json
{ "type": "error", "message": "<reason>" }
```

Possible error reasons:
- `"Need to enter a code and a username."` — missing query params
- `"No game with this code."` — game code not found
- `"Username taken in this lobby."` — username already in use
- `"This game has already started"` — game is past the lobby phase

The connection is closed immediately after an error message.

---

### `GET /trivia/files`

Returns the list of trivia data filenames available on the server.

**Response `200 OK`** — JSON array of filenames

```json
["world-capitals.json", "us-presidents.json"]
```

Returns `[]` if no files are found.

---

### `GET /trivia/keys`

Returns the top-level category keys contained in a trivia file. Used to populate the game creation UI.

**Query parameters**

| Param | Required | Description |
|---|---|---|
| `file` | yes | Filename (with or without `.json` extension) |

**Response `200 OK`** — JSON array of category names

```json
["Africa", "Asia", "Europe", "Americas"]
```

Returns `[]` if the file does not exist. Defends against path traversal by accepting only the base filename.

---

## Internal Endpoints

These endpoints are called **server-to-server only** and must not be called from the frontend.

### `POST /internal/create-game`

Creates a game directly on the receiving server without consulting Redis. Called by `POST /create-game` when the routing decision points to a different server. Accepts the same request body as `/create-game` plus an optional `code` field. When `code` is supplied the game is created with that exact code (to match the one already registered in Redis).

**Request body**

| Field | Type | Required | Description |
|---|---|---|---|
| `title` | string | yes | Same as `/create-game` |
| `lobbyTime` | number | yes | Same as `/create-game` |
| `gameTime` | number | yes | Same as `/create-game` |
| `code` | string | no | Pre-assigned code from the routing server |

**Response `200 OK`** — same shape as `/create-game`

```json
{ "code": "A3BX9Z", "serverAddr": "server-2:8080" }
```

---

## WebSocket Messages

All WebSocket frames carry JSON. **Server → client** frames use the `GameEvent` structure; **client → server** frames use the `PlayerRequest` structure.

### Client → Server: `PlayerRequest`

Sent whenever the player submits an answer, and once when the game ends.

| Field | Type | Description |
|---|---|---|
| `username` | string | Player's username (must match the one used at `/get-ws-url`) |
| `code` | string | Game code |
| `Item` | string | The answer being submitted |

```json
{
  "username": "alice",
  "code": "A3BX9Z",
  "Item": "Paris"
}
```

**Special value:** when the client receives a `Leaderboard` event it sends `"Item": "GAME_OVER"` to signal the server that it has acknowledged the end of the game.

```json
{ "username": "alice", "code": "A3BX9Z", "Item": "GAME_OVER" }
```

---

### Server → Client: `GameEvent`

All server-push messages share this top-level structure. Only the fields relevant to each `Type` are populated.

| Field | Type | Always present | Description |
|---|---|---|---|
| `Type` | string | yes | Event type identifier (see below) |
| `TimeLeft` | number | no | Seconds remaining in the current phase |
| `Players` | object | no | Map of `username → PlayerMeta` |
| `State` | object | no | Map of `item → PlayerMeta \| null` (the board) |
| `Leaderboard` | array | no | Ordered leaderboard entries |

#### `PlayerMeta` object

```json
{ "username": "alice", "color": "356 75% 57%" }
```

`color` is an HSL string without the `hsl()` wrapper.

---

#### Event type: `Time`

Broadcast every second. Carries the countdown for the active phase (lobby or game).

```json
{ "Type": "Time", "TimeLeft": 42 }
```

Sent during both the lobby and game phases.

---

#### Event type: `Players`

Broadcast every second during the **lobby phase**. Contains the full current player roster.

```json
{
  "Type": "Players",
  "Players": {
    "alice": { "username": "alice", "color": "356 75% 57%" },
    "bob":   { "username": "bob",   "color": "27 87% 67%"  }
  }
}
```

---

#### Event type: `Start`

Broadcast once when the lobby countdown reaches zero and the game begins. No payload beyond `Type`.

```json
{ "Type": "Start" }
```

The client navigates from `/lobby` to `/game` upon receiving this event.

---

#### Event type: `Board`

Broadcast every second during the **game phase**, and immediately after a player successfully claims a square. Contains the full board state — every item mapped to the player who claimed it, or `null` if unclaimed.

```json
{
  "Type": "Board",
  "State": {
    "Paris":   { "username": "alice", "color": "356 75% 57%" },
    "Berlin":  null,
    "Madrid":  { "username": "bob",   "color": "27 87% 67%"  }
  }
}
```

Answer matching is **case-insensitive** on the server.

---

#### Event type: `Leaderboard`

Broadcast when the game timer reaches zero **or** all squares have been claimed. Contains the final standings (up to the top 3 ranks, all ties included).

```json
{
  "Type": "Leaderboard",
  "Leaderboard": [
    { "username": "alice", "color": "356 75% 57%", "correct": 12, "rank": 1, "isTied": false },
    { "username": "bob",   "color": "27 87% 67%",  "correct":  7, "rank": 2, "isTied": false }
  ]
}
```

| Field | Type | Description |
|---|---|---|
| `username` | string | Player display name |
| `color` | string | HSL color string |
| `correct` | number | Number of squares claimed |
| `rank` | number | Final rank (1-indexed; tied players share a rank) |
| `isTied` | boolean | `true` when another player shares this rank |

After receiving this event the client sends `GAME_OVER` and navigates to `/podium`.

---

## Game Lifecycle Summary

```
Client                              Server
  |                                    |
  |-- POST /create-game -------------->|  Returns { code, serverAddr }
  |                                    |
  |-- GET /get-ws-url?code&username -->|  Returns { url }
  |                                    |
  |-- GET /ws?game&user  (upgrade) --->|  Sends { type: "success", message: title }
  |                                    |  (or { type: "error", message } + close)
  |                                    |
  |          [Lobby phase]             |
  |<-- Time (every 1s) ---------------|
  |<-- Players (every 1s) ------------|
  |                                    |
  |          [Game starts]             |
  |<-- Start --------------------------|  Client navigates to /game
  |                                    |
  |          [Game phase]              |
  |<-- Time (every 1s) ---------------|
  |<-- Board (every 1s + on claim) ---|
  |                                    |
  |-- PlayerRequest (answer) -------->|
  |                                    |
  |          [Game ends]               |
  |<-- Leaderboard --------------------|
  |-- PlayerRequest (GAME_OVER) ----->|  Client navigates to /podium
```
