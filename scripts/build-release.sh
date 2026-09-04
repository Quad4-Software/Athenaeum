#!/usr/bin/env bash
# Cross-compile Athenaeum for every supported release target.
# Env: VERSION (default: dev), OUT_DIR (default: dist), SLIM=1 for athenaeum-slim-*
# names (frontend must already be built with VITE_SLIM=1), optional LDFLAGS extras.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
# shellcheck source=platforms.sh
source "${ROOT}/scripts/platforms.sh"

VERSION="${VERSION:-dev}"
OUT_DIR="${OUT_DIR:-${ROOT}/dist}"
export ROOT VERSION OUT_DIR

while read -r goos goarch goarm _suffix; do
  [[ -z "${goos}" || "${goos}" =~ ^# ]] && continue
  GOOS="${goos}" GOARCH="${goarch}" GOARM="${goarm}" platforms_build_one
done < <(platforms_list)
