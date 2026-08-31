<script lang="ts">
  import { untrack } from "svelte";
  import * as pdfjs from "pdfjs-dist";
  import type { PDFDocumentLoadingTask, PDFDocumentProxy, RenderTask } from "pdfjs-dist";
  import type { TextLayer } from "pdfjs-dist";
  import { pdfOpenOptions } from "$lib/reader/pdf-options";
  import "$lib/reader/PdfReader.css";
  import workerUrl from "pdfjs-dist/build/pdf.worker.min.mjs?url";
  import ReaderShortcuts from "$lib/components/ReaderShortcuts.svelte";
  import ReaderAnnotations from "$lib/components/ReaderAnnotations.svelte";
  import PdfToolbar from "$lib/reader/PdfToolbar.svelte";
  import PdfReaderSurface from "$lib/reader/PdfReaderSurface.svelte";
  import { handlePageKeys } from "$lib/reader/reader-keys";
  import {
    pdfPageCacheKey,
    prevSpreadStart,
    nextSpreadStart,
    clampSpreadStart,
    shouldCommitSpreadMount,
  } from "$lib/reader/reader-navigation";
  import { buildPreloadPlan, buildVisibleSpreadPlan } from "$lib/reader/pdf-spread-plan";
  import { PdfPageCache } from "$lib/reader/pdf-page-cache";
  import {
    bindPdfTextLayerEvents,
    createPdfPageWrap,
    renderPdfPageInto,
  } from "$lib/reader/pdf-page-render";
  import {
    clampPdfPage,
    nextPdfScale,
    parsePagesPerView,
    parsePdfJumpPage,
    pdfNextDisabled,
    pdfPageLabel,
    pdfPrevDisabled,
    pdfProgressRatio,
    pdfSelectionLocation,
    pdfSlotWidth,
    pdfZoomLabel,
    prevPdfScale,
    PDF_LAYOUT_PAD,
    PDF_SHORTCUT_ITEMS,
    type PagesPerView,
  } from "$lib/reader/pdf-reader";
  import {
    loadPdfChapters,
    searchPdf,
    type ReaderChapter,
    type ReaderSearchHit,
  } from "$lib/reader/reader-search";
  import { storageKey } from "$lib/brand/storage";
  import { api, ApiError } from "$lib/api/client";
  import { toast } from "$lib/stores/toast.svelte";
  import type { Highlight } from "$lib/api/types";

  const PAGES_KEY = storageKey("pdf-pages-per-view");

  pdfjs.GlobalWorkerOptions.workerSrc = workerUrl;

  interface Props {
    url: string;
    bookId?: number;
    initialPage?: number;
    onProgress?: (page: number, percent: number) => void;
  }

  let { url, bookId, initialPage = 1, onProgress }: Props = $props();

  let pagesRow = $state<HTMLDivElement>();
  let spreadPage = $state(untrack(() => initialPage));
  let doc = $state<PDFDocumentProxy | undefined>();
  let loadingTask: PDFDocumentLoadingTask | undefined;
  let renderTask: RenderTask | undefined;
  let textLayerTask: TextLayer | undefined;

  let page = $state(untrack(() => initialPage));
  let total = $state(0);
  let scale = $state(1);
  let loading = $state(true);
  let autoFit = $state(true);
  let fitTick = $state(0);
  let pagesPerView = $state<PagesPerView>(
    untrack(() =>
      typeof localStorage === "undefined" ? 1 : parsePagesPerView(localStorage.getItem(PAGES_KEY)),
    ),
  );
  let chapters = $state<ReaderChapter[]>([]);
  let viewport = $state<HTMLDivElement>();
  let selectionLocation = $state("");
  let selectionText = $state("");

  const pageCache = new PdfPageCache();
  let spreadMountId = 0;

  let shortcutsOpen = $state(false);
  let annotationsOpen = $state(false);
  let moreOpen = $state(false);
  let annotationsRevision = $state(0);
  let savedHighlights = $state<Highlight[]>([]);

  let pageLabel = $derived(pdfPageLabel(spreadPage, pagesPerView, total));
  let zoomShort = $derived(pdfZoomLabel(autoFit, scale, "short"));
  let zoomLong = $derived(pdfZoomLabel(autoFit, scale, "long"));
  let prevDisabled = $derived(pdfPrevDisabled(spreadPage));
  let nextDisabled = $derived(pdfNextDisabled(spreadPage, total));

  $effect(() => {
    const task = pdfjs.getDocument(pdfOpenOptions(url));
    loadingTask = task;
    loading = true;
    task.promise
      .then((d) => {
        doc = d;
        total = d.numPages;
        autoFit = true;
        scale = 1;
        page = clampPdfPage(
          untrack(() => initialPage),
          d.numPages,
        );
        spreadPage = page;
        loading = false;
        void loadPdfChapters(d).then((items) => {
          chapters = items;
        });
      })
      .catch(() => {
        loading = false;
        toast.error("Failed to load PDF");
      });

    return () => {
      renderTask?.cancel();
      textLayerTask?.cancel();
      void loadingTask?.destroy();
      pageCache.clear();
      doc = undefined;
      loadingTask = undefined;
    };
  });

  $effect(() => {
    void spreadPage;
    const targetScale = scale;
    const row = pagesRow;
    const outer = viewport;
    void fitTick;
    void pagesPerView;
    void annotationsRevision;
    void loading;
    if (loading || !doc || !row || !outer || total <= 0) return;

    const mountId = ++spreadMountId;
    let cancelled = false;
    const activeRenderTasks: RenderTask[] = [];
    renderTask?.cancel();
    textLayerTask?.cancel();

    const pad = PDF_LAYOUT_PAD;
    const slots = Math.min(pagesPerView, total - spreadPage + 1);
    if (slots <= 0) return;
    const slotWidth = pdfSlotWidth(outer.clientWidth, slots);
    const outerHeight = outer.clientHeight;
    const layout = {
      targetScale,
      autoFit,
      slotWidth,
      viewportHeight: outerHeight,
    };
    const renderLayout = {
      slotWidth,
      outerHeight,
      targetScale,
      autoFit,
      pad,
    };
    const visiblePlan = buildVisibleSpreadPlan(spreadPage, pagesPerView, total, layout, (key) =>
      pageCache.has(key),
    );
    if (visiblePlan.length === 0) return;

    const mountSpread = async () => {
      const staging = document.createDocumentFragment();
      const pending: Promise<void>[] = [];

      for (const entry of visiblePlan) {
        const cached = pageCache.take(entry.cacheKey);
        if (cached) {
          const layer = cached.querySelector(".pdf-text-layer") as HTMLDivElement | null;
          if (layer) bindPdfTextLayerEvents(layer, entry.page, captureSelectionForPage);
          staging.appendChild(cached);
          continue;
        }

        const wrap = createPdfPageWrap(entry.page, captureSelectionForPage);
        staging.appendChild(wrap);
        pending.push(
          renderPdfPageInto(
            wrap,
            entry.page,
            entry.cacheKey,
            doc,
            renderLayout,
            () => cancelled,
            activeRenderTasks,
            false,
            pageCache,
            savedHighlights,
            (task) => {
              renderTask = task;
            },
            (tl) => {
              textLayerTask = tl;
            },
          ),
        );
      }

      if (pending.length > 0) await Promise.all(pending);
      if (!shouldCommitSpreadMount(mountId, spreadMountId, cancelled)) return;

      row.replaceChildren(...Array.from(staging.childNodes));

      const lastPage = visiblePlan[visiblePlan.length - 1].page;
      page = lastPage;
      onProgress?.(lastPage, pdfProgressRatio(lastPage, total));

      for (const entry of buildPreloadPlan(spreadPage, pagesPerView, total, layout, (key) =>
        pageCache.has(key),
      )) {
        const wrap = createPdfPageWrap(entry.page, captureSelectionForPage);
        void renderPdfPageInto(
          wrap,
          entry.page,
          entry.cacheKey,
          doc,
          renderLayout,
          () => cancelled,
          activeRenderTasks,
          true,
          pageCache,
          savedHighlights,
          (task) => {
            renderTask = task;
          },
          (tl) => {
            textLayerTask = tl;
          },
        ).catch(() => undefined);
      }
    };

    void mountSpread();

    return () => {
      cancelled = true;
      for (const task of activeRenderTasks) task.cancel();
      textLayerTask?.cancel();
      if (row) {
        for (const child of row.children) {
          if (!(child instanceof HTMLDivElement)) continue;
          const pageNum = Number(child.dataset.page);
          if (!pageNum) continue;
          const cacheKey = pdfPageCacheKey(pageNum, targetScale, autoFit, slotWidth, outerHeight);
          if (!pageCache.has(cacheKey)) pageCache.set(cacheKey, child);
        }
      }
    };
  });

  function prev() {
    spreadPage = prevSpreadStart(spreadPage, pagesPerView);
    page = spreadPage;
  }
  function next() {
    spreadPage = nextSpreadStart(spreadPage, pagesPerView, total);
    page = spreadPage;
  }
  function zoomIn() {
    autoFit = false;
    scale = nextPdfScale(scale);
  }
  function zoomOut() {
    autoFit = false;
    scale = prevPdfScale(scale);
  }
  function resetFit() {
    autoFit = true;
    scale = 1;
  }

  function onKey(event: KeyboardEvent) {
    handlePageKeys(event, {
      prev,
      next,
      zoomIn,
      zoomOut,
      shortcuts: () => (shortcutsOpen = true),
    });
  }

  function setPagesPerView(value: PagesPerView) {
    pagesPerView = value;
    localStorage.setItem(PAGES_KEY, String(value));
    spreadPage = clampSpreadStart(spreadPage, value, total);
    page = spreadPage;
  }

  async function runSearch(query: string) {
    if (!doc) return [];
    return searchPdf(doc, query);
  }

  function jumpToChapter(chapter: ReaderChapter) {
    jumpTo(chapter.location);
  }

  function jumpToSearch(hit: ReaderSearchHit) {
    jumpTo(hit.location);
  }

  async function addBookmark() {
    if (!bookId) return;
    try {
      await api.createBookmark(bookId, String(page), `Page ${page}`);
      toast.success("Bookmark saved");
    } catch (e) {
      toast.error(e instanceof ApiError ? e.message : "Failed to save bookmark");
    }
  }

  function captureSelectionForPage(pageNum: number) {
    const sel = window.getSelection();
    const text = sel?.toString()?.trim() ?? "";
    if (!text) {
      selectionLocation = "";
      selectionText = "";
      return;
    }
    selectionText = text;
    selectionLocation = pdfSelectionLocation(pageNum, text);
  }

  $effect(() => {
    if (!bookId) return;
    void api.listHighlights(bookId).then((list) => {
      savedHighlights = list;
    });
  });

  async function addHighlight() {
    if (!bookId || !selectionLocation) return;
    try {
      const hl = await api.createHighlight(bookId, selectionLocation, selectionText);
      savedHighlights = [...savedHighlights, hl];
      toast.success("Highlight saved");
      selectionLocation = "";
      selectionText = "";
      window.getSelection()?.removeAllRanges();
      annotationsRevision += 1;
      fitTick += 1;
    } catch (e) {
      toast.error(e instanceof ApiError ? e.message : "Failed to save highlight");
    }
  }

  function jumpTo(location: string) {
    const target = parsePdfJumpPage(location, total);
    if (target != null) {
      spreadPage = target;
      page = target;
    }
    annotationsOpen = false;
  }

  $effect(() => {
    const wrap = viewport;
    if (!wrap) return;
    const ro = new ResizeObserver(() => {
      if (autoFit) fitTick += 1;
    });
    ro.observe(wrap);
    return () => ro.disconnect();
  });
</script>

<svelte:window onkeydown={onKey} />

<ReaderShortcuts
  open={shortcutsOpen}
  onclose={() => (shortcutsOpen = false)}
  items={PDF_SHORTCUT_ITEMS}
/>

<div class="relative flex h-full flex-col">
  <PdfToolbar
    {bookId}
    {pageLabel}
    {zoomShort}
    {zoomLong}
    {autoFit}
    {pagesPerView}
    {prevDisabled}
    {nextDisabled}
    {selectionLocation}
    {chapters}
    bind:moreOpen
    onSearch={doc ? runSearch : undefined}
    onPrev={prev}
    onNext={next}
    onZoomOut={zoomOut}
    onZoomIn={zoomIn}
    onResetFit={resetFit}
    onPagesPerView={setPagesPerView}
    onChapterSelect={jumpToChapter}
    onSearchSelect={jumpToSearch}
    onOpenShortcuts={() => (shortcutsOpen = true)}
    onBookmark={() => void addBookmark()}
    onHighlight={() => void addHighlight()}
    onToggleAnnotations={() => (annotationsOpen = !annotationsOpen)}
  />

  <PdfReaderSurface {loading} onPrev={prev} onNext={next} bind:viewport bind:pagesRow />

  {#if bookId}
    <ReaderAnnotations
      {bookId}
      open={annotationsOpen}
      revision={annotationsRevision}
      locationType="page"
      onclose={() => (annotationsOpen = false)}
      onJump={jumpTo}
    />
  {/if}
</div>
