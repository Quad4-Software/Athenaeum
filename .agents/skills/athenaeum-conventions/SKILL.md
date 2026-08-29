---
name: athenaeum-conventions
description: >-
  Athenaeum codebase conventions for Go, Svelte, icons, constants, file size,
  lint, and tests. Use when editing or reviewing code in this repository.
---

# Athenaeum conventions

## Hard rules

1. **No emojis** in code, comments, UI copy, markdown, or commit messages you draft.
2. **No emoji arrows.** Write `->`, `<-`, `=>`, or words (`to`, `from`).
3. **Icons from Lucide only.**
   - Web UI: `import { Name } from "@lucide/svelte"`
   - Docs site: `import { Name } from "lucide-react"`
   - Do not invent inline SVG icons when Lucide has a match.
4. **No god files.** One concern per module when growth hurts readability. Extract shared helpers instead of duplicating.
5. **Constants over magic values.** Shared paths, query keys, route fragments, timeouts, and status strings belong in named constants or existing config packages.
6. **Readable, maintainable, low-to-medium cognitive load.** Prefer clear names and short functions over clever nesting.
7. **Format and lint** after edits. Fix issues you cause.
8. **Run tests after each major feature.** Prefer `task test` (or targeted `go test` / `pnpm test:run` plus `task test:contract` when APIs move).

## Stack map

| Area | Stack |
| ---- | ----- |
| Backend | Go `net/http`, packages under `internal/` |
| DB | SQLite (modernc) by default, optional Postgres |
| Frontend | Svelte 5, TypeScript, Tailwind v4, Vite, pnpm |
| Icons | `@lucide/svelte` |
| Docs | Docusaurus + React, `lucide-react` for icons |
| Auth | Sessions, optional TOTP/OIDC, guest invites |

## Go

- Keep handlers thin. Put logic in focused packages under `internal/`.
- Reuse `internal/config`, `internal/models`, and existing helpers before adding parallel types.
- Match existing error and logging style in the package you touch.
- After API route changes: `task generate` and `task test:contract`.

## Svelte / web

- Follow existing Svelte 5 runes patterns in neighboring files.
- Reuse components under `web/src/lib/components/` and reader modules under `web/src/lib/reader/`.
- New UI strings go in `web/src/lib/i18n/locales/en.json`, then sync locales.
- Do not hand-roll icons or emoji as decoration.

## Split and reuse

Before growing a file past comfortable review size:

- Move pure helpers next to the feature (`*.ts` / small Go package).
- Share constants in a dedicated module rather than scattering literals.
- Prefer composition over a single mega-component or mega-handler.

## Prose in this repo

User-facing sentences follow `no-ai-slop`. Project docs follow `athenaeum-docs`.
