---
sidebar_position: 1
title: Introduction
description: Self-hosted library for EPUB, PDF, comics, and audiobooks.
---

# Athenaeum

Self-hosted library for EPUB, PDF, comics, and audiobooks. One binary or
Docker image. Point it at a media folder for a web reader, search, shelves,
optional multi-user auth, and OPDS. See the [roadmap](/roadmap) for apps and
more formats.

## Good fit if you want

- Your own library on a NAS, VPS, or home PC without a cloud subscription
- Browser reading for common ebook and audiobook formats
- EPUB narration with Kokoro TTS (in-browser, or an optional sidecar)
- KOReader-friendly OPDS / KOSync, share links, and optional send-to-Kindle
- One process to back up, with Docker and host service units when you need them

## Not required

You do not need Node at runtime. Auth, OIDC, Prometheus, S3 library mounts, and
the Kokoro sidecar are optional. SQLite is the default database.

## Next steps

1. [Getting started](./getting-started) - Docker, binary, installer, or source
2. [Features](./features) - what the product does
3. [Deploying](./deploying) - Compose, services, backups, sidecars
4. [Configuration](./configuration) - flags and environment variables

Release notes: [`CHANGELOG.md`](https://github.com/ivan/reader/blob/main/CHANGELOG.md)
