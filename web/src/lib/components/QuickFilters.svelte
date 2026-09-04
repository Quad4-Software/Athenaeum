<script lang="ts">
  import { BookText, Files, FileText, Headphones } from "@lucide/svelte";
  import { library } from "$lib/stores/library.svelte";
  import { router } from "$lib/router.svelte";
  import { i18n } from "$lib/stores/i18n.svelte";
  import type { BookFormat } from "$lib/api/types";

  type QuickId = BookFormat;

  const items: {
    id: QuickId;
    labelKey: string;
    icon: typeof BookText;
  }[] = [
    { id: "epub", labelKey: "nav.epub", icon: BookText },
    { id: "pdf", labelKey: "nav.pdf", icon: Files },
    { id: "papers", labelKey: "nav.papers", icon: FileText },
    { id: "audio", labelKey: "nav.audiobooks", icon: Headphones },
  ];

  function isActive(id: QuickId): boolean {
    if (router.current.name !== "library" && router.current.name !== "collection") return false;
    return library.formatFilter === id;
  }

  function select(id: QuickId) {
    if (library.formatFilter === id) {
      library.setFormat("");
    } else {
      library.setFormat(id);
    }
    if (router.current.name !== "library") router.navigate("/");
  }
</script>

<div
  class="quick-filters mb-3 -mx-1 flex gap-2 overflow-x-auto px-1 pb-0.5 md:hidden"
  role="toolbar"
  aria-label={i18n.t("library.quickFilters")}
>
  {#each items as item (item.id)}
    {@const active = isActive(item.id)}
    <button
      type="button"
      class="quick-chip"
      class:quick-chip--active={active}
      aria-pressed={active}
      onclick={() => select(item.id)}
    >
      <item.icon size={14} />
      <span>{i18n.t(item.labelKey)}</span>
    </button>
  {/each}
</div>

<style>
  .quick-filters {
    scrollbar-width: none;
    -webkit-overflow-scrolling: touch;
  }

  .quick-filters::-webkit-scrollbar {
    display: none;
  }

  .quick-chip {
    display: inline-flex;
    align-items: center;
    gap: 0.375rem;
    flex-shrink: 0;
    min-height: 2rem;
    padding: 0.25rem 0.625rem;
    border: 0;
    border-radius: 9999px;
    font-size: 0.6875rem;
    font-weight: 500;
    color: var(--color-muted);
    background: var(--color-surface);
    box-shadow: inset 0 0 0 1px var(--color-border);
    cursor: pointer;
    white-space: nowrap;
    transition:
      color 120ms ease,
      background-color 120ms ease,
      box-shadow 120ms ease;
  }

  .quick-chip:hover {
    color: var(--color-fg);
    background: var(--color-surface-hover);
  }

  .quick-chip--active {
    color: var(--color-primary-fg);
    background: var(--color-primary);
    box-shadow: none;
  }

  .quick-chip--active:hover {
    color: var(--color-primary-fg);
    background: var(--color-primary-hover);
  }
</style>
