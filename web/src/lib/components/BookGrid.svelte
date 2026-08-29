<script lang="ts">
  import BookCard from "./BookCard.svelte";
  import BookCardSkeleton from "./BookCardSkeleton.svelte";
  import Skeleton from "./Skeleton.svelte";
  import { density } from "$lib/stores/density.svelte";
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
      ? "grid grid-cols-3 gap-x-2 gap-y-4 sm:grid-cols-4 sm:gap-x-3 sm:gap-y-5 md:grid-cols-5 lg:grid-cols-6 xl:grid-cols-8"
      : "grid grid-cols-2 gap-x-3 gap-y-6 sm:grid-cols-3 sm:gap-x-4 sm:gap-y-6 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6",
  );

  function sentinel(node: HTMLElement) {
    const observer = new IntersectionObserver(
      (entries) => {
        if (entries[0]?.isIntersecting && hasMore && !loading) onLoadMore();
      },
      { rootMargin: "600px" },
    );
    observer.observe(node);
    return () => observer.disconnect();
  }
</script>

{#if initialLoading}
  <div class={gridClass}>
    {#each Array(density.value === "compact" ? 16 : 12) as _, i (i)}
      <BookCardSkeleton />
    {/each}
  </div>
{:else}
  <div class={gridClass} data-density={density.value}>
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
  <div {@attach sentinel} class="flex h-16 items-center justify-center">
    {#if loading}
      <Skeleton width="6rem" height="0.75rem" rounded="full" />
    {/if}
  </div>
{/if}
