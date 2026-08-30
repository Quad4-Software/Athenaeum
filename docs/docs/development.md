---
sidebar_position: 9
title: Development
description: Task recipes, testing methods, and CI workflows.
---

# Development

```sh
task setup            # Go modules + web pnpm + docs pnpm
task tools:install    # air, golangci-lint, govulncheck, deadcode, lefthook, git-cliff -> ./bin
task hooks:install    # lefthook pre-commit / pre-push (or: lefthook install)
task doctor           # toolchain, optional tools, ports
task dev              # Vite :5173 + Go live reload (air) on :8080
task build            # frontend + single binary -> ./bin/athenaeum
task build:slim       # same without in-browser Kokoro WASM -> ./bin/athenaeum-slim
task run              # build then run local server
task demo             # Go server with --demo seeded library
task reset:data       # wipe ./data
task reset:demo       # wipe ./data and start --demo
task build:web        # production SPA into internal/assets/dist
task build:demo       # static SPA -> ./site (offline / GitHub Pages)
task docs:dev         # Docusaurus at http://localhost:3000
task docs:build       # docs/build
task docs:serve       # preview production docs build
task showcase         # Playwright screenshots -> ./showcase
task generate         # OpenAPI JSON + typed web/src/lib/api/generated
task generate:check   # assert generated artifacts are up to date
task i18n:sync        # fill missing locale keys from en.json
task i18n:check       # locale key parity
task changelog        # draft unreleased notes via git-cliff
task fmt              # gofmt
task vendor           # go mod vendor
task clean            # remove build artifacts
task fetch-samples    # optional sample media
task test             # Go + Vitest
task test:race        # Go tests with the race detector
task test:property    # Go testing/quick + frontend fast-check properties
task test:contract    # OpenAPI/route/i18n/env drift + generate/i18n checks
task test:fuzz        # Go native fuzz targets (FUZZTIME=10s)
task test:fuzz:long
task test:coverage    # Go coverage (fails below COVERAGE_MIN, default 45)
task test:coverage:web # Vitest coverage with thresholds
task test:bench       # Go benchmarks
task test:mutation    # Gremlins mutation testing (Go)
task test:mutation:web # Stryker mutation testing (web utils)
task test:e2e         # Playwright UI tests (builds binary first)
task test:all         # unit + race + property + contract + short fuzz + coverage
task test:lighthouse  # Lighthouse CI vs production binary (needs Chromium)
task lint             # gofmt + golangci-lint + eslint + prettier + svelte-check
task lint:go          # Go linters only
task lint:web         # frontend linters only
task security         # gosec + govulncheck + pnpm audit
task deadcode         # unused Go symbols
task knip             # unused frontend dependency scan
task profile          # CPU profile from ATHENAEUM_PPROF (default 127.0.0.1:6060)
task docker:build     # docker compose build
task docker:up        # docker compose up -d --build
task docker:down      # docker compose down
task docker:logs      # follow container logs
```

See also [CONTRIBUTING.md](https://github.com/ivan/reader/blob/main/CONTRIBUTING.md) and `.devcontainer/` for a full
devcontainer. VS Code / Cursor debug configs live in `.vscode/launch.json`.

task test:lighthouse builds the binary if needed, serves it on :18080, and runs
[Lighthouse CI](https://github.com/GoogleChrome/lighthouse-ci)
(web/lighthouserc.cjs) with performance / accessibility / best-practices / SEO
score gates (default 90 or higher). Reports land in web/.lighthouseci/. Override with
LIGHTHOUSE_MIN_SCORE or CHROME_PATH.

For a production-style UI without re-embedding into the binary, build the frontend
(`task build:web`) and run with `--web-dir ./internal/assets/dist` (or any Vite
output directory that contains `index.html`).

Local profiling: start with `--pprof 127.0.0.1:6060` (or `ATHENAEUM_PPROF`), then
`task profile` or `go tool pprof http://127.0.0.1:6060/debug/pprof/profile`.

### Testing methods

| Method | How to run | Notes |
| ------ | ---------- | ----- |
| Unit | `task test` | Go `testing` + Vitest |
| Race | `task test:race` | Also on Linux in CI |
| Property | `task test:property` | `testing/quick` (Go), `fast-check` (web) |
| Contract / drift | `task test:contract` | OpenAPI vs routes, FE API paths, i18n keys, `.env.example`, generated client |
| Fuzz | `task test:fuzz` | Go native fuzz (`Fuzz*` in `internal/`) |
| Coverage | `task test:coverage` / `test:coverage:web` | Atomic Go profile; Vitest v8 thresholds |
| Bench | `task test:bench` | Existing `Benchmark*` targets |
| Mutation (Go) | `task test:mutation` | [Gremlins](https://gremlins.dev) on auth/library |
| Mutation (web) | `task test:mutation:web` | Stryker on sanitize/password/format utils |
| Oracle | included in `task test` | `*_oracle_test.go` / `*.oracle.test.ts` |
| E2E / UI | `task test:e2e` | Playwright (also a CI job) |
| Lighthouse | `task test:lighthouse` | Score budgets |

Fuzz duration defaults to 10s per target (`FUZZTIME`, minimum 5s). Mutation efficacy /
coverage thresholds are in `.gremlins.yaml` and `web/stryker.config.js`
(override with `MUTATION_EFFICACY`, `MUTATION_MCOVER`, or `MUTATION_DIFF` for
changed-files-only runs). Go coverage minimum defaults to 45% (`COVERAGE_MIN`).

## Continuous integration

Workflows live under .github/workflows/ (actions pinned to full commit SHAs):

| Workflow | Trigger | What it does |
| -------- | ------- | ------------ |
| ci.yml | push / PR / manual | Lint (gofmt, golangci-lint, eslint, prettier, svelte-check), Go tests, fuzz, coverage gate, cross-compile (Linux/macOS/Windows/BSD/armv6/armv7/riscv64), `--self-check` on native + QEMU arches, Vitest + coverage, govulncheck, pnpm audit, gosec, CodeQL, multi-arch Docker + container self-check, Lighthouse, Playwright, generate check |
| codeql.yml | push / PR / weekly / manual | CodeQL for Go and JavaScript/TypeScript |
| release.yml | v* tags / manual | Multi-arch single binaries + GitHub Release (SBOM + license report), multi-arch image to GHCR (`linux/amd64`, `arm64`) |
| pages.yml | push (docs/web) / manual | Docs site + offline demo -> GitHub Pages |

Default GITHUB_TOKEN permissions are read-only. Jobs elevate only what they
need (security-events, packages, pages, contents: write for releases).
Dependabot keeps Actions, Go modules, and web/ npm deps updated weekly.
