# Deploying the backend to a free-tier EC2 instance

This guide walks through hosting the Go backend on a single free-tier AWS EC2 instance, with the frontend deployed separately on Vercel (HTTPS). It favors the simplest setup that actually works end-to-end, not a scaled-out architecture.

## Why it's set up this way

- **Redis runs locally on the same box** (`127.0.0.1:6379`, never exposed publicly). The backend uses Redis for server-registration/game-routing bookkeeping (`server/redis/routing.go`) even with a single instance — it's required to boot, but with only one server registered it isn't doing any real load-balancing yet. A managed Redis (ElastiCache) only starts paying for itself once you have *multiple* app servers that need to share one Redis; for one instance, self-hosting it alongside the app is simplest and free indefinitely.
- **TLS is mandatory, not optional.** The frontend is served over HTTPS, and browsers block a page from opening an insecure `ws://` connection (mixed content). The backend supports this via the `WS_SCHEME` env var (see `server/game-init/handlers.go`), which controls the scheme used when building WebSocket URLs returned to clients — set it to `wss` once you have TLS in front of the app. [Caddy](https://caddyserver.com/) is used here because it terminates TLS and issues/renews Let's Encrypt certs automatically with a 3-line config, with no separate certbot setup.
- **One port, not one per game.** The backend multiplexes every game and every WebSocket connection through a single port (`SERVER_BASE_URL`, default `:8080`). Only that one port (behind the reverse proxy, on 443) needs to be reachable — never expose 6379 or the raw 8080 app port publicly.

This assumes no domain name is owned yet — using a free [DuckDNS](https://www.duckdns.org/) subdomain pointed at the instance's Elastic IP. All AWS steps are via the Console; adapt to the AWS CLI if you prefer.

## Part 1 — Provision the EC2 instance

1. **EC2 → Launch instance**: Ubuntu 22.04/24.04 LTS, instance type `t2.micro` or `t3.micro` (free tier eligible), default 8GB gp3 root volume, create/select a key pair for SSH.
2. **Security group** — create one with:
  - SSH (22) from your IP only
  - HTTP (80) from `0.0.0.0/0` (needed for the Let's Encrypt HTTP-01 challenge)
  - HTTPS (443) from `0.0.0.0/0`
  - nothing else open — no 6379, no 8080; those stay internal to the box
3. **Elastic IP** — allocate one and associate it with the instance, so the public IP is stable across reboots.
4. **DuckDNS** — create a free subdomain (e.g. `sporacle.duckdns.org`) pointed at the Elastic IP.



## Part 2 — Install Redis and the app on the instance

SSH into the instance, then:

1. Install and start Redis:
  ```bash
   sudo apt update && sudo apt install -y redis-server
   # confirm `bind 127.0.0.1` in /etc/redis/redis.conf (default)
   sudo systemctl enable --now redis-server
  ```
2. Build the backend binary. Either install the Go toolchain on the instance, or cross-compile locally and `scp` the binary over (lighter on a free-tier box's disk/RAM):
  ```bash
   # on your laptop, from the repo root
   cd server
   make gen
   GOOS=linux GOARCH=amd64 go build -o sporacle-server .
   scp -i ~/.ssh/<your-key>.pem sporacle-server ubuntu@<elastic-ip>:/tmp/
   scp -i ~/.ssh/<your-key>.pem -r ../trivia ubuntu@<elastic-ip>:/tmp/
  ```
3. On the instance, create a service user and move things into place. The backend reads trivia JSON from `../trivia` relative to its working directory, so mirror the repo layout: the binary lives in `/opt/sporacle/server/` and `trivia/` sits next to it at `/opt/sporacle/trivia/`.
  ```bash
   sudo useradd --system --no-create-home --shell /usr/sbin/nologin sporacle
   sudo mkdir -p /opt/sporacle/server
   sudo mv /tmp/sporacle-server /opt/sporacle/server/
   sudo mv /tmp/trivia /opt/sporacle/trivia
   sudo chown -R sporacle:sporacle /opt/sporacle
  ```
4. Create `/opt/sporacle/server/.env`:
  ```
   SERVER_BASE_URL=:8080
   SERVER_ADDR=sporacle.duckdns.org:443
   REDIS_ADDR=localhost:6379
   WS_SCHEME=wss
   LOBBY_TIME=60
  ```
5. Create a systemd unit at `/etc/systemd/system/sporacle.service`:
  ```ini
   [Unit]
   Description=Sporacle backend
   After=network.target redis-server.service
   Requires=redis-server.service

   [Service]
   WorkingDirectory=/opt/sporacle/server
   EnvironmentFile=/opt/sporacle/server/.env
   ExecStart=/opt/sporacle/server/sporacle-server
   Restart=on-failure
   User=sporacle

   [Install]
   WantedBy=multi-user.target
  ```
   Then:



## Part 3 — TLS reverse proxy (Caddy)

1. Install Caddy from [their official apt repo](https://caddyserver.com/docs/install#debian-ubuntu-raspbian).
2. Set `/etc/caddy/Caddyfile`:
  ```
   sporacle.duckdns.org {
       reverse_proxy 127.0.0.1:8080
   }
  ```
   Caddy handles the WebSocket upgrade passthrough automatically and issues/renews the Let's Encrypt cert with no extra config.
3. Restart and verify:
  ```bash
   sudo systemctl restart caddy
   curl https://sporacle.duckdns.org/trivia/files   # should return JSON
   journalctl -u caddy   # should show successful cert issuance
  ```



## Part 4 — Point the frontend at it

1. In Vercel project settings, set `VITE_SERVER_URLS=https://sporacle.duckdns.org` (or `VITE_SERVER_BASE_URL` — see the `VITE_SERVER_URLS`/`VITE_SERVER_BASE_URL` precedence note in the root `README.md`; if you set `VITE_SERVER_URLS`, make sure it only lists servers that are actually running, since the client picks randomly among all of them).
2. Redeploy the Vercel frontend so the new env var is baked into the build.



## Verification

1. `redis-cli -h localhost ping` on the instance → `PONG`.
2. `systemctl status sporacle` → active; `journalctl -u sporacle -f` shows `Registered as sporacle.duckdns.org:443` and `Listening on :8080`.
3. `curl https://sporacle.duckdns.org/trivia/files` from your laptop → 200 JSON.
4. From the deployed Vercel app: create a game, check the Network tab that `/get-ws-url` returns a `wss://` URL, join, and confirm the WebSocket connects with no mixed-content console error and board/timer events flow.
5. `cd server && make test` locally, to confirm nothing about the deployment config broke the existing test suite.



## Updating after a change

- **Go code changed:** rebuild and replace the binary, then restart.
  ```bash
  cd server
  GOOS=linux GOARCH=amd64 go build -o sporacle-server .
  scp -i ~/.ssh/<your-key>.pem sporacle-server ubuntu@<elastic-ip>:/tmp/
  ssh -i ~/.ssh/<your-key>.pem ubuntu@<elastic-ip> \
    'sudo mv /tmp/sporacle-server /opt/sporacle/server/sporacle-server && sudo chown sporacle:sporacle /opt/sporacle/server/sporacle-server && sudo systemctl restart sporacle'
  ```
- `.env` **changed on the instance:** `sudo systemctl restart sporacle`.
- **Trivia JSON changed:** copy the file into `/opt/sporacle/trivia/`. It is read at game creation, so no restart is needed.
- **Frontend changed:** redeploy on Vercel; the instance isn't involved.



## Scaling beyond one instance

The Redis-based routing (`server_load`, `game_servers` in `server/redis/routing.go`) and the client's `VITE_SERVER_URLS` list already support running multiple backend instances behind a shared Redis — see the "distributed server management" discussion. That's out of scope for this single-instance guide; the short version is: point every instance's `REDIS_ADDR` at one shared Redis (not one per box), give each instance its own `SERVER_ADDR`/domain/TLS cert, and list all of their public URLs in `VITE_SERVER_URLS`.