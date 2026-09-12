---
sidebar_position: 10
title: Project layout
description: Repository directory map for Athenaeum.
---

# Project layout

```
cmd/athenaeum/         Entry point
cmd/genapi/            OpenAPI / typed path generator
internal/
  altcha/              ALTCHA proof-of-work widget config and Sentinel client
  auth/                Password hashing, sessions, API keys, TOTP
  backup/              Admin backup/restore zip helpers
  brand/               Product identity constants for forks
  cli/                 Offline user management (`athenaeum users`)
  config/              Flags and environment
  demo/                Generated demo library seeder
  i18n/                Locale loading and custom language support
  invite/              Invite tokens and guest/permanent acceptance
  libfs/               Local and S3-capable library filesystem backends
  library/             Scanner, EPUB/PDF/audio/comic metadata, watcher
  logging/             slog setup (level, optional log file)
  models/              Shared domain types and permissions
  oidc/                OIDC / SSO client helpers
  opds/                OPDS Atom / OPDS 2 generators
  pocketid/            Pocket ID admin API connector
  pprofserve/          Loopback pprof helper
  sandbox/             Optional Linux Landlock + seccomp-bpf
  storage/             SQLite (default) or PostgreSQL
  server/              HTTP API, auth middleware, embedded SPA
  system/              Cross-platform host stats for admin UI
  telemetry/           Sentry/GlitchTip integration
  term/                TTY-aware CLI color helpers
  version/             Build-time version and web version constants
  assets/              go:embed frontend bundle (overridable via --web-dir)
web/                   Svelte 5 frontend (PWA service worker at build time)
docs/                  Docusaurus site (Markdown in docs/docs/)
deploy/                systemd / OpenRC / runit / dinit / s6 units + env example
docker/kokoro/         Optional Kokoro TTS sidecar image
scripts/               backup, restore, doctor, coverage, showcase helpers
showcase/              README / marketing screenshots (task showcase)
site/                  Offline demo SPA output (task build:demo, gitignored)
CHANGELOG.md           Release notes
CONTRIBUTING.md        Contributor guide
```

Compose files at the repo root: `docker-compose.yml` (with `altcha`,
`kokoro`, and `postgres` profiles) and `docker-compose.coolify.yml`.
