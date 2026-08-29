#!/usr/bin/env bash
# Build (if needed), serve the production binary, run Lighthouse CI against it.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
BIN="$ROOT/bin/athenaeum"
DATA_DIR="$(mktemp -d)"
LIBRARY_DIR="$ROOT/library"
PORT="${LIGHTHOUSE_PORT:-18080}"
URL="http://127.0.0.1:${PORT}"
MIN_SCORE="${LIGHTHOUSE_MIN_SCORE:-0.9}"

cleanup() {
  if [[ -n "${SERVER_PID:-}" ]] && kill -0 "$SERVER_PID" 2>/dev/null; then
    kill "$SERVER_PID" 2>/dev/null || true
    wait "$SERVER_PID" 2>/dev/null || true
  fi
  rm -rf "$DATA_DIR"
}
trap cleanup EXIT

if [[ ! -x "$BIN" ]]; then
  echo "Building athenaeum binary for Lighthouse CI..."
  (cd "$ROOT" && task build)
fi

mkdir -p "$LIBRARY_DIR"

# Prefer an explicit CHROME_PATH, then common Chromium/Chrome binaries.
if [[ -z "${CHROME_PATH:-}" ]]; then
  for candidate in chromium chromium-browser google-chrome google-chrome-stable; do
    if command -v "$candidate" >/dev/null 2>&1; then
      export CHROME_PATH
      CHROME_PATH="$(command -v "$candidate")"
      break
    fi
  done
fi

if [[ -z "${CHROME_PATH:-}" ]]; then
  echo "No Chrome/Chromium found. Install Chromium or set CHROME_PATH." >&2
  exit 1
fi

"$BIN" --addr ":${PORT}" --library "$LIBRARY_DIR" --data "$DATA_DIR" >/tmp/athenaeum-lighthouse.log 2>&1 &
SERVER_PID=$!

for _ in $(seq 1 60); do
  if curl -fsS "$URL/api/health" >/dev/null 2>&1; then
    break
  fi
  sleep 0.5
done

if ! curl -fsS "$URL/api/health" >/dev/null 2>&1; then
  echo "Server failed to start on $URL"
  cat /tmp/athenaeum-lighthouse.log
  exit 1
fi

cd "$ROOT/web"
export LIGHTHOUSE_URL="$URL"
export LIGHTHOUSE_MIN_SCORE="$MIN_SCORE"
pnpm exec lhci autorun --config=./lighthouserc.cjs
