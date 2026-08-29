# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.0] - 2026-08-30

### Added

- Single-binary EPUB/PDF/audiobook/comic library server with embedded Svelte UI and installable PWA
- SQLite FTS5 search (optional PostgreSQL), collections, tags, star ratings, reading progress
- Optional multi-user auth: TOTP 2FA, OIDC, public registration, guest accounts, invites
- OPDS 1.2 and 2.0, KOReader KOSync, share links, SMTP send-to-Kindle/email
- Resumable uploads, webhooks, S3/MinIO library mounts, backup/restore
- EPUB narrator (in-browser Kokoro + SpeechSynthesis), optional Kokoro TTS sidecar
- Pocket ID Admin API connector and OIDC helpers
- ALTCHA proof-of-work on login/setup, Prometheus metrics, Sentry/GlitchTip reporting
- CLI (`athenaeum`) with colored help, structured logging, optional Linux sandbox
- `--demo` / offline GitHub Pages demo, Docusaurus docs site, Playwright showcase
- Docker Compose profiles (`altcha`, `kokoro`, `postgres`), Coolify compose file
- CI/CD (lint, tests, fuzz, coverage, Docker, Lighthouse, e2e, GHCR releases)
- Developer toolkit: Air, golangci-lint, lefthook, Taskfile, Dev Containers, VS Code configs
