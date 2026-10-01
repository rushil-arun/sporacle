// Ramps concurrent POST /create-game requests to find the instance's breaking
// point for game creation. Each VU loops creating a new game as fast as it can.
//
// Usage:
//   BASE_URL=https://sporacle.duckdns.org k6 run --out json=reports/raw/create_game.ndjson scripts/loadtest/01_create_game.js
//
// Env overrides:
//   BASE_URL     target origin (default http://localhost:8080)
//   MAX_VUS      peak concurrent VUs (default 200)
//   STAGE_SECS   duration of each ramp stage in seconds (default 20)
import http from 'k6/http';
import { check, sleep } from 'k6';
import { Trend, Counter } from 'k6/metrics';
import { BASE_URL, randomTitle } from './lib/common.js';

const MAX_VUS = parseInt(__ENV.MAX_VUS || '200', 10);
const STAGE = `${__ENV.STAGE_SECS || '20'}s`;

const createGameLatency = new Trend('create_game_latency', true);
const createGameErrors = new Counter('create_game_errors');

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

export default function () {
  const payload = JSON.stringify({ title: randomTitle(), lobbyTime: 10, gameTime: 10 });
  const params = { headers: { 'Content-Type': 'application/json' }, timeout: '15s' };

  const res = http.post(`${BASE_URL}/create-game`, payload, params);

  createGameLatency.add(res.timings.duration);
  const ok = check(res, { 'status is 200': (r) => r.status === 200 });
  if (!ok) {
    createGameErrors.add(1);
    // Back off so a downed/crashed server doesn't turn this into a tight
    // hot-fail loop that pegs the CPU and floods logs for the rest of the run.
    sleep(1);
  }
}
