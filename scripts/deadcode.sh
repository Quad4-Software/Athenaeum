#!/usr/bin/env bash
# Report unused Go code with deadcode.
set -euo pipefail

DEADCODE_VERSION="${DEADCODE_VERSION:-v0.52.0}"

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

if ! command -v deadcode >/dev/null 2>&1; then
  echo "installing deadcode ${DEADCODE_VERSION} into ./bin"
  GOBIN="${ROOT}/bin" go install "golang.org/x/tools/cmd/deadcode@${DEADCODE_VERSION}"
  export PATH="${ROOT}/bin:${PATH}"
fi

# Exclude generated and test-only packages from noise.
deadcode -test=false ./cmd/... ./internal/...
