---
sidebar_position: 12
title: Library and readers
description: Formats, metadata, narration, sharing, SMTP/Kindle, and offline grants.
---

# Library and readers

## Supported formats

| Format | In-browser | Notes |
| ------ | ---------- | ----- |
| EPUB | Yes (epub.js) | Metadata, covers, bookmarks, highlights, narration |
| PDF | Yes (pdf.js) | Info dictionary and embedded covers |
| MOBI / AZW / AZW3 | Yes | Section navigation via `/api/books/{id}/mobi-sections` |
| KFX | Download | Served as a file download |
| CBZ / CBR | Yes | Dual-page, fit modes, RTL/manga, page API |
| MP3 / M4B / M4A / OGG / FLAC | Yes | HTML5 audio, chapters/tracks, range streaming |

Multi-file audiobook folders are merged automatically. Local library mounts
are watched for changes in addition to periodic auto-scan. S3 mounts use
manual rescan or the admin auto-scan interval (no filesystem watch).

## S3 library mounts

Under Settings -> Library, choose backend **S3** when adding a mount. Provide
endpoint, bucket, optional prefix, access key, secret key, region, path-style,
and TLS. Athenaeum talks to AWS S3, MinIO, Cloudflare R2, Garage, and other
MinIO-compatible APIs.

- Display path is `s3://bucket/prefix`
- Uploads, downloads, delete, convert, and scan go through the object store
- Covers stay under `ATHENAEUM_DATA/covers`
- `POST /api/libraries/test-s3` validates credentials before saving

Example create body:

```json
{
  "name": "Cloud",
  "backend": "s3",
  "s3": {
    "endpoint": "minio:9000",
    "region": "us-east-1",
    "bucket": "athenaeum",
    "prefix": "main/",
    "accessKey": "...",
    "secretKey": "...",
    "usePathStyle": true,
    "tls": false
  }
}
```

## Metadata

Admins and users with `edit_metadata` can edit titles, authors, series, and
covers from the book page or `PUT /api/books/{id}`. Covers support upload,
URL import, and delete.

External metadata search and apply:

- `POST /api/books/{id}/metadata/search`
- `POST /api/books/{id}/metadata/apply`
- Bulk match: `POST /api/library/metadata/match` plus status endpoint
- Providers: `GET /api/metadata/providers`

Optional format conversion: `POST /api/books/{id}/convert?target=...`.

## Tags, ratings, favorites

- Tags: create globally, attach per book, filter library with `?tag=`
- Ratings: 1-5 stars per user (`GET/PUT /api/books/{id}/rating`)
- Favorites: `GET/PUT /api/books/{id}/favorite` and `GET /api/favorites`

## Reader preferences

EPUB font, theme, spacing, and spread sync via
`GET/PUT /api/auth/reader-prefs`. The browser keeps a localStorage cache for
offline/first paint.

## Narration (TTS)

EPUB narration uses in-browser Kokoro (ONNX Runtime Web: WebGPU with WASM
fallback) when WebAssembly is available, or the browser SpeechSynthesis API.
An optional Kokoro sidecar remains available for advanced/server setups
(`docker compose --profile kokoro`) under
Settings -> Administration -> Narration (TTS).

APIs: `GET/PUT /api/admin/tts`, `POST /api/admin/tts/test`,
`GET /api/tts/status`, `GET /api/tts/voices`, `POST /api/tts/synthesize`
(proxied through Athenaeum). See [Deploying](./deploying).

## Sharing

Create time-limited or persistent public download links from the book page
(`POST /api/books/{id}/share`). Recipients use `/share/{token}/download`
without signing in. Revoke with `DELETE /api/books/{id}/share/{shareId}`.

## SMTP and Kindle

Configure outbound SMTP under Settings -> Administration -> SMTP
(`GET/PUT /api/admin/smtp`). Users set a delivery address under Profile
(`GET/PUT /api/auth/kindle-email`). Send a book with
`POST /api/books/{id}/send`.

## Offline grants

Server-side offline grants (`GET/POST/DELETE /api/offline`) complement the
PWA shell and per-track audiobook cache in the client. Grants authorize
keeping a book available offline for the signed-in user.
