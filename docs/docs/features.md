---
sidebar_position: 3
title: Features
description: Readers, narration, library tools, auth, OPDS, and operations.
---

# Features

## Read your files

- **Browser readers** for EPUB, PDF, MOBI/AZW/AZW3, and comics (CBZ/CBR). KFX
  is download-only.
- **Audiobooks** via HTML5 audio with range streaming. Multi-file folders merge
  into one book with track-aware resume.
- **Comic modes**: dual-page spread, fit options, RTL/manga direction, bookmarks.
- **Reading tools**: EPUB bookmarks and highlights, reading-time stats, synced
  EPUB font/theme/spacing prefs.
- **PWA**: install to home screen. Service worker stays update-safe and does not
  cache API responses.

## Narration with Kokoro

- Listen to EPUBs with **in-browser Kokoro TTS** (ONNX Runtime Web: WebGPU with
  WASM fallback) when WebAssembly is available. The full release binary embeds
  that runtime; the **slim** binary (`athenaeum-slim-*`) leaves it out.
- Fall back to the browser **SpeechSynthesis** voice if needed.
- Optional **Kokoro sidecar** for server-side TTS
  (`docker compose --profile kokoro`). Works with full and slim binaries.
  Configure under Settings -> Administration -> Narration (TTS).

Details: [Library and readers](./library) and [Deploying](./deploying).

## Library and search

- Concurrent scan, incremental indexing, SQLite FTS5 search (optional PostgreSQL
  tsvector). WAL mode and paginated browsing for large libraries.
- Manual shelves, smart collections, and auto shelves after scan (Recently
  Added, By Author).
- Continue reading, tags, 1-5 star ratings, favorites.
- Multiple library mounts: local folders or MinIO-compatible S3. Per-user mount
  access for non-admins.
- Optional EPUB full-text content index (admin job).
- Resumable chunked uploads with SHA-256 duplicate detection.
- Metadata edit, external metadata search/apply, optional format conversion.

## Catalogs and sharing

- OPDS 1.2 at `/opds/` and OPDS 2 at `/opds/v2/` for e-reader apps.
- KOSync progress sync at `/kosync/`.
- Public share download links.
- SMTP send-to-email / Kindle.

See [OPDS and KOSync](./catalogs) and [Library and readers](./library).

## Users and access

- Optional auth: first-run wizard or CLI/env bootstrap.
- Sessions, profiles, password policy, audit log, guest accounts with expiry.
- Invites (permanent or guest), optional SMTP and Pocket ID passkey provisioning.
- TOTP 2FA, optional self-registration, OIDC and admin groups.
- Optional ALTCHA proof-of-work on login and setup.
- Admin webhooks for user, invite, upload, and scan events (HMAC + delivery log).

See [Authentication](./authentication).

## Operations

- Single binary with embedded Svelte UI (optional `--web-dir` for a disk SPA).
  Releases also publish a slim binary without in-browser Kokoro WASM.
- Docker Compose, `./install.sh`, and host units (systemd, OpenRC, runit, dinit, s6).
- Prometheus `/metrics` (admin toggle, optional Basic Auth), extended health.
- Admin backup/restore zip, JSON config export/import.
- Optional Sentry / GlitchTip.
- Release binaries across Linux, macOS, Windows, and BSD. Multi-arch GHCR image.

See [Operations](./operations) and [Deploying](./deploying).

## Tech stack

| Layer | Technology |
| ----- | ---------- |
| Backend | Go (`net/http`, `go:embed`), SQLite (modernc), optional PostgreSQL (pgx) |
| Search | SQLite FTS5 or PostgreSQL tsvector/GIN |
| Auth | bcrypt sessions |
| Frontend | Svelte 5, TypeScript, Tailwind CSS v4, Vite, pnpm, PWA |
| Icons | `@lucide/svelte` |
| Readers | epub.js, pdf.js, HTML5 audio |
| Narration | Kokoro (in-browser or sidecar), SpeechSynthesis |
| Errors | Sentry / GlitchTip |
| Tooling | Task, Vitest, Playwright, ESLint, Prettier, golangci-lint, GitHub Actions |
