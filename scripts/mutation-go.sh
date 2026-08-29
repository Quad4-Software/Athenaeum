#!/usr/bin/env bash
# Run Go mutation testing with Gremlins on security-sensitive packages.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

GREMLINS_VERSION="${GREMLINS_VERSION:-v0.6.0}"
EFFICACY="${MUTATION_EFFICACY:-70}"
MCOVER="${MUTATION_MCOVER:-40}"

if ! command -v gremlins >/dev/null 2>&1; then
  echo "Installing gremlins ${GREMLINS_VERSION}"
  GOBIN="${ROOT}/bin" go install "github.com/go-gremlins/gremlins/cmd/gremlins@${GREMLINS_VERSION}"
  export PATH="${ROOT}/bin:${PATH}"
fi

EXTRA_ARGS=()
if [[ -n "${MUTATION_DIFF:-}" ]]; then
  EXTRA_ARGS+=(--diff "${MUTATION_DIFF}")
fi

exec gremlins unleash \
  --config="${ROOT}/.gremlins.yaml" \
  --threshold-efficacy="${EFFICACY}" \
  --threshold-mcover="${MCOVER}" \
  --coverpkg=./internal/library,./internal/auth \
  "${EXTRA_ARGS[@]}" \
  "$@"
