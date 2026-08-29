# AGENTS.md

Guidelines for AI agents working in this repository (Athenaeum).

## Repository

Athenaeum is a universal library tool. Today the Go backend embeds a Svelte 5
web UI as a self-hosted EPUB, PDF, comic, and audiobook catalog and reader.
Docs are a Docusaurus site under `docs/`. The roadmap covers apps, papers, ZIM,
feeds, peer sharing, and opt-in local-first AI.

```text
cmd/athenaeum/     binary entrypoint
internal/          Go packages (server, auth, library, opds, ...)
web/               Svelte 5 + Vite frontend
docs/              Docusaurus docs site
deploy/            service units and env examples
docker/            optional sidecars (Kokoro TTS, ...)
```

## Skills

Installable agent skills live under `.agents/skills/`:

| Skill | Use when |
| ----- | -------- |
| `no-ai-slop` | Writing or editing any prose (docs, README, PR text, comments that are sentences) |
| `rossmann-voice` | Explicitly asked for that voice profile |
| `athenaeum-conventions` | Any code or UI change in this repo |
| `athenaeum-docs` | Editing `docs/` or user-facing install/feature copy |

Full no-slop rule list: `.agents/references/no-ai-slop-rules.md`
Banned-words reference: `.agents/skills/no-ai-slop/references/ai-writing-detection.md`
Tooling cheat sheet: `.agents/tools.md`

## Operating rules

1. Before writing or editing prose, read `.agents/skills/no-ai-slop/SKILL.md`.
2. Before returning prose, self-check against the banned-words reference.
3. For code or UI work, follow `.agents/skills/athenaeum-conventions/SKILL.md`.
4. For docs site or install guides, also follow `.agents/skills/athenaeum-docs/SKILL.md`.
5. Do not invent features, flags, or API routes. Confirm against the tree or tests.
6. Prefer reusable packages and shared constants over one-off copies.
7. After a major feature, run format/lint and the relevant tests (see `.agents/tools.md`).

## Non-negotiable project conventions

1. No emojis anywhere (code, markdown, UI strings, docs).
2. No emoji arrows. Use `->`, `<-`, `=>`, or plain words.
3. Icons come from the Lucide pack (`@lucide/svelte` in `web/`, `lucide-react` in `docs/`). Do not hand-roll SVG icons.
4. No god files. Split by concern and reuse shared helpers where it pays off.
5. Prefer named constants over magic strings and numbers.
6. Keep code readable and maintainable at low-to-medium cognitive load.
7. Format and lint changed code. Fix issues you introduce.
8. Run tests after each major feature (`task test`, plus targeted packages when faster).

## Docs audience

Primary readers are self-hosters and household admins, not only Go contributors.
Lead with Docker / release binary / `install.sh`. Put "build from source" after
the run paths. Prefer plain language over stack buzzwords on the homepage and
intro pages. Keep deep API and package detail on the operate/contribute pages.
