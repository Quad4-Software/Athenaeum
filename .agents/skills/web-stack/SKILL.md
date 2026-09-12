---
name: web-stack
description: >-
  The web frontend stack as actually installed: Svelte 5 runes, runed,
  bits-ui v2, Tailwind v4, Vite 8, TypeScript 6, Vitest 5. Use when writing
  or editing anything under web/, and to look up current upstream docs
  instead of relying on stale training knowledge.
---

# Web stack

Model training data predates most of this stack. The installed versions are
newer than what you probably know, so when an API detail matters, fetch the
upstream doc instead of recalling it. Sources and llms.txt endpoints are
listed per package below.

## Installed versions (verified from node_modules, September 2026)

| Package | Version |
| ------- | ------- |
| svelte | 5.57 |
| runed | 0.37.1 |
| bits-ui | 2.19 |
| tailwindcss | 4.3.3 (`@tailwindcss/vite` plugin) |
| vite | 8.2 (`@sveltejs/vite-plugin-svelte` 7.3) |
| typescript | 6.0.3 |
| vitest | 5.0 (`@vitest/coverage-v8`) |
| @testing-library/svelte | 5.4 |
| fast-check | 4.9 |
| @stryker-mutator/core + vitest-runner | 10.0 |
| @playwright/test | 1.63 |
| eslint | 10.10 (`eslint-plugin-svelte` 3.23, typescript-eslint 8.70) |
| svelte-check | 4.7 |
| @lucide/svelte | 1.42 |
| @sentry/svelte | 10.73 |
| vite-plugin-pwa + workbox-window | 1.3 / 7.4 |
| epubjs | 0.3.93 |
| pdfjs-dist | 6.3 |
| kokoro-js + @huggingface/transformers | 1.2 / 4.2 |
| altcha | 3.2 |
| qrcode | 1.5 |

Check `web/package.json` for the current pins; this table ages.

## Svelte 5: this codebase is runes-only

The repo has zero legacy syntax: no `createEventDispatcher`, no `on:click`,
no `<slot>`. Keep it that way. Anything you write must use runes.

| Legacy (do not write) | Runes equivalent |
| --------------------- | ---------------- |
| `export let foo` | `let { foo } = $props()` with a `Props` interface |
| `$: doubled = x * 2` | `const doubled = $derived(x * 2)` |
| `$: { sideEffect(x) }` | `$effect(() => sideEffect(x))` |
| `createEventDispatcher` + `on:save` | callback prop: `onsave?: () => void` |
| `on:click={fn}` | `onclick={fn}` |
| `<slot />` | `children: Snippet` prop + `{@render children()}` |
| named slots | `{#snippet name(args)}` + typed `Snippet<[T]>` props |
| `bind:value` into child props | `let { value = $bindable() } = $props()` |
| `svelte/store` `writable` | `$state` / `$derived` in a `.svelte.ts` module |

Measured usage in this repo: `$state` 273x, `$derived` 94x, `$props` 60x,
`$effect` 59x, `$bindable` 30x, `$derived.by` 16x, `$effect.root` 5x,
`{@render}`/`{#snippet}` 55 sites.

### Runes rules that matter here

- Shared state lives in `.svelte.ts` modules as exported classes or objects
  using runes (`router.svelte.ts`, `navigation.svelte.ts` are the pattern).
  Plain `.ts` files cannot use runes.
- `$derived` for anything computable from other state. `$effect` only for
  syncing with systems outside the reactivity graph: DOM APIs, epub.js,
  pdf.js, timers, media. If you are updating one piece of state from
  another inside `$effect`, you wanted `$derived` or an event handler.
- Do not destructure or eagerly read `$derived`/`$state` values into plain
  `let` bindings; you lose reactivity. Read them at point of use.
- Callback props over event dispatching. Type them optional
  (`onclose?: () => void`) and call with `onclose?.()`.
- `untrack`, `flushSync`, `tick` exist for escape hatches; reach for them
  rarely and with a reason.
- `children` is a `Snippet`, not implicit slot content:
  `import type { Snippet } from "svelte"`.

### Upstream docs (fetch these; do not guess API shape)

- `https://svelte.dev/llms.txt` -> index of LLM doc sets
- `https://svelte.dev/llms-small.txt` -> minimal full docs
- `https://svelte.dev/llms-medium.txt` -> abridged with examples
- `https://svelte.dev/llms-full.txt` -> complete docs
- `https://svelte.dev/docs/svelte/llms.txt` -> Svelte package docs only

## bits-ui v2: headless components

bits-ui v2 is runes-based and its API differs from v1 (what older training
data knows). Components used here: Dialog, DropdownMenu, ContextMenu,
Popover, Select, Slider, Toggle, Switch, Tooltip, Command.

House pattern: compose bits-ui inside the wrappers in
`web/src/lib/components/` (Modal, Dropdown, SelectMenu, Popover, Slider,
Toggle, Tooltip, ContextMenu, CommandPalette, ConfirmDialogHost, ToastHost).
Feature code uses the wrappers, not bits-ui directly.

API shapes to get right (see `Modal.svelte`, `Dropdown.svelte`):

- Namespace imports: `import { Dialog } from "bits-ui"`, then
  `Dialog.Root`, `Dialog.Portal`, `Dialog.Overlay`, `Dialog.Content`,
  `Dialog.Title`.
- Controlled state via `bind:open` plus an `onOpenChange` callback prop on
  the wrapper that calls `onclose?.()` when it transitions closed.
- Render delegation: triggers take `{#snippet child({ props })}` so the
  consumer supplies the trigger element while bits-ui wires ARIA and
  events. `{@render children()}` for content.
- Styling is plain `class` props and `data-*` attributes; components ship
  unstyled.

Upstream docs:

- `https://bits-ui.com/llms.txt` -> index of all LLM-friendly pages
- `https://bits-ui.com/docs/llms.txt` -> consolidated full docs
- Append `/llms.txt` to any doc page URL, e.g.
  `https://bits-ui.com/docs/components/dialog/llms.txt`

## runed: rune utilities

`PersistedState` is the workhorse (16 uses): localStorage-backed reactive
state for UI preferences and recents. Pattern: `new PersistedState<T>(KEY,
defaultValue)` at module scope in a `.svelte.ts` or library file, with the
storage key in a named constant (see `src/lib/commands/recent.ts`).

Also in use: `watch`, `useEventListener`, `useIntersectionObserver`,
`useDebounce`, `useResizeObserver`, `useInterval`.

Available and worth reaching for before hand-rolling: `Debounced`,
`Throttled`, `Previous`, `StateHistory`, `Context`, `FiniteStateMachine`,
`ElementSize`, `ElementRect`, `IsInViewport`, `IsFocusWithin`,
`IsDocumentVisible`, `IsIdle`, `IsMounted`, `PressedKeys`,
`onClickOutside`, `activeElement`, `resource`, `extract`,
`useMutationObserver`, `useSearchParams`, `boolAttr`, `onCleanup`.

Upstream docs (no llms.txt; read the page):
`https://runed.dev/docs` and `https://runed.dev/docs/utilities/<name>`,
e.g. `.../utilities/persisted-state`.

## Tailwind v4: CSS-first, no JS config

There is no `tailwind.config.js`. All config lives in `src/app.css`:

- `@import "tailwindcss"` brings in the framework.
- `@theme inline` maps runtime CSS variables into utility namespaces, so
  `bg-surface`, `text-muted`, `border-border` resolve to `--surface`,
  `--fg-muted`, `--border` and friends. Palette tokens live on
  `:root` / `[data-theme]` and in `web/src/lib/brand/themes/`; do not
  hardcode hex or oklch values in components.
- `@custom-variant` registers `dark`, `light`, and `starting` variants
  (`starting` wraps `@starting-style` for entry transitions).

v4 renames that older training data gets wrong: `shadow-sm` ->
`shadow-xs`, `shadow` -> `shadow-sm`, `rounded-sm` -> `rounded-xs`,
`rounded` -> `rounded-sm`, `outline-none` -> `outline-hidden`, `ring` ->
`ring-3`, `blur` -> `blur-sm`, `blur-sm` -> `blur-xs`, `flex-shrink-*` ->
`shrink-*`, `flex-grow-*` -> `grow-*`, `bg-opacity-*` is gone (use
`bg-black/50` opacity syntax), `space-x/y-*` margin handling unchanged but
prefer `gap-*` in flex/grid. When unsure, check
`https://tailwindcss.com/docs` and the v4 upgrade guide.

## Toolchain notes

- TypeScript 6: `baseUrl` in tsconfig is deprecated and removed in TS 7;
  the docs-site lint error about it comes from the editor using a newer TS
  than docs' pinned 5.6. Do not add `baseUrl` to web configs.
- Vite 8 via `@sveltejs/vite-plugin-svelte` 7; env flags `VITE_SLIM`,
  `VITE_DEMO` gate the slim and demo builds (`pnpm build:slim`,
  `pnpm build:demo`).
- PWA: `vite-plugin-pwa` + `workbox-window`; service worker code is in
  `src/lib/pwa/`.
- Icons: `@lucide/svelte` only. Never inline an SVG icon.
- Readers: `epubjs` (EPUB), `pdfjs-dist` (PDF), `kokoro-js` +
  `@huggingface/transformers` (TTS narration). These have their own
  lifecycle quirks; follow `src/lib/reader/` and `src/lib/narrator/`
  patterns rather than the libraries' README defaults.

## Verify before finishing

```sh
cd web
pnpm check          # svelte-check, required before build
pnpm lint           # eslint
pnpm test:run       # vitest
pnpm format:check   # prettier
```

`task lint:web` and `task test:web` from the repo root run the same.

## Self-check for web changes

1. Any `on:`, `<slot>`, `createEventDispatcher`, `export let`, or `$:` in
   what I wrote? Rewrite with runes.
2. Did I hand-roll something runed provides? Check the list above.
3. Did I compose bits-ui directly in feature code instead of using the
   `components/` wrappers? Use the wrapper, or extend it if it lacks the
   prop you need.
4. Any hardcoded color instead of a token utility (`bg-surface`,
   `text-muted`, ...)? Fix it.
5. Any inline SVG icon? Use `@lucide/svelte`.
6. Did I rely on remembered API details for bits-ui/runed/Svelte? Fetch
   the llms.txt or doc page and confirm.
