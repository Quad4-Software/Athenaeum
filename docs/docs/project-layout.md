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
  auth/                Password hashing, sessions, OIDC helpers
  brand/               Product identity constants for forks
  cli/                 Offline user management (`athenaeum users`)
  config/              Flags and environment
  demo/                Generated demo library seeder
  libfs/               Local and S3-capable library filesystem backends
  library/             Scanner, EPUB/PDF/audio/comic metadata, watcher
  logging/             slog setup (level, optional log file)
  models/              Shared domain types and permissions
  opds/                OPDS Atom / OPDS 2 generators
  pprofserve/          Loopback pprof helper
  sandbox/             Optional Linux Landlock + seccomp-bpf
  storage/             SQLite (default) or PostgreSQL
  server/              HTTP API, auth middleware, embedded SPA
  system/              Cross-platform host stats for admin UI
  telemetry/           Sentry/GlitchTip integration
  term/                TTY-aware CLI color helpers
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

Compose overlays at the repo root: `docker-compose.yml`,
`docker-compose.prod.yml`, `docker-compose.kokoro.yml`,
`docker-compose.postgres.yml`, `docker-compose.coolify.yml`.
