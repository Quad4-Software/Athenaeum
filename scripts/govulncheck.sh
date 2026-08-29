#!/usr/bin/env bash
# Run govulncheck against the Athenaeum module (vendored deps scanned via go.mod).
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

if ! command -v govulncheck >/dev/null 2>&1; then
  echo "installing govulncheck into ./bin"
  GOBIN="${ROOT}/bin" go install golang.org/x/vuln/cmd/govulncheck@latest
  export PATH="${ROOT}/bin:${PATH}"
fi

govulncheck ./...
