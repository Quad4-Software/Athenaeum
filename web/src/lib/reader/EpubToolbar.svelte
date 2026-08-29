<script lang="ts">
  import ReaderTools from "$lib/components/ReaderTools.svelte";
  import EpubDisplayMenu from "$lib/reader/EpubDisplayMenu.svelte";
  import EpubMoreMenu from "$lib/reader/EpubMoreMenu.svelte";
  import type { EpubFontId, StoredCustomFont } from "$lib/reader/epub-fonts";
  import type { ReaderTheme } from "$lib/reader/epub-theme";
  import type { EpubSpreadMode } from "$lib/reader/epub-reader";
  import type { ReaderChapter, ReaderSearchHit } from "$lib/reader/reader-search";

  interface Props {
    bookId?: number;
    chapters: ReaderChapter[];
    fontPct: number;
    fontId: EpubFontId;
    customFont: StoredCustomFont | null;
    spreadMode: EpubSpreadMode;
    selectionCfi: string;
    displayOpen: boolean;
    moreOpen: boolean;
    readerTheme: ReaderTheme;
    lineHeight: number;
    marginPx: number;
    onChapterSelect: (chapter: ReaderChapter) => void;
    onSearch: (query: string) => Promise<ReaderSearchHit[]>;
    onSearchSelect: (hit: ReaderSearchHit) => void;
    onSmallerFont: () => void;
    onLargerFont: () => void;
    onFontChange: (fontId: string) => void;
    onFontUpload: (event: Event) => void;
    onRemoveCustomFont: () => void;
    onSpreadMode: (mode: EpubSpreadMode) => void;
    onNarrate: () => void;
    onBookmark: () => void;
    onHighlight: () => void;
    onToggleAnnotations: () => void;
    onOpenShortcuts: () => void;
  }

  let {
    bookId,
    chapters,
    fontPct,
    fontId,
    customFont,
    spreadMode,
    selectionCfi,
    displayOpen = $bindable(false),
    moreOpen = $bindable(false),
    readerTheme = $bindable(),
    lineHeight = $bindable(),
    marginPx = $bindable(),
    onChapterSelect,
    onSearch,
    onSearchSelect,
    onSmallerFont,
    onLargerFont,
    onFontChange,
    onFontUpload,
    onRemoveCustomFont,
    onSpreadMode,
    onNarrate,
    onBookmark,
    onHighlight,
    onToggleAnnotations,
    onOpenShortcuts,
  }: Props = $props();
</script>

<div
  class="flex items-center justify-between gap-1 border-b border-border bg-bg/80 px-2 py-1.5 backdrop-blur sm:px-3"
>
  <div class="flex min-w-0 items-center gap-0.5 sm:gap-1">
    <ReaderTools {chapters} {onChapterSelect} {onSearch} {onSearchSelect} />
  </div>

  <div class="flex shrink-0 items-center gap-0.5 sm:gap-1">
    <EpubDisplayMenu
      bind:open={displayOpen}
      {fontPct}
      bind:readerTheme
      {fontId}
      {customFont}
      bind:lineHeight
      bind:marginPx
      {spreadMode}
      {onSmallerFont}
      {onLargerFont}
      {onFontChange}
      {onFontUpload}
      {onRemoveCustomFont}
      {onSpreadMode}
    />
    <EpubMoreMenu
      bind:open={moreOpen}
      {bookId}
      {selectionCfi}
      {onNarrate}
      {onBookmark}
      {onHighlight}
      {onToggleAnnotations}
      {onOpenShortcuts}
    />
  </div>
</div>
