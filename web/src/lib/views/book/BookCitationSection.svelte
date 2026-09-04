<script lang="ts">
  import { Copy, FileText } from "@lucide/svelte";
  import Button from "$lib/components/Button.svelte";
  import { i18n } from "$lib/stores/i18n.svelte";
  import type { Book } from "$lib/api/types";
  import { citationVolumeLine } from "./book-view-actions";

  interface Props {
    book: Book;
    bibtexBusy: boolean;
    oncopybibtex: () => void;
    ondownloadbibtex: () => void;
  }

  let { book, bibtexBusy, oncopybibtex, ondownloadbibtex }: Props = $props();

  let volumeLine = $derived(citationVolumeLine(book));
</script>

<div class="mt-6 space-y-2">
  <p class="text-sm font-medium text-fg">{i18n.t("book.citation")}</p>
  {#if book.journal}
    <p class="text-sm text-muted">
      <span class="text-subtle">{i18n.t("book.journal")}:</span>
      {book.journal}
    </p>
  {/if}
  {#if book.publishedYear}
    <p class="text-sm text-muted">
      <span class="text-subtle">{i18n.t("book.year")}:</span>
      {book.publishedYear}
    </p>
  {/if}
  {#if volumeLine}
    <p class="text-sm text-muted">{volumeLine}</p>
  {/if}
  {#if book.doi}
    <p class="text-sm text-muted">
      <span class="text-subtle">{i18n.t("book.doi")}:</span>
      <a
        class="text-primary underline-offset-2 hover:underline"
        href={`https://doi.org/${book.doi}`}
        target="_blank"
        rel="noopener noreferrer"
      >
        {book.doi}
      </a>
    </p>
  {/if}
  {#if book.arxivId}
    <p class="text-sm text-muted">
      <span class="text-subtle">{i18n.t("book.arxiv")}:</span>
      <a
        class="text-primary underline-offset-2 hover:underline"
        href={`https://arxiv.org/abs/${book.arxivId}`}
        target="_blank"
        rel="noopener noreferrer"
      >
        {book.arxivId}
      </a>
    </p>
  {/if}
  {#if book.pubmedId}
    <p class="text-sm text-muted">
      <span class="text-subtle">{i18n.t("book.pubmed")}:</span>
      <a
        class="text-primary underline-offset-2 hover:underline"
        href={`https://pubmed.ncbi.nlm.nih.gov/${book.pubmedId}`}
        target="_blank"
        rel="noopener noreferrer"
      >
        {book.pubmedId}
      </a>
    </p>
  {/if}
  <div class="flex flex-wrap gap-2 pt-1">
    <Button
      variant="ghost"
      class="min-h-10 ring-1 ring-border text-xs"
      loading={bibtexBusy}
      onclick={oncopybibtex}
    >
      <Copy size={14} />
      {i18n.t("book.copyBibtex")}
    </Button>
    <Button
      variant="ghost"
      class="min-h-10 ring-1 ring-border text-xs"
      loading={bibtexBusy}
      onclick={ondownloadbibtex}
    >
      <FileText size={14} />
      {i18n.t("book.downloadBibtex")}
    </Button>
  </div>
</div>
