# Athenaeum

A fast, self-hosted EPUB, PDF and audiobook library server that ships as a
single static binary. Point it at a folder of media and it gives you a clean
web reader with search, collections, optional multi-user auth, OPDS, and
light/dark theming.

<p align="center">
  <img src="showcase/library-theme-split.png" alt="Library in dark and light themes" width="100%" />
</p>

## Quick start

Requires Go 1.26+, Node 20+, pnpm 11+, and [Task](https://taskfile.dev).

```sh
task setup
task build
./bin/athenaeum --addr :8080 --library /path/to/books --data ./data
```

Or use the interactive installer (binary, Docker, or source, plus optional
service units):

```sh
./install.sh
```

Open http://localhost:8080. On first visit create an admin in the setup wizard,
or pass --admin-user and --admin-pass on the CLI when no users exist yet.

Demo library (no own media needed):

```sh
task demo
```

## Documentation

Docs are a Docusaurus site in [docs/](docs/) (content in [docs/docs/](docs/docs/)):

```sh
task docs:dev      # http://localhost:3000
task docs:build    # docs/build (GitHub Pages artifact)
```

- [Features](docs/docs/features.md)
- [Getting started](docs/docs/getting-started.md)
- [Authentication](docs/docs/authentication.md)
- [CLI users](docs/docs/cli-users.md)
- [Library and readers](docs/docs/library.md)
- [OPDS and KOSync](docs/docs/catalogs.md)
- [Operations](docs/docs/operations.md)
- [Development](docs/docs/development.md)
- [Deploying](docs/docs/deploying.md)
- [Configuration](docs/docs/configuration.md)
- [HTTP API](docs/docs/http-api.md)
- [Project layout](docs/docs/project-layout.md)

GitHub Pages deploys the docs site (offline demo at `/demo`) via
`.github/workflows/pages.yml`. Contributing guide: [CONTRIBUTING.md](CONTRIBUTING.md).
Security policy: [SECURITY.md](SECURITY.md).

## Changelog

See [CHANGELOG.md](CHANGELOG.md).

## Security

See [SECURITY.md](SECURITY.md) for supported versions and how to report
vulnerabilities privately.

## License

[MIT](LICENSE)
