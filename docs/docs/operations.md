---
sidebar_position: 7
title: Operations
description: Metrics, maintenance, i18n, PWA, profiling, server settings, and backup.
---

# Operations

## Metrics and health

- `GET /metrics`. Prometheus exposition format. Disabled by default. Enable
  under Settings -> Administration -> Server. Optional HTTP Basic Auth
  protects the endpoint when enabled.
- `GET /api/health`. Public JSON health probe: status, database, scanning,
  lastScan, diskFreeBytes, version, webVersion. When error reporting is
  configured, includes telemetry with the browser Sentry/GlitchTip DSN and
  environment. Suitable for Kubernetes liveness/readiness or systemd alongside
  `/metrics`.
- `GET /api/system/stats`. Admin host CPU, memory, disk, and version summary
  for the Administration UI.

## Maintenance tasks

Settings -> Administration -> Maintenance runs library jobs (also available as
APIs under `/api/admin/tasks/*` and `POST /api/admin/content-index`):

| Task | Effect |
| ---- | ------ |
| Verify | Integrity check of indexed files |
| Prune missing | Remove DB rows for files that no longer exist |
| Cleanup covers | Remove orphan cover cache files |
| Regenerate covers | Rebuild cover images from source files |
| Cleanup series | Normalize / clean series metadata |
| Cleanup text | Normalize text fields |
| Content index | Index EPUB full text for content search |

`GET /api/admin/tasks/status` reports in-flight job state.

## Custom locales (i18n)

Bundled UI locales ship with the SPA. Drop custom JSON message files under
`{data}/i18n/` to override or add languages. Public endpoints:

- `GET /api/i18n/locales`
- `GET /api/i18n/template`
- `GET /api/i18n/{locale}`

Custom locale files are included in the admin backup zip.

## Error tracking (Sentry / GlitchTip)

Optional error reporting for the Go server and the embedded SPA.
[GlitchTip](https://glitchtip.com/) and other Sentry-compatible backends use
the same DSN format.

Set in `.env` or pass flags at startup:

```env
ATHENAEUM_SENTRY_DSN=https://{key}@errors.example.com/{project-id}
ATHENAEUM_SENTRY_ENVIRONMENT=production
# ATHENAEUM_SENTRY_DSN_PUBLIC=   # browser DSN if different from server DSN
# ATHENAEUM_SENTRY_RELEASE=      # defaults to app version
# ATHENAEUM_SENTRY_TRACES_SAMPLE_RATE=0
```

The server reports panics and HTTP 5xx responses. The browser initializes from
`/api/health` and reports unhandled errors plus API failures. When CSP is
enabled, the ingest host is added to `connect-src` automatically. Leave
`ATHENAEUM_SENTRY_DSN` unset to disable entirely.

## Progressive Web App

After building the frontend into the binary (or pointing `--web-dir` at a Vite
production output directory), the app is installable (Add to Home Screen /
Install app in Chromium). A service worker precaches hashed assets, fetches
`index.html` with network-first semantics, and never intercepts `/api/*`. When
you deploy a new version, users see a Reload / Later banner instead of stale
cached UI.

## Local profiling (pprof)

For development, start with `--pprof 127.0.0.1:6060` or `ATHENAEUM_PPROF`.
Only loopback addresses are accepted. Then:

```sh
task profile                  # writes coverage/cpu.pprof
go tool pprof -http=:0 ./coverage/cpu.pprof
```

## Server settings

Admins configure reverse-proxy trust, CORS, CSP, library auto-scan interval,
and metrics via Settings -> Administration -> Server or
`GET/PUT /api/admin/server`:

| Field | Description |
| ----- | ----------- |
| metricsEnabled | Expose `/metrics` |
| metricsAuth | Require Basic Auth on `/metrics` |
| trustedProxies | Comma-separated IPs/CIDRs that may send X-Forwarded-* headers |
| corsEnabled / corsOrigins | CORS for cross-origin API access |
| cspEnabled / cspPolicy | Content-Security-Policy header (empty = default) |
| autoScanEnabled / autoScanIntervalSec | Background library rescan (min 60s) |

## Webhooks

Configure outbound event URLs under Settings -> Administration -> Webhooks
(`GET/POST /api/admin/webhooks`). Each delivery POSTs JSON:

```json
{ "id": "...", "event": "user.create", "createdAt": "...", "data": { } }
```

When a secret is set, Athenaeum sends
`X-Athenaeum-Signature: sha256=<hmac-hex>` of the raw body. Failed deliveries
retry up to three times. Inspect recent attempts with
`GET /api/admin/webhooks/{id}/deliveries`.

v1 events: `user.create`, `user.delete`, `invite.created`, `invite.accepted`,
`book.upload`, `library.scan.complete`.

## Backup and config export

Settings -> Administration -> Backup downloads a zip of cached covers, custom
locale files, and a `config.json` snapshot. With the default SQLite driver the
zip also includes `athenaeum.db`. With PostgreSQL it includes a `DATABASE.txt`
note instead (use `pg_dump` for the database). Restore uploads a zip into the
data directory (restart required). Config export/import
(`GET /api/admin/config/export`, `POST /api/admin/config/import`) covers server
settings, OIDC metadata (no secrets), and library mount names.

Also back up the entire `--data` directory on disk (SQLite file and/or cover
cache). When using PostgreSQL, run `pg_dump` as well. Library files on disk
are not modified by Athenaeum. CLI helpers: `./scripts/backup.sh` and
`./scripts/restore.sh` (see [Deploying](./deploying)).
