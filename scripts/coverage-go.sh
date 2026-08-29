#!/usr/bin/env bash
# Write Go coverage profiles and fail when total coverage is below COVERAGE_MIN.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

OUT_DIR="${COVERAGE_DIR:-./coverage}"
MIN="${COVERAGE_MIN:-45}"
mkdir -p "${OUT_DIR}"
PROFILE="${OUT_DIR}/coverage.out"

go test -mod=vendor -count=1 -covermode=atomic -coverprofile="${PROFILE}" \
  -coverpkg=./internal/... ./internal/...

go tool cover -func="${PROFILE}" | tee "${OUT_DIR}/coverage.txt"
go tool cover -html="${PROFILE}" -o "${OUT_DIR}/coverage.html"

total_line="$(tail -n 1 "${OUT_DIR}/coverage.txt")"
echo "${total_line}"
echo "HTML report: ${OUT_DIR}/coverage.html"

pct="$(echo "${total_line}" | awk '{print $NF}' | tr -d '%')"
awk -v pct="${pct}" -v min="${MIN}" 'BEGIN {
  if (pct+0 < min+0) {
    printf "coverage %.1f%% is below minimum %s%%\n", pct, min > "/dev/stderr"
    exit 1
  }
  printf "coverage %.1f%% meets minimum %s%%\n", pct, min
}'
