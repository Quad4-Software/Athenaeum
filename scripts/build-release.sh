#!/usr/bin/env bash
# Cross-compile Athenaeum for every supported release target.
# Env: VERSION (default: dev), OUT_DIR (default: dist), optional LDFLAGS extras.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
# shellcheck source=platforms.sh
source "${ROOT}/scripts/platforms.sh"

VERSION="${VERSION:-dev}"
OUT_DIR="${OUT_DIR:-${ROOT}/dist}"
mkdir -p "${OUT_DIR}"

LDFLAGS="-s -w -X athenaeum/internal/version.Version=${VERSION} -X athenaeum/internal/version.WebVersion=${VERSION}"

while read -r goos goarch goarm suffix; do
  [[ -z "${goos}" || "${goos}" =~ ^# ]] && continue
  out="$(platforms_artifact_name "${VERSION}" "${goos}" "${goarch}" "${goarm}")"
  echo "building ${out}"
  env_args=(CGO_ENABLED=0 "GOOS=${goos}" "GOARCH=${goarch}")
  if [[ "${goarm}" != "-" ]]; then
    env_args+=("GOARM=${goarm}")
  fi
  env "${env_args[@]}" go build -mod=vendor -trimpath -ldflags "${LDFLAGS}" \
    -o "${OUT_DIR}/${out}" "${ROOT}/cmd/athenaeum"
  (
    cd "${OUT_DIR}"
    if command -v sha256sum >/dev/null 2>&1; then
      sha256sum "${out}" > "${out}.sha256"
    elif command -v shasum >/dev/null 2>&1; then
      shasum -a 256 "${out}" > "${out}.sha256"
    fi
  )
done < <(platforms_list)
