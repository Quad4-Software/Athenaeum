---
sidebar_position: 3
title: Features
description: Core library, readers, auth, OPDS, and operations features.
---

# Features

## Core

- Single binary. The Svelte frontend is embedded into the Go executable by
  default. Optional `--web-dir` / `ATHENAEUM_WEB_DIR` serves a built SPA from
  disk instead (no Node at runtime either way). No CDN, no Google Fonts.
- Massive libraries. Concurrent filesystem scanning, incremental indexing,
  SQLite FTS5 search (optional PostgreSQL tsvector), WAL mode, and paginated
  infinite-scroll browsing.
- Slow-network friendly. HTTP range streaming, lazy-loaded reader engines,
  immutable asset caching, and a PWA with offline shell support.
- EPUB, PDF and audiobooks. EPUB metadata and covers parsed with stdlib only.
  PDF info dictionary and embedded JPEG/PNG covers. Audiobook ID3/Vorbis/MP4
  tags with embedded art. EPUB and PDF in the browser (epub.js / pdf.js).
  Audiobooks via HTML5 audio with range streaming.
- Collections and shelves. Manual shelves, user-defined smart collections
  (format, author, recency filters), and auto shelves after each scan
  (Recently Added, By Author).
- Continue reading. Per-user progress with a dedicated sidebar filter for
  in-progress books.
- Multiple library mounts. Add local folder roots or MinIO-compatible S3
  buckets from Settings, switch between them in the sidebar, and reorder or
  hide sidebar sections. S3 credentials are stored per mount (secret key is
  write-only in the API).

## Auth and access

- Optional auth. First-run setup wizard in the browser, or bootstrap via CLI
  flag. Sessions, profiles, admin user management, configurable password
  strength policy, audit logging, and guest accounts with automatic expiry.
  Optional [ALTCHA](https://altcha.org/docs/) proof-of-work on login and setup.
- Guest accounts. Admins create temporary users with one-time passwords, bulk
  revoke, extend expiry, invite links, and an expiring-soon list under
  Settings -> Administration -> Users.
- Invites. Tokenized permanent or guest invites with copyable `/invite/{token}`
  links, optional SMTP email, and optional Pocket ID passkey provisioning.
- Webhooks. Admin-configured outbound HTTPS hooks for user, invite, upload, and
  scan events with HMAC signatures and a delivery log.
- Pocket ID connector. Admin API key integration for user provisioning and
  OIDC issuer apply (see [Authentication](./authentication)).
- Per-user library access. Admins can restrict non-admin users to specific
  library mounts.
- 2FA and SSO groups. TOTP authenticator support, optional self-registration,
  OIDC admin groups.

## Readers and library

- Reading features. Per-user EPUB bookmarks and highlights (API-synced) and
  reading-time stats. EPUB narration via in-browser Kokoro (WASM/WebGPU) or
  browser SpeechSynthesis, with an optional Kokoro TTS sidecar under
  Settings -> Administration -> Narration (TTS).
- Synced reader prefs. EPUB font, theme, spacing, and spread sync via
  `GET/PUT /api/auth/reader-prefs` (localStorage cache for first paint).
- MOBI / AZW / AZW3. In-browser reader with section navigation. KFX remains
  download-only.
- Comic reader. Dual-page spread, fit modes, RTL/manga direction, bookmarks.
- Multi-file audiobooks. Track-aware resume and per-track offline cache.
- Tags and ratings. User-defined tags on books, library filter by tag, and
  1-5 star ratings.
- Content search. Optional EPUB full-text index via admin content-index job.
- Resumable uploads. Chunked PATCH uploads into library mounts with duplicate
  detection via SHA-256.

## Catalogs and sharing

- OPDS 1.2 catalog at `/opds/` for e-reader apps (KOReader and others), including
  series, comics, and Kindle-oriented feeds, cover thumbnails, per-user library
  filtering, and reading progress in entry summaries.
- KOSync and OPDS 2. KOReader progress sync at `/kosync/` and OPDS 2 JSON catalog
  at `/opds/v2/`. See [OPDS and KOSync](./catalogs).
- Sharing and delivery. Public share download links, SMTP send-to-email/Kindle.
  See [Library and readers](./library).

## Operations and UI

- Operations. Prometheus /metrics (admin toggle plus optional Basic Auth),
  extended GET /api/health, backup/restore zip in the admin UI, JSON config
  export/import, and optional Sentry/GlitchTip error reporting.
- Installable PWA. Add to home screen on mobile/desktop. Update-safe service
  worker (network-first HTML, never caches API responses).
- Modern, themeable UI. Svelte 5 and Tailwind v4 design tokens, red/black
  theme, collapsible sidebars, series/author filters, mobile support.

## Tech stack

| Layer    | Technology |
| -------- | ---------- |
| Backend  | Go (net/http, go:embed), pure-Go SQLite (modernc), optional PostgreSQL (pgx), Sentry SDK |
| Search   | SQLite FTS5 or PostgreSQL tsvector/GIN (title, author, series, description, content) |
| Auth     | bcrypt sessions (golang.org/x/crypto) |
| Frontend | Svelte 5, TypeScript, Tailwind CSS v4, Vite, pnpm, PWA |
| Icons    | @lucide/svelte |
| Readers  | epubjs, pdfjs-dist, HTML5 audio |
| Errors   | Sentry / GlitchTip (@sentry/svelte, sentry-go) |
| Tooling  | Task, Vitest, Playwright, ESLint, Prettier, Lighthouse CI, gosec, GitHub Actions |
