---
sidebar_position: 2
title: Getting started
description: Build and run Athenaeum with Make, manual Go/pnpm, or Task.
---

# Getting started

## Requirements

- Go 1.26+
- Node 22+ and pnpm 11+

[Task](https://taskfile.dev) is optional. Make covers build and install. Raw
`go` / `pnpm` commands work too.

## Build and run

### Make

```sh
cd web && pnpm install && cd ..
make build
./bin/athenaeum --addr :8080 --library /path/to/books --data ./data
```

`make help` lists targets (`build`, `install`, `clean`, and so on).

### Manual

```sh
go mod download
cd web && pnpm install && pnpm build && cd ..
mkdir -p bin
go build -trimpath -o bin/athenaeum ./cmd/athenaeum
./bin/athenaeum --addr :8080 --library /path/to/books --data ./data
```

### Task (optional)

```sh
task setup
task build
./bin/athenaeum --addr :8080 --library /path/to/books --data ./data
```

### Installer

Interactive install (binary, Docker, or source, plus optional service units):

```sh
./install.sh
```

Supported library files: `.epub`, `.pdf`, `.mobi`, `.azw`, `.azw3` (in-browser),
`.kfx` (download only), `.cbz`, `.cbr`, `.mp3`, `.m4b`, `.m4a`, `.ogg`, `.flac`.
Multi-file audiobook folders are merged automatically. Local library mounts are
watched for filesystem changes in addition to periodic auto-scan (remote
`s3://` mounts are not watched).

Open http://localhost:8080. On first visit create an admin in the setup wizard,
or pass `--admin-user` and `--admin-pass` on the CLI when no users exist yet.
See [Authentication](./authentication).

Verify a build with `./bin/athenaeum --self-check` (dirs, database, HTTP).
Release tags publish single binaries for Linux (amd64/arm64/armv6/armv7/riscv64),
macOS, Windows, FreeBSD, OpenBSD, and NetBSD, plus a multi-arch image to GHCR.

## Next steps

- [Features](./features) - full capability list
- [Deploying](./deploying) - Docker, installer, and host services
- [Configuration](./configuration) - flags and environment variables
- [Development](./development) - contributor workflow and tests
