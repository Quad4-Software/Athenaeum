#!/usr/bin/env bash
# Backup Athenaeum data (and optionally the library) to a compressed archive.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

DRY_RUN=0
INCLUDE_LIBRARY=0
DOCKER=0
DATA_DIR="${ATHENAEUM_DATA:-}"
LIBRARY_DIR="${ATHENAEUM_LIBRARY:-}"
OUT_FILE=""
COMPOSE_FILE=""
VOLUME_NAME=""
NO_COLOR=0
FORCE_COLOR=0

usage() {
  cat <<'EOF'
Usage: backup.sh [options]

  -d, --data DIR          Data directory (default: ATHENAEUM_DATA or ./data)
  -l, --library DIR       Library directory (default: ATHENAEUM_LIBRARY or ./library)
  -o, --out FILE          Output archive path (.tar.gz)
      --include-library   Also pack the library tree
      --docker            Backup Docker named volume instead of a host data dir
      --volume NAME       Docker volume name (default: <project>_athenaeum-data)
      --compose FILE      Compose file used to resolve the project name
      --dry-run           Print actions without writing
      --no-color          Disable ANSI colors
      --color             Force ANSI colors
  -h, --help              Show this help

Examples:
  ./scripts/backup.sh -o ./backups/athenaeum-$(date +%Y%m%d).tar.gz
  ./scripts/backup.sh --include-library -o /tmp/full.tar.gz
  ./scripts/backup.sh --docker -o ./backups/docker-data.tar.gz
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

while [[ $# -gt 0 ]]; do
  case "$1" in
    -d|--data) DATA_DIR="$2"; shift 2 ;;
    -l|--library) LIBRARY_DIR="$2"; shift 2 ;;
    -o|--out) OUT_FILE="$2"; shift 2 ;;
    --include-library) INCLUDE_LIBRARY=1; shift ;;
    --docker) DOCKER=1; shift ;;
    --volume) VOLUME_NAME="$2"; shift 2 ;;
    --compose) COMPOSE_FILE="$2"; shift 2 ;;
    --dry-run) DRY_RUN=1; shift ;;
    --no-color) NO_COLOR=1; shift ;;
    --color) FORCE_COLOR=1; shift ;;
    -h|--help) usage; exit 0 ;;
    *) die "unknown option: $1" ;;
  esac
done

if [[ -n "${NO_COLOR:-}" ]]; then NO_COLOR=1; fi

DATA_DIR="${DATA_DIR:-$ROOT/data}"
LIBRARY_DIR="${LIBRARY_DIR:-$ROOT/library}"
mkdir -p "$(dirname "${OUT_FILE:-$ROOT/backups/x}")" 2>/dev/null || true

if [[ -z "$OUT_FILE" ]]; then
  stamp="$(date -u +%Y%m%dT%H%M%SZ)"
  OUT_FILE="$ROOT/backups/athenaeum-${stamp}.tar.gz"
fi

if [[ $DOCKER -eq 1 ]]; then
  command -v docker >/dev/null 2>&1 || die "docker not found"
  if [[ -z "$VOLUME_NAME" ]]; then
    project="$(basename "$ROOT")"
    if [[ -n "$COMPOSE_FILE" ]] && command -v docker >/dev/null 2>&1; then
      project="$(docker compose ${COMPOSE_FILE:+-f "$COMPOSE_FILE"} config --format json 2>/dev/null \
        | sed -n 's/.*"name"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' | head -n1 || true)"
      project="${project:-$(basename "$ROOT")}"
    fi
    VOLUME_NAME="${COMPOSE_PROJECT_NAME:-$project}_athenaeum-data"
  fi
  info "backing up Docker volume: $VOLUME_NAME"
  out_abs="$(cd "$(dirname "$OUT_FILE")" && pwd)/$(basename "$OUT_FILE")"
  out_dir="$(dirname "$out_abs")"
  out_base="$(basename "$out_abs")"
  if [[ $DRY_RUN -eq 1 ]]; then
    info "dry-run: docker run --rm -v ${VOLUME_NAME}:/data:ro -v ${out_dir}:/backup alpine tar czf /backup/${out_base} -C /data ."
    ok "would write $out_abs"
    exit 0
  fi
  docker volume inspect "$VOLUME_NAME" >/dev/null 2>&1 || die "volume not found: $VOLUME_NAME"
  docker run --rm \
    -v "${VOLUME_NAME}:/data:ro" \
    -v "${out_dir}:/backup" \
    alpine \
    tar czf "/backup/${out_base}" -C /data .
  ok "wrote $out_abs"
  exit 0
fi

[[ -d "$DATA_DIR" ]] || die "data directory missing: $DATA_DIR"

tmp="$(mktemp -d "${TMPDIR:-/tmp}/athenaeum-backup.XXXXXX")"
cleanup() { rm -rf "$tmp"; }
trap cleanup EXIT

info "staging data from $DATA_DIR"
run mkdir -p "$tmp/data"
if [[ $DRY_RUN -eq 0 ]]; then
  db=""
  for candidate in "$DATA_DIR/athenaeum.db" "$DATA_DIR/reader.db"; do
    if [[ -f "$candidate" ]]; then
      db="$candidate"
      break
    fi
  done
  if [[ -n "$db" ]] && command -v sqlite3 >/dev/null 2>&1; then
    local_db_name="$(basename "$db")"
    info "snapshotting SQLite ($local_db_name) with sqlite3 .backup"
    sqlite3 "$db" ".backup '$tmp/data/${local_db_name}'"
    if command -v rsync >/dev/null 2>&1; then
      rsync -a \
        --exclude 'athenaeum.db' --exclude 'athenaeum.db-wal' --exclude 'athenaeum.db-shm' \
        --exclude 'reader.db' --exclude 'reader.db-wal' --exclude 'reader.db-shm' \
        "$DATA_DIR"/ "$tmp/data/"
    else
      cp -a "$DATA_DIR"/. "$tmp/data/"
      rm -f "$tmp/data/athenaeum.db-wal" "$tmp/data/athenaeum.db-shm" \
        "$tmp/data/reader.db-wal" "$tmp/data/reader.db-shm"
      sqlite3 "$db" ".backup '$tmp/data/${local_db_name}'"
    fi
  else
    if [[ -n "$db" ]]; then
      warn "sqlite3 not found; copying data directory as-is (stop the server for a consistent backup)"
    fi
    cp -a "$DATA_DIR"/. "$tmp/data/"
  fi
  {
    printf 'created=%s\n' "$(date -u +%Y-%m-%dT%H:%M:%SZ)"
    printf 'data=%s\n' "$DATA_DIR"
    if [[ $INCLUDE_LIBRARY -eq 1 ]]; then
      printf 'library=%s\n' "$LIBRARY_DIR"
    fi
  } >"$tmp/BACKUP_META.txt"
fi

if [[ $INCLUDE_LIBRARY -eq 1 ]]; then
  [[ -d "$LIBRARY_DIR" ]] || die "library directory missing: $LIBRARY_DIR"
  info "staging library from $LIBRARY_DIR"
  run mkdir -p "$tmp/library"
  if [[ $DRY_RUN -eq 0 ]]; then
    cp -a "$LIBRARY_DIR"/. "$tmp/library/"
  fi
fi

info "writing $OUT_FILE"
run mkdir -p "$(dirname "$OUT_FILE")"
if [[ $DRY_RUN -eq 1 ]]; then
  ok "would write $OUT_FILE"
  exit 0
fi
tar czf "$OUT_FILE" -C "$tmp" .
ok "wrote $OUT_FILE ($(du -h "$OUT_FILE" | awk '{print $1}'))"
