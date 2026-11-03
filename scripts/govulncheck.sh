#!/usr/bin/env bash
# Run govulncheck against the Athenaeum module (vendored deps scanned via go.mod).
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

GOVULNCHECK_VERSION="${GOVULNCHECK_VERSION:-v1.1.4}"

if ! command -v govulncheck >/dev/null 2>&1; then
  echo "installing govulncheck ${GOVULNCHECK_VERSION} into ./bin"
  GOBIN="${ROOT}/bin" go install "golang.org/x/vuln/cmd/govulncheck@${GOVULNCHECK_VERSION}"
  export PATH="${ROOT}/bin:${PATH}"
fi

govulncheck ./...
