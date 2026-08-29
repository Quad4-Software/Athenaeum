---
sidebar_position: 1
title: Introduction
description: Athenaeum is a universal library tool with a self-hosted catalog and reader.
---

# Athenaeum

Athenaeum is a universal library tool. Today it ships as a self-hosted catalog
and reader for EPUB, PDF, comics, and audiobooks: one static binary or Docker
image, point it at a folder of media, and you get a web reader, search, shelves,
optional multi-user auth, and OPDS for e-reader apps.

The longer aim is one place for the media people actually keep: books, papers,
offline knowledge archives, feeds, and listening, runnable as a household server
or as a local desktop and mobile app. See the [roadmap](/roadmap).

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
