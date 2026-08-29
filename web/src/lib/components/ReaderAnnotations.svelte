<script lang="ts">
  import { Bookmark, Highlighter, Trash2, X } from "@lucide/svelte";
  import { api, ApiError } from "$lib/api/client";
  import type { Bookmark as BookmarkRow, Highlight } from "$lib/api/types";
  import EmptyState from "$lib/components/EmptyState.svelte";
  import { toast } from "$lib/stores/toast.svelte";
  import { i18n } from "$lib/stores/i18n.svelte";

  interface Props {
    bookId: number;
    open?: boolean;
    revision?: number;
    onclose?: () => void;
    onJump?: (location: string) => void;
    locationType?: "cfi" | "page";
  }

  let {
    bookId,
    open = false,
    revision = 0,
    onclose,
    onJump,
    locationType = "cfi",
  }: Props = $props();

  let bookmarks = $state<BookmarkRow[]>([]);
  let highlights = $state<Highlight[]>([]);
  let tab = $state<"bookmarks" | "highlights">("bookmarks");
  let loading = $state(false);

  $effect(() => {
    void revision;
    if (!open || !bookId) return;
    loading = true;
    Promise.all([api.listBookmarks(bookId), api.listHighlights(bookId)])
      .then(([bm, hl]) => {
        bookmarks = bm;
        highlights = hl;
      })
      .catch((e) => {
        toast.error(e instanceof ApiError ? e.message : "Failed to load annotations");
      })
      .finally(() => {
        loading = false;
      });
  });

  async function removeBookmark(id: number) {
    try {
      await api.deleteBookmark(bookId, id);
      bookmarks = bookmarks.filter((b) => b.id !== id);
    } catch (e) {
      toast.error(e instanceof ApiError ? e.message : "Delete failed");
    }
  }

  async function removeHighlight(id: number) {
    try {
      await api.deleteHighlight(bookId, id);
      highlights = highlights.filter((h) => h.id !== id);
    } catch (e) {
      toast.error(e instanceof ApiError ? e.message : "Delete failed");
    }
  }

  function bookmarkLabel(loc: string) {
    if (locationType === "page") return `Page ${loc}`;
    return loc.length > 24 ? `${loc.slice(0, 24)}...` : loc;
  }
</script>

{#if open}
  <aside class="annotations-panel">
    <div class="annotations-head">
      <div class="flex gap-2">
        <button
          type="button"
          class="annotations-tab"
          class:annotations-tab--active={tab === "bookmarks"}
          onclick={() => (tab = "bookmarks")}
        >
          <Bookmark size={14} /> Bookmarks
        </button>
        <button
          type="button"
          class="annotations-tab"
          class:annotations-tab--active={tab === "highlights"}
          onclick={() => (tab = "highlights")}
        >
          <Highlighter size={14} /> Highlights
        </button>
      </div>
      <button type="button" class="btn btn-ghost" aria-label="Close" onclick={() => onclose?.()}>
        <X size={16} />
      </button>
    </div>
    {#if loading}
      <p class="p-3 text-sm text-muted">{i18n.t("common.loading")}</p>
    {:else if tab === "bookmarks"}
      {#if bookmarks.length === 0}
        <EmptyState
          size="sm"
          class="px-3"
          title={i18n.t("reader.bookmarksEmptyTitle")}
          body={i18n.t("reader.bookmarksEmptyBody")}
        >
          {#snippet icon(size)}
            <Bookmark {size} />
          {/snippet}
        </EmptyState>
      {:else}
        <ul class="annotations-list">
          {#each bookmarks as bm (bm.id)}
            <li class="annotations-item">
              <button type="button" class="annotations-jump" onclick={() => onJump?.(bm.location)}>
                <span class="font-medium">{bm.label || i18n.t("reader.bookmarkLabel")}</span>
                <span class="text-xs text-muted">{bookmarkLabel(bm.location)}</span>
              </button>
              <button
                type="button"
                class="btn btn-ghost text-xs"
                aria-label="Delete bookmark"
                onclick={() => removeBookmark(bm.id)}
              >
                <Trash2 size={14} />
              </button>
            </li>
          {/each}
        </ul>
      {/if}
    {:else if highlights.length === 0}
      <EmptyState
        size="sm"
        class="px-3"
        title={i18n.t("reader.highlightsEmptyTitle")}
        body={i18n.t("reader.highlightsEmptyBody")}
      >
        {#snippet icon(size)}
          <Highlighter {size} />
        {/snippet}
      </EmptyState>
    {:else}
      <ul class="annotations-list">
        {#each highlights as hl (hl.id)}
          <li class="annotations-item annotations-item--stack">
            <button type="button" class="annotations-jump" onclick={() => onJump?.(hl.location)}>
              <span class="highlight-excerpt">{hl.excerpt || "Highlight"}</span>
              {#if hl.note}
                <span class="text-xs text-muted">{hl.note}</span>
              {/if}
            </button>
            <button
              type="button"
              class="btn btn-ghost text-xs"
              aria-label="Delete highlight"
              onclick={() => removeHighlight(hl.id)}
            >
              <Trash2 size={14} />
            </button>
          </li>
        {/each}
      </ul>
    {/if}
  </aside>
{/if}

<style>
  .annotations-panel {
    position: absolute;
    top: 0;
    right: 0;
    z-index: 20;
    display: flex;
    width: min(100%, 18rem);
    height: 100%;
    flex-direction: column;
    border-left: 1px solid var(--color-border);
    background: var(--color-bg-elevated);
    padding-bottom: env(safe-area-inset-bottom);
  }

  @media (max-width: 767px) {
    .annotations-panel {
      top: auto;
      right: 0;
      left: 0;
      width: 100%;
      height: min(72dvh, 100%);
      border-left: none;
      border-top: 1px solid var(--color-border);
      border-radius: 0.75rem 0.75rem 0 0;
      box-shadow: var(--shadow);
    }
  }

  @media (min-width: 768px) and (max-width: 1023px) {
    .annotations-panel {
      width: min(100%, 20rem);
    }
  }
  .annotations-head {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 0.5rem;
    border-bottom: 1px solid var(--color-border);
    padding: 0.5rem;
  }
  .annotations-tab {
    display: inline-flex;
    align-items: center;
    gap: 0.25rem;
    border-radius: 0.375rem;
    padding: 0.25rem 0.5rem;
    font-size: 0.75rem;
    color: var(--color-muted);
  }
  .annotations-tab--active {
    background: var(--color-surface-hover);
    color: var(--color-fg);
  }
  .annotations-list {
    overflow-y: auto;
    flex: 1;
  }
  .annotations-item {
    display: flex;
    align-items: flex-start;
    gap: 0.25rem;
    border-bottom: 1px solid var(--color-border);
    padding: 0.5rem;
  }
  .annotations-item--stack {
    flex-direction: column;
  }
  .annotations-jump {
    flex: 1;
    text-align: left;
    font-size: 0.8125rem;
  }
  .annotations-jump:hover {
    color: var(--color-primary);
  }
  .highlight-excerpt {
    display: block;
    border-left: 3px solid var(--color-primary);
    padding-left: 0.5rem;
  }
</style>
