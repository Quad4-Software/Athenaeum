#!/usr/bin/env bash
# Reset local ./data (and optionally reseed demo library).
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

DATA_DIR="${ATHENAEUM_DATA:-./data}"
LIBRARY_DIR="${ATHENAEUM_LIBRARY:-./library}"
DEMO="${1:-}"

if [[ -e "${DATA_DIR}" ]]; then
  rm -rf "${DATA_DIR}"
  echo "removed ${DATA_DIR}"
fi
mkdir -p "${DATA_DIR}" "${LIBRARY_DIR}"

if [[ "${DEMO}" == "--demo" || "${DEMO}" == "demo" ]]; then
  echo "starting demo seed (Ctrl+C after ready, or use task demo)"
  exec go run ./cmd/athenaeum --addr "${ATHENAEUM_ADDR:-:8080}" --library "${LIBRARY_DIR}" --data "${DATA_DIR}" --demo
fi

echo "data directory reset. Run: task demo   or   task run"
