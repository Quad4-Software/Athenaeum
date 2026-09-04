#!/usr/bin/env bash
# Cross-compile one Athenaeum target (CI and local).
# Env: GOOS, GOARCH, VERSION (see platforms_build_one), optional GOARM,
#   OUT_DIR, SLIM, NAME_STYLE, EMBED_VERSION, CHECKSUM, LDFLAGS, CGO_ENABLED
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
# shellcheck source=platforms.sh
source "${ROOT}/scripts/platforms.sh"
export ROOT
platforms_build_one
