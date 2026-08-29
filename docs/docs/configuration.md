---
sidebar_position: 5
title: Configuration
description: CLI flags and ATHENAEUM_* environment variables.
---

# Configuration

Flags override environment variables. Shell exports override values loaded from
`.env`. Optional alternate dotenv path: `ATHENAEUM_ENV_FILE`.

| Flag | Env | Default | Description |
| ---- | --- | ------- | ----------- |
| --addr | ATHENAEUM_ADDR | :8080 | HTTP listen address |
| --library | ATHENAEUM_LIBRARY | ./library | Root directory scanned for books |
| --data | ATHENAEUM_DATA | ./data | Database and cover cache location |
| --database-driver | ATHENAEUM_DATABASE_DRIVER | sqlite | Database backend: sqlite or postgres |
| --database-url | ATHENAEUM_DATABASE_URL | | PostgreSQL connection URL (required for postgres) |
| --web-dir | ATHENAEUM_WEB_DIR | | Serve frontend from this directory instead of embedded assets |
| --admin-user | ATHENAEUM_ADMIN_USER | | Bootstrap admin username |
| --admin-pass | ATHENAEUM_ADMIN_PASS | | Bootstrap admin password (min 8) |
| --upload-max-bytes | ATHENAEUM_UPLOAD_MAX_BYTES | 2147483648 | Max upload size in bytes (2 GB) |
| --scan-workers | ATHENAEUM_SCAN_WORKERS | 2 | Parallel library index workers |
| --log-level | ATHENAEUM_LOG_LEVEL | info | Log level: debug, info, warn, error |
| --log-file | ATHENAEUM_LOG_FILE | | Also append logs to this file |
| --debug | ATHENAEUM_DEBUG | false | Shortcut for --log-level=debug |
| --pprof | ATHENAEUM_PPROF | | Loopback pprof listen address (e.g. 127.0.0.1:6060) |
| --color | ATHENAEUM_COLOR | auto | CLI color: auto, always, never |
| --no-color | ATHENAEUM_NO_COLOR | false | Disable ANSI color (NO_COLOR also works) |
| --sandbox | ATHENAEUM_SANDBOX | off | Linux Landlock/seccomp: off, try, strict |
| --sandbox-landlock | ATHENAEUM_SANDBOX_LANDLOCK | true | Toggle Landlock when sandbox is on |
| --sandbox-seccomp | ATHENAEUM_SANDBOX_SECCOMP | true | Toggle seccomp-bpf when sandbox is on |
| --demo | ATHENAEUM_DEMO | false | Seed generated demo books, audiobooks, and covers |
| --sentry-dsn | ATHENAEUM_SENTRY_DSN | | Sentry/GlitchTip DSN (server) |
| --sentry-dsn-public | ATHENAEUM_SENTRY_DSN_PUBLIC | | Browser DSN (defaults to server DSN) |
| --sentry-environment | ATHENAEUM_SENTRY_ENVIRONMENT | | Sentry environment tag |
| --sentry-release | ATHENAEUM_SENTRY_RELEASE | app version | Sentry release name |
| --sentry-traces-sample-rate | ATHENAEUM_SENTRY_TRACES_SAMPLE_RATE | 0 | Performance trace sample rate (0-1) |
| --altcha | ATHENAEUM_ALTCHA_ENABLED | false | Require ALTCHA PoW on protected auth forms |
| --altcha-mode | ATHENAEUM_ALTCHA_MODE | builtin | builtin or sentinel |
| --altcha-hmac-secret | ATHENAEUM_ALTCHA_HMAC_SECRET | auto | Builtin challenge HMAC (persisted under data/) |
| --altcha-hmac-key-secret | ATHENAEUM_ALTCHA_HMAC_KEY_SECRET | | Optional secondary HMAC key secret |
| --altcha-challenge-url | ATHENAEUM_ALTCHA_CHALLENGE_URL | | Widget challenge URL override |
| --altcha-sentinel-url | ATHENAEUM_ALTCHA_SENTINEL_URL | | Sentinel base URL |
| --altcha-verify-url | ATHENAEUM_ALTCHA_VERIFY_URL | | Sentinel verify URL |
| --altcha-api-key-secret | ATHENAEUM_ALTCHA_API_KEY_SECRET | | Sentinel API key secret |
| --altcha-cost | ATHENAEUM_ALTCHA_COST | 5000 | Builtin PoW cost |
| --altcha-expires | ATHENAEUM_ALTCHA_EXPIRES | 300 | Challenge expiry seconds |
| --altcha-protect | ATHENAEUM_ALTCHA_PROTECT | login,setup | Comma-separated forms to protect |
| --altcha-widget-* | ATHENAEUM_ALTCHA_WIDGET_* | | Theme, display, type, auto, language, name, workers, hide logo/footer |
| --password-min-length | ATHENAEUM_PASSWORD_MIN_LENGTH | 8 | Minimum password length |
| --password-long-length | ATHENAEUM_PASSWORD_LONG_LENGTH | 12 | Length that satisfies diversity without min-kinds (0 disables) |
| --password-min-kinds | ATHENAEUM_PASSWORD_MIN_KINDS | 3 | Minimum character classes (0 disables diversity rule) |
| --password-require-lower | ATHENAEUM_PASSWORD_REQUIRE_LOWER | false | Require a lowercase letter |
| --password-require-upper | ATHENAEUM_PASSWORD_REQUIRE_UPPER | false | Require an uppercase letter |
| --password-require-digit | ATHENAEUM_PASSWORD_REQUIRE_DIGIT | false | Require a digit |
| --password-require-symbol | ATHENAEUM_PASSWORD_REQUIRE_SYMBOL | false | Require a symbol |

CLI-only (see [CLI users](./cli-users)):

| Env | Description |
| --- | ----------- |
| ATHENAEUM_PASSWORD | Password for `athenaeum users` when `--password` is omitted |
| ATHENAEUM_ENV_FILE | Alternate dotenv path loaded at startup |

Upload parts are stored under `{data}/uploads/` until complete.
Short-lived S3 scan downloads use `{data}/tmp/`.

Library S3 backends are configured per mount in Settings (or via
`POST /api/libraries`), not through environment variables. See
[Library and readers](./library#s3-library-mounts).

## Database

SQLite is the default and recommended for single-node installs. The database
file lives at `{data}/athenaeum.db` (or legacy `reader.db`).

PostgreSQL is optional for operators who want an external database:

```sh
ATHENAEUM_DATABASE_DRIVER=postgres
ATHENAEUM_DATABASE_URL=postgres://user:pass@host:5432/athenaeum?sslmode=require
```

Full-text search uses SQLite FTS5 on the default backend, and PostgreSQL
`tsvector` with GIN indexes (prefix queries) when using postgres. Schema is
created automatically on first connect. Covers and uploads still use
`ATHENAEUM_DATA`.

## Sandbox (Linux)

On Linux, `--sandbox=try` applies Landlock V9 via
[go-landlock](https://pkg.go.dev/github.com/landlock-lsm/go-landlock/landlock)
in BestEffort mode (filesystem allowlist for data/library, optional `--web-dir`,
plus common system paths) and a seccomp-bpf denylist for dangerous syscalls.
Use `strict` to require Landlock V9 without BestEffort fallback.
