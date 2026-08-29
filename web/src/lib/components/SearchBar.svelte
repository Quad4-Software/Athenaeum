<script lang="ts">
  import { Search } from "@lucide/svelte";
  import { commandPalette } from "$lib/stores/commandPalette.svelte";
  import { i18n } from "$lib/stores/i18n.svelte";

  function openPalette() {
    commandPalette.show();
  }
</script>

<button
  type="button"
  class="search-trigger"
  aria-keyshortcuts="Control+K Meta+K"
  aria-label={i18n.t("commands.paletteTitle")}
  onclick={openPalette}
>
  <Search size={16} class="shrink-0 text-subtle" />
  <span class="min-w-0 flex-1 truncate text-left text-sm text-subtle">
    {i18n.t("commands.palettePlaceholder")}
  </span>
  <kbd class="search-kbd" aria-hidden="true">{i18n.t("search.shortcut")}</kbd>
</button>

<style>
  .search-trigger {
    display: flex;
    width: 100%;
    max-width: 28rem;
    align-items: center;
    gap: 0.6rem;
    height: 2.5rem;
    border-radius: 0.55rem;
    border: 1px solid var(--border);
    background: var(--bg);
    padding: 0 0.65rem 0 0.85rem;
    color: inherit;
    cursor: pointer;
    transition:
      border-color 120ms ease,
      background-color 120ms ease;
  }

  .search-trigger:hover {
    border-color: var(--border-strong);
    background: var(--surface);
  }

  .search-kbd {
    display: none;
    border-radius: 0.3rem;
    border: 1px solid var(--border);
    background: var(--bg-elevated);
    padding: 0.12rem 0.4rem;
    font-size: 0.625rem;
    font-weight: 600;
    color: var(--fg-subtle);
  }

  @media (min-width: 640px) {
    .search-kbd {
      display: inline-flex;
    }
  }
</style>
