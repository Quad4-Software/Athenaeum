# Documentation

Athenaeum docs are a [Docusaurus](https://docusaurus.io/) site.

## Local development

```sh
task docs:setup
task docs:dev
```

Open http://localhost:3000. Markdown lives in [`docs/`](./docs/) (the content
folder inside this package). Theme and homepage are under [`src/`](./src/).

```sh
task docs:build   # output -> docs/build
task docs:serve   # preview production build
```

## Content map

| Page | Topic |
| ---- | ----- |
| intro | Product overview |
| getting-started | Build, run, demo modes |
| features | Capability list |
| deploying | Installer, Docker profiles, host services, backup |
| configuration | Flags and env vars |
| authentication | Sessions, guests, permissions, TOTP, OIDC, ALTCHA |
| cli-users | `athenaeum users` offline account management |
| library | Formats, metadata, narration, sharing, SMTP/Kindle |
| catalogs | OPDS 1.2 / 2 and KOSync |
| operations | Metrics, maintenance, i18n, PWA, backup |
| http-api | Full route reference |
| development | Task recipes and CI |
| project-layout | Repository map |

## GitHub Pages

`.github/workflows/pages.yml` builds this site and embeds the offline SPA demo
at `/demo`. Enable **Settings -> Pages -> GitHub Actions** on the repository.

| Variable | Purpose |
| -------- | ------- |
| `DOCUSAURUS_URL` | Canonical site origin |
| `DOCUSAURUS_BASE_URL` | Path prefix (for example `/reader/`) |
| `GITHUB_REPOSITORY` | `owner/repo` for edit links and navbar |
