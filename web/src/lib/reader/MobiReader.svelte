<script lang="ts">
  import { Bookmark, ChevronLeft, ChevronRight, List } from "@lucide/svelte";
  import { api, ApiError } from "$lib/api/client";
  import type { MobiSection } from "$lib/api/types";
  import ReaderAnnotations from "$lib/components/ReaderAnnotations.svelte";
  import { readerGestures } from "$lib/reader/reader-touch";
  import { i18n } from "$lib/stores/i18n.svelte";
  import { toast } from "$lib/stores/toast.svelte";

  interface Props {
    bookId: number;
    initialSection?: number;
    onProgress?: (section: number, percent: number) => void;
  }

  let { bookId, initialSection = 0, onProgress }: Props = $props();

  let sections = $state<MobiSection[]>([]);
  let index = $state(0);
  let loading = $state(true);
  let error = $state<string | null>(null);
  let annotationsOpen = $state(false);
  let annotationsRevision = $state(0);

  $effect(() => {
    const id = bookId;
    loading = true;
    error = null;
    api
      .getMobiSections(id)
      .then((s) => {
        sections = s;
        index = Math.min(Math.max(initialSection, 0), Math.max(s.length - 1, 0));
        loading = false;
        report();
      })
      .catch(() => {
        error = "Could not load MOBI content. Try converting to EPUB in book settings.";
        loading = false;
      });
  });

  function report() {
    const pct = sections.length ? (index + 1) / sections.length : 0;
    onProgress?.(index, pct);
  }

  function prev() {
    if (index > 0) {
      index -= 1;
      report();
    }
  }
  function next() {
    if (index < sections.length - 1) {
      index += 1;
      report();
    }
  }

  function onKey(event: KeyboardEvent) {
    if (event.key === "ArrowLeft") prev();
    if (event.key === "ArrowRight") next();
  }

  async function addBookmark() {
    if (!bookId) return;
    try {
      const label = sections[index]?.title || `${i18n.t("reader.bookmarkLabel")} ${index + 1}`;
      await api.createBookmark(bookId, String(index), label);
      toast.success(i18n.t("reader.bookmarkSaved"));
      annotationsRevision += 1;
    } catch (e) {
      toast.error(e instanceof ApiError ? e.message : i18n.t("reader.bookmarkFailed"));
    }
  }

  function jumpTo(location: string) {
    const n = Number(location);
    if (Number.isFinite(n) && n >= 0 && n < sections.length) {
      index = n;
      report();
    }
    annotationsOpen = false;
  }
</script>

<svelte:window onkeydown={onKey} />

<div class="relative flex h-full flex-col">
  <div
    class="flex items-center justify-center gap-2 border-b border-border bg-bg/80 px-2 py-1.5 text-sm sm:gap-3 sm:px-3 sm:py-2"
  >
    <button class="btn btn-ghost" aria-label="Previous" onclick={prev} disabled={index <= 0}>
      <ChevronLeft size={18} />
    </button>
    <span class="tabular-nums text-muted">
      {sections.length ? index + 1 : 0} / {sections.length}
    </span>
    <button
      class="btn btn-ghost"
      aria-label="Next"
      onclick={next}
      disabled={sections.length === 0 || index >= sections.length - 1}
    >
      <ChevronRight size={18} />
    </button>

    {#if bookId}
      <span class="mx-1 h-5 w-px bg-border"></span>
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
    class="flex-1 overflow-y-auto p-3 sm:p-4"
    use:readerGestures={{
      onSwipeLeft: next,
      onSwipeRight: prev,
    }}
  >
    {#if loading}
      <p class="text-sm text-muted">Loading...</p>
    {:else if error}
      <p class="text-sm text-muted">{error}</p>
    {:else if sections[index]}
      <article class="prose-mobi max-w-3xl mx-auto text-fg">
        {#if sections[index].title}
          <h1 class="mb-4 text-lg font-semibold">{sections[index].title}</h1>
        {/if}
        <!-- eslint-disable-next-line svelte/no-at-html-tags -->
        {@html sections[index].html}
      </article>
    {/if}
  </div>

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

<style>
  :global(.prose-mobi p) {
    margin-bottom: 1em;
    line-height: 1.6;
  }
</style>
