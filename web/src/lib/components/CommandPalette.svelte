<script lang="ts">
  import { tick } from "svelte";
  import { Search, CornerDownLeft } from "@lucide/svelte";
  import { api } from "$lib/api/client";
  import { router } from "$lib/router.svelte";
  import { commandPalette } from "$lib/stores/commandPalette.svelte";
  import { keybindings } from "$lib/stores/keybindings.svelte";
  import { i18n } from "$lib/stores/i18n.svelte";
  import { formatChord, isMacPlatform } from "$lib/commands/chords";
  import { listRecentBooks, rememberBook } from "$lib/commands/recent";
  import type { RecentBook } from "$lib/commands/recent";
  import { runCommand, visibleCommands } from "$lib/commands/registry";
  import type { CommandDef } from "$lib/commands/types";
  import type { Book } from "$lib/api/types";

  type CommandRow = { kind: "command"; id: string; command: CommandDef; title: string };
  type BookRow = { kind: "book"; id: string; book: RecentBook; title: string };
  type PaletteRow = CommandRow | BookRow;

  let inputEl = $state<HTMLInputElement | null>(null);
  let listEl = $state<HTMLDivElement | null>(null);
  let activeIndex = $state(0);
  let bookHits = $state<Book[]>([]);
  let searching = $state(false);
  let searchTimer: ReturnType<typeof setTimeout> | null = null;

  const isMac = isMacPlatform();

  let query = $derived(commandPalette.query);

  let commandRows = $derived.by((): CommandRow[] => {
    const q = query.trim().toLowerCase();
    const cmds = visibleCommands().filter((c) => c.id !== "palette.open");
    const matched = q
      ? cmds.filter((c) => i18n.t(c.titleKey).toLowerCase().includes(q))
      : cmds.filter((c) => c.section === "navigate" || c.section === "actions");
    return matched.slice(0, 10).map((command) => ({
      kind: "command" as const,
      id: `cmd:${command.id}`,
      command,
      title: i18n.t(command.titleKey),
    }));
  });

  let bookRows = $derived.by((): BookRow[] => {
    const q = query.trim();
    const source: RecentBook[] = q ? bookHits.slice(0, 8) : listRecentBooks();
    return source.map((book) => ({
      kind: "book" as const,
      id: `book:${book.id}`,
      book,
      title: book.title,
    }));
  });

  let rows = $derived<PaletteRow[]>([...bookRows, ...commandRows]);

  $effect(() => {
    if (!commandPalette.open) return;
    activeIndex = 0;
    void tick().then(() => inputEl?.focus());
  });

  $effect(() => {
    const q = query.trim();
    if (!commandPalette.open) return;
    if (searchTimer) clearTimeout(searchTimer);
    if (!q) {
      bookHits = [];
      searching = false;
      return;
    }
    searching = true;
    searchTimer = setTimeout(() => {
      void api
        .listBooks({ search: q, limit: 8, offset: 0, sort: "title" })
        .then((page) => {
          bookHits = page.items;
        })
        .catch(() => {
          bookHits = [];
        })
        .finally(() => {
          searching = false;
        });
    }, 180);
    return () => {
      if (searchTimer) clearTimeout(searchTimer);
    };
  });

  $effect(() => {
    void rows;
    if (activeIndex >= rows.length) activeIndex = Math.max(0, rows.length - 1);
  });

  function close() {
    commandPalette.hide();
  }

  function onOverlayKeydown(event: KeyboardEvent) {
    if (event.key === "Escape") {
      event.preventDefault();
      close();
    }
  }

  function onInputKeydown(event: KeyboardEvent) {
    if (event.key === "Escape") {
      event.preventDefault();
      close();
      return;
    }
    if (event.key === "ArrowDown") {
      event.preventDefault();
      if (rows.length === 0) return;
      activeIndex = (activeIndex + 1) % rows.length;
      scrollActive();
      return;
    }
    if (event.key === "ArrowUp") {
      event.preventDefault();
      if (rows.length === 0) return;
      activeIndex = (activeIndex - 1 + rows.length) % rows.length;
      scrollActive();
      return;
    }
    if (event.key === "Enter") {
      event.preventDefault();
      const row = rows[activeIndex];
      if (row) void activate(row);
    }
  }

  function scrollActive() {
    void tick().then(() => {
      const el = listEl?.querySelector<HTMLElement>(`[data-index="${activeIndex}"]`);
      el?.scrollIntoView({ block: "nearest" });
    });
  }

  async function activate(row: PaletteRow) {
    if (row.kind === "book") {
      rememberBook({
        id: row.book.id,
        title: row.book.title,
        author: row.book.author,
        hasCover: row.book.hasCover,
        modifiedAt: row.book.modifiedAt,
      });
      close();
      router.navigate(`/book/${row.book.id}`);
      return;
    }
    close();
    await runCommand(row.command.id);
  }

  function chordLabel(command: CommandDef): string {
    const chord = keybindings.bindingFor(command.id);
    return chord ? formatChord(chord, isMac) : "";
  }
</script>

{#if commandPalette.open}
  <div class="palette-root" role="presentation" onkeydown={onOverlayKeydown}>
    <button
      type="button"
      class="palette-backdrop"
      aria-label={i18n.t("commands.close")}
      onclick={close}
    ></button>
    <div
      class="palette-panel"
      role="dialog"
      aria-modal="true"
      aria-label={i18n.t("commands.paletteTitle")}
    >
      <div class="palette-input-wrap">
        <Search size={18} class="text-subtle" />
        <input
          bind:this={inputEl}
          class="palette-input"
          type="search"
          placeholder={i18n.t("commands.palettePlaceholder")}
          value={query}
          oninput={(e) => {
            commandPalette.query = e.currentTarget.value;
            activeIndex = 0;
          }}
          onkeydown={onInputKeydown}
          autocomplete="off"
          spellcheck="false"
        />
        <kbd class="palette-kbd">{isMac ? "esc" : "Esc"}</kbd>
      </div>

      <div
        class="palette-list"
        bind:this={listEl}
        role="listbox"
        aria-label={i18n.t("commands.results")}
      >
        {#if rows.length === 0}
          <p class="palette-empty">
            {searching ? i18n.t("common.loading") : i18n.t("commands.noResults")}
          </p>
        {:else}
          {#if bookRows.length > 0}
            <p class="palette-section">
              {query.trim() ? i18n.t("commands.sectionBooks") : i18n.t("commands.sectionRecent")}
            </p>
            {#each bookRows as row, i (row.id)}
              <button
                type="button"
                role="option"
                data-index={i}
                class="palette-row"
                class:palette-row--active={i === activeIndex}
                aria-selected={i === activeIndex}
                onclick={() => void activate(row)}
                onmouseenter={() => (activeIndex = i)}
              >
                <span class="palette-thumb">
                  {#if row.book.hasCover}
                    <img
                      src={api.coverUrl(row.book.id, row.book.modifiedAt)}
                      alt=""
                      loading="lazy"
                    />
                  {:else}
                    <span class="palette-thumb-fallback"></span>
                  {/if}
                </span>
                <span class="min-w-0 flex-1 text-left">
                  <span class="block truncate text-sm text-fg">{row.title}</span>
                  {#if row.book.author}
                    <span class="block truncate text-xs text-muted">{row.book.author}</span>
                  {/if}
                </span>
                <CornerDownLeft size={14} class="text-subtle opacity-60" />
              </button>
            {/each}
          {/if}

          {#if commandRows.length > 0}
            <p class="palette-section">{i18n.t("commands.sectionCommands")}</p>
            {#each commandRows as row, j (row.id)}
              {@const i = bookRows.length + j}
              <button
                type="button"
                role="option"
                data-index={i}
                class="palette-row"
                class:palette-row--active={i === activeIndex}
                aria-selected={i === activeIndex}
                onclick={() => void activate(row)}
                onmouseenter={() => (activeIndex = i)}
              >
                <span class="min-w-0 flex-1 text-left text-sm text-fg">{row.title}</span>
                {#if chordLabel(row.command)}
                  <kbd class="palette-chord">{chordLabel(row.command)}</kbd>
                {/if}
              </button>
            {/each}
          {/if}
        {/if}
      </div>
    </div>
  </div>
{/if}

<style>
  .palette-root {
    position: fixed;
    inset: 0;
    z-index: 80;
    display: grid;
    place-items: start center;
    padding: 12vh 1rem 1rem;
  }

  .palette-backdrop {
    position: absolute;
    inset: 0;
    border: 0;
    background: var(--overlay);
    cursor: default;
  }

  .palette-panel {
    position: relative;
    z-index: 1;
    width: min(36rem, 100%);
    overflow: hidden;
    border-radius: 0.85rem;
    border: 1px solid var(--border);
    background: var(--bg-elevated);
    box-shadow: var(--shadow);
    animation: fade-in 160ms ease-out;
  }

  .palette-input-wrap {
    display: flex;
    align-items: center;
    gap: 0.65rem;
    border-bottom: 1px solid var(--border);
    padding: 0.85rem 1rem;
  }

  .palette-input {
    flex: 1;
    min-width: 0;
    border: 0;
    background: transparent;
    color: var(--fg);
    font-size: 1rem;
    outline: none;
  }

  .palette-input::placeholder {
    color: var(--fg-subtle);
  }

  .palette-kbd,
  .palette-chord {
    flex-shrink: 0;
    border-radius: 0.35rem;
    border: 1px solid var(--border);
    background: var(--bg);
    padding: 0.15rem 0.4rem;
    font-family: var(--font-mono);
    font-size: 0.65rem;
    color: var(--fg-subtle);
  }

  .palette-list {
    max-height: min(24rem, 55vh);
    overflow: auto;
    padding: 0.4rem;
  }

  .palette-section {
    margin: 0.35rem 0.5rem 0.2rem;
    font-family: var(--font-display);
    font-size: 0.7rem;
    font-weight: 560;
    letter-spacing: 0.04em;
    text-transform: uppercase;
    color: var(--fg-subtle);
  }

  .palette-row {
    display: flex;
    width: 100%;
    align-items: center;
    gap: 0.75rem;
    border: 0;
    border-radius: 0.55rem;
    background: transparent;
    padding: 0.55rem 0.65rem;
    cursor: pointer;
    color: inherit;
  }

  .palette-row--active {
    background: color-mix(in oklch, var(--primary) 14%, var(--surface-hover));
  }

  .palette-thumb {
    width: 2rem;
    height: 2.85rem;
    flex-shrink: 0;
    overflow: hidden;
    border-radius: 0.25rem;
    background: var(--bg);
  }

  .palette-thumb img,
  .palette-thumb-fallback {
    display: block;
    width: 100%;
    height: 100%;
    object-fit: cover;
  }

  .palette-thumb-fallback {
    background: color-mix(in oklch, var(--border) 55%, var(--bg));
  }

  .palette-empty {
    margin: 0;
    padding: 1.5rem 1rem;
    text-align: center;
    font-size: 0.875rem;
    color: var(--fg-muted);
  }
</style>
