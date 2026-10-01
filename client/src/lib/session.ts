// Persists the info needed to reclaim a player after a refresh or dropped
// connection. sessionStorage keeps it per-tab, so several players can be
// tested from one browser.
export interface GameSession {
  username: string;
  code: string;
  wsUrl: string;
  token: string;
  title: string;
}

const KEY = 'sporacle-session';

export function saveSession(s: GameSession) {
  try {
    sessionStorage.setItem(KEY, JSON.stringify(s));
  } catch {
    // storage unavailable; reconnect-on-refresh just won't work
  }
}

export function loadSession(): GameSession | null {
  try {
    const raw = sessionStorage.getItem(KEY);
    return raw ? (JSON.parse(raw) as GameSession) : null;
  } catch {
    return null;
  }
}

export function clearSession() {
  try {
    sessionStorage.removeItem(KEY);
  } catch {
    // ignore
  }
}
