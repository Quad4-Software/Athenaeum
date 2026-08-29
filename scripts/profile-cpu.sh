#!/usr/bin/env bash
# Capture a 30s CPU profile from a running pprof endpoint.
set -euo pipefail

ADDR="${ATHENAEUM_PPROF:-127.0.0.1:6060}"
OUT="${1:-./coverage/cpu.pprof}"
SECONDS_PROFILE="${PROFILE_SECONDS:-30}"

mkdir -p "$(dirname "${OUT}")"
URL="http://${ADDR}/debug/pprof/profile?seconds=${SECONDS_PROFILE}"
echo "fetching ${URL}"
curl -fsS "${URL}" -o "${OUT}"
echo "wrote ${OUT}"
echo "analyze with: go tool pprof -http=:0 ${OUT}"
