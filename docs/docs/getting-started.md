---
sidebar_position: 2
title: Getting started
description: Run Athenaeum with Docker, a release binary, the installer, or a source build.
---

# Getting started

Pick a run path below. Most self-hosters use Docker or the installer.
Building from source is for development.

## Installer

```sh
curl -fsSL https://raw.githubusercontent.com/Quad4-Software/Athenaeum/master/install.sh | bash
```

Or download and run it interactively:

```sh
curl -fsSL -o install.sh https://raw.githubusercontent.com/Quad4-Software/Athenaeum/master/install.sh
chmod +x install.sh
./install.sh
```

Non-interactive examples and flags are in [Deploying](./deploying).

## Docker (GHCR)

```sh
docker run -d --name athenaeum \
  -p 8080:8080 \
  -v athenaeum-data:/data \
  -v /path/to/books:/library \
  ghcr.io/quad4-software/athenaeum:latest
```

`latest` tracks `master`. Pin a release with a `v*` tag.

## Docker Compose

```sh
git clone https://github.com/Quad4-Software/Athenaeum.git
cd Athenaeum
cp .env.example .env
# set ATHENAEUM_LIBRARY_HOST_PATH to your media folder
docker compose up -d
```

Open http://localhost:8080 (or the port in `ATHENAEUM_PUBLISH_PORT`).

`docker compose up -d` pulls `ghcr.io/quad4-software/athenaeum:latest`.
Use `docker compose up -d --build` to build locally. Full Compose options,
profiles (Postgres, Kokoro, ALTCHA), and Coolify notes are in
[Deploying](./deploying).

## Release binary

Download a release for your OS/arch from GitHub Releases, then:

```sh
chmod +x athenaeum
./athenaeum --addr :8080 --library /path/to/books --data ./data
```

Linux (amd64/arm64/armv6/armv7/riscv64), macOS, Windows, FreeBSD, OpenBSD, and
NetBSD builds are published on release tags. Multi-arch images also go to GHCR.

## First login

Open http://localhost:8080. Create an admin in the setup wizard, or pass
`--admin-user` / `--admin-pass` (or the matching env vars) when no users exist
yet. See [Authentication](./authentication).

Check a running install with:

```sh
./athenaeum --self-check
```

(or the same flag on your installed binary path).

## What files it indexes

| Kind | Extensions |
| ---- | ---------- |
| Ebooks | `.epub`, `.pdf`, `.mobi`, `.azw`, `.azw3` (read in browser), `.kfx` (download) |
| Comics | `.cbz`, `.cbr` |
| Audio | `.mp3`, `.m4b`, `.m4a`, `.ogg`, `.flac` |

Multi-file audiobook folders merge into one book. Local mounts are watched for
changes (plus periodic auto-scan). Remote `s3://` mounts are not filesystem-watched.

## Build from source

Needs Go 1.26+, Node 22+, and pnpm 11+. [Task](https://taskfile.dev) is optional.

```sh
git clone https://github.com/Quad4-Software/Athenaeum.git
cd Athenaeum
```

### Make

```sh
cd web && pnpm install && cd ..
make build
./bin/athenaeum --addr :8080 --library /path/to/books --data ./data
```

### Manual

```sh
go mod download
cd web && pnpm install && pnpm build && cd ..
mkdir -p bin
go build -trimpath -o bin/athenaeum ./cmd/athenaeum
./bin/athenaeum --addr :8080 --library /path/to/books --data ./data
```

### Task

```sh
task setup
task build
./bin/athenaeum --addr :8080 --library /path/to/books --data ./data
```

Contributor workflow: [Development](./development).

## Next steps

- [Features](./features)
- [Deploying](./deploying)
- [Configuration](./configuration)
