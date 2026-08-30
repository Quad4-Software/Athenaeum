#!/usr/bin/env bash
# Shared release/CI target matrix helpers for Athenaeum cross builds.
# Usage:
#   source scripts/platforms.sh
#   platforms_list            # print "goos goarch goarm suffix" lines
#   platforms_artifact_name VERSION GOOS GOARCH GOARM
set -euo pipefail

platforms_list() {
  cat <<'EOF'
linux amd64 - amd64
linux arm64 - arm64
linux arm 6 armv6
linux arm 7 armv7
linux riscv64 - riscv64
darwin amd64 - amd64
darwin arm64 - arm64
windows amd64 - amd64
windows arm64 - arm64
freebsd amd64 - amd64
freebsd arm64 - arm64
openbsd amd64 - amd64
openbsd arm64 - arm64
netbsd amd64 - amd64
EOF
}

# platforms_artifact_name VERSION GOOS GOARCH [GOARM]
# Prints athenaeum[-slim]-VERSION-GOOS-SUFFIX[.exe]
# Set SLIM=1 to prefix the artifact with athenaeum-slim.
platforms_artifact_name() {
  local ver="$1" goos="$2" goarch="$3" goarm="${4:--}"
  local suffix
  case "${goos}/${goarch}/${goarm}" in
    linux/arm/6) suffix="armv6" ;;
    linux/arm/7) suffix="armv7" ;;
    *) suffix="${goarch}" ;;
  esac
  local prefix="athenaeum"
  case "${SLIM:-0}" in
    1|true|yes|YES|TRUE) prefix="athenaeum-slim" ;;
  esac
  local out="${prefix}-${ver}-${goos}-${suffix}"
  if [[ "${goos}" == "windows" ]]; then
    out="${out}.exe"
  fi
  printf '%s\n' "${out}"
}
