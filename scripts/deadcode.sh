#!/usr/bin/env bash
# Report unused Go code with deadcode.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

if ! command -v deadcode >/dev/null 2>&1; then
  echo "installing deadcode into ./bin"
  GOBIN="${ROOT}/bin" go install golang.org/x/tools/cmd/deadcode@latest
  export PATH="${ROOT}/bin:${PATH}"
fi

# Exclude generated and test-only packages from noise.
deadcode -test=false ./cmd/... ./internal/...
