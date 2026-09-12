<script lang="ts">
  import type { Book } from "$lib/api/types";
  import type { BookEditFields } from "$lib/components/metadata/fields";

  interface Props {
    book: Book;
    fields: BookEditFields;
  }

  let { book, fields }: Props = $props();
</script>

<div class="grid min-w-0 flex-1 gap-3 sm:grid-cols-2">
  {#if book.contentHash}
    <div class="sm:col-span-2 rounded-lg border border-border bg-bg-elevated px-3 py-2">
      <p class="text-xs text-muted">Content hash</p>
      <p class="mt-0.5 break-all font-mono text-xs text-fg">{book.contentHash}</p>
    </div>
  {/if}
  <label class="block sm:col-span-2">
    <span class="text-xs text-muted">Title</span>
    <input class="input mt-1 w-full" bind:value={fields.title} required />
  </label>
  <label class="block">
    <span class="text-xs text-muted">Author</span>
    <input class="input mt-1 w-full" bind:value={fields.author} />
  </label>
  <label class="block">
    <span class="text-xs text-muted">Language</span>
    <input class="input mt-1 w-full" bind:value={fields.language} />
  </label>
  <label class="block">
    <span class="text-xs text-muted">Series</span>
    <input class="input mt-1 w-full" bind:value={fields.series} />
  </label>
  <label class="block">
    <span class="text-xs text-muted">Series index</span>
    <input
      class="input mt-1 w-full"
      type="number"
      step="any"
      min="0"
      bind:value={fields.seriesIndex}
    />
  </label>
  <label class="block">
    <span class="text-xs text-muted">Journal</span>
    <input class="input mt-1 w-full" bind:value={fields.journal} />
  </label>
  <label class="block">
    <span class="text-xs text-muted">Year</span>
    <input class="input mt-1 w-full" type="number" min="0" bind:value={fields.publishedYear} />
  </label>
  <label class="block">
    <span class="text-xs text-muted">Volume</span>
    <input class="input mt-1 w-full" bind:value={fields.volume} />
  </label>
  <label class="block">
    <span class="text-xs text-muted">Issue</span>
    <input class="input mt-1 w-full" bind:value={fields.issue} />
  </label>
  <label class="block">
    <span class="text-xs text-muted">Pages</span>
    <input class="input mt-1 w-full" bind:value={fields.pages} />
  </label>
  <label class="block">
    <span class="text-xs text-muted">DOI</span>
    <input class="input mt-1 w-full" bind:value={fields.doi} />
  </label>
  <label class="block">
    <span class="text-xs text-muted">arXiv ID</span>
    <input class="input mt-1 w-full" bind:value={fields.arxivId} />
  </label>
  <label class="block">
    <span class="text-xs text-muted">PubMed ID</span>
    <input class="input mt-1 w-full" bind:value={fields.pubmedId} />
  </label>
  <label class="block sm:col-span-2">
    <span class="text-xs text-muted">Description</span>
    <textarea class="input mt-1 min-h-24 w-full resize-y" bind:value={fields.description}
    ></textarea>
  </label>
</div>
