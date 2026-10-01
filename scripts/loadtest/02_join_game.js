// Ramps concurrent players joining lobbies: each VU creates (or reuses) a game
// lobby, resolves its WS URL via /get-ws-url, opens the WebSocket, waits for
// the handshake success message, then disconnects. Measures join/handshake
// latency under concurrent load, independent of actual gameplay.
//
// Usage:
//   BASE_URL=https://sporacle.duckdns.org k6 run --out json=reports/raw/join_game.ndjson scripts/loadtest/02_join_game.js
//
// Env overrides:
//   BASE_URL       target origin (default http://localhost:8080)
//   MAX_VUS        peak concurrent VUs (default 200)
//   STAGE_SECS     duration of each ramp stage in seconds (default 20)
//   PLAYERS_PER_GAME  players joining the same lobby before moving to the next (default 20)
import ws from 'k6/ws';
import { check, sleep } from 'k6';
import { Trend, Counter } from 'k6/metrics';
import http from 'k6/http';
import { BASE_URL, randomTitle, randomUsername } from './lib/common.js';

const MAX_VUS = parseInt(__ENV.MAX_VUS || '200', 10);
const STAGE = `${__ENV.STAGE_SECS || '20'}s`;
const PLAYERS_PER_GAME = parseInt(__ENV.PLAYERS_PER_GAME || '20', 10);

const joinHandshakeLatency = new Trend('join_handshake_latency', true);
const joinErrors = new Counter('join_errors');

function stagesUpTo(max, stageDuration) {
  const steps = [10, 25, 50, 100, 200, 400, 800, 1500, 3000].filter((n) => n <= max);
  if (steps[steps.length - 1] !== max) steps.push(max);
  const stages = steps.map((target) => ({ duration: stageDuration, target }));
  stages.push({ duration: stageDuration, target: 0 });
  return stages;
}

export const options = {
  scenarios: {
    ramp: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: stagesUpTo(MAX_VUS, STAGE),
      gracefulRampDown: '5s',
    },
  },
  thresholds: {},
};

// Pre-create enough lobbies to hold MAX_VUS players without overcrowding any one
// board (games never start here since we disconnect right after handshake, but
// keep lobbies believable in size).
export function setup() {
  const numGames = Math.max(1, Math.ceil(MAX_VUS / PLAYERS_PER_GAME));
  const codes = [];
  for (let i = 0; i < numGames; i++) {
    const payload = JSON.stringify({ title: randomTitle(), lobbyTime: 120, gameTime: 30 });
    const res = http.post(`${BASE_URL}/create-game`, payload, {
      headers: { 'Content-Type': 'application/json' },
      timeout: '15s',
    });
    if (res.status === 200) {
      codes.push(JSON.parse(res.body).code);
    }
  }
  return { codes };
}

export default function (data) {
  if (!data.codes || data.codes.length === 0) return;
  const code = data.codes[__VU % data.codes.length];
  const username = randomUsername('joiner');

  const wsUrlRes = http.get(
    `${BASE_URL}/get-ws-url?code=${encodeURIComponent(code)}&username=${encodeURIComponent(username)}`,
    { timeout: '10s' }
  );
  if (wsUrlRes.status !== 200) {
    joinErrors.add(1);
    sleep(1);
    return;
  }
  const wsUrl = JSON.parse(wsUrlRes.body).url;

  const start = Date.now();
  let handshakeOk = false;
  const res = ws.connect(wsUrl, {}, function (socket) {
    socket.on('message', function (msg) {
      const data = JSON.parse(msg);
      if (data.type === 'success') {
        joinHandshakeLatency.add(Date.now() - start);
        handshakeOk = true;
        socket.close();
      } else if (data.type === 'error') {
        socket.close();
      }
    });
    socket.setTimeout(function () {
      socket.close();
    }, 10000);
  });

  const ok = check(res, { 'ws connected': (r) => r && r.status === 101 });
  if (!ok || !handshakeOk) joinErrors.add(1);
}
