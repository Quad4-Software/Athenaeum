---
sidebar_position: 2
title: Getting started
description: Build, run, and try Athenaeum demo modes.
---

# Getting started

## Requirements

- Go 1.26+
- Node 22+ and pnpm 11+
- [Task](https://taskfile.dev) (task)

## Build and run

```sh
task setup
task build
./bin/athenaeum --addr :8080 --library /path/to/books --data ./data
```

Supported library files: `.epub`, `.pdf`, `.mobi`, `.azw`, `.azw3` (in-browser),
`.kfx` (download only), `.cbz`, `.cbr`, `.mp3`, `.m4b`, `.m4a`, `.ogg`, `.flac`.
Multi-file audiobook folders are merged automatically. Local library mounts are
watched for filesystem changes in addition to periodic auto-scan (remote
`s3://` mounts are not watched).

Open http://localhost:8080. On first visit you can create an admin account in
the setup wizard, or use the CLI bootstrap described in
[Authentication](./authentication).

## Demo mode (generated library)

Seed a public-domain demo library (Project Gutenberg classics with Open Library
covers, shelves, and progress) without needing your own media:

```sh
task demo
```

Or:

```sh
./bin/athenaeum --demo --addr :8080 --library ./library --data ./data
```

Titles include Alice, Pride and Prejudice, Frankenstein, Sherlock Holmes,
Dracula, Moby-Dick, and more. Stub media files are written under library/demo/.
Real covers are fetched from Open Library when the network is available
(generated fallbacks otherwise). EPUB stubs are replaced with Project Gutenberg
downloads when reachable. Seeding is idempotent (catalog=pd-v2 marker).
Metadata search still works against these titles.

## Offline web demo (no Go backend)

The SPA can run as a static site with an in-browser mock API (for GitHub Pages
showcases):

```sh
task build:demo
cd site && python -m http.server 4173
```

Then open http://localhost:4173/?demo=1.

In GitHub: enable Pages -> GitHub Actions, then push to main. The workflow
`.github/workflows/pages.yml` builds the Docusaurus site and embeds the offline
demo under `/demo`. Locally: `task docs:build` or `task build:demo`.

Capture README screenshots with Playwright:

```sh
task showcase
```

That builds the demo SPA and writes PNGs to ./showcase.
