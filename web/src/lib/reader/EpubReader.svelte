<script lang="ts">
  import ePub, { type Rendition } from "epubjs";
  import {
    epubOpenOptions,
    epubRenderOptions,
    ensureEpubDocumentHead,
  } from "$lib/reader/epub-options";
  import { theme } from "$lib/stores/theme.svelte";
  import { api } from "$lib/api/client";
  import { toast } from "$lib/stores/toast.svelte";
  import { i18n } from "$lib/stores/i18n.svelte";
  import { narrator } from "$lib/stores/narrator.svelte";
  import { audioPlayer } from "$lib/stores/audioPlayer.svelte";
  import { buildUtteranceQueue, paragraphsFromContents } from "$lib/narrator/text";
  import { isBrowserTTSAvailable } from "$lib/narrator/browser";
  import ReaderShortcuts from "$lib/components/ReaderShortcuts.svelte";
  import ReaderAnnotations from "$lib/components/ReaderAnnotations.svelte";
  import EpubToolbar from "$lib/reader/EpubToolbar.svelte";
  import EpubReaderSurface from "$lib/reader/EpubReaderSurface.svelte";
  import { handlePageKeys } from "$lib/reader/reader-keys";
  import { bindDocumentGestures } from "$lib/reader/reader-touch";
  import {
    loadEpubChapters,
    searchEpub,
    type ReaderChapter,
    type ReaderSearchHit,
  } from "$lib/reader/reader-search";
  import type { Book } from "epubjs";
  import { storageKey } from "$lib/brand/storage";
  import {
    applyEpubFont,
    clearCustomFont,
    loadCustomFont,
    loadFontPreference,
    saveCustomFont,
    saveFontPreference,
    type EpubFontId,
    type StoredCustomFont,
  } from "$lib/reader/epub-fonts";
  import {
    EPUB_PALETTE,
    injectEpubContentBackground,
    resolvedEpubTheme,
    type ReaderTheme,
  } from "$lib/reader/epub-theme";
  import { patchEpubArchiveEncoding } from "$lib/reader/epub-encoding";
  import {
    applyEpubFontPct,
    applyEpubSurfaceBackground,
    applyEpubThemeOverrides,
    buildEpubPrefKeys,
    createEpubPrefsSaver,
    canSelectEpubFont,
    decideEpubNarration,
    epubFontUploadErrorKey,
    epubjsSpread,
    epubLoadErrorMessage,
    epubNarrationParagraphs,
    epubPageKeyHandlers,
    epubPercentFromCfi,
    EPUB_SHORTCUT_ITEMS,
    errorMessage,
    loadAndPaintEpubHighlights,
    loadInitialEpubDisplayPrefs,
    mergeRemoteEpubPrefs,
    nextEpubFontPct,
    paintEpubHighlight,
    persistEpubDisplayPrefs,
    persistEpubSpreadMode,
    persistEpubThemePrefs,
    preloadEpubSpineSections,
    prevEpubFontPct,
    readEpubSelectionText,
    resolveLoadedCustomFont,
    scheduleEpubLocationsGenerate,
    takeFileInput,
    type EpubSpreadMode,
  } from "$lib/reader/epub-reader";

  const PREF_KEYS = buildEpubPrefKeys(storageKey);
  const palette = EPUB_PALETTE;
  const initialPrefs = loadInitialEpubDisplayPrefs(
    typeof localStorage !== "undefined" ? localStorage : null,
    PREF_KEYS,
  );

  interface Props {
    url: string;
    bookId?: number;
    title?: string;
    initialLocation?: string;
    onProgress?: (location: string, percent: number) => void;
  }

  let { url, bookId, title = "", initialLocation, onProgress }: Props = $props();

  let container = $state<HTMLDivElement>();
  let rendition: Rendition | undefined;
  let currentCfi = $state("");
  let ready = $state(false);
  let loadError = $state<string | null>(null);
  let fontPct = $state(initialPrefs.fontPct);
  let readerTheme = $state<ReaderTheme>(initialPrefs.theme);
  let lineHeight = $state(initialPrefs.lineHeight);
  let marginPx = $state(initialPrefs.marginPx);
  let spreadMode = $state<EpubSpreadMode>(initialPrefs.spread);
  let prefsLoaded = $state(false);
  let epubBook = $state<Book | undefined>();
  let chapters = $state<ReaderChapter[]>([]);
  let fontId = $state<EpubFontId>(loadFontPreference());
  let customFont = $state<StoredCustomFont | null>(null);
  let shortcutsOpen = $state(false);
  let displayOpen = $state(false);
  let moreOpen = $state(false);
  let annotationsOpen = $state(false);
  let annotationsRevision = $state(0);
  let selectionCfi = $state("");
  let selectionText = $state("");

  const prefsSaver = createEpubPrefsSaver({
    isReady: () => prefsLoaded,
    getPrefs: () => ({
      fontPct,
      theme: readerTheme,
      lineHeight,
      marginPx,
      spread: spreadMode,
    }),
    save: (prefs) => api.saveReaderPrefs(prefs),
  });

  $effect(() => {
    void api
      .getReaderPrefs()
      .then((res) => {
        const merged = mergeRemoteEpubPrefs(
          { fontPct, theme: readerTheme, lineHeight, marginPx, spread: spreadMode },
          res.prefs,
        );
        fontPct = merged.fontPct;
        readerTheme = merged.theme;
        lineHeight = merged.lineHeight;
        marginPx = merged.marginPx;
        spreadMode = merged.spread;
        persistEpubDisplayPrefs(PREF_KEYS, merged);
      })
      .catch(() => undefined)
      .finally(() => {
        prefsLoaded = true;
      });
  });

  function pageKeyHandlers() {
    return epubPageKeyHandlers({
      prev,
      next,
      largerFont,
      smallerFont,
      openShortcuts: () => (shortcutsOpen = true),
    });
  }

  function resolvedTheme(): ReaderTheme {
    return resolvedEpubTheme(readerTheme, theme.mode === "dark");
  }

  const epubSurfaceBg = $derived(palette[resolvedTheme()].bg);

  function applyReaderTheme() {
    if (!rendition) return;
    const { fg, bg } = palette[resolvedTheme()];
    applyEpubThemeOverrides(rendition.themes, { fg, bg, lineHeight, marginPx });
    applyEpubSurfaceBackground(container, () => rendition?.getContents(), bg);
    persistEpubThemePrefs(PREF_KEYS, { theme: readerTheme, lineHeight, marginPx });
    prefsSaver.queue();
  }

  function applyFontFamily() {
    if (!rendition) return;
    applyEpubFont(rendition, fontId, customFont);
  }

  function applyFont(pct = fontPct) {
    if (!rendition) return;
    applyEpubFontPct(rendition.themes, pct, PREF_KEYS.font);
    prefsSaver.queue();
  }

  async function addHighlight() {
    if (!bookId || !selectionCfi) return;
    try {
      const hl = await api.createHighlight(bookId, selectionCfi, selectionText);
      if (rendition)
        paintEpubHighlight(rendition.annotations, hl.location, hl.id, hl.color || "yellow");
      toast.success("Highlight saved");
      selectionCfi = "";
      selectionText = "";
      annotationsRevision += 1;
    } catch {
      toast.error("Failed to save highlight");
    }
  }

  function jumpTo(location: string) {
    rendition?.display(location);
    annotationsOpen = false;
  }

  async function runSearch(query: string) {
    if (!epubBook) return [];
    return searchEpub(epubBook, query);
  }

  function setSpreadMode(mode: EpubSpreadMode) {
    spreadMode = mode;
    persistEpubSpreadMode(PREF_KEYS.spread, mode);
    prefsSaver.queue();
  }

  function onFontChange(value: string) {
    if (!canSelectEpubFont(value, !!customFont)) return;
    fontId = value as EpubFontId;
    saveFontPreference(value as EpubFontId);
  }

  async function onFontUpload(event: Event) {
    const file = takeFileInput(event);
    if (!file) return;
    try {
      customFont = await saveCustomFont(file);
      fontId = "custom";
      saveFontPreference("custom");
      toast.success(i18n.t("reader.fontUploaded"));
    } catch (e) {
      toast.error(i18n.t(epubFontUploadErrorKey(errorMessage(e))));
    }
  }

  async function removeCustomFont() {
    await clearCustomFont();
    customFont = null;
    if (fontId === "custom") {
      fontId = "book";
      saveFontPreference("book");
    }
  }

  async function addBookmark() {
    if (!bookId || !currentCfi) return;
    try {
      await api.createBookmark(bookId, currentCfi, i18n.t("reader.bookmarkLabel"));
      toast.success(i18n.t("reader.bookmarkSaved"));
    } catch {
      toast.error(i18n.t("reader.bookmarkFailed"));
    }
  }

  $effect(() => {
    void loadCustomFont().then((stored) => {
      customFont = stored;
      const preference = loadFontPreference();
      const resolved = resolveLoadedCustomFont({ preference, hasCustomFont: !!stored });
      fontId = resolved.fontId as EpubFontId;
      if (resolved.shouldResetPreference) saveFontPreference("book");
    });
  });

  $effect(() => {
    void fontId;
    void customFont;
    applyFontFamily();
  });

  $effect(() => {
    void spreadMode;
    const node = container;
    const startLocation = initialLocation;
    const spread = epubjsSpread(spreadMode);
    if (!node) return;

    loadError = null;
    ready = false;
    chapters = [];

    const book = ePub(url, epubOpenOptions());
    book.spine.hooks.content.register((doc: Document) => {
      ensureEpubDocumentHead(doc);
    });
    epubBook = book;
    const r = book.renderTo(node, epubRenderOptions(spread));
    rendition = r;
    r.on("keydown", (event: KeyboardEvent) => {
      handlePageKeys(event, pageKeyHandlers());
    });

    r.hooks.content.register((contents: { document: Document }) => {
      injectEpubContentBackground(contents.document, palette[resolvedTheme()].bg);
      bindDocumentGestures(contents.document, {
        onSwipeLeft: next,
        onSwipeRight: prev,
        onTapLeft: prev,
        onTapRight: next,
      });
    });

    let cancelled = false;

    const fail = (message?: string) => {
      if (cancelled) return;
      const text = epubLoadErrorMessage(message, i18n.t("reader.loadFailed"));
      loadError = text;
      toast.error(text);
    };

    book.on("openFailed", (err: Error) => {
      fail(err?.message);
    });

    void book.opened
      .then(() => {
        patchEpubArchiveEncoding(book);
        return r.display(startLocation || undefined);
      })
      .then(() => {
        if (cancelled) return;
        ready = true;
        applyReaderTheme();
        applyFont();
        applyFontFamily();
        if (bookId) {
          const id = bookId;
          void loadAndPaintEpubHighlights(r.annotations, () => api.listHighlights(id));
        }
        void loadEpubChapters(book).then((items) => {
          if (!cancelled) chapters = items;
        });
      })
      .catch((err: unknown) => {
        fail(err instanceof Error ? err.message : undefined);
      });

    book.ready
      .then(() => {
        scheduleEpubLocationsGenerate(() => {
          if (cancelled) return;
          void book.locations.generate(2048).catch(() => undefined);
        });
      })
      .catch(() => undefined);

    r.on("relocated", (location: { start: { cfi: string; index: number } }) => {
      currentCfi = location.start.cfi;
      preloadEpubSpineSections(book, location.start.index);
      onProgress?.(
        location.start.cfi,
        epubPercentFromCfi((value) => book.locations.percentageFromCfi(value), location.start.cfi),
      );
    });

    r.on("selected", (cfiRange: string) => {
      selectionCfi = cfiRange;
      selectionText = readEpubSelectionText(() => r.getContents().window?.getSelection());
    });

    return () => {
      cancelled = true;
      narrator.stop();
      r.destroy();
      book.destroy();
      rendition = undefined;
      epubBook = undefined;
      ready = false;
      loadError = null;
    };
  });

  $effect(() => {
    void theme.mode;
    void readerTheme;
    void lineHeight;
    void marginPx;
    applyReaderTheme();
  });

  $effect(() => {
    applyFont(fontPct);
  });

  function prev() {
    rendition?.prev();
  }
  function next() {
    rendition?.next();
  }
  function smallerFont() {
    fontPct = prevEpubFontPct(fontPct);
  }
  function largerFont() {
    fontPct = nextEpubFontPct(fontPct);
  }

  async function startNarration() {
    if (!rendition) return;
    if (audioPlayer.active) audioPlayer.stop();
    if (narrator.active) {
      narrator.stop();
      return;
    }
    await narrator.refreshStatus();
    const decision = decideEpubNarration({
      narratorActive: narrator.active,
      provider: narrator.provider,
      kokoroEnabled: narrator.kokoroEnabled,
      browserAvailable: isBrowserTTSAvailable(),
    });
    if (decision.kind === "toggle-off") {
      narrator.stop();
      return;
    }
    if (decision.kind === "error") {
      toast.error(i18n.t(decision.key));
      return;
    }
    if (decision.switchToBrowser) narrator.setProvider("browser");
    const paragraphs = epubNarrationParagraphs(
      () =>
        rendition!.getContents() as
          | { document?: Document; window?: Window }
          | { document?: Document; window?: Window }[]
          | undefined,
      paragraphsFromContents,
    );
    await narrator.start(buildUtteranceQueue(paragraphs), { title });
  }

  function onKey(event: KeyboardEvent) {
    handlePageKeys(event, pageKeyHandlers());
  }
</script>

<svelte:window onkeydown={onKey} />

<ReaderShortcuts
  open={shortcutsOpen}
  onclose={() => (shortcutsOpen = false)}
  items={[...EPUB_SHORTCUT_ITEMS]}
/>

<div class="relative flex h-full w-full flex-col">
  <EpubToolbar
    {bookId}
    {chapters}
    {fontPct}
    {fontId}
    {customFont}
    {spreadMode}
    {selectionCfi}
    bind:displayOpen
    bind:moreOpen
    bind:readerTheme
    bind:lineHeight
    bind:marginPx
    onChapterSelect={(c: ReaderChapter) => jumpTo(c.location)}
    onSearch={runSearch}
    onSearchSelect={(h: ReaderSearchHit) => jumpTo(h.location)}
    onSmallerFont={smallerFont}
    onLargerFont={largerFont}
    {onFontChange}
    {onFontUpload}
    onRemoveCustomFont={() => void removeCustomFont()}
    onSpreadMode={setSpreadMode}
    onNarrate={startNarration}
    onBookmark={addBookmark}
    onHighlight={addHighlight}
    onToggleAnnotations={() => (annotationsOpen = !annotationsOpen)}
    onOpenShortcuts={() => (shortcutsOpen = true)}
  />

  <EpubReaderSurface
    {spreadMode}
    surfaceBg={epubSurfaceBg}
    {ready}
    {loadError}
    onPrev={prev}
    onNext={next}
    bind:container
  />

  {#if bookId}
    <ReaderAnnotations
      {bookId}
      open={annotationsOpen}
      revision={annotationsRevision}
      onclose={() => (annotationsOpen = false)}
      onJump={jumpTo}
    />
  {/if}
</div>

<style src="./EpubReader.css"></style>
