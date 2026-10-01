// Ramps concurrent, actively-played games: each iteration creates its own
// short lobby, joins it as the sole player, waits for the lobby timer to
// auto-start the game, then claims a board square and measures how long it
// takes for the resulting board-state broadcast to come back.
//
// This models "many games being played concurrently" from the server's
// perspective (N live Manager.Run() goroutines each ticking their timer and
// broadcasting state), which is the dimension that actually stresses the
// backend. Shared-lobby join contention (many players in the *same* game) is
// covered separately by 02_join_game.js's PLAYERS_PER_GAME setting.
//
// Usage:
//   BASE_URL=https://sporacle.duckdns.org k6 run --out json=reports/raw/play_game.ndjson scripts/loadtest/03_play_game.js
//
// Env overrides:
//   BASE_URL     target origin (default http://localhost:8080)
//   MAX_VUS      peak concurrent VUs/games (default 100)
//   STAGE_SECS   duration of each ramp stage in seconds (default 30 -- long
//                enough to cover a full lobby+game session per iteration)
//   LOBBY_SECS   lobby countdown before auto-start (default 10, the server minimum)
//   GAME_SECS    game duration once started (default 10, the server minimum)
import ws from 'k6/ws';
import http from 'k6/http';
import { check, sleep } from 'k6';
import { Trend, Counter } from 'k6/metrics';
import { BASE_URL, randomTitle, randomUsername } from './lib/common.js';

const MAX_VUS = parseInt(__ENV.MAX_VUS || '100', 10);
const STAGE = `${__ENV.STAGE_SECS || '30'}s`;
const LOBBY_SECS = parseInt(__ENV.LOBBY_SECS || '10', 10);
const GAME_SECS = parseInt(__ENV.GAME_SECS || '10', 10);

const createLatency = new Trend('play_create_latency', true);
const joinLatency = new Trend('play_join_latency', true);
const claimRoundTrip = new Trend('play_claim_round_trip', true);
const sessionErrors = new Counter('play_session_errors');
const sessionsCompleted = new Counter('play_sessions_completed');

function stagesUpTo(max, stageDuration) {
  const steps = [5, 10, 25, 50, 100, 200, 400].filter((n) => n <= max);
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
      gracefulRampDown: '10s',
    },
  },
  thresholds: {},
};

export default function () {
  const createStart = Date.now();
  const createRes = http.post(
    `${BASE_URL}/create-game`,
    JSON.stringify({ title: randomTitle(), lobbyTime: LOBBY_SECS, gameTime: GAME_SECS }),
    { headers: { 'Content-Type': 'application/json' }, timeout: '15s' }
  );
  if (createRes.status !== 200) {
    sessionErrors.add(1);
    // Back off so a downed/crashed server doesn't turn this into a tight
    // hot-fail loop that pegs the CPU and floods logs for the rest of the run.
    sleep(1);
    return;
  }
  createLatency.add(Date.now() - createStart);
  const code = JSON.parse(createRes.body).code;
  const username = randomUsername('player');

  const wsUrlRes = http.get(
    `${BASE_URL}/get-ws-url?code=${encodeURIComponent(code)}&username=${encodeURIComponent(username)}`,
    { timeout: '10s' }
  );
  if (wsUrlRes.status !== 200) {
    sessionErrors.add(1);
    sleep(1);
    return;
  }
  const wsUrl = JSON.parse(wsUrlRes.body).url;

  const joinStart = Date.now();
  let claimSentAt = 0;
  let claimed = false;
  let completed = false;

  const res = ws.connect(wsUrl, {}, function (socket) {
    socket.on('message', function (raw) {
      const msg = JSON.parse(raw);

      if (msg.type === 'success') {
        joinLatency.add(Date.now() - joinStart);
        return;
      }
      if (msg.type === 'error') {
        sessionErrors.add(1);
        socket.close();
        return;
      }
      if (msg.Type === 'Board' || msg.type === 'Board') {
        const board = msg.State || {};
        if (!claimed) {
          const openItems = Object.keys(board).filter((k) => !board[k]);
          if (openItems.length > 0) {
            const item = openItems[Math.floor(Math.random() * openItems.length)];
            claimSentAt = Date.now();
            socket.send(JSON.stringify({ username, code, Item: item }));
            claimed = true;
          }
        } else if (claimSentAt > 0) {
          claimRoundTrip.add(Date.now() - claimSentAt);
          claimSentAt = 0; // only measure the first post-claim broadcast
        }
        return;
      }
      if (msg.Type === 'Leaderboard' || msg.type === 'Leaderboard') {
        completed = true;
        socket.close();
      }
    });

    socket.setTimeout(function () {
      socket.close();
    }, (LOBBY_SECS + GAME_SECS + 15) * 1000);
  });

  const ok = check(res, { 'ws connected': (r) => r && r.status === 101 });
  if (!ok) {
    sessionErrors.add(1);
  } else if (completed) {
    sessionsCompleted.add(1);
  } else {
    sessionErrors.add(1);
  }
}
