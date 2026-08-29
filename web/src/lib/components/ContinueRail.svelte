<script lang="ts">
  import { link } from "$lib/router.svelte";
  import Cover from "./Cover.svelte";
  import BookCoverProgress from "./BookCoverProgress.svelte";
  import { i18n } from "$lib/stores/i18n.svelte";
  import type { Book } from "$lib/api/types";

  interface Props {
    books: Book[];
  }

  let { books }: Props = $props();
  let items = $derived(books.filter((b) => (b.progressPercent ?? 0) > 0 && (b.progressPercent ?? 0) < 100).slice(0, 12));
</script>

{#if items.length > 0}
  <section class="mb-8" aria-label={i18n.t("library.continueReading")}>
    <div class="mb-3 flex items-baseline justify-between gap-3">
      <h2 class="font-display text-lg font-semibold text-fg">{i18n.t("library.continueReading")}</h2>
    </div>
    <div class="continue-rail -mx-1 flex gap-3 overflow-x-auto px-1 pb-2">
      {#each items as book (book.id)}
        <a
          href={`/book/${book.id}`}
          use:link={`/book/${book.id}`}
          class="group w-[7.5rem] shrink-0 sm:w-36"
        >
          <div
            class="relative overflow-hidden rounded-[var(--radius-card)] shadow-sm transition duration-150 group-hover:-translate-y-0.5"
          >
            <Cover {book} />
            <BookCoverProgress percent={book.progressPercent ?? 0} />
          </div>
          <p class="mt-2 truncate text-xs font-medium text-fg" title={book.title}>{book.title}</p>
        </a>
      {/each}
    </div>
  </section>
{/if}

<style>
  .continue-rail {
    scrollbar-width: thin;
  }
</style>
