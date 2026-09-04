# Documentation

Athenaeum docs are a [Docusaurus](https://docusaurus.io/) site.

## Local development

Requires Node 22+ and pnpm 11+.

### Manual

```sh
cd docs
pnpm install
pnpm start
```

Open http://localhost:3000. Markdown lives in [`docs/`](./docs/) (the content
folder inside this package). Theme and homepage are under [`src/`](./src/).

```sh
cd docs && pnpm install && pnpm build   # output -> docs/build
cd docs && pnpm serve                   # preview production build
```

### Task (optional)

```sh
task docs:setup
task docs:dev
task docs:build   # output -> docs/build
task docs:serve   # preview production build
```

## Content map

| Page | Topic |
| ---- | ----- |
| intro | Product overview |
| getting-started | Build and run (Make, manual, Task) |
| features | Capability list |
| deploying | Installer, Docker profiles, host services, backup |
| configuration | Flags and env vars |
| authentication | Sessions, guests, permissions, TOTP, OIDC, ALTCHA |
| cli-users | `athenaeum users` offline account management |
| library | Formats, metadata, narration, sharing, SMTP/Kindle |
| catalogs | OPDS 1.2 / 2 and KOSync |
| operations | Metrics, maintenance, i18n, PWA, backup |
| http-api | Full route reference |
| development | Contributor workflow and CI |
| project-layout | Repository map |

## GitHub Pages

Site URL: https://athenaeum.quad4.io (`docs/static/CNAME`).
`.github/workflows/pages.yml` builds this site and embeds the offline SPA under
`/demo`. Enable **Settings -> Pages -> GitHub Actions** and set the custom
domain to `athenaeum.quad4.io` (DNS CNAME to the Pages host).

| Variable | Purpose | Production |
| -------- | ------- | ---------- |
| `DOCUSAURUS_URL` | Canonical site origin | `https://athenaeum.quad4.io` |
| `DOCUSAURUS_BASE_URL` | Path prefix | `/` |
| `GITHUB_REPOSITORY` | `owner/repo` for edit links and navbar | from Actions |
