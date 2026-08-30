#!/usr/bin/env bash
# Athenaeum interactive installer: binary, Docker, or source build.
# Supports dry-run, rollback on failure, and service unit installation.
set -euo pipefail

# curl | bash leaves BASH_SOURCE unset (and set -u would abort on [0]).
if [[ -n "${BASH_SOURCE[0]:-}" ]]; then
  SCRIPT_PATH="$(readlink -f "${BASH_SOURCE[0]}" 2>/dev/null || realpath "${BASH_SOURCE[0]}" 2>/dev/null || echo "${BASH_SOURCE[0]}")"
  ROOT="$(cd "$(dirname "$SCRIPT_PATH")" && pwd)"
else
  SCRIPT_PATH=""
  ROOT="$(pwd)"
fi
DEPLOY_DIR="$ROOT/deploy"

VERSION="latest"
METHOD=""
LISTEN_IP=""
LISTEN_PORT="8080"
LIBRARY_DIR=""
DATA_DIR=""
PREFIX="/usr/local"
INSTALL_USER="athenaeum"
SERVICE_KIND=""
CREATE_USER=1
INSTALL_SERVICE=1
BOOTSTRAP_ADMIN=0
ADMIN_USER=""
ADMIN_PASS=""
DOCKER_IMAGE=""
RELEASE_BASE=""
REPO_SLUG=""
COMPOSE_FILE="$ROOT/docker-compose.yml"
DRY_RUN=0
ASSUME_YES=0
NO_COLOR=0
FORCE_COLOR=0
NONINTERACTIVE=0

# Rollback stack (newline-separated undo commands, executed in reverse).
UNDO_STACK=()
ROLLBACK_DONE=0
INSTALL_STATE_DIR=""

usage() {
  cat <<'EOF'
Usage: install.sh [options]

Install methods:
  --method binary|docker|source   Install path (prompted if omitted)

Network and paths:
  --ip ADDR                       Listen IP (0.0.0.0, 127.0.0.1, or interface IP)
  --port N                        Listen port (default 8080)
  --library DIR                   Library root
  --data DIR                      Data directory
  --prefix DIR                    Install prefix for binary (default /usr/local)

Service:
  --service systemd|openrc|runit|dinit|s6|none
  --user NAME                     System user (default athenaeum)
  --no-user                       Do not create a system user
  --no-service                    Skip service unit installation

Binary / Docker:
  --version TAG                   Release tag or latest (default latest)
  --repo OWNER/NAME               Release repository slug
  --release-base URL              Base URL for release assets
  --image NAME                    Container image (docker method)
  --compose FILE                  Compose file (default docker-compose.yml)

Admin bootstrap (optional):
  --admin-user NAME
  --admin-pass PASS

Behavior:
  --dry-run                       Print actions without changing the system
  -y, --yes                       Accept defaults / non-interactive where possible
  --no-color                      Disable ANSI colors
  --color                         Force ANSI colors
  -h, --help                      Show this help

Environment:
  NO_COLOR, FORCE_COLOR, ATHENAEUM_REPO, ATHENAEUM_RELEASE_BASE,
  ATHENAEUM_IMAGE, ATHENAEUM_VERSION
EOF
}

# --- colors -----------------------------------------------------------------

color_enabled() {
  if [[ $NO_COLOR -eq 1 ]]; then return 1; fi
  if [[ -n "${NO_COLOR:-}" && $FORCE_COLOR -eq 0 ]]; then return 1; fi
  if [[ $FORCE_COLOR -eq 1 ]]; then return 0; fi
  [[ -t 1 ]]
}

c() {
  local code="$1"; shift
  if color_enabled; then printf '\033[%sm%s\033[0m' "$code" "$*"; else printf '%s' "$*"; fi
}
bold() { c 1 "$*"; }
dim()  { c 2 "$*"; }
cyan() { c 36 "$*"; }
green(){ c 32 "$*"; }
yellow(){ c 33 "$*"; }
red()  { c 31 "$*"; }
magenta(){ c 35 "$*"; }

banner() {
  printf '\n%s\n' "$(bold "$(magenta "Athenaeum installer")")"
  printf '%s\n\n' "$(dim "binary · docker · source · services · rollback")"
}

info()  { printf '%s %s\n' "$(cyan "info")" "$*" >&2; }
ok()    { printf '%s %s\n' "$(green "ok")" "$*" >&2; }
warn()  { printf '%s %s\n' "$(yellow "warn")" "$*" >&2; }
err()   { printf '%s %s\n' "$(red "error")" "$*" >&2; }
die()   { err "$*"; exit 1; }

# --- rollback / dry-run -----------------------------------------------------

push_undo() {
  UNDO_STACK+=("$*")
}

run_rollback() {
  [[ $ROLLBACK_DONE -eq 1 ]] && return 0
  ROLLBACK_DONE=1
  if [[ ${#UNDO_STACK[@]} -eq 0 ]]; then
    warn "nothing to roll back"
    return 0
  fi
  warn "rolling back ${#UNDO_STACK[@]} action(s)"
  local i
  for ((i=${#UNDO_STACK[@]}-1; i>=0; i--)); do
    info "undo: ${UNDO_STACK[$i]}"
    if [[ $DRY_RUN -eq 0 ]]; then
      # shellcheck disable=SC2086
      eval "${UNDO_STACK[$i]}" || warn "undo failed: ${UNDO_STACK[$i]}"
    fi
  done
  ok "rollback finished"
}

on_error() {
  local ec=$?
  err "install failed (exit $ec)"
  run_rollback
  exit "$ec"
}

trap on_error ERR

do_cmd() {
  if [[ $DRY_RUN -eq 1 ]]; then
    info "dry-run: $*"
    return 0
  fi
  "$@"
}

do_shell() {
  if [[ $DRY_RUN -eq 1 ]]; then
    info "dry-run: $*"
    return 0
  fi
  eval "$*"
}

write_file() {
  local dest="$1"
  local mode="${2:-0644}"
  local content
  content="$(cat)"
  if [[ $DRY_RUN -eq 1 ]]; then
    info "dry-run: write $dest (mode $mode)"
    return 0
  fi
  local dir
  dir="$(dirname "$dest")"
  mkdir -p "$dir"
  if [[ -e "$dest" ]]; then
    local bak
    bak="${dest}.bak.$(date -u +%Y%m%dT%H%M%SZ)"
    cp -a "$dest" "$bak"
    push_undo "mv -f '$bak' '$dest'"
  else
    push_undo "rm -f '$dest'"
  fi
  printf '%s' "$content" >"$dest"
  chmod "$mode" "$dest"
}

# --- prompts ----------------------------------------------------------------

need_tty() {
  [[ $NONINTERACTIVE -eq 1 ]] && return 1
  [[ $ASSUME_YES -eq 1 ]] && return 1
  [[ -t 0 ]]
}

ask() {
  local prompt="$1"
  local default="${2:-}"
  local reply=""
  if ! need_tty; then
    printf '%s\n' "$default"
    return 0
  fi
  if [[ -n "$default" ]]; then
    printf '%s [%s]: ' "$prompt" "$default" >&2
  else
    printf '%s: ' "$prompt" >&2
  fi
  read -r reply || true
  if [[ -z "$reply" ]]; then
    printf '%s\n' "$default"
  else
    printf '%s\n' "$reply"
  fi
}

choose() {
  local prompt="$1"
  shift
  local options=("$@")
  local i default_idx=0
  if ! need_tty; then
    printf '%s\n' "${options[0]}"
    return 0
  fi
  printf '%s\n' "$(bold "$prompt")" >&2
  for i in "${!options[@]}"; do
    printf '  %s) %s\n' "$((i+1))" "${options[$i]}" >&2
  done
  local reply
  reply="$(ask "Choice" "1")"
  if [[ "$reply" =~ ^[0-9]+$ ]] && (( reply >= 1 && reply <= ${#options[@]} )); then
    printf '%s\n' "${options[$((reply-1))]}"
  else
    # allow typing the value
    for opt in "${options[@]}"; do
      if [[ "$reply" == "$opt" ]]; then
        printf '%s\n' "$opt"
        return 0
      fi
    done
    printf '%s\n' "${options[$default_idx]}"
  fi
}

confirm() {
  local prompt="$1"
  if [[ $ASSUME_YES -eq 1 ]]; then return 0; fi
  if ! need_tty; then return 0; fi
  local ans
  ans="$(ask "$prompt [y/N]" "n")"
  [[ "$ans" == "y" || "$ans" == "Y" || "$ans" == "yes" ]]
}

# --- detection --------------------------------------------------------------

detect_os_arch() {
  local os arch
  os="$(uname -s | tr '[:upper:]' '[:lower:]')"
  arch="$(uname -m)"
  case "$arch" in
    x86_64|amd64) arch="amd64" ;;
    aarch64|arm64) arch="arm64" ;;
    armv6l) arch="armv6" ;;
    armv7l) arch="armv7" ;;
    riscv64) arch="riscv64" ;;
    *) die "unsupported architecture: $arch" ;;
  esac
  case "$os" in
    linux|darwin) ;;
    freebsd|openbsd|netbsd) ;;
    mingw*|msys*|cygwin*|windows_nt) os="windows" ;;
    *) die "unsupported OS: $os (use docker or build from source)" ;;
  esac
  printf '%s %s\n' "$os" "$arch"
}

list_ips() {
  local ips=()
  if command -v ip >/dev/null 2>&1; then
    while IFS= read -r line; do
      [[ -n "$line" ]] && ips+=("$line")
    done < <(ip -4 -o addr show scope global 2>/dev/null | awk '{print $4}' | cut -d/ -f1)
  elif command -v ifconfig >/dev/null 2>&1; then
    while IFS= read -r line; do
      [[ -n "$line" ]] && ips+=("$line")
    done < <(ifconfig 2>/dev/null | awk '/inet / && $2 != "127.0.0.1" {print $2}' | sed 's/addr://')
  fi
  printf '%s\n' "${ips[@]}"
}

detect_repo_slug() {
  if [[ -n "${ATHENAEUM_REPO:-}" ]]; then
    printf '%s\n' "$ATHENAEUM_REPO"
    return 0
  fi
  if [[ -n "$REPO_SLUG" ]]; then
    printf '%s\n' "$REPO_SLUG"
    return 0
  fi
  if [[ -d "$ROOT/.git" ]] && command -v git >/dev/null 2>&1; then
    local url
    url="$(git -C "$ROOT" remote get-url origin 2>/dev/null || true)"
    if [[ "$url" =~ [:/]([^/:]+)/([^/]+)(\.git)?$ ]]; then
      printf '%s/%s\n' "${BASH_REMATCH[1]}" "${BASH_REMATCH[2]%.git}"
      return 0
    fi
  fi
  # Fallback for curl | bash and checkouts without a remote.
  printf 'Quad4-Software/Athenaeum\n'
}

detect_init() {
  if command -v systemctl >/dev/null 2>&1 && [[ -d /run/systemd/system ]]; then
    printf 'systemd\n'
  elif command -v rc-service >/dev/null 2>&1 || [[ -d /etc/init.d ]]; then
    printf 'openrc\n'
  elif command -v dinitctl >/dev/null 2>&1; then
    printf 'dinit\n'
  elif [[ -d /etc/sv || -d /var/service ]]; then
    printf 'runit\n'
  elif command -v s6-svscan >/dev/null 2>&1 || [[ -d /service ]]; then
    printf 's6\n'
  else
    printf 'none\n'
  fi
}

have_root() {
  [[ "$(id -u)" -eq 0 ]]
}

ensure_cmd() {
  local bin="$1"
  command -v "$bin" >/dev/null 2>&1 || die "required command not found: $bin"
}

# --- parse args -------------------------------------------------------------

while [[ $# -gt 0 ]]; do
  case "$1" in
    --method) METHOD="$2"; shift 2 ;;
    --ip) LISTEN_IP="$2"; shift 2 ;;
    --port) LISTEN_PORT="$2"; shift 2 ;;
    --library) LIBRARY_DIR="$2"; shift 2 ;;
    --data) DATA_DIR="$2"; shift 2 ;;
    --prefix) PREFIX="$2"; shift 2 ;;
    --service) SERVICE_KIND="$2"; shift 2 ;;
    --user) INSTALL_USER="$2"; shift 2 ;;
    --no-user) CREATE_USER=0; shift ;;
    --no-service) INSTALL_SERVICE=0; SERVICE_KIND="none"; shift ;;
    --version) VERSION="$2"; shift 2 ;;
    --repo) REPO_SLUG="$2"; shift 2 ;;
    --release-base) RELEASE_BASE="$2"; shift 2 ;;
    --image) DOCKER_IMAGE="$2"; shift 2 ;;
    --compose) COMPOSE_FILE="$2"; shift 2 ;;
    --admin-user) ADMIN_USER="$2"; BOOTSTRAP_ADMIN=1; shift 2 ;;
    --admin-pass) ADMIN_PASS="$2"; BOOTSTRAP_ADMIN=1; shift 2 ;;
    --dry-run) DRY_RUN=1; shift ;;
    -y|--yes) ASSUME_YES=1; NONINTERACTIVE=1; shift ;;
    --no-color) NO_COLOR=1; shift ;;
    --color) FORCE_COLOR=1; shift ;;
    -h|--help) usage; exit 0 ;;
    *) die "unknown option: $1 (see --help)" ;;
  esac
done

VERSION="${ATHENAEUM_VERSION:-$VERSION}"
RELEASE_BASE="${ATHENAEUM_RELEASE_BASE:-$RELEASE_BASE}"
DOCKER_IMAGE="${ATHENAEUM_IMAGE:-$DOCKER_IMAGE}"
REPO_SLUG="$(detect_repo_slug)"

# When install.sh is run via curl | bash, ROOT is not the repo tree.
# Clone into a durable directory so docker/source methods and unit templates work.
ensure_repo_tree() {
  if [[ -f "$ROOT/go.mod" && -f "$ROOT/docker-compose.yml" ]]; then
    return 0
  fi
  command -v git >/dev/null 2>&1 || die "git is required to fetch Athenaeum when not running from a checkout"
  local dest="${ATHENAEUM_INSTALL_DIR:-$HOME/athenaeum}"
  if [[ -f "$dest/go.mod" && -f "$dest/docker-compose.yml" ]]; then
    info "using existing checkout at $dest"
  else
    info "cloning https://github.com/${REPO_SLUG}.git into $dest"
    if [[ $DRY_RUN -eq 1 ]]; then
      info "dry-run: git clone https://github.com/${REPO_SLUG}.git $dest"
    else
      mkdir -p "$(dirname "$dest")"
      if [[ -d "$dest/.git" ]]; then
        git -C "$dest" fetch --depth 1 origin master
        git -C "$dest" checkout -f FETCH_HEAD
      else
        rm -rf "$dest"
        git clone --depth 1 "https://github.com/${REPO_SLUG}.git" "$dest"
      fi
    fi
  fi
  ROOT="$dest"
  DEPLOY_DIR="$ROOT/deploy"
  COMPOSE_FILE="$ROOT/docker-compose.yml"
}

ensure_repo_tree

# --- interactive configuration ----------------------------------------------

apply_nonroot_defaults() {
  have_root && return 0

  if [[ $CREATE_USER -eq 1 ]]; then
    warn "not root: skipping system user creation (re-run with sudo for a dedicated user)"
    CREATE_USER=0
    INSTALL_USER="$(id -un)"
  fi

  # Piped / non-interactive installs must not default to systemd without root.
  if [[ $INSTALL_SERVICE -eq 1 && -z "$SERVICE_KIND" ]] && ! need_tty; then
    warn "not root: skipping service unit installation"
    INSTALL_SERVICE=0
    SERVICE_KIND="none"
  fi

  if [[ "$PREFIX" == "/usr/local" ]] && [[ ! -w /usr/local/bin && ! -w /usr/local ]]; then
    PREFIX="${XDG_BIN_HOME:-$HOME/.local}"
    warn "not root: install prefix set to $PREFIX"
  fi
}

configure() {
  banner

  if [[ $DRY_RUN -eq 1 ]]; then
    warn "dry-run mode: no files will be changed"
  fi

  apply_nonroot_defaults

  if [[ -z "$METHOD" ]]; then
    METHOD="$(choose "Install method" "binary" "docker" "source")"
  fi
  case "$METHOD" in
    binary|docker|source) ;;
    *) die "invalid --method: $METHOD" ;;
  esac

  local ips=()
  mapfile -t ips < <(list_ips)
  if [[ -z "$LISTEN_IP" ]]; then
    local ip_choices=("0.0.0.0 (all interfaces)" "127.0.0.1 (loopback only)" "custom")
    local ip
    for ip in "${ips[@]}"; do
      ip_choices+=("$ip")
    done
    local picked
    picked="$(choose "Listen address" "${ip_choices[@]}")"
    case "$picked" in
      "0.0.0.0 (all interfaces)") LISTEN_IP="0.0.0.0" ;;
      "127.0.0.1 (loopback only)") LISTEN_IP="127.0.0.1" ;;
      custom) LISTEN_IP="$(ask "Custom IP" "0.0.0.0")" ;;
      *) LISTEN_IP="$picked" ;;
    esac
  fi

  if need_tty || [[ -z "${LISTEN_PORT}" ]]; then
    LISTEN_PORT="$(ask "Listen port" "$LISTEN_PORT")"
  fi
  [[ "$LISTEN_PORT" =~ ^[0-9]+$ ]] || die "invalid port: $LISTEN_PORT"
  (( LISTEN_PORT >= 1 && LISTEN_PORT <= 65535 )) || die "port out of range: $LISTEN_PORT"

  local default_data="/var/lib/athenaeum/data"
  local default_library="/var/lib/athenaeum/library"
  if ! have_root && [[ "$METHOD" != "docker" ]]; then
    default_data="$ROOT/data"
    default_library="$ROOT/library"
  fi
  if [[ "$METHOD" == "docker" ]]; then
    default_data="docker-volume:athenaeum-data"
    default_library="${ATHENAEUM_LIBRARY_HOST_PATH:-$ROOT/library}"
  fi

  DATA_DIR="${DATA_DIR:-$(ask "Data directory" "$default_data")}"
  LIBRARY_DIR="${LIBRARY_DIR:-$(ask "Library directory" "$default_library")}"

  if [[ $INSTALL_SERVICE -eq 1 && -z "$SERVICE_KIND" ]]; then
    local detected
    detected="$(detect_init)"
    if ! have_root; then
      SERVICE_KIND="none"
      INSTALL_SERVICE=0
      warn "not root: skipping service unit installation"
    else
      SERVICE_KIND="$(choose "Service manager (detected: $detected)" "systemd" "openrc" "runit" "dinit" "s6" "none")"
    fi
  fi
  SERVICE_KIND="${SERVICE_KIND:-none}"

  if [[ $BOOTSTRAP_ADMIN -eq 0 ]] && need_tty; then
    if confirm "Bootstrap an admin user now?"; then
      BOOTSTRAP_ADMIN=1
      ADMIN_USER="$(ask "Admin username" "admin")"
      ADMIN_PASS="$(ask "Admin password" "")"
      [[ -n "$ADMIN_PASS" ]] || die "admin password required"
    fi
  fi

  printf '\n%s\n' "$(bold "Summary")"
  printf '  method:   %s\n' "$METHOD"
  printf '  listen:   %s:%s\n' "$LISTEN_IP" "$LISTEN_PORT"
  printf '  data:     %s\n' "$DATA_DIR"
  printf '  library:  %s\n' "$LIBRARY_DIR"
  printf '  prefix:   %s\n' "$PREFIX"
  printf '  user:     %s\n' "$INSTALL_USER"
  printf '  service:  %s\n' "$SERVICE_KIND"
  printf '  version:  %s\n' "$VERSION"
  if [[ $BOOTSTRAP_ADMIN -eq 1 ]]; then
    printf '  admin:    %s\n' "$ADMIN_USER"
  fi
  if [[ $DRY_RUN -eq 1 ]]; then
    printf '  mode:     %s\n' "$(yellow dry-run)"
  fi
  printf '\n'
  confirm "Proceed with installation?" || die "aborted"
}

# --- filesystem helpers -----------------------------------------------------

create_system_user() {
  [[ $CREATE_USER -eq 1 ]] || return 0
  [[ "$METHOD" == "docker" ]] && return 0
  if id "$INSTALL_USER" >/dev/null 2>&1; then
    ok "user $INSTALL_USER already exists"
    return 0
  fi
  if ! have_root; then
    if [[ $DRY_RUN -eq 1 ]]; then
      info "dry-run: would create user $INSTALL_USER"
      return 0
    fi
    die "root required to create user $INSTALL_USER (or pass --no-user)"
  fi
  info "creating system user $INSTALL_USER"
  if command -v useradd >/dev/null 2>&1; then
    do_cmd useradd --system --home-dir /var/lib/athenaeum --shell /usr/sbin/nologin --user-group "$INSTALL_USER"
  elif command -v adduser >/dev/null 2>&1; then
    do_cmd adduser -S -H -h /var/lib/athenaeum -s /sbin/nologin -D "$INSTALL_USER" 2>/dev/null \
      || do_cmd adduser --system --home /var/lib/athenaeum --shell /usr/sbin/nologin --group "$INSTALL_USER"
  else
    die "cannot create user: no useradd/adduser"
  fi
  push_undo "userdel '$INSTALL_USER' 2>/dev/null || true"
}

ensure_dirs() {
  [[ "$METHOD" == "docker" ]] && return 0
  local dirs=()
  [[ "$DATA_DIR" != docker-volume:* ]] && dirs+=("$DATA_DIR")
  dirs+=("$LIBRARY_DIR")
  local d
  for d in "${dirs[@]}"; do
    info "ensuring directory $d"
    do_cmd mkdir -p "$d"
    if have_root && id "$INSTALL_USER" >/dev/null 2>&1; then
      do_cmd chown -R "$INSTALL_USER:$INSTALL_USER" "$d"
    fi
    if [[ $DRY_RUN -eq 0 && ! -d "$d" ]]; then
      push_undo "rmdir '$d' 2>/dev/null || true"
    fi
  done
}

write_env_file() {
  local dest="$1"
  local addr="${LISTEN_IP}:${LISTEN_PORT}"
  if [[ "$LISTEN_IP" == "0.0.0.0" ]]; then
    addr=":${LISTEN_PORT}"
  fi
  local data="$DATA_DIR"
  local library="$LIBRARY_DIR"
  if [[ "$METHOD" == "docker" ]]; then
    data="/data"
    library="/library"
  fi
  local body
  body="$(cat <<EOF
# Generated by Athenaeum install.sh on $(date -u +%Y-%m-%dT%H:%M:%SZ)
ATHENAEUM_ADDR=${addr}
ATHENAEUM_DATA=${data}
ATHENAEUM_LIBRARY=${library}
ATHENAEUM_UPLOAD_MAX_BYTES=2147483648
ATHENAEUM_LOG_LEVEL=info
ATHENAEUM_COLOR=never
EOF
)"
  if [[ $BOOTSTRAP_ADMIN -eq 1 ]]; then
    body+=$'\n'"ATHENAEUM_ADMIN_USER=${ADMIN_USER}"
    body+=$'\n'"ATHENAEUM_ADMIN_PASS=${ADMIN_PASS}"
  fi
  printf '%s\n' "$body" | write_file "$dest" 0640
  if have_root && [[ $DRY_RUN -eq 0 ]]; then
    chown root:"$INSTALL_USER" "$dest" 2>/dev/null || true
  fi
  ok "wrote env $dest"
}

install_binary_to_prefix() {
  local src="$1"
  local dest="$PREFIX/bin/athenaeum"
  info "installing binary to $dest"
  do_cmd mkdir -p "$PREFIX/bin"
  if [[ -e "$dest" && $DRY_RUN -eq 0 ]]; then
    local bak
    bak="${dest}.bak.$(date -u +%Y%m%dT%H%M%SZ)"
    cp -a "$dest" "$bak"
    push_undo "mv -f '$bak' '$dest'"
  elif [[ $DRY_RUN -eq 0 ]]; then
    push_undo "rm -f '$dest'"
  fi
  do_cmd install -m 0755 "$src" "$dest"
  if [[ -d "$ROOT/man" ]]; then
    do_cmd mkdir -p "$PREFIX/share/man/man1"
    do_cmd install -m 0644 "$ROOT/man/athenaeum.1" "$PREFIX/share/man/man1/" 2>/dev/null || true
    do_cmd install -m 0644 "$ROOT/man/athenaeum-users.1" "$PREFIX/share/man/man1/" 2>/dev/null || true
  fi
}

# --- install methods --------------------------------------------------------

resolve_release_tag() {
  local tag="$VERSION"
  if [[ "$tag" == "latest" ]]; then
    if [[ -n "$RELEASE_BASE" ]]; then
      printf 'latest\n'
      return 0
    fi
    if command -v curl >/dev/null 2>&1 && [[ -n "$REPO_SLUG" ]]; then
      local api="https://api.github.com/repos/${REPO_SLUG}/releases/latest"
      local resolved
      resolved="$(curl -fsSL "$api" 2>/dev/null | sed -n 's/.*"tag_name"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' | head -n1 || true)"
      if [[ -n "$resolved" ]]; then
        printf '%s\n' "$resolved"
        return 0
      fi
    fi
    # Fall back to local package version when offline / no releases yet.
    if [[ -f "$ROOT/web/package.json" ]]; then
      sed -n 's/.*"version"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' "$ROOT/web/package.json" | head -n1
      return 0
    fi
    printf 'latest\n'
    return 0
  fi
  printf '%s\n' "$tag"
}

download_binary() {
  ensure_cmd curl
  local os arch tag ver asset url
  read -r os arch <<<"$(detect_os_arch)"
  tag="$(resolve_release_tag)"
  ver="${tag#v}"
  asset="athenaeum-${ver}-${os}-${arch}"
  [[ "$os" == "windows" ]] && asset="${asset}.exe"

  INSTALL_STATE_DIR="$(mktemp -d "${TMPDIR:-/tmp}/athenaeum-install.XXXXXX")"
  push_undo "rm -rf '$INSTALL_STATE_DIR'"

  if [[ -n "$RELEASE_BASE" ]]; then
    url="${RELEASE_BASE%/}/${asset}"
  elif [[ -n "$REPO_SLUG" ]]; then
    url="https://github.com/${REPO_SLUG}/releases/download/${tag}/${asset}"
  else
    die "cannot resolve download URL (set --repo or --release-base)"
  fi

  info "downloading $url"
  if [[ $DRY_RUN -eq 1 ]]; then
    info "dry-run: curl -fL $url -o $INSTALL_STATE_DIR/$asset"
    printf '%s\n' "$INSTALL_STATE_DIR/$asset"
    return 0
  fi
  if ! curl -fL --retry 3 --retry-delay 2 "$url" -o "$INSTALL_STATE_DIR/$asset"; then
    return 1
  fi
  if curl -fsSL "${url}.sha256" -o "$INSTALL_STATE_DIR/${asset}.sha256" 2>/dev/null; then
    info "verifying checksum"
    (cd "$INSTALL_STATE_DIR" && sha256sum -c "${asset}.sha256") || die "checksum mismatch"
  else
    warn "no .sha256 sidecar found; skipping checksum verification"
  fi
  chmod +x "$INSTALL_STATE_DIR/$asset"
  printf '%s\n' "$INSTALL_STATE_DIR/$asset"
}

github_release_available() {
  [[ -n "$RELEASE_BASE" ]] && return 0
  [[ -n "$REPO_SLUG" ]] || return 1
  local api="https://api.github.com/repos/${REPO_SLUG}/releases/latest"
  curl -fsSL "$api" >/dev/null 2>&1
}

install_binary_method() {
  local bin=""
  if [[ -x "$ROOT/bin/athenaeum" ]] && need_tty \
    && confirm "Use existing ./bin/athenaeum instead of downloading?"; then
    bin="$ROOT/bin/athenaeum"
    install_binary_to_prefix "$bin"
    return 0
  fi

  if ! github_release_available && [[ -z "$RELEASE_BASE" ]]; then
    if [[ -x "$ROOT/bin/athenaeum" ]]; then
      info "no GitHub release found; using existing $ROOT/bin/athenaeum"
      install_binary_to_prefix "$ROOT/bin/athenaeum"
      return 0
    fi
    warn "no GitHub release found; building from source"
    if [[ $DRY_RUN -eq 1 ]]; then
      info "dry-run: would build from source in $ROOT"
      return 0
    fi
    install_source_method
    return 0
  fi

  if bin="$(download_binary)"; then
    install_binary_to_prefix "$bin"
    return 0
  fi

  if [[ -x "$ROOT/bin/athenaeum" ]]; then
    warn "download failed; using existing $ROOT/bin/athenaeum"
    install_binary_to_prefix "$ROOT/bin/athenaeum"
    return 0
  fi
  warn "download failed; building from source"
  if [[ $DRY_RUN -eq 1 ]]; then
    info "dry-run: would build from source in $ROOT"
    return 0
  fi
  install_source_method
}

install_source_method() {
  ensure_cmd go
  ensure_cmd node
  if command -v corepack >/dev/null 2>&1; then
    info "activating packageManager via corepack"
    do_cmd corepack enable
    if [[ -f "$ROOT/web/package.json" ]]; then
      local pm
      pm="$(sed -n 's/.*"packageManager"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' "$ROOT/web/package.json" | head -n1)"
      if [[ -n "$pm" ]]; then
        do_cmd corepack prepare "$pm" --activate
      fi
    fi
  fi
  ensure_cmd pnpm
  info "building from source in $ROOT"
  if command -v task >/dev/null 2>&1; then
    do_cmd task -d "$ROOT" setup
    do_cmd task -d "$ROOT" build
  else
    do_cmd bash -lc "cd '$ROOT' && go mod download && cd web && pnpm install && pnpm build"
    do_cmd bash -lc "cd '$ROOT' && mkdir -p bin && go build -trimpath -o bin/athenaeum ./cmd/athenaeum"
  fi
  [[ $DRY_RUN -eq 1 || -x "$ROOT/bin/athenaeum" ]] || die "build produced no binary"
  install_binary_to_prefix "$ROOT/bin/athenaeum"
}

install_docker_method() {
  ensure_cmd docker
  command -v docker >/dev/null 2>&1 || die "docker not found"
  local image="${DOCKER_IMAGE}"
  if [[ -z "$image" ]]; then
    if [[ -n "$REPO_SLUG" ]]; then
      image="ghcr.io/$(echo "$REPO_SLUG" | tr '[:upper:]' '[:lower:]'):${VERSION#v}"
      [[ "$VERSION" == "latest" ]] && image="ghcr.io/$(echo "$REPO_SLUG" | tr '[:upper:]' '[:lower:]'):latest"
    else
      image="athenaeum:local"
    fi
  fi

  local env_path="$ROOT/.env"
  if [[ ! -f "$env_path" ]]; then
    if [[ -f "$ROOT/.env.example" ]]; then
      info "creating .env from .env.example"
      if [[ $DRY_RUN -eq 0 ]]; then
        cp "$ROOT/.env.example" "$env_path"
        push_undo "rm -f '$env_path'"
      else
        info "dry-run: cp .env.example .env"
      fi
    else
      write_file "$env_path" 0600 <<EOF
ATHENAEUM_LIBRARY_HOST_PATH=${LIBRARY_DIR}
ATHENAEUM_PUBLISH_PORT=${LISTEN_PORT}
EOF
    fi
  fi

  # Patch compose-oriented env keys.
  if [[ $DRY_RUN -eq 0 ]]; then
    grep -q '^ATHENAEUM_LIBRARY_HOST_PATH=' "$env_path" 2>/dev/null \
      && sed -i.bak "s|^ATHENAEUM_LIBRARY_HOST_PATH=.*|ATHENAEUM_LIBRARY_HOST_PATH=${LIBRARY_DIR}|" "$env_path" \
      || printf '\nATHENAEUM_LIBRARY_HOST_PATH=%s\n' "$LIBRARY_DIR" >>"$env_path"
    grep -q '^ATHENAEUM_PUBLISH_PORT=' "$env_path" 2>/dev/null \
      && sed -i "s|^ATHENAEUM_PUBLISH_PORT=.*|ATHENAEUM_PUBLISH_PORT=${LISTEN_PORT}|" "$env_path" \
      || printf 'ATHENAEUM_PUBLISH_PORT=%s\n' "$LISTEN_PORT" >>"$env_path"
    if [[ $BOOTSTRAP_ADMIN -eq 1 ]]; then
      grep -q '^ATHENAEUM_ADMIN_USER=' "$env_path" 2>/dev/null \
        && sed -i "s|^ATHENAEUM_ADMIN_USER=.*|ATHENAEUM_ADMIN_USER=${ADMIN_USER}|" "$env_path" \
        || printf 'ATHENAEUM_ADMIN_USER=%s\n' "$ADMIN_USER" >>"$env_path"
      grep -q '^ATHENAEUM_ADMIN_PASS=' "$env_path" 2>/dev/null \
        && sed -i "s|^ATHENAEUM_ADMIN_PASS=.*|ATHENAEUM_ADMIN_PASS=${ADMIN_PASS}|" "$env_path" \
        || printf 'ATHENAEUM_ADMIN_PASS=%s\n' "$ADMIN_PASS" >>"$env_path"
    fi
    rm -f "${env_path}.bak"
  else
    info "dry-run: update $env_path library/port/admin settings"
  fi

  do_cmd mkdir -p "$LIBRARY_DIR"

  if [[ "$image" == "athenaeum:local" || "$image" == *:local ]]; then
    info "building local image"
    do_cmd docker compose -f "$COMPOSE_FILE" build
  else
    info "pulling $image"
    do_cmd docker pull "$image" || warn "pull failed; will try compose build"
  fi

  info "starting compose stack"
  do_cmd env ATHENAEUM_LIBRARY_HOST_PATH="$LIBRARY_DIR" ATHENAEUM_PUBLISH_PORT="$LISTEN_PORT" \
    docker compose -f "$COMPOSE_FILE" up -d
  push_undo "docker compose -f '$COMPOSE_FILE' down"
  ok "docker stack is up"
}

# --- services ---------------------------------------------------------------

install_service_units() {
  [[ "$SERVICE_KIND" == "none" ]] && return 0
  [[ "$METHOD" == "docker" ]] && { warn "skipping host service units for docker method"; return 0; }

  local env_dest="/etc/athenaeum/athenaeum.env"
  if ! have_root; then
    env_dest="$ROOT/deploy/env/athenaeum.env"
    warn "not root: writing env to $env_dest (copy to /etc/athenaeum/ when ready)"
  fi
  do_cmd mkdir -p "$(dirname "$env_dest")"
  write_env_file "$env_dest"

  case "$SERVICE_KIND" in
    systemd)
      local unit_src="$DEPLOY_DIR/systemd/athenaeum.service"
      local unit_dst="/etc/systemd/system/athenaeum.service"
      [[ -f "$unit_src" ]] || die "missing $unit_src"
      if ! have_root; then
        warn "not root: service file is at $unit_src"
        return 0
      fi
      # Rewrite Exec path / env for prefix
      local unit
      unit="$(sed "s|/usr/local/bin/athenaeum|${PREFIX}/bin/athenaeum|g" "$unit_src")"
      printf '%s\n' "$unit" | write_file "$unit_dst" 0644
      do_cmd systemctl daemon-reload
      if confirm "Enable and start athenaeum.service now?"; then
        do_cmd systemctl enable --now athenaeum.service
        push_undo "systemctl disable --now athenaeum.service 2>/dev/null || true"
      fi
      ;;
    openrc)
      local src="$DEPLOY_DIR/openrc/athenaeum"
      local dst="/etc/init.d/athenaeum"
      if ! have_root; then warn "not root: openrc script at $src"; return 0; fi
      local body
      body="$(sed "s|/usr/local/bin/athenaeum|${PREFIX}/bin/athenaeum|g" "$src")"
      printf '%s\n' "$body" | write_file "$dst" 0755
      if confirm "Add athenaeum to default runlevel and start?"; then
        do_cmd rc-update add athenaeum default
        do_cmd rc-service athenaeum start
        push_undo "rc-service athenaeum stop 2>/dev/null || true; rc-update del athenaeum default 2>/dev/null || true"
      fi
      ;;
    runit)
      local src="$DEPLOY_DIR/runit/athenaeum"
      local dst="/etc/sv/athenaeum"
      if ! have_root; then warn "not root: runit service at $src"; return 0; fi
      do_cmd mkdir -p /etc/sv
      if [[ $DRY_RUN -eq 0 ]]; then
        rm -rf "$dst"
        cp -a "$src" "$dst"
        sed -i "s|/usr/local/bin/athenaeum|${PREFIX}/bin/athenaeum|g" "$dst/run"
        chmod +x "$dst/run" "$dst/finish" "$dst/log/run" 2>/dev/null || true
        push_undo "rm -rf '$dst'"
      fi
      if confirm "Enable runit service (link into /var/service)?"; then
        do_cmd mkdir -p /var/service
        do_cmd ln -sfn "$dst" /var/service/athenaeum
        push_undo "rm -f /var/service/athenaeum"
      fi
      ;;
    dinit)
      local src="$DEPLOY_DIR/dinit/athenaeum"
      local dst="/etc/dinit.d/athenaeum"
      if ! have_root; then warn "not root: dinit service at $src"; return 0; fi
      local body
      body="$(sed "s|/usr/local/bin/athenaeum|${PREFIX}/bin/athenaeum|g" "$src")"
      printf '%s\n' "$body" | write_file "$dst" 0644
      if confirm "Enable and start dinit service?"; then
        do_cmd dinitctl enable athenaeum || true
        do_cmd dinitctl start athenaeum || true
        push_undo "dinitctl stop athenaeum 2>/dev/null || true; dinitctl disable athenaeum 2>/dev/null || true"
      fi
      ;;
    s6)
      local src="$DEPLOY_DIR/s6/athenaeum"
      local dst="/etc/s6/athenaeum"
      if ! have_root; then warn "not root: s6 service at $src"; return 0; fi
      do_cmd mkdir -p /etc/s6
      if [[ $DRY_RUN -eq 0 ]]; then
        rm -rf "$dst"
        cp -a "$src" "$dst"
        sed -i "s|/usr/local/bin/athenaeum|${PREFIX}/bin/athenaeum|g" "$dst/run"
        chmod +x "$dst/run" "$dst/finish" 2>/dev/null || true
        push_undo "rm -rf '$dst'"
      fi
      if [[ -d /service ]] && confirm "Link into /service for s6-svscan?"; then
        do_cmd ln -sfn "$dst" /service/athenaeum
        push_undo "rm -f /service/athenaeum"
      else
        info "service files installed under $dst (link into your scan directory)"
      fi
      ;;
    *)
      die "unknown service kind: $SERVICE_KIND"
      ;;
  esac
  ok "service configuration for $SERVICE_KIND ready"
}

print_next_steps() {
  local addr_url
  if [[ "$LISTEN_IP" == "0.0.0.0" || "$LISTEN_IP" == "127.0.0.1" ]]; then
    addr_url="http://127.0.0.1:${LISTEN_PORT}"
  else
    addr_url="http://${LISTEN_IP}:${LISTEN_PORT}"
  fi

  printf '\n%s\n' "$(bold "Next steps")"
  case "$METHOD" in
    docker)
      printf '  %s\n' "Open ${addr_url}"
      printf '  %s\n' "Logs: docker compose -f ${COMPOSE_FILE} logs -f"
      printf '  %s\n' "Backup: ./scripts/backup.sh --docker -o ./backups/data.tar.gz"
      ;;
    *)
      printf '  %s\n' "Open ${addr_url}"
      if [[ "$SERVICE_KIND" == "none" ]]; then
        local addr="${LISTEN_IP}:${LISTEN_PORT}"
        [[ "$LISTEN_IP" == "0.0.0.0" ]] && addr=":${LISTEN_PORT}"
        printf '  %s\n' "Run: ${PREFIX}/bin/athenaeum --addr ${addr} --library ${LIBRARY_DIR} --data ${DATA_DIR}"
      fi
      printf '  %s\n' "Backup: ./scripts/backup.sh -d ${DATA_DIR} -o ./backups/data.tar.gz"
      printf '  %s\n' "Restore: ./scripts/restore.sh --from ./backups/data.tar.gz -d ${DATA_DIR}"
      ;;
  esac
  printf '  %s\n' "Service units live under deploy/{systemd,openrc,runit,dinit,s6}"
  printf '\n'
}

# --- main -------------------------------------------------------------------

main() {
  configure
  create_system_user
  ensure_dirs

  case "$METHOD" in
    binary) install_binary_method ;;
    source) install_source_method ;;
    docker) install_docker_method ;;
  esac

  if [[ "$METHOD" != "docker" && "$SERVICE_KIND" == "none" ]]; then
    local env_host="/etc/athenaeum/athenaeum.env"
    if ! have_root; then
      env_host="$ROOT/deploy/env/athenaeum.env"
    fi
    do_cmd mkdir -p "$(dirname "$env_host")"
    write_env_file "$env_host"
  fi

  install_service_units

  # Clear undo stack on success so EXIT does not roll back a good install.
  UNDO_STACK=()
  trap - ERR

  ok "installation complete"
  print_next_steps
}

main "$@"
