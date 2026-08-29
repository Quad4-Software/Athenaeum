<script lang="ts">
  import {
    Bookmark,
    ChevronLeft,
    ChevronRight,
    Columns2,
    FlipHorizontal2,
    Images,
    List,
    Maximize,
    StretchHorizontal,
    StretchVertical,
  } from "@lucide/svelte";
  import { api, ApiError } from "$lib/api/client";
  import EmptyState from "$lib/components/EmptyState.svelte";
  import Popover from "$lib/components/Popover.svelte";
  import ReaderAnnotations from "$lib/components/ReaderAnnotations.svelte";
  import { readerGestures } from "$lib/reader/reader-touch";
  import {
    comicFitClass,
    comicSpreadPages,
    nextComicPage,
    prevComicPage,
    type ComicFit,
  } from "$lib/reader/comic-reader";
  import { bookOfflineCache } from "$lib/offline/book-cache";
  import { storageKey } from "$lib/brand/storage";
  import { i18n } from "$lib/stores/i18n.svelte";
  import { toast } from "$lib/stores/toast.svelte";

  const FIT_KEY = storageKey("comic-fit");
  const SPREAD_KEY = storageKey("comic-spread");
  const RTL_KEY = storageKey("comic-rtl");
  const WIDE_QUERY = "(min-width: 768px)";

  interface Props {
    bookId: number;
    initialPage?: number;
    onProgress?: (page: number, percent: number) => void;
  }

  let { bookId, initialPage = 0, onProgress }: Props = $props();

  let total = $state(0);
  let page = $state(0);
  let loading = $state(true);
  let loadedCount = $state(0);
  let displayOpen = $state(false);
  let annotationsOpen = $state(false);
  let annotationsRevision = $state(0);
  let pageSrc = $state<Record<number, string>>({});
  let wide = $state(typeof window !== "undefined" ? window.matchMedia(WIDE_QUERY).matches : false);

  let fit = $state<ComicFit>(
    (typeof localStorage !== "undefined" ? (localStorage.getItem(FIT_KEY) as ComicFit) : null) ||
      "contain",
  );
  let spreadEnabled = $state(
    typeof localStorage !== "undefined" ? localStorage.getItem(SPREAD_KEY) === "1" : false,
  );
  let rtl = $state(
    typeof localStorage !== "undefined" ? localStorage.getItem(RTL_KEY) === "1" : false,
  );

  let spreadPages = $derived(comicSpreadPages(page, total, spreadEnabled, wide));
  let displayPages = $derived(rtl ? [...spreadPages].reverse() : spreadPages);
  let spreadKey = $derived(spreadPages.join("-"));
  let prefetchPages = $derived(
    Array.from(
      new Set(
        [
          ...comicSpreadPages(
            prevComicPage(page, total, spreadEnabled, wide),
            total,
            spreadEnabled,
            wide,
          ),
          ...comicSpreadPages(
            nextComicPage(page, total, spreadEnabled, wide),
            total,
            spreadEnabled,
            wide,
          ),
        ].filter((p) => !spreadPages.includes(p)),
      ),
    ),
  );

  $effect(() => {
    const id = bookId;
    loading = true;
    api
      .getComicManifest(id)
      .then((m) => {
        total = m.total;
        page = Math.min(Math.max(initialPage, 0), Math.max(m.total - 1, 0));
        loading = false;
        report();
      })
      .catch(() => {
        loading = false;
      });
  });

  $effect(() => {
    if (typeof window === "undefined") return;
    const mq = window.matchMedia(WIDE_QUERY);
    const apply = () => (wide = mq.matches);
    apply();
    mq.addEventListener("change", apply);
    return () => mq.removeEventListener("change", apply);
  });

  $effect(() => {
    void spreadKey;
    loadedCount = 0;
  });

  $effect(() => {
    const id = bookId;
    const pages = [...displayPages, ...prefetchPages];
    let cancelled = false;
    for (const p of pages) {
      const online = api.comicPageUrl(id, p);
      void bookOfflineCache.resolvePageUrl(id, p, online).then((url) => {
        if (cancelled) return;
        pageSrc = { ...pageSrc, [p]: url };
      });
    }
    return () => {
      cancelled = true;
    };
  });

  function pageUrl(p: number): string {
    return pageSrc[p] ?? api.comicPageUrl(bookId, p);
  }

  function report() {
    onProgress?.(page, total ? (page + 1) / total : 0);
  }

  function prev() {
    const target = prevComicPage(page, total, spreadEnabled, wide);
    if (target === page) return;
    page = target;
    report();
  }

  function next() {
    const target = nextComicPage(page, total, spreadEnabled, wide);
    if (target === page) return;
    page = target;
    report();
  }

  function onKey(event: KeyboardEvent) {
    if (event.key === "ArrowLeft") (rtl ? next : prev)();
    if (event.key === "ArrowRight") (rtl ? prev : next)();
  }

  function setFit(value: ComicFit) {
    fit = value;
    localStorage.setItem(FIT_KEY, value);
  }

  function toggleSpread() {
    spreadEnabled = !spreadEnabled;
    localStorage.setItem(SPREAD_KEY, spreadEnabled ? "1" : "0");
  }

  function toggleRtl() {
    rtl = !rtl;
    localStorage.setItem(RTL_KEY, rtl ? "1" : "0");
  }

  async function addBookmark() {
    if (!bookId) return;
    try {
      const label = `${i18n.t("reader.bookmarkLabel")} ${page + 1}`;
      await api.createBookmark(bookId, String(page + 1), label);
      toast.success(i18n.t("reader.bookmarkSaved"));
      annotationsRevision += 1;
    } catch (e) {
      toast.error(e instanceof ApiError ? e.message : i18n.t("reader.bookmarkFailed"));
    }
  }

  function jumpTo(location: string) {
    const n = Number(location);
    if (Number.isFinite(n) && n >= 1 && n <= total) {
      page = n - 1;
      report();
    }
    annotationsOpen = false;
  }
</script>

<svelte:window onkeydown={onKey} />

<div class="flex h-full flex-col bg-bg-elevated">
  <div
    class="flex items-center justify-center gap-2 border-b border-border bg-bg/80 px-2 py-1.5 text-sm sm:gap-3 sm:px-3 sm:py-2"
  >
    <button
      class="btn btn-ghost"
      aria-label={rtl ? "Next page" : "Previous page"}
      onclick={rtl ? next : prev}
      disabled={rtl ? total === 0 || page >= total - 1 : page <= 0}
    >
      <ChevronLeft size={18} />
    </button>
    <span class="tabular-nums text-muted">{total ? page + 1 : 0} / {total}</span>
    <button
      class="btn btn-ghost"
      aria-label={rtl ? "Previous page" : "Next page"}
      onclick={rtl ? prev : next}
      disabled={rtl ? page <= 0 : total === 0 || page >= total - 1}
    >
      <ChevronRight size={18} />
    </button>

    <span class="mx-1 h-5 w-px bg-border"></span>

    <Popover bind:open={displayOpen} placement="bottom" align="end" minWidth={220}>
      {#snippet trigger(toggle)}
        <button
          type="button"
          class="btn btn-ghost text-xs"
          class:ring-1={displayOpen}
          class:ring-border={displayOpen}
          aria-expanded={displayOpen}
          aria-label={i18n.t("reader.display")}
          onclick={toggle}
        >
          <Columns2 size={16} />
        </button>
      {/snippet}

      <div class="space-y-3 p-1">
        <div>
          <p class="mb-1.5 text-xs font-medium text-muted">{i18n.t("reader.comicFit")}</p>
          <div class="flex items-center gap-1">
            <button
              type="button"
              class="btn btn-ghost flex-1"
              class:ring-1={fit === "contain"}
              class:ring-border={fit === "contain"}
              class:text-primary={fit === "contain"}
              aria-pressed={fit === "contain"}
              aria-label={i18n.t("reader.comicFitContain")}
              onclick={() => setFit("contain")}
            >
              <Maximize size={16} />
            </button>
            <button
              type="button"
              class="btn btn-ghost flex-1"
              class:ring-1={fit === "width"}
              class:ring-border={fit === "width"}
              class:text-primary={fit === "width"}
              aria-pressed={fit === "width"}
              aria-label={i18n.t("reader.comicFitWidth")}
              onclick={() => setFit("width")}
            >
              <StretchHorizontal size={16} />
            </button>
            <button
              type="button"
              class="btn btn-ghost flex-1"
              class:ring-1={fit === "height"}
              class:ring-border={fit === "height"}
              class:text-primary={fit === "height"}
              aria-pressed={fit === "height"}
              aria-label={i18n.t("reader.comicFitHeight")}
              onclick={() => setFit("height")}
            >
              <StretchVertical size={16} />
            </button>
          </div>
        </div>

        <div class="flex items-center justify-between gap-2">
          <span class="text-xs font-medium text-muted">{i18n.t("reader.comicDualPage")}</span>
          <button
            type="button"
            class="btn btn-ghost text-xs ring-1 ring-border"
            class:text-primary={spreadEnabled}
            aria-pressed={spreadEnabled}
            aria-label={i18n.t("reader.comicDualPage")}
            onclick={toggleSpread}
          >
            <Columns2 size={16} />
          </button>
        </div>

        <div class="flex items-center justify-between gap-2">
          <span class="text-xs font-medium text-muted">{i18n.t("reader.comicDirection")}</span>
          <button
            type="button"
            class="btn btn-ghost text-xs ring-1 ring-border"
            class:text-primary={rtl}
            aria-pressed={rtl}
            aria-label={i18n.t("reader.comicDirection")}
            onclick={toggleRtl}
          >
            <FlipHorizontal2 size={16} />
            {rtl ? i18n.t("reader.comicDirectionRtl") : i18n.t("reader.comicDirectionLtr")}
          </button>
        </div>
      </div>
    </Popover>

    {#if bookId}
      <button
        class="btn btn-ghost text-xs"
        aria-label={i18n.t("reader.bookmarkLabel")}
        onclick={addBookmark}
      >
        <Bookmark size={16} />
      </button>
      <button
        class="btn btn-ghost text-xs"
        aria-label="Annotations"
        onclick={() => (annotationsOpen = !annotationsOpen)}
      >
        <List size={16} />
      </button>
    {/if}
  </div>
  <div
    class="relative flex flex-1 items-center justify-center overflow-auto p-2"
    use:readerGestures={{
      onSwipeLeft: rtl ? prev : next,
      onSwipeRight: rtl ? next : prev,
      onTapLeft: rtl ? next : prev,
      onTapRight: rtl ? prev : next,
    }}
  >
    {#if loading}
      <p class="text-sm text-muted">Loading...</p>
    {:else if total > 0}
      {#each prefetchPages as prefetchPage (prefetchPage)}
        <img
          src={pageUrl(prefetchPage)}
          alt=""
          class="pointer-events-none absolute h-0 w-0 overflow-hidden opacity-0"
          decoding="async"
          fetchpriority="low"
          aria-hidden="true"
        />
      {/each}
      {#key spreadKey}
        <div class="flex h-full w-full items-center justify-center gap-1">
          {#each displayPages as p (p)}
            <div class="flex h-full min-w-0 flex-1 items-center justify-center">
              <img
                src={pageUrl(p)}
                alt="Page {p + 1}"
                class="{comicFitClass(fit)} shadow-[var(--shadow)]"
                class:opacity-0={loadedCount < displayPages.length}
                decoding="async"
                fetchpriority="high"
                onload={() => (loadedCount += 1)}
              />
            </div>
          {/each}
        </div>
      {/key}
      {#if loadedCount < displayPages.length}
        <p class="absolute text-sm text-muted">Loading page...</p>
      {/if}
    {:else}
      <EmptyState
        size="sm"
        title={i18n.t("reader.noPagesTitle")}
        body={i18n.t("reader.noPagesBody")}
      >
        {#snippet icon(size)}
          <Images {size} />
        {/snippet}
      </EmptyState>
    {/if}
  </div>

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
