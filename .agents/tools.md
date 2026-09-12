# Tools

Commands agents should use in this repo. Prefer Task when available.

## Setup

```sh
task setup              # go mod download + web/docs pnpm install
task tools:install      # air, golangci-lint, govulncheck, deadcode, lefthook, git-cliff
task hooks:install      # git hooks
task doctor             # toolchain and ports
```

## Day to day

```sh
task dev                # Vite :5173 + Go (air) :8080
task demo               # seeded demo library
task build              # web + Go binary -> ./bin/athenaeum
task build:slim         # without Kokoro WASM -> ./bin/athenaeum-slim
task generate           # OpenAPI + typed API paths
```

## Format and lint

```sh
task fmt                # gofmt + web prettier
task lint               # gofmt check, golangci-lint, eslint, prettier, svelte-check
task lint:go
task lint:web
```

Run these on touched packages/files before finishing a change. Fix failures.

## Tests

```sh
task test               # Go + Vitest
task test:go            # PKG=./internal/auth narrows scope
task test:web
task test:contract      # OpenAPI / route / i18n / env drift
task test:e2e           # Playwright (builds binary if needed)
task test:race          # Go race detector
task test:property      # Go testing/quick + web fast-check suites
task test:fuzz          # Go fuzz targets, 10s each (test:fuzz:long for 1m)
task test:coverage      # Go coverage report (test:coverage:web for Vitest)
task test:mutation      # Gremlins on Go (test:mutation:web for Stryker)
task test:bench         # Go benchmarks
task test:all           # unit + race + property + contract + short fuzz + coverage
```

After a major feature: `task fmt` (or format via lint tools), `task lint`, then
`task test` (or the smallest suite that covers the change, then expand if risk
is high). API or route changes also need `task generate` and `task test:contract`.
When you add or change tests, see `.agents/skills/testing/SKILL.md`; mutation
and property targets above are how you verify the tests actually constrain
behavior.

## Hygiene and security

```sh
task generate:check     # generated files are in sync
task deadcode           # unreachable Go code
task knip               # unused web exports/deps
task security           # gosec + govulncheck + pnpm audit
```

Run `deadcode`/`knip` after refactors that may have orphaned code.

## Docs site

```sh
task docs:dev           # http://localhost:3000
task docs:build         # docs/build
```

Or from `docs/`: `pnpm start`, `pnpm build`, `pnpm typecheck`.

## Manual equivalents

```sh
go test ./...
cd web && pnpm test:run && pnpm lint && pnpm format:check && pnpm check
cd docs && pnpm typecheck && pnpm build
```
