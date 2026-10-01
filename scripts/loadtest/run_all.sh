#!/usr/bin/env bash
# Runs all three load-test scenarios (create-game ramp, join-game ramp,
# play-game ramp) against a target server and builds an HTML report with
# graphs from the results.
#
# Usage:
#   scripts/loadtest/run_all.sh [base_url]
#   scripts/loadtest/run_all.sh https://sporacle.duckdns.org
#   scripts/loadtest/run_all.sh                # defaults to http://localhost:8080
#
# Run a single scenario instead of all three:
#   scripts/loadtest/run_all.sh https://sporacle.duckdns.org create
#   scripts/loadtest/run_all.sh https://sporacle.duckdns.org join
#   scripts/loadtest/run_all.sh https://sporacle.duckdns.org play
#
# Env overrides (apply to all scenarios unless a per-test var is set):
#   MAX_VUS, STAGE_SECS
#   CREATE_MAX_VUS, JOIN_MAX_VUS, PLAY_MAX_VUS override MAX_VUS per scenario
set -euo pipefail

ROOT="$(cd "$(dirname "$0")" && pwd)"
BASE_URL="${1:-${BASE_URL:-http://localhost:8080}}"
ONLY="${2:-all}"

MAX_VUS="${MAX_VUS:-200}"
STAGE_SECS="${STAGE_SECS:-20}"

TS="$(date +%Y%m%d-%H%M%S)"
OUT="$ROOT/reports/$TS"
mkdir -p "$OUT/raw" "$OUT/summary" "$OUT/logs"

run_scenario() {
  local name="$1" script="$2" max_vus="$3" stage_secs="$4" extra_env="${5:-}"
  echo "==> Running $name against $BASE_URL (max_vus=$max_vus, stage=${stage_secs}s)"
  env BASE_URL="$BASE_URL" MAX_VUS="$max_vus" STAGE_SECS="$stage_secs" $extra_env \
    k6 run \
      --out "json=$OUT/raw/$name.ndjson" \
      --summary-export "$OUT/summary/$name.json" \
      --summary-trend-stats="avg,min,med,p(90),p(95),p(99),max" \
      "$ROOT/$script" 2>&1 | tee "$OUT/logs/$name.log"
}

case "$ONLY" in
  create|all)
    run_scenario create_game 01_create_game.js "${CREATE_MAX_VUS:-$MAX_VUS}" "$STAGE_SECS"
    ;;
esac
case "$ONLY" in
  join|all)
    run_scenario join_game 02_join_game.js "${JOIN_MAX_VUS:-$MAX_VUS}" "$STAGE_SECS"
    ;;
esac
case "$ONLY" in
  play|all)
    run_scenario play_game 03_play_game.js "${PLAY_MAX_VUS:-$MAX_VUS}" 30
    ;;
esac

echo "==> Building report"
node "$ROOT/report.mjs" "$OUT"

echo "==> Done: $OUT/report.html"
