// Shared helpers for the loadtest k6 scripts. Import with relative paths, e.g.
//   import { BASE_URL, WS_BASE_URL, randomTitle } from './lib/common.js';

export const TITLES = [
  'African Countries', 'Asian Countries', 'European Countries',
  'MLB Teams', 'NBA MVPs', 'NBA Teams', 'NFL MVPs', 'NFL Teams',
  'US Capitals', 'US States',
];

export function randomTitle() {
  return TITLES[Math.floor(Math.random() * TITLES.length)];
}

// BASE_URL is the HTTP(S) origin, e.g. https://sporacle.duckdns.org or http://localhost:8080.
export const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';

// Derive the ws(s):// equivalent of BASE_URL for direct socket connects when needed.
export const WS_BASE_URL = BASE_URL.replace(/^http/, 'ws');

export function randomUsername(prefix) {
  return `${prefix}_${__VU}_${__ITER}_${Math.floor(Math.random() * 1e6)}`;
}
