<script lang="ts">
  import { ArrowLeft } from "@lucide/svelte";
  import ErrorView from "$lib/views/ErrorView.svelte";
  import { api, ApiError } from "$lib/api/client";
  import { router } from "$lib/router.svelte";
  import { toast } from "$lib/stores/toast.svelte";
  import { audioPlayer } from "$lib/stores/audioPlayer.svelte";
  import { narrator } from "$lib/stores/narrator.svelte";
  import { bookOfflineCache } from "$lib/offline/book-cache";
  import {
    isAudioFormat,
    isComicFormat,
    isMobiFormat,
    type Book,
    type Progress,
  } from "$lib/api/types";

  type EpubComponent = (typeof import("$lib/reader/EpubReader.svelte"))["default"];
  type PdfComponent = (typeof import("$lib/reader/PdfReader.svelte"))["default"];
  type AudioComponent = (typeof import("$lib/reader/AudioReader.svelte"))["default"];
  type MobiComponent = (typeof import("$lib/reader/MobiReader.svelte"))["default"];
  type ComicComponent = (typeof import("$lib/reader/ComicReader.svelte"))["default"];

  interface Props {
    id: number;
  }

  let { id }: Props = $props();

  let book = $state<Book | null>(null);
  let progress = $state<Progress | null>(null);
  let error = $state<string | null>(null);
  let errorCode = $state<number | null>(null);
  let percent = $state(0);
  let fileUrl = $state("");

  let EpubReader = $state<EpubComponent | null>(null);
  let PdfReader = $state<PdfComponent | null>(null);
  let AudioReader = $state<AudioComponent | null>(null);
  let MobiReader = $state<MobiComponent | null>(null);
  let ComicReader = $state<ComicComponent | null>(null);

  let saveTimer: ReturnType<typeof setTimeout> | null = null;
  let readTimer: ReturnType<typeof setInterval> | null = null;

  $effect(() => {
    const bookId = id;
    readTimer = setInterval(() => {
      void api.addReadingTime(bookId, 60).catch(() => undefined);
    }, 60_000);
    return () => {
      if (readTimer) clearInterval(readTimer);
    };
  });

  $effect(() => {
    const bookId = id;
    book = null;
    progress = null;
    error = null;
    errorCode = null;
    EpubReader = null;
    PdfReader = null;
    AudioReader = null;
    MobiReader = null;
    ComicReader = null;
    Promise.all([api.getBook(bookId), api.getProgress(bookId)])
      .then(([b, p]) => {
        if (bookId !== id) return;
        book = b;
        progress = p;
        percent = p.percent;
        const stream = api.fileUrl(b.id);
        void bookOfflineCache.resolveFileUrl(b.id, stream, b.fileSize, b.modifiedAt).then((url) => {
          if (bookId === id) fileUrl = url;
        });
        fileUrl = stream;
        if (isAudioFormat(b.format)) {
          void import("$lib/reader/AudioReader.svelte").then((m) => {
            if (bookId === id) AudioReader = m.default;
          });
        } else if (b.format === "epub") {
          void import("$lib/reader/EpubReader.svelte").then((m) => {
            if (bookId === id) EpubReader = m.default;
          });
        } else if (b.format === "pdf") {
          void import("$lib/reader/PdfReader.svelte").then((m) => {
            if (bookId === id) PdfReader = m.default;
          });
        } else if (isMobiFormat(b.format)) {
          void import("$lib/reader/MobiReader.svelte").then((m) => {
            if (bookId === id) MobiReader = m.default;
          });
        } else if (isComicFormat(b.format)) {
          void import("$lib/reader/ComicReader.svelte").then((m) => {
            if (bookId === id) ComicReader = m.default;
          });
        }
      })
      .catch((e) => {
        if (bookId !== id) return;
        if (e instanceof ApiError) {
          errorCode = e.status;
          error = e.message;
          if (e.status !== 401) toast.error(e.message);
        } else {
          error = "Failed to open book";
          toast.error(error);
        }
      });
    return () => {
      if (saveTimer) clearTimeout(saveTimer);
      saveTimer = null;
    };
  });

  function persist(location: string, pct: number) {
    percent = pct;
    if (saveTimer) clearTimeout(saveTimer);
    const bookId = id;
    saveTimer = setTimeout(() => {
      void api.saveProgress(bookId, { location, percent: pct }).catch((e) => {
        toast.error(e instanceof ApiError ? e.message : "Failed to save progress");
      });
    }, 800);
  }

  function onEpubProgress(location: string, pct: number) {
    persist(location, pct);
  }

  function onPdfProgress(page: number, pct: number) {
    persist(String(page), pct);
  }

  function onAudioProgress(location: string, pct: number) {
    persist(location, pct);
  }

  function onMobiProgress(section: number, pct: number) {
    persist(String(section), pct);
  }

  function onComicProgress(page: number, pct: number) {
    persist(String(page), pct);
  }

  let audioMiniPad = $derived((audioPlayer.active && !audioPlayer.expanded) || narrator.showBar);
</script>

<div
  class="flex h-[100dvh] flex-col bg-bg pt-[env(safe-area-inset-top)]"
  class:pb-bottom-chrome={audioMiniPad}
>
  <header
    class="flex items-center gap-2 border-b border-border px-2 py-1.5 sm:gap-3 sm:px-3 sm:py-2"
  >
    <button
      type="button"
      class="btn btn-ghost"
      aria-label="Close reader"
      onclick={() => router.navigate(book ? `/book/${book.id}` : "/")}
    >
      <ArrowLeft size={18} />
    </button>
    <div class="min-w-0 flex-1">
      <p class="truncate text-sm font-medium text-fg">{book?.title ?? "Loading..."}</p>
    </div>
    <span class="text-xs tabular-nums text-muted">{Math.round(percent * 100)}%</span>
  </header>

  <div class="relative h-px flex-none bg-border">
    <div class="h-full bg-primary transition-[width]" style:width={`${percent * 100}%`}></div>
  </div>

  <div class="min-h-0 flex-1">
    {#if errorCode}
      <ErrorView code={errorCode} message={error ?? undefined} compact />
    {:else if book && progress && isAudioFormat(book.format) && AudioReader}
      {@const Reader = AudioReader}
      <Reader
        {book}
        url={api.fileUrl(book.id)}
        initialLocation={progress.location ?? ""}
        onProgress={onAudioProgress}
      />
    {:else if book && progress && book.format === "epub" && EpubReader}
      {@const Reader = EpubReader}
      <Reader
        url={fileUrl || api.fileUrl(book.id)}
        bookId={book.id}
        title={book.title}
        initialLocation={progress.location}
        onProgress={onEpubProgress}
      />
    {:else if book && progress && book.format === "pdf" && PdfReader}
      {@const Reader = PdfReader}
      <Reader
        url={fileUrl || api.fileUrl(book.id)}
        bookId={book.id}
        initialPage={progress.location ? Number(progress.location) || 1 : 1}
        onProgress={onPdfProgress}
      />
    {:else if book && progress && isMobiFormat(book.format) && MobiReader}
      {@const Reader = MobiReader}
      <Reader
        bookId={book.id}
        initialSection={progress.location ? Number(progress.location) || 0 : 0}
        onProgress={onMobiProgress}
      />
    {:else if book && progress && isComicFormat(book.format) && ComicReader}
      {@const Reader = ComicReader}
      <Reader
        bookId={book.id}
        initialPage={progress.location ? Number(progress.location) || 0 : 0}
        onProgress={onComicProgress}
      />
    {:else if book?.format === "kfx"}
      <div class="grid h-full place-items-center p-6 text-center text-sm text-muted">
        KFX files are DRM-protected and cannot be read in-browser. Download the file or convert with
        Calibre on your desktop.
      </div>
    {:else}
      <div class="grid h-full place-items-center text-sm text-muted">Loading...</div>
    {/if}
  </div>
</div>
