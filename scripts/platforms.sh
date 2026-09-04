#!/usr/bin/env bash
# Shared release/CI target matrix helpers for Athenaeum cross builds.
# Usage:
#   source scripts/platforms.sh
#   platforms_list            # print "goos goarch goarm suffix" lines
#   platforms_gha_include     # JSON array for strategy.matrix.include
#   platforms_artifact_name VERSION GOOS GOARCH GOARM
#   platforms_build_one       # build one target (see env below)
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


# platforms_selfcheck_list prints self-check matrix rows:
# name goos goarch goarm suffix runner qemu
# qemu is "-" when unused (native runner).
platforms_selfcheck_list() {
  cat <<'EOF'
linux-amd64 linux amd64 - amd64 ubuntu-latest -
linux-arm64 linux arm64 - arm64 ubuntu-latest qemu-aarch64
linux-armv7 linux arm 7 armv7 ubuntu-latest qemu-arm
linux-armv6 linux arm 6 armv6 ubuntu-latest qemu-arm
linux-riscv64 linux riscv64 - riscv64 ubuntu-latest qemu-riscv64
darwin-arm64 darwin arm64 - arm64 macos-latest -
windows-amd64 windows amd64 - amd64 windows-latest -
EOF
}

# platforms_selfcheck_include prints JSON for strategy.matrix.include.
platforms_selfcheck_include() {
  local first=1 name goos goarch goarm suffix runner qemu goarm_json qemu_json
  printf '['
  while read -r name goos goarch goarm suffix runner qemu; do
    [[ -z "${name}" || "${name}" =~ ^# ]] && continue
    goarm_json="${goarm}"
    if [[ "${goarm}" == "-" ]]; then
      goarm_json=""
    fi
    qemu_json="${qemu}"
    if [[ "${qemu}" == "-" ]]; then
      qemu_json=""
    fi
    if [[ "${first}" -eq 0 ]]; then
      printf ','
    fi
    first=0
    printf '{"name":"%s","goos":"%s","goarch":"%s","goarm":"%s","suffix":"%s","runner":"%s","qemu":"%s"}' \
      "${name}" "${goos}" "${goarch}" "${goarm_json}" "${suffix}" "${runner}" "${qemu_json}"
  done < <(platforms_selfcheck_list)
  printf ']\n'
}

# platforms_gha_include prints a compact JSON array for GitHub Actions
# strategy.matrix.include. goarm is "" when unused (matches prior YAML).
platforms_gha_include() {
  local first=1 goos goarch goarm suffix name goarm_json
  printf '['
  while read -r goos goarch goarm suffix; do
    [[ -z "${goos}" || "${goos}" =~ ^# ]] && continue
    name="${goos}-${suffix}"
    goarm_json="${goarm}"
    if [[ "${goarm}" == "-" ]]; then
      goarm_json=""
    fi
    if [[ "${first}" -eq 0 ]]; then
      printf ','
    fi
    first=0
    printf '{"name":"%s","goos":"%s","goarch":"%s","goarm":"%s","suffix":"%s"}' \
      "${name}" "${goos}" "${goarch}" "${goarm_json}" "${suffix}"
  done < <(platforms_list)
  printf ']\n'
}

# platforms_artifact_name VERSION GOOS GOARCH [GOARM]
# Prints athenaeum[-slim]-VERSION-GOOS-SUFFIX[.exe]
# Set SLIM=1 to prefix the artifact with athenaeum-slim.
# Set NAME_STYLE=ci to omit VERSION (athenaeum-GOOS-SUFFIX[.exe]).
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
  local out
  case "${NAME_STYLE:-release}" in
    ci) out="${prefix}-${goos}-${suffix}" ;;
    *) out="${prefix}-${ver}-${goos}-${suffix}" ;;
  esac
  if [[ "${goos}" == "windows" ]]; then
    out="${out}.exe"
  fi
  printf '%s\n' "${out}"
}

# platforms_repo_root resolves the repository root from this file's location.
platforms_repo_root() {
  local here
  here="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
  printf '%s\n' "${here}"
}

# platforms_build_one cross-compiles one target into OUT_DIR.
# Required env: GOOS, GOARCH, VERSION (unless NAME_STYLE=ci and EMBED_VERSION=0)
# Optional env: GOARM, OUT_DIR (default: <repo>/dist), SLIM, NAME_STYLE,
#   EMBED_VERSION (default 1), CHECKSUM (default 1), LDFLAGS (override),
#   CGO_ENABLED (default 0), ROOT
platforms_build_one() {
  local root="${ROOT:-$(platforms_repo_root)}"
  local goos="${GOOS:?GOOS is required}"
  local goarch="${GOARCH:?GOARCH is required}"
  local goarm="${GOARM:-}"
  if [[ -z "${goarm}" ]]; then
    goarm="-"
  fi
  local version="${VERSION:-dev}"
  local out_dir="${OUT_DIR:-${root}/dist}"
  mkdir -p "${out_dir}"

  local out
  out="$(platforms_artifact_name "${version}" "${goos}" "${goarch}" "${goarm}")"
  echo "building ${out}"

  local ldflags
  if [[ -n "${LDFLAGS:-}" ]]; then
    ldflags="${LDFLAGS}"
  else
    case "${EMBED_VERSION:-1}" in
      0|false|no|NO|FALSE) ldflags="-s -w" ;;
      *)
        ldflags="-s -w -X athenaeum/internal/version.Version=${version} -X athenaeum/internal/version.WebVersion=${version}"
        ;;
    esac
  fi

  local env_args=(CGO_ENABLED="${CGO_ENABLED:-0}" "GOOS=${goos}" "GOARCH=${goarch}")
  if [[ "${goarm}" != "-" ]]; then
    env_args+=("GOARM=${goarm}")
  fi

  env "${env_args[@]}" go build -mod=vendor -trimpath -ldflags "${ldflags}" \
    -o "${out_dir}/${out}" "${root}/cmd/athenaeum"

  case "${CHECKSUM:-1}" in
    0|false|no|NO|FALSE) return 0 ;;
  esac
  (
    cd "${out_dir}"
    if command -v sha256sum >/dev/null 2>&1; then
      sha256sum "${out}" > "${out}.sha256"
    elif command -v shasum >/dev/null 2>&1; then
      shasum -a 256 "${out}" > "${out}.sha256"
    fi
  )
}

if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
  case "${1:-}" in
    gha-include)
      platforms_gha_include
      ;;
    selfcheck-include)
      platforms_selfcheck_include
      ;;
    list)
      platforms_list
      ;;
    *)
      echo "usage: $0 {gha-include|selfcheck-include|list}" >&2
      exit 1
      ;;
  esac
fi
