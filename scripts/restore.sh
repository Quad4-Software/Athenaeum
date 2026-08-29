#!/usr/bin/env bash
# Restore Athenaeum data (and optionally library) from a backup archive.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

DRY_RUN=0
FORCE=0
KEEP_SAFETY=1
DOCKER=0
DATA_DIR="${ATHENAEUM_DATA:-}"
LIBRARY_DIR="${ATHENAEUM_LIBRARY:-}"
FROM_FILE=""
VOLUME_NAME=""
COMPOSE_FILE=""
NO_COLOR=0
FORCE_COLOR=0

usage() {
  cat <<'EOF'
Usage: restore.sh --from ARCHIVE [options]

  -f, --from FILE         Backup archive produced by backup.sh
  -d, --data DIR          Target data directory (default: ATHENAEUM_DATA or ./data)
  -l, --library DIR       Target library directory (only if archive contains library/)
      --docker            Restore into a Docker named volume
      --volume NAME       Docker volume name (default: <project>_athenaeum-data)
      --compose FILE      Compose file used to resolve the project name
      --force             Replace existing data without prompting
      --no-safety-backup  Skip automatic pre-restore snapshot of current data
      --dry-run           Print actions without writing
      --no-color          Disable ANSI colors
      --color             Force ANSI colors
  -h, --help              Show this help

Stop Athenaeum before restoring. A restart is required after restore.
EOF
}

color_ok() {
  if [[ $NO_COLOR -eq 1 ]]; then return 1; fi
  if [[ $FORCE_COLOR -eq 1 ]]; then return 0; fi
  [[ -t 1 ]] && [[ -z "${NO_COLOR:-}" ]]
}

c() {
  local code="$1"; shift
  if color_ok; then printf '\033[%sm%s\033[0m' "$code" "$*"; else printf '%s' "$*"; fi
}
info()  { printf '%s %s\n' "$(c 36 info)" "$*"; }
ok()    { printf '%s %s\n' "$(c 32 ok)" "$*"; }
warn()  { printf '%s %s\n' "$(c 33 warn)" "$*" >&2; }
die()   { printf '%s %s\n' "$(c 31 error)" "$*" >&2; exit 1; }
run() {
  if [[ $DRY_RUN -eq 1 ]]; then
    info "dry-run: $*"
    return 0
  fi
  "$@"
}

confirm() {
  local prompt="$1"
  if [[ $FORCE -eq 1 ]]; then return 0; fi
  if [[ ! -t 0 ]]; then die "refusing non-interactive restore without --force"; fi
  local ans=""
  printf '%s [y/N] ' "$prompt"
  read -r ans
  [[ "$ans" == "y" || "$ans" == "Y" || "$ans" == "yes" ]]
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    -f|--from) FROM_FILE="$2"; shift 2 ;;
    -d|--data) DATA_DIR="$2"; shift 2 ;;
    -l|--library) LIBRARY_DIR="$2"; shift 2 ;;
    --docker) DOCKER=1; shift ;;
    --volume) VOLUME_NAME="$2"; shift 2 ;;
    --compose) COMPOSE_FILE="$2"; shift 2 ;;
    --force) FORCE=1; shift ;;
    --no-safety-backup) KEEP_SAFETY=0; shift ;;
    --dry-run) DRY_RUN=1; shift ;;
    --no-color) NO_COLOR=1; shift ;;
    --color) FORCE_COLOR=1; shift ;;
    -h|--help) usage; exit 0 ;;
    *) die "unknown option: $1" ;;
  esac
done

if [[ -n "${NO_COLOR:-}" ]]; then NO_COLOR=1; fi
[[ -n "$FROM_FILE" ]] || die "--from is required"
[[ -f "$FROM_FILE" ]] || die "archive not found: $FROM_FILE"

DATA_DIR="${DATA_DIR:-$ROOT/data}"
LIBRARY_DIR="${LIBRARY_DIR:-$ROOT/library}"
FROM_ABS="$(cd "$(dirname "$FROM_FILE")" && pwd)/$(basename "$FROM_FILE")"

tmp="$(mktemp -d "${TMPDIR:-/tmp}/athenaeum-restore.XXXXXX")"
cleanup() { rm -rf "$tmp"; }
trap cleanup EXIT

info "inspecting archive"
tar tzf "$FROM_ABS" >/dev/null 2>&1 || die "not a readable gzip tar archive"
tar xzf "$FROM_ABS" -C "$tmp"

has_data=0
has_library=0
[[ -d "$tmp/data" ]] && has_data=1
[[ -d "$tmp/library" ]] && has_library=1

if [[ $has_data -eq 0 && $DOCKER -eq 0 ]]; then
  # Archives from docker volume backup are rooted at /
  if [[ -f "$tmp/athenaeum.db" || -f "$tmp/reader.db" || -d "$tmp/covers" || -d "$tmp/uploads" ]]; then
    mkdir -p "$tmp/data"
    # Move root contents into data/
    shopt -s dotglob nullglob
    for item in "$tmp"/*; do
      base="$(basename "$item")"
      [[ "$base" == "data" || "$base" == "library" || "$base" == "BACKUP_META.txt" ]] && continue
      mv "$item" "$tmp/data/"
    done
    shopt -u dotglob nullglob
    has_data=1
  fi
fi

[[ $has_data -eq 1 || $DOCKER -eq 1 ]] || die "archive has no data/ payload"

if [[ $DOCKER -eq 1 ]]; then
  command -v docker >/dev/null 2>&1 || die "docker not found"
  if [[ -z "$VOLUME_NAME" ]]; then
    project="$(basename "$ROOT")"
    if [[ -n "$COMPOSE_FILE" ]] && command -v docker >/dev/null 2>&1; then
      project="$(docker compose -f "$COMPOSE_FILE" config --format json 2>/dev/null \
        | sed -n 's/.*"name"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' | head -n1 || true)"
      project="${project:-$(basename "$ROOT")}"
    fi
    VOLUME_NAME="${COMPOSE_PROJECT_NAME:-$project}_athenaeum-data"
  fi
  confirm "Restore into Docker volume ${VOLUME_NAME}? This replaces volume contents." || die "aborted"
  if [[ $DRY_RUN -eq 1 ]]; then
    info "dry-run: extract archive into volume $VOLUME_NAME"
    exit 0
  fi
  docker volume inspect "$VOLUME_NAME" >/dev/null 2>&1 || docker volume create "$VOLUME_NAME" >/dev/null
  src="$tmp/data"
  if [[ ! -d "$src" ]]; then src="$tmp"; fi
  docker run --rm \
    -v "${VOLUME_NAME}:/data" \
    -v "${src}:/restore:ro" \
    alpine \
    sh -c 'rm -rf /data/* /data/.[!.]* /data/..?* 2>/dev/null; cp -a /restore/. /data/'
  ok "restored into volume $VOLUME_NAME"
  warn "restart the container to pick up the restored database"
  exit 0
fi

confirm "Restore into ${DATA_DIR}? Existing files will be replaced." || die "aborted"

if [[ $KEEP_SAFETY -eq 1 && -d "$DATA_DIR" && -n "$(ls -A "$DATA_DIR" 2>/dev/null || true)" ]]; then
  safety="$ROOT/backups/pre-restore-$(date -u +%Y%m%dT%H%M%SZ).tar.gz"
  info "writing safety backup to $safety"
  if [[ $DRY_RUN -eq 0 ]]; then
    mkdir -p "$(dirname "$safety")"
    tar czf "$safety" -C "$DATA_DIR" .
  else
    info "dry-run: tar czf $safety -C $DATA_DIR ."
  fi
fi

info "restoring data into $DATA_DIR"
run mkdir -p "$DATA_DIR"
if [[ $DRY_RUN -eq 0 ]]; then
  # Clear existing contents carefully
  find "$DATA_DIR" -mindepth 1 -maxdepth 1 -exec rm -rf {} +
  cp -a "$tmp/data"/. "$DATA_DIR/"
fi

if [[ $has_library -eq 1 ]]; then
  info "restoring library into $LIBRARY_DIR"
  run mkdir -p "$LIBRARY_DIR"
  if [[ $DRY_RUN -eq 0 ]]; then
    find "$LIBRARY_DIR" -mindepth 1 -maxdepth 1 -exec rm -rf {} +
    cp -a "$tmp/library"/. "$LIBRARY_DIR/"
  fi
fi

ok "restore complete"
warn "start or restart Athenaeum to use the restored data"
if [[ -f "$tmp/BACKUP_META.txt" ]]; then
  info "backup metadata:"
  sed 's/^/  /' "$tmp/BACKUP_META.txt"
fi
