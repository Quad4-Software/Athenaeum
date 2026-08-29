<script lang="ts">
  import {
    ChevronLeft,
    ChevronRight,
    ZoomIn,
    ZoomOut,
    Bookmark,
    Highlighter,
    List,
  } from "@lucide/svelte";
  import ReaderTools from "$lib/components/ReaderTools.svelte";
  import PdfMoreMenu from "$lib/reader/PdfMoreMenu.svelte";
  import type { PagesPerView } from "$lib/reader/pdf-reader";
  import type { ReaderChapter, ReaderSearchHit } from "$lib/reader/reader-search";

  interface Props {
    bookId?: number;
    pageLabel: string;
    zoomShort: string;
    zoomLong: string;
    autoFit: boolean;
    pagesPerView: PagesPerView;
    prevDisabled: boolean;
    nextDisabled: boolean;
    selectionLocation: string;
    chapters: ReaderChapter[];
    moreOpen: boolean;
    onSearch?: (query: string) => Promise<ReaderSearchHit[]>;
    onPrev: () => void;
    onNext: () => void;
    onZoomOut: () => void;
    onZoomIn: () => void;
    onResetFit: () => void;
    onPagesPerView: (value: PagesPerView) => void;
    onChapterSelect: (chapter: ReaderChapter) => void;
    onSearchSelect: (hit: ReaderSearchHit) => void;
    onOpenShortcuts: () => void;
    onBookmark: () => void;
    onHighlight: () => void;
    onToggleAnnotations: () => void;
  }

  let {
    bookId,
    pageLabel,
    zoomShort,
    zoomLong,
    autoFit,
    pagesPerView,
    prevDisabled,
    nextDisabled,
    selectionLocation,
    chapters,
    moreOpen = $bindable(false),
    onSearch,
    onPrev,
    onNext,
    onZoomOut,
    onZoomIn,
    onResetFit,
    onPagesPerView,
    onChapterSelect,
    onSearchSelect,
    onOpenShortcuts,
    onBookmark,
    onHighlight,
    onToggleAnnotations,
  }: Props = $props();
</script>

<div
  class="flex items-center justify-between gap-1 border-b border-border bg-bg/80 px-2 py-1.5 backdrop-blur md:hidden"
>
  <button
    class="btn btn-ghost shrink-0"
    aria-label="Previous page"
    onclick={onPrev}
    disabled={prevDisabled}
  >
    <ChevronLeft size={18} />
  </button>
  <span class="min-w-0 truncate text-center text-xs tabular-nums text-muted sm:text-sm">
    {pageLabel}
  </span>
  <div class="flex shrink-0 items-center gap-0.5">
    <button class="btn btn-ghost" aria-label="Next page" onclick={onNext} disabled={nextDisabled}>
      <ChevronRight size={18} />
    </button>
    <ReaderTools {chapters} {onChapterSelect} {onSearch} {onSearchSelect} />
    <PdfMoreMenu
      bind:open={moreOpen}
      {bookId}
      {autoFit}
      {zoomLong}
      {pagesPerView}
      {selectionLocation}
      {onZoomOut}
      {onZoomIn}
      {onResetFit}
      {onPagesPerView}
      {onOpenShortcuts}
      {onBookmark}
      {onHighlight}
      {onToggleAnnotations}
    />
  </div>
</div>

<div
  class="hidden items-center justify-center gap-2 border-b border-border bg-bg/80 px-3 py-2 text-sm md:flex md:flex-wrap"
>
  <button class="btn btn-ghost" aria-label="Previous page" onclick={onPrev} disabled={prevDisabled}>
    <ChevronLeft size={18} />
  </button>
  <span class="tabular-nums text-muted">{pageLabel}</span>
  <button class="btn btn-ghost" aria-label="Next page" onclick={onNext} disabled={nextDisabled}>
    <ChevronRight size={18} />
  </button>
  <span class="mx-2 h-5 w-px bg-border"></span>
  <button class="btn btn-ghost" aria-label="Zoom out" onclick={onZoomOut}>
    <ZoomOut size={18} />
  </button>
  <button
    class="btn btn-ghost text-xs {autoFit ? 'text-primary' : ''}"
    aria-label="Fit page"
    onclick={onResetFit}
  >
    Fit
  </button>
  <span class="tabular-nums text-muted">{zoomShort}</span>
  <button class="btn btn-ghost" aria-label="Zoom in" onclick={onZoomIn}>
    <ZoomIn size={18} />
  </button>
  <select
    class="field-input w-auto py-1 text-xs"
    value={String(pagesPerView)}
    onchange={(e) => onPagesPerView(Number(e.currentTarget.value) as PagesPerView)}
  >
    <option value="1">1 page</option>
    <option value="2">2 pages</option>
    <option value="3">3 pages</option>
  </select>
  <ReaderTools {chapters} {onChapterSelect} {onSearch} {onSearchSelect} />
  <button class="btn btn-ghost text-xs" aria-label="Shortcuts" onclick={onOpenShortcuts}>
    ?
  </button>
  {#if bookId}
    <button class="btn btn-ghost text-xs" aria-label="Bookmark page" onclick={onBookmark}>
      <Bookmark size={16} />
    </button>
    <button
      class="btn btn-ghost text-xs"
      aria-label="Highlight selection"
      onclick={onHighlight}
      disabled={!selectionLocation}
    >
      <Highlighter size={16} />
    </button>
    <button class="btn btn-ghost text-xs" aria-label="Annotations" onclick={onToggleAnnotations}>
      <List size={16} />
    </button>
  {/if}
</div>
