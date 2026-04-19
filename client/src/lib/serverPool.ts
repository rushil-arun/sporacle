const SERVER_URLS = (
  import.meta.env.VITE_SERVER_URLS ||
  import.meta.env.VITE_SERVER_BASE_URL ||
  'http://localhost:8080'
)
  .split(',')
  .map((s: string) => s.trim())
  .filter(Boolean);

export function pickRandomServer(): string {
  return SERVER_URLS[Math.floor(Math.random() * SERVER_URLS.length)];
}
