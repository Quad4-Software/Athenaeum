---
sidebar_position: 4
title: Deploying
description: Installer, Docker Compose, host services, and backups.
---

# Deploying

## Installer

`./install.sh` walks through binary download, Docker Compose, or a source
build. It can pick a listen IP from host interfaces, set library and data
paths, write `/etc/athenaeum/athenaeum.env`, and install a service unit.

```sh
./install.sh                  # interactive
./install.sh --dry-run        # print actions only
./install.sh --method docker --port 8080 --library /media/books -y
./install.sh --method binary --service systemd --ip 127.0.0.1
./install.sh --method source --prefix /usr/local --no-service
./install.sh --no-color
```

| Option | Description |
| ------ | ----------- |
| `--method` | `binary`, `docker`, or `source` (prompted if omitted) |
| `--ip` / `--port` | Listen address |
| `--library` / `--data` | Library root and data directory |
| `--prefix` | Binary install prefix (default `/usr/local`) |
| `--service` | `systemd`, `openrc`, `runit`, `dinit`, `s6`, or `none` |
| `--user` / `--no-user` | System user (default `athenaeum`) |
| `--no-service` | Skip service unit installation |
| `--version` / `--repo` / `--release-base` | Release asset selection |
| `--image` / `--compose` | Docker image and compose file |
| `--admin-user` / `--admin-pass` | Optional first-admin bootstrap |
| `--dry-run` / `-y` / `--no-color` | Behavior flags |

Failed installs roll back tracked filesystem and service changes. Use
`--method source` when no release assets are published yet.

Service unit templates live under `deploy/`:

| Init     | Path |
| -------- | ---- |
| systemd  | `deploy/systemd/athenaeum.service` |
| OpenRC   | `deploy/openrc/athenaeum` |
| runit    | `deploy/runit/athenaeum/` |
| dinit    | `deploy/dinit/athenaeum` |
| s6       | `deploy/s6/athenaeum/` |

Shared environment example: `deploy/env/athenaeum.env.example`.

## Docker

Published images are multi-arch on GHCR
(`ghcr.io/<owner>/athenaeum`): `linux/amd64` and `linux/arm64`. Release binaries
also cover `armv6`, `armv7`, and `riscv64` (and BSD) outside the container.

Copy .env.example to .env and set optional bootstrap credentials. Point
ATHENAEUM_LIBRARY_HOST_PATH at your media folder (default: ./library).

```sh
cp .env.example .env
docker compose up -d --build
```

Open http://localhost:8080 (or the host port from ATHENAEUM_PUBLISH_PORT). The
container always uses /data for the database and /library for the initial scan
mount. Host paths in ATHENAEUM_DATA / ATHENAEUM_LIBRARY apply to local runs only, not
Docker.

Additional library mounts (local paths inside the container, or MinIO-compatible
S3 backends) are configured in Settings after first login. S3 credentials are
stored in the database per mount, not as compose env vars.

| Variable | Default | Description |
| -------- | ------- | ----------- |
| ATHENAEUM_LIBRARY_HOST_PATH | ./library | Host folder bind-mounted to /library |
| ATHENAEUM_PUBLISH_PORT | 8080 | Host port published to container :8080 |
| ATHENAEUM_VERSION | 0.1.0 | Image build version label |
| ATHENAEUM_ADMIN_USER / ATHENAEUM_ADMIN_PASS | | Optional first-admin bootstrap via .env |
| ATHENAEUM_SENTRY_DSN | | Optional Sentry/GlitchTip DSN (passed through env_file) |
| ATHENAEUM_ALTCHA_ENABLED | false | Optional ALTCHA PoW on login/setup |

Optional sidecars use Compose profiles on the same file (combine freely, or set
`COMPOSE_PROFILES` in `.env`):

```sh
docker compose --profile altcha up -d --build
docker compose --profile kokoro up -d --build
docker compose --profile postgres up -d --build
```

- **altcha**: ALTCHA Sentinel backend (trial/license). Set Athenaeum to
  `ATHENAEUM_ALTCHA_MODE=sentinel` and point challenge/verify URLs at the
  Sentinel service. See ALTCHA docs.
- **kokoro**: Optional Kokoro TTS sidecar (EPUB narration defaults to
  in-browser Kokoro WASM/WebGPU; browser SpeechSynthesis also works). In
  Settings -> Administration -> Narration (TTS) set base URL to
  `http://kokoro:8880` (same Compose network) or `http://127.0.0.1:8880`
  from the host. First start downloads model weights and may take a few minutes.
- **postgres**: Postgres 16 instead of SQLite. Also set in `.env`:
  `ATHENAEUM_DATABASE_DRIVER=postgres` and
  `ATHENAEUM_DATABASE_URL=postgres://athenaeum:athenaeum@postgres:5432/athenaeum?sslmode=disable`.
  Back up with `pg_dump` (the admin ZIP backup skips the SQLite file and notes
  that an external dump is required).

Coolify deployments should use `docker-compose.coolify.yml` (same profiles).

Persistent state lives in the athenaeum-data Docker volume. Back it up with
`./scripts/backup.sh --docker` (see Backup below) or:

```sh
docker run --rm -v "${COMPOSE_PROJECT_NAME:-$(basename "$PWD")}_athenaeum-data":/data -v "$PWD":/backup alpine tar czf /backup/athenaeum-data.tgz -C /data .
```

Put a reverse proxy in front for TLS and forward X-Forwarded-For for audit
logs.

ALTCHA setup details are in [Authentication](./authentication).

## Host services

Prefer the units under `deploy/` plus `/etc/athenaeum/athenaeum.env`. Example
systemd unit (also installed by `./install.sh --service systemd`):

```ini
[Unit]
Description=Athenaeum media library server
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
User=athenaeum
Group=athenaeum
EnvironmentFile=-/etc/athenaeum/athenaeum.env
WorkingDirectory=/var/lib/athenaeum
ExecStart=/usr/local/bin/athenaeum
Restart=on-failure

[Install]
WantedBy=multi-user.target
```

Put a reverse proxy (Caddy, nginx, Traefik) in front for TLS. Forward
X-Forwarded-For so audit logs record client IPs correctly.

To serve a frontend build from disk instead of the binary-embedded assets, set
`ATHENAEUM_WEB_DIR` in the env file (or pass `--web-dir`). The directory must
contain the Vite production output (`index.html` and `assets/`).

## Backup

Back up the entire `--data` directory. With the default SQLite driver it
contains the database file and cached cover images. When using PostgreSQL,
also run `pg_dump` against ATHENAEUM_DATABASE_URL. Library files on disk are
not modified by Athenaeum.

```sh
./scripts/backup.sh -d /var/lib/athenaeum/data -o ./backups/data.tar.gz
./scripts/backup.sh --include-library -l /media/books -o ./backups/full.tar.gz
./scripts/backup.sh --docker -o ./backups/docker-data.tar.gz
./scripts/restore.sh --from ./backups/data.tar.gz -d /var/lib/athenaeum/data
```

`backup.sh` uses `sqlite3 .backup` when available for a consistent database
snapshot. Stop the service before restore, then restart it.
