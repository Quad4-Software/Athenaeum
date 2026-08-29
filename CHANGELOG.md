# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- Invites: tokenized permanent or guest invites with copyable `/invite/{token}`
  links, optional SMTP email, and optional Pocket ID passkey provisioning.
- Webhooks: admin outbound HTTPS subscriptions for `user.create`,
  `user.delete`, `invite.created`, `invite.accepted`, `book.upload`, and
  `library.scan.complete`, with HMAC signatures and a delivery log.
- Pocket ID Admin API connector (base URL + API key), OIDC apply helper, and
  signup-token proxies.
- S3 (MinIO-compatible) library mounts: per-library credentials, scan/stream/
  upload/delete/convert through object storage, Settings UI, and
  `POST /api/libraries/test-s3`.
- Optional PostgreSQL backend (`ATHENAEUM_DATABASE_DRIVER=postgres` plus
  `ATHENAEUM_DATABASE_URL`) with tsvector/GIN full-text search equivalent to
  SQLite FTS5. SQLite remains the default. Compose profile: `postgres`.
- Docusaurus documentation site under `docs/` with an Athenaeum-themed homepage,
  showcase gallery, expanded operator guides (auth, CLI users, library, OPDS),
  full HTTP route reference, and GitHub Pages workflow (`pages.yml`) that also
  serves the offline demo at `/demo`.
- EPUB narrator: in-browser Kokoro via `kokoro-js` (WebGPU with WASM/q8
  fallback) without admin sidecar config, plus browser SpeechSynthesis.
  Optional Kokoro TTS sidecar (`docker compose --profile kokoro`, Admin ->
  Narration settings) remains for advanced/server proxy use.
- Single `docker-compose.yml` with profiles `altcha`, `kokoro`, and `postgres`
  (Coolify keeps `docker-compose.coolify.yml`).
- Developer experience toolkit: Air live reload, golangci-lint, govulncheck,
  pnpm audit, coverage gates, lefthook, VS Code launch configs, Dev Containers,
  `task doctor` / `reset:*` / `profile` / `generate` / `i18n:sync` / `changelog`,
  knip/deadcode, CONTRIBUTING plus issue/PR templates, release SBOM, and optional
  loopback `--pprof` / `ATHENAEUM_PPROF`.

## [0.1.0] - 2026-08-30

### Added

- Initial release: single-binary EPUB/PDF/audiobook library server with embedded Svelte UI.
- SQLite FTS5 search, collections, reading progress, optional multi-user auth, OPDS 1.2,
  resumable uploads, guest accounts, Prometheus metrics, and backup/restore.
- Installable PWA with offline-cached UI assets and optional Sentry/GlitchTip error reporting.
- CLI (`athenaeum`) with colored help/status, structured logging, and optional Linux sandbox
  (Landlock + seccomp).
- Cross-platform host stats for the admin UI, metadata cleanup, and shared outbound HTTP pooling.
- `--demo` / `ATHENAEUM_DEMO` seeds a public-domain demo library (Gutenberg classics,
  Open Library covers, shelves). Offline web demo (`VITE_DEMO=1` / `task build:demo`) runs
  the SPA without a backend for GitHub Pages. `task showcase` captures Playwright
  screenshots into `./showcase`.
- Optional [ALTCHA](https://altcha.org/docs/) proof-of-work protection on login/setup
  (builtin challenge endpoint or ALTCHA Sentinel), with widget customization and
  Compose `--profile altcha` for the Sentinel service.
- Lighthouse CI (`@lhci/cli`) with score budgets for frontend performance.
- GitHub Actions CI/CD (lint, matrix Go tests/builds, race, fuzz, coverage,
  optional mutation, web tests, gosec, CodeQL, Docker, Lighthouse, Playwright e2e,
  tagged releases to GHCR) with SHA-pinned actions, least-privilege permissions,
  concurrency, and `workflow_dispatch`.
- Property tests (`testing/quick` / `fast-check`), native Go fuzz targets, and
  mutation testing via Gremlins (Go) and Stryker (web utils).
- Contract/drift checks for OpenAPI vs registered routes, frontend API paths,
  locale key parity, and `.env.example` coverage of `ATHENAEUM_*` config keys.
- Expanded Playwright UI coverage for login, theme toggle, error pages, and
  public OpenAPI/health contracts.
- Hardened HTML sanitizer so broken markup cannot leave nested `script` tags.
- Tags and per-user star ratings, synced EPUB reader preferences, comic dual-page/RTL/fit modes,
  MOBI/comic bookmarks.
- Multi-file audiobook track resume and per-track offline cache.
- TOTP 2FA, optional public registration, OIDC group-to-admin mapping.
- Share download links, SMTP send-to-Kindle/email, KOReader KOSync endpoints,
  OPDS 2 catalog (`/opds/v2/`), EPUB full-text content indexing (`POST /api/admin/content-index`),
  and offline book grant API.
- `--web-dir` / `ATHENAEUM_WEB_DIR` serves the SPA from a filesystem directory
  instead of the assets embedded in the binary.
