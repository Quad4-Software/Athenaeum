#!/usr/bin/env bash
# Discover and run Go native fuzz targets for a bounded duration.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

FUZZTIME="${FUZZTIME:-10s}"
SCOPE="${SCOPE:-./internal/...}"

# Short budgets flake with "context deadline exceeded" during worker shutdown.
# Keep a floor so local/CI smoke runs stay reliable.
case "${FUZZTIME}" in
  *ms) ;;
  *)
    secs="${FUZZTIME%s}"
    if [[ "${secs}" =~ ^[0-9]+$ ]] && [[ "${secs}" -lt 5 ]]; then
      echo "FUZZTIME=${FUZZTIME} is below 5s floor; using 5s"
      FUZZTIME=5s
    fi
    ;;
esac

# Give go test headroom beyond fuzztime for corpus IO and worker drain.
TIMEOUT="${FUZZ_TIMEOUT:-30m}"

echo "Discovering fuzz targets under ${SCOPE}"
failed=0
found=0

while IFS= read -r pkg; do
  [[ -z "${pkg}" ]] && continue
  mapfile -t names < <(go test -mod=vendor -list '^Fuzz' "${pkg}" 2>/dev/null | grep '^Fuzz' || true)
  for name in "${names[@]:-}"; do
    [[ -z "${name}" ]] && continue
    found=$((found + 1))
    echo "==> ${pkg} ${name} (${FUZZTIME})"
    # -run=^$ skips ordinary tests so oracle/unit output does not mix in.
    # -fuzzminimizetime=0 keeps the budget on exploration, not minimize.
    if ! go test -mod=vendor -count=1 -timeout="${TIMEOUT}" -run='^$' "-fuzz=^${name}$" \
      -fuzztime="${FUZZTIME}" -fuzzminimizetime=0 "${pkg}"; then
      echo "FAIL ${pkg} ${name}"
      failed=1
    fi
  done
done < <(go list -mod=vendor ${SCOPE})

if [[ "${found}" -eq 0 ]]; then
  echo "No fuzz targets found under ${SCOPE}"
  exit 1
fi

if [[ "${failed}" -ne 0 ]]; then
  exit 1
fi
echo "All ${found} fuzz target(s) completed"
