---
sidebar_position: 1
title: Introduction
description: Athenaeum is a self-hosted EPUB, PDF, and audiobook library server.
---

# Athenaeum

A fast, self-hosted EPUB, PDF, and audiobook library that ships as a **single
static binary**. Point it at a folder of media and you get a clean web reader
with search, collections, optional multi-user auth, OPDS, and light/dark
theming.

## Why Athenaeum

- **One process.** Go backend embeds the Svelte UI. No CDN, no Node at runtime.
- **Real formats.** EPUB, PDF, MOBI/AZW/AZW3 in the browser, KFX download,
  comics (CBZ/CBR), and audiobooks with multi-file folder merge.
- **Readers included.** Browser EPUB/PDF/MOBI engines, HTML5 audio with range
  streaming, comic dual-page/RTL modes, optional EPUB narration.
- **Ops-friendly.** Prometheus metrics, backup/restore, sandboxing, Docker and
  host service units, installable PWA.

## Next steps

1. [Getting started](./getting-started) - build, run, and try demo mode
2. [Features](./features) - full capability list
3. [Deploying](./deploying) - installer, Docker, and host services
4. [Authentication](./authentication) - users, TOTP, OIDC, API keys
5. [Configuration](./configuration) - flags and environment variables

Release notes live in [`CHANGELOG.md`](https://github.com/ivan/reader/blob/main/CHANGELOG.md)
at the repository root.
