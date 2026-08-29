<script lang="ts">
  import BookCard from "./BookCard.svelte";
  import BookCardSkeleton from "./BookCardSkeleton.svelte";
  import Skeleton from "./Skeleton.svelte";
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
  <div
    class="grid grid-cols-2 gap-x-3 gap-y-7 sm:grid-cols-3 sm:gap-x-4 sm:gap-y-6 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6"
  >
    {#each Array(12) as _, i (i)}
      <BookCardSkeleton />
    {/each}
  </div>
{:else}
  <div
    class="grid grid-cols-2 gap-x-3 gap-y-7 sm:grid-cols-3 sm:gap-x-4 sm:gap-y-6 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6"
  >
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
