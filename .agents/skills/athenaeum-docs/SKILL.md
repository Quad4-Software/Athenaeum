---
name: athenaeum-docs
description: >-
  Writing and structuring Athenaeum docs for self-hosters. Use when editing
  docs/, README install sections, homepage feature cards, or deploy guides.
---

# Athenaeum docs

Audience: people who want a private library on their NAS, VPS, or home lab.
Lead with run paths. Put contributor build steps later.

Also apply `no-ai-slop` to every sentence.

## Page jobs

| Page | Job |
| ---- | --- |
| Homepage cards | Benefits in plain language. One idea per card. |
| Intro | What it is and who it is for. Link to start paths. |
| Getting started | Run it today (Docker, binary, installer). Source build last. |
| Features | Capability list grouped by what users do. |
| Deploying | Docker, services, backups, optional sidecars. |
| Operate docs | Flags, auth, library, OPDS, API. Precise and skimmable. |

## Homepage feature cards

- Short title. One or two plain sentences. No stack soup on the card.
- Narration via **Kokoro** is its own card (not buried under "readers").
- Prefer Lucide icons (`lucide-react`). No emoji. No hand-rolled SVG.
- Headings name the content (`What you get`), not movie-poster teasers.

## Language

- Say "point it at a folder of books" before "concurrent FTS5 indexing".
- Name products people know: Docker, KOReader, Kindle, OPDS, PWA.
- Keep deep internals (WASM, WebGPU, pgx, modernc) on operate/tech sections.
- Use `->` in paths like `Settings -> Administration`, never emoji arrows.

## Self-hoster checklist

When editing install or feature copy, ask:

1. Can someone who only knows Docker Compose succeed from this page?
2. Is the first code block a run path, not a toolchain install?
3. Are optional power features labeled optional?
4. Did I remove filler and marketing abstractions?

## Related paths

- Site: `docs/src/pages/index.tsx`, `docs/src/components/HomepageFeatures/`
- Content: `docs/docs/*.md`
- Sidebar: `docs/sidebars.ts`
