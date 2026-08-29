<script lang="ts">
  import { Bookmark, Highlighter, List, MoreVertical, ZoomIn, ZoomOut } from "@lucide/svelte";
  import Popover from "$lib/components/Popover.svelte";
  import type { PagesPerView } from "$lib/reader/pdf-reader";

  interface Props {
    open: boolean;
    bookId?: number;
    autoFit: boolean;
    zoomLong: string;
    pagesPerView: PagesPerView;
    selectionLocation: string;
    onZoomOut: () => void;
    onZoomIn: () => void;
    onResetFit: () => void;
    onPagesPerView: (value: PagesPerView) => void;
    onOpenShortcuts: () => void;
    onBookmark: () => void;
    onHighlight: () => void;
    onToggleAnnotations: () => void;
  }

  let {
    open = $bindable(false),
    bookId,
    autoFit,
    zoomLong,
    pagesPerView,
    selectionLocation,
    onZoomOut,
    onZoomIn,
    onResetFit,
    onPagesPerView,
    onOpenShortcuts,
    onBookmark,
    onHighlight,
    onToggleAnnotations,
  }: Props = $props();
</script>

<Popover bind:open placement="bottom" align="end" minWidth={220}>
  {#snippet trigger(toggle)}
    <button
      type="button"
      class="btn btn-ghost"
      class:ring-1={open}
      class:ring-border={open}
      aria-expanded={open}
      aria-label="More options"
      onclick={toggle}
    >
      <MoreVertical size={18} />
    </button>
  {/snippet}
  <div class="flex flex-col gap-1 p-1">
    <div class="flex items-center justify-between gap-2 px-2 py-1">
      <button class="btn btn-ghost flex-1" aria-label="Zoom out" onclick={onZoomOut}>
        <ZoomOut size={16} />
      </button>
      <button
        class="btn btn-ghost flex-1 text-xs {autoFit ? 'text-primary' : ''}"
        aria-label="Fit page"
        onclick={onResetFit}
      >
        Fit
      </button>
      <button class="btn btn-ghost flex-1" aria-label="Zoom in" onclick={onZoomIn}>
        <ZoomIn size={16} />
      </button>
    </div>
    <p class="px-2 text-center text-xs tabular-nums text-muted">
      {zoomLong}
    </p>
    <label class="block px-2 py-1">
      <span class="mb-1 block text-xs text-muted">Pages per view</span>
      <select
        class="field-input w-full py-1 text-xs"
        value={String(pagesPerView)}
        onchange={(e) => onPagesPerView(Number(e.currentTarget.value) as PagesPerView)}
      >
        <option value="1">1 page</option>
        <option value="2">2 pages</option>
        <option value="3">3 pages</option>
      </select>
    </label>
    <button
      class="btn btn-ghost w-full justify-start text-xs"
      aria-label="Shortcuts"
      onclick={() => {
        open = false;
        onOpenShortcuts();
      }}
    >
      Keyboard shortcuts
    </button>
    {#if bookId}
      <button
        class="btn btn-ghost w-full justify-start text-xs"
        aria-label="Bookmark page"
        onclick={() => {
          open = false;
          onBookmark();
        }}
      >
        <Bookmark size={14} /> Bookmark
      </button>
      <button
        class="btn btn-ghost w-full justify-start text-xs"
        aria-label="Highlight selection"
        onclick={() => {
          open = false;
          onHighlight();
        }}
        disabled={!selectionLocation}
      >
        <Highlighter size={14} /> Highlight
      </button>
      <button
        class="btn btn-ghost w-full justify-start text-xs"
        aria-label="Annotations"
        onclick={() => {
          open = false;
          onToggleAnnotations();
        }}
      >
        <List size={14} /> Annotations
      </button>
    {/if}
  </div>
</Popover>
