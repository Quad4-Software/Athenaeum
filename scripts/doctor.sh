#!/usr/bin/env bash
# Check that required developer tools and versions are available.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

ok=0
warn=0

pass() { printf 'OK   %s\n' "$*"; ok=$((ok + 1)); }
fail() { printf 'FAIL %s\n' "$*"; exit 1; }
note() { printf 'WARN %s\n' "$*"; warn=$((warn + 1)); }

need_cmd() {
  local bin="$1"
  if ! command -v "$bin" >/dev/null 2>&1; then
    fail "missing command: $bin"
  fi
  local ver
  case "$bin" in
    go) ver="$(go version 2>&1)" ;;
    *) ver="$("$bin" --version 2>&1 | head -n1)" ;;
  esac
  pass "$bin (${ver})"
}

need_cmd go
need_cmd node
need_cmd pnpm
need_cmd task

go_ver="$(go env GOVERSION 2>/dev/null || true)"
case "${go_ver}" in
  go1.26*|go1.27*|go1.28*|devel*) pass "Go toolchain ${go_ver}" ;;
  *) note "Go ${go_ver:-unknown} (README asks for Go 1.26+)" ;;
esac

node_major="$(node -p "process.versions.node.split('.')[0]")"
if [[ "${node_major}" -ge 22 ]]; then
  pass "Node major ${node_major}"
else
  fail "Node ${node_major} is too old (need 22+)"
fi

if [[ -d web/node_modules ]]; then
  pass "web/node_modules present"
else
  note "web/node_modules missing (run: task setup)"
fi

if [[ -f go.sum ]]; then
  pass "go.sum present"
else
  fail "go.sum missing"
fi

for opt in air golangci-lint govulncheck lefthook; do
  if command -v "${opt}" >/dev/null 2>&1; then
    pass "optional ${opt}"
  else
    note "optional ${opt} not installed (task tools:install)"
  fi
done

if command -v chromium >/dev/null 2>&1 || command -v google-chrome >/dev/null 2>&1 || command -v chromium-browser >/dev/null 2>&1; then
  pass "Chromium/Chrome available for Lighthouse/e2e"
else
  note "no Chromium/Chrome found (needed for task test:lighthouse / Playwright)"
fi

if ss -ltn 2>/dev/null | grep -qE '[:.]8080\s'; then
  note "port 8080 appears in use"
else
  pass "port 8080 looks free"
fi

if ss -ltn 2>/dev/null | grep -qE '[:.]5173\s'; then
  note "port 5173 appears in use"
else
  pass "port 5173 looks free"
fi

printf '\nDoctor finished: %d checks ok' "${ok}"
if [[ "${warn}" -gt 0 ]]; then
  printf ', %d warnings' "${warn}"
fi
printf '\n'
