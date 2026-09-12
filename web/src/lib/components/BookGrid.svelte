<script lang="ts">
  import { useIntersectionObserver } from "runed";
  import BookCard from "./BookCard.svelte";
  import BookCardSkeleton from "./BookCardSkeleton.svelte";
  import Skeleton from "./Skeleton.svelte";
  import { density } from "$lib/stores/density.svelte";
  import { coverColumnMinPx, coverSize } from "$lib/stores/cover-size.svelte";
  import type { Book } from "$lib/api/types";

  interface Props {
    books: Book[];
    hasMore: boolean;
    loading: boolean;
    initialLoading?: boolean;
    selectMode?: boolean;
    selected?: Set<number>;
    onToggleSelect?: (id: number) => void;
    onLoadMore: () => void;
  }

  let {
    books,
    hasMore,
    loading,
    initialLoading = false,
    selectMode = false,
    selected,
    onToggleSelect,
    onLoadMore,
  }: Props = $props();

  let gridClass = $derived(
    density.value === "compact"
      ? "grid gap-x-2 gap-y-4 sm:gap-x-3 sm:gap-y-5"
      : "grid gap-x-3 gap-y-6 sm:gap-x-4 sm:gap-y-6",
  );

  // Cover width drives the column count via auto-fill; min(100%, ...) keeps a
  // single column from overflowing narrow viewports.
  let columnMinPx = $derived(coverColumnMinPx(coverSize.value, density.value === "compact"));
  let gridTemplate = $derived(`repeat(auto-fill, minmax(min(100%, ${columnMinPx}px), 1fr))`);

  let sentinelEl = $state<HTMLElement | null>(null);

  useIntersectionObserver(
    () => sentinelEl,
    (entries) => {
      if (entries[0]?.isIntersecting && hasMore && !loading) onLoadMore();
    },
    { rootMargin: "600px", threshold: 0 },
  );
</script>

{#if initialLoading}
  <div class={gridClass} style:grid-template-columns={gridTemplate}>
    {#each Array(density.value === "compact" ? 16 : 12) as _, i (i)}
      <BookCardSkeleton />
    {/each}
  </div>
{:else}
  <div class={gridClass} style:grid-template-columns={gridTemplate} data-density={density.value}>
    {#each books as book (book.id)}
      <BookCard
        {book}
        {selectMode}
        selected={selected?.has(book.id) ?? false}
        onToggleSelect={() => onToggleSelect?.(book.id)}
      />
    {/each}
  </div>
{/if}

{#if hasMore}
  <div bind:this={sentinelEl} class="flex h-16 items-center justify-center">
    {#if loading}
      <Skeleton width="6rem" height="0.75rem" rounded="full" />
    {/if}
  </div>
{/if}
