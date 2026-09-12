<script lang="ts">
  import { Pencil } from "@lucide/svelte";
  import { Collapsible } from "bits-ui";
  import { tick } from "svelte";
  import MetadataIdentifyPanel from "$lib/components/metadata/MetadataIdentifyPanel.svelte";
  import CoverUpload from "$lib/components/metadata/CoverUpload.svelte";
  import BookEditForm from "$lib/components/metadata/BookEditForm.svelte";
  import {
    bookToEditFields,
    bookToSearchFields,
    emptyEditFields,
    emptySearchFields,
    hasSearchInput,
    providerLabel,
    type BookEditFields,
    type MetadataSearchFields,
  } from "$lib/components/metadata/fields";
  import { api, ApiError } from "$lib/api/client";
  import { toast } from "$lib/stores/toast.svelte";
  import { i18n } from "$lib/stores/i18n.svelte";
  import type { Book, MetadataMatch, MetadataProvider } from "$lib/api/types";

  interface Props {
    book: Book;
    onsaved?: (book: Book) => void;
    open?: boolean;
    panel?: "edit" | "identify";
  }

  let {
    book,
    onsaved,
    open = $bindable(false),
    panel = $bindable<"edit" | "identify">("edit"),
  }: Props = $props();

  let identifySection = $state<HTMLElement>();
  let saving = $state(false);
  let searching = $state(false);

  let fields = $state<BookEditFields>(emptyEditFields());
  let search = $state<MetadataSearchFields>(emptySearchFields());
  let providers = $state<MetadataProvider[]>([]);
  let selectedProviders = $state<string[]>([]);
  let matches = $state<MetadataMatch[]>([]);
  let matchCoverUrl = $state("");
  let applyCoverOnSave = $state(true);

  let coverKey = $state(0);
  let formSyncKey = $derived(open ? book.id : 0);

  function loadForm(b: Book) {
    fields = bookToEditFields(b);
    search = bookToSearchFields(b);
    matches = [];
    matchCoverUrl = "";
  }

  $effect(() => {
    const key = formSyncKey;
    if (!key) return;
    loadForm(book);
  });

  $effect(() => {
    if (!open || providers.length > 0) return;
    void api.listMetadataProviders().then((list) => {
      providers = list;
      selectedProviders = list.map((p) => p.id);
    });
  });

  $effect(() => {
    if (!open || panel !== "identify") return;
    void tick().then(() => identifySection?.scrollIntoView({ block: "nearest" }));
  });

  async function runSearch() {
    if (!hasSearchInput(search)) {
      toast.error(i18n.t("book.searchNeedInput"));
      return;
    }
    if (selectedProviders.length === 0) {
      toast.error(i18n.t("book.searchNeedProvider"));
      return;
    }
    searching = true;
    matches = [];
    try {
      const res = await api.searchMetadata(book.id, {
        title: search.title.trim(),
        author: search.author.trim(),
        isbn: search.isbn.trim(),
        asin: search.asin.trim(),
        doi: search.doi.trim(),
        arxivId: search.arxivId.trim(),
        pubmedId: search.pubmedId.trim(),
        providers: selectedProviders,
      });
      matches = res.matches;
      if (matches.length === 0) {
        toast.error(i18n.t("book.searchNoMatches"));
      }
    } catch (e) {
      toast.error(e instanceof ApiError ? e.message : i18n.t("book.metadataSearchFailed"));
    } finally {
      searching = false;
    }
  }

  function useMatch(match: MetadataMatch) {
    fields.title = match.title;
    fields.author = match.author;
    if (match.series) fields.series = match.series;
    if (match.seriesIndex != null && match.seriesIndex > 0) {
      fields.seriesIndex = String(match.seriesIndex);
    }
    if (match.language) fields.language = match.language;
    if (match.description) fields.description = match.description;
    if (match.doi) fields.doi = match.doi;
    if (match.arxivId) fields.arxivId = match.arxivId;
    if (match.pubmedId) fields.pubmedId = match.pubmedId;
    if (match.journal) fields.journal = match.journal;
    if (match.volume) fields.volume = match.volume;
    if (match.issue) fields.issue = match.issue;
    if (match.pages) fields.pages = match.pages;
    if (match.publishedYear != null && match.publishedYear > 0) {
      fields.publishedYear = String(match.publishedYear);
    }
    matchCoverUrl = match.coverUrl ?? "";
    toast.success(i18n.t("book.filledFrom", { provider: providerLabel(providers, match.source) }));
  }

  async function applyMatchNow(match: MetadataMatch) {
    saving = true;
    try {
      const updated = await api.applyMetadataMatch(book.id, match, applyCoverOnSave);
      coverKey += 1;
      loadForm(updated);
      onsaved?.(updated);
      toast.success(i18n.t("book.metadataApplied"));
    } catch (e) {
      toast.error(e instanceof ApiError ? e.message : i18n.t("book.metadataApplyFailed"));
    } finally {
      saving = false;
    }
  }

  async function save(event: Event) {
    event.preventDefault();
    if (!fields.title.trim()) {
      toast.error(i18n.t("book.titleRequired"));
      return;
    }
    saving = true;
    try {
      const idx = fields.seriesIndex.trim() ? Number(fields.seriesIndex) : 0;
      const year = fields.publishedYear.trim() ? Number(fields.publishedYear) : 0;
      let updated = await api.updateBook(book.id, {
        title: fields.title.trim(),
        author: fields.author.trim(),
        series: fields.series.trim(),
        seriesIndex: Number.isFinite(idx) ? idx : 0,
        language: fields.language.trim(),
        description: fields.description.trim(),
        doi: fields.doi.trim(),
        arxivId: fields.arxivId.trim(),
        pubmedId: fields.pubmedId.trim(),
        journal: fields.journal.trim(),
        volume: fields.volume.trim(),
        issue: fields.issue.trim(),
        pages: fields.pages.trim(),
        publishedYear: Number.isFinite(year) ? year : 0,
      });
      if (applyCoverOnSave && matchCoverUrl) {
        updated = await api.coverFromUrl(book.id, matchCoverUrl);
        coverKey += 1;
        matchCoverUrl = "";
      }
      loadForm(updated);
      onsaved?.(updated);
      toast.success(i18n.t("book.metadataSaved"));
      open = false;
    } catch (e) {
      toast.error(e instanceof ApiError ? e.message : i18n.t("book.metadataSaveFailed"));
    } finally {
      saving = false;
    }
  }
</script>

<Collapsible.Root bind:open class="mt-6 rounded-[var(--radius-card)] border border-border">
  <Collapsible.Trigger
    class="flex w-full items-center justify-between px-4 py-3 text-sm font-medium text-fg"
  >
    <span class="inline-flex items-center gap-2">
      <Pencil size={16} />
      Edit metadata & cover
    </span>
    <span class="text-muted">{open ? "Hide" : "Show"}</span>
  </Collapsible.Trigger>

  <Collapsible.Content>
    <form class="space-y-4 border-t border-border px-4 py-4" onsubmit={save}>
      <MetadataIdentifyPanel
        {book}
        {providers}
        bind:selectedProviders
        {search}
        {matches}
        {searching}
        {saving}
        bind:applyCoverOnSave
        bind:sectionEl={identifySection}
        onsearch={() => void runSearch()}
        onuse={useMatch}
        onapply={(match) => void applyMatchNow(match)}
      />

      <div class="flex flex-col gap-4 sm:flex-row">
        <CoverUpload
          {book}
          {coverKey}
          onuploaded={(updated) => {
            coverKey += 1;
            matchCoverUrl = "";
            onsaved?.(updated);
          }}
          onremoved={(updated) => {
            coverKey += 1;
            onsaved?.(updated);
          }}
        />
        <BookEditForm {book} {fields} />
      </div>

      <div class="flex justify-end gap-2">
        <button type="button" class="btn btn-ghost" onclick={() => (open = false)}>Cancel</button>
        <button type="submit" class="btn btn-primary" disabled={saving}>
          {saving ? "Saving..." : "Save metadata"}
        </button>
      </div>
    </form>
  </Collapsible.Content>
</Collapsible.Root>
