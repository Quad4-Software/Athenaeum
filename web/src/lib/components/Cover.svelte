<script lang="ts">
  import { BookText, FileText, Headphones } from "@lucide/svelte";
  import { api } from "$lib/api/client";
  import { isAudioFormat, type Book } from "$lib/api/types";

  interface Props {
    book: Book;
  }

  let { book }: Props = $props();
  let failed = $state(false);
  let visible = $state(true);

  let showImage = $derived(book.hasCover && !failed);
  let coverSrc = $derived(
    visible && showImage ? api.coverUrl(book.id, book.modifiedAt) : undefined,
  );

  function unloadOffscreen(node: HTMLElement) {
    const observer = new IntersectionObserver(
      (entries) => {
        visible = entries[0]?.isIntersecting ?? true;
      },
      { rootMargin: "300px" },
    );
    observer.observe(node);
    return () => observer.disconnect();
  }
</script>

<div
  class="relative flex aspect-[2/3] w-full items-center justify-center overflow-hidden rounded-[var(--radius-card)] bg-bg-elevated"
  {@attach unloadOffscreen}
>
  {#if showImage}
    <img
      src={coverSrc}
      alt={`Cover of ${book.title}`}
      loading="lazy"
      decoding="async"
      fetchpriority="low"
      class="h-full w-full object-cover"
      onerror={() => (failed = true)}
    />
  {:else}
    <div class="flex flex-col items-center gap-3 p-4 text-center text-muted">
      {#if isAudioFormat(book.format)}
        <Headphones size={36} strokeWidth={1.5} />
      {:else if book.format === "pdf"}
        <FileText size={36} strokeWidth={1.5} />
      {:else}
        <BookText size={36} strokeWidth={1.5} />
      {/if}
      <span class="line-clamp-3 text-xs font-medium text-subtle">{book.title}</span>
    </div>
  {/if}
</div>
