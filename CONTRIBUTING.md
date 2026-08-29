# Contributing to Athenaeum

## Prerequisites

- Go 1.26+
- Node 22+
- pnpm 11+
- [Task](https://taskfile.dev)

Optional but recommended: `air`, `golangci-lint`, `govulncheck`, `lefthook`, Chromium.

```sh
task tools:install   # air, golangci-lint, govulncheck, deadcode, lefthook, git-cliff
task setup           # go mod download + web/docs pnpm install
task hooks:install   # git hooks (fmt/lint/contract on commit/push)
task doctor          # verify toolchain and ports
```

## Day-to-day

```sh
task dev             # Vite :5173 + Go with live reload (air) on :8080
task demo            # seeded public-domain style demo library
task lint            # gofmt check, golangci-lint, eslint, prettier, svelte-check
task test            # Go + Vitest
task test:contract   # OpenAPI / route / i18n / env drift
task generate        # regenerate OpenAPI + typed API paths
```

VS Code / Cursor: use the **Full stack** launch compound in `.vscode/launch.json`.
Dev Containers: open the repo in a container via `.devcontainer/`.

## Pull requests

1. Keep changes focused. Prefer small PRs.
2. Run `task lint` and `task test:contract` before pushing.
3. Update `CHANGELOG.md` under `[Unreleased]` for user-visible changes
   (`task changelog` drafts from conventional commits via git-cliff).
4. If you change API routes or docs, run `task generate` and commit
   `web/src/lib/api/generated/`.
5. If you add UI strings, update `web/src/lib/i18n/locales/en.json` then
   `task i18n:sync` so other locales gain the new keys.

## Code style

- Go: `gofmt`, `golangci-lint` (see `.golangci.yml`)
- Web: ESLint + Prettier + `svelte-check`
- No emojis in code or docs
- Match existing naming and package layout (`docs/docs/project-layout.md`)

## Security

Do not commit secrets. Use `.env` locally (gitignored) and `.env.example` for
new `ATHENAEUM_*` keys. Report vulnerabilities privately if possible.

## License

By contributing you agree your work is released under the MIT license.
