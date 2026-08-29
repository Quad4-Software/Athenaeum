<script lang="ts">
  import { Pencil, Search, Trash2, Upload } from "@lucide/svelte";
  import { tick } from "svelte";
  import Cover from "$lib/components/Cover.svelte";
  import { api, ApiError } from "$lib/api/client";
  import { toast } from "$lib/stores/toast.svelte";
  import {
    isAudioFormat,
    type Book,
    type MetadataMatch,
    type MetadataProvider,
  } from "$lib/api/types";

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
  let uploading = $state(false);
  let searching = $state(false);

  let title = $state("");
  let author = $state("");
  let series = $state("");
  let seriesIndex = $state("");
  let language = $state("");
  let description = $state("");

  let searchTitle = $state("");
  let searchAuthor = $state("");
  let searchISBN = $state("");
  let searchASIN = $state("");
  let providers = $state<MetadataProvider[]>([]);
  let selectedProviders = $state<string[]>([]);
  let matches = $state<MetadataMatch[]>([]);
  let matchCoverUrl = $state("");
  let applyCoverOnSave = $state(true);

  let providerHint = $derived(
    providers.length > 0
      ? `Search ${providers.map((p) => p.label).join(", ")} and pick a match.`
      : "Search external catalogs and pick a match.",
  );

  let needsAsin = $derived(
    isAudioFormat(book.format) ||
      providers.some((p) => p.requiresAsin && selectedProviders.includes(p.id)),
  );

  let coverKey = $state(0);
  let formSyncKey = $derived(open ? book.id : 0);

  function loadForm(b: Book) {
    title = b.title;
    author = b.author;
    series = b.series ?? "";
    seriesIndex = b.seriesIndex != null && b.seriesIndex > 0 ? String(b.seriesIndex) : "";
    language = b.language ?? "";
    description = b.description ?? "";
    searchTitle = b.title;
    searchAuthor = b.author;
    searchISBN = "";
    searchASIN = "";
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

  function toggle() {
    open = !open;
  }

  $effect(() => {
    if (!open || panel !== "identify") return;
    void tick().then(() => identifySection?.scrollIntoView({ block: "nearest" }));
  });

  function toggleProvider(id: string) {
    if (selectedProviders.includes(id)) {
      selectedProviders = selectedProviders.filter((p) => p !== id);
    } else {
      selectedProviders = [...selectedProviders, id];
    }
  }

  function providerLabel(id: string): string {
    return providers.find((p) => p.id === id)?.label ?? id;
  }

  async function runSearch() {
    if (!searchTitle.trim() && !searchAuthor.trim() && !searchISBN.trim() && !searchASIN.trim()) {
      toast.error("Enter a title, author, ISBN, or ASIN to search");
      return;
    }
    if (selectedProviders.length === 0) {
      toast.error("Select at least one metadata source");
      return;
    }
    searching = true;
    matches = [];
    try {
      const res = await api.searchMetadata(book.id, {
        title: searchTitle.trim(),
        author: searchAuthor.trim(),
        isbn: searchISBN.trim(),
        asin: searchASIN.trim(),
        providers: selectedProviders,
      });
      matches = res.matches;
      if (matches.length === 0) {
        toast.error("No matches found. Try fewer words or a different source.");
      }
    } catch (e) {
      toast.error(e instanceof ApiError ? e.message : "Metadata search failed");
    } finally {
      searching = false;
    }
  }

  function useMatch(match: MetadataMatch) {
    title = match.title;
    author = match.author;
    if (match.series) series = match.series;
    if (match.seriesIndex != null && match.seriesIndex > 0) {
      seriesIndex = String(match.seriesIndex);
    }
    if (match.language) language = match.language;
    if (match.description) description = match.description;
    matchCoverUrl = match.coverUrl ?? "";
    toast.success(`Filled from ${providerLabel(match.source)}`);
  }

  async function applyMatchNow(match: MetadataMatch) {
    saving = true;
    try {
      const updated = await api.applyMetadataMatch(book.id, match, applyCoverOnSave);
      title = updated.title;
      author = updated.author;
      series = updated.series ?? "";
      seriesIndex =
        updated.seriesIndex != null && updated.seriesIndex > 0 ? String(updated.seriesIndex) : "";
      language = updated.language ?? "";
      description = updated.description ?? "";
      coverKey += 1;
      matchCoverUrl = "";
      loadForm(updated);
      onsaved?.(updated);
      toast.success("Metadata applied");
    } catch (e) {
      toast.error(e instanceof ApiError ? e.message : "Failed to apply metadata");
    } finally {
      saving = false;
    }
  }

  async function save(event: Event) {
    event.preventDefault();
    if (!title.trim()) {
      toast.error("Title is required");
      return;
    }
    saving = true;
    try {
      const idx = seriesIndex.trim() ? Number(seriesIndex) : 0;
      let updated = await api.updateBook(book.id, {
        title: title.trim(),
        author: author.trim(),
        series: series.trim(),
        seriesIndex: Number.isFinite(idx) ? idx : 0,
        language: language.trim(),
        description: description.trim(),
      });
      if (applyCoverOnSave && matchCoverUrl) {
        updated = await api.coverFromUrl(book.id, matchCoverUrl);
        coverKey += 1;
        matchCoverUrl = "";
      }
      loadForm(updated);
      onsaved?.(updated);
      toast.success("Metadata saved");
      open = false;
    } catch (e) {
      toast.error(e instanceof ApiError ? e.message : "Failed to save metadata");
    } finally {
      saving = false;
    }
  }

  async function onCoverSelected(event: Event) {
    const input = event.currentTarget as HTMLInputElement;
    const file = input.files?.[0];
    input.value = "";
    if (!file) return;
    uploading = true;
    try {
      const updated = await api.uploadCover(book.id, file);
      coverKey += 1;
      matchCoverUrl = "";
      onsaved?.(updated);
      toast.success("Cover updated");
    } catch (e) {
      toast.error(e instanceof ApiError ? e.message : "Failed to upload cover");
    } finally {
      uploading = false;
    }
  }

  async function removeCover() {
    uploading = true;
    try {
      const updated = await api.deleteCover(book.id);
      coverKey += 1;
      onsaved?.(updated);
      toast.success("Cover removed");
    } catch (e) {
      toast.error(e instanceof ApiError ? e.message : "Failed to remove cover");
    } finally {
      uploading = false;
    }
  }
</script>

<div class="mt-6 rounded-[var(--radius-card)] border border-border">
  <button
    type="button"
    class="flex w-full items-center justify-between px-4 py-3 text-sm font-medium text-fg"
    onclick={toggle}
  >
    <span class="inline-flex items-center gap-2">
      <Pencil size={16} />
      Edit metadata & cover
    </span>
    <span class="text-muted">{open ? "Hide" : "Show"}</span>
  </button>

  {#if open}
    <form class="space-y-4 border-t border-border px-4 py-4" onsubmit={save}>
      <section class="identify-panel" bind:this={identifySection}>
        <div class="identify-header">
          <Search size={16} />
          <div>
            <p class="identify-title">Identify from external sources</p>
            <p class="identify-hint">{providerHint}</p>
          </div>
        </div>

        <div class="grid gap-3 sm:grid-cols-2">
          <label class="block sm:col-span-2">
            <span class="text-xs text-muted">Search title</span>
            <input class="input mt-1 w-full" bind:value={searchTitle} />
          </label>
          <label class="block">
            <span class="text-xs text-muted">Search author</span>
            <input class="input mt-1 w-full" bind:value={searchAuthor} />
          </label>
          <label class="block">
            <span class="text-xs text-muted">ISBN</span>
            <input class="input mt-1 w-full" bind:value={searchISBN} placeholder="Optional" />
          </label>
          {#if needsAsin}
            <label class="block sm:col-span-2">
              <span class="text-xs text-muted">ASIN (Audnexus / Audible)</span>
              <input class="input mt-1 w-full" bind:value={searchASIN} placeholder="B00XXXXXXXX" />
            </label>
          {/if}
        </div>

        {#if providers.length > 0}
          <div class="provider-row">
            {#each providers as provider (provider.id)}
              <label class="provider-chip">
                <input
                  type="checkbox"
                  checked={selectedProviders.includes(provider.id)}
                  onchange={() => toggleProvider(provider.id)}
                />
                <span>{provider.label}</span>
              </label>
            {/each}
          </div>
        {/if}

        <div class="identify-actions">
          <label class="cover-check">
            <input type="checkbox" bind:checked={applyCoverOnSave} />
            <span>Apply cover from match</span>
          </label>
          <button
            type="button"
            class="btn btn-ghost ring-1 ring-border"
            disabled={searching}
            onclick={runSearch}
          >
            {searching ? "Searching..." : "Search"}
          </button>
        </div>

        {#if matches.length > 0}
          <ul class="match-list">
            {#each matches as match, i (i)}
              <li class="match-item">
                {#if match.coverUrl}
                  <img src={match.coverUrl} alt="" class="match-cover" loading="lazy" />
                {:else}
                  <div class="match-cover match-cover--empty"></div>
                {/if}
                <div class="match-body">
                  <p class="match-title">{match.title}</p>
                  {#if match.author}
                    <p class="match-author">{match.author}</p>
                  {/if}
                  <p class="match-meta">
                    <span class="match-source">{providerLabel(match.source)}</span>
                    {#if match.publishedYear}
                      <span>{match.publishedYear}</span>
                    {/if}
                    {#if match.isbn}
                      <span>ISBN {match.isbn}</span>
                    {/if}
                    {#if match.asin}
                      <span>ASIN {match.asin}</span>
                    {/if}
                  </p>
                  {#if match.description}
                    <p class="match-desc">{match.description}</p>
                  {/if}
                </div>
                <div class="match-actions">
                  <button
                    type="button"
                    class="btn btn-ghost text-xs"
                    onclick={() => useMatch(match)}
                  >
                    Use
                  </button>
                  <button
                    type="button"
                    class="btn btn-primary text-xs"
                    disabled={saving}
                    onclick={() => applyMatchNow(match)}
                  >
                    Apply
                  </button>
                </div>
              </li>
            {/each}
          </ul>
        {/if}
      </section>

      <div class="flex flex-col gap-4 sm:flex-row">
        <div class="w-32 shrink-0">
          {#key coverKey}
            <Cover book={{ ...book, modifiedAt: book.modifiedAt }} />
          {/key}
          <div class="mt-2 flex flex-col gap-2">
            <label class="btn btn-ghost ring-1 ring-border cursor-pointer text-xs">
              <Upload size={14} />
              {uploading ? "Uploading..." : "Upload cover"}
              <input
                type="file"
                accept="image/jpeg,image/png,image/webp"
                class="sr-only"
                disabled={uploading}
                onchange={onCoverSelected}
              />
            </label>
            {#if book.hasCover}
              <button
                type="button"
                class="btn btn-ghost text-xs text-danger"
                disabled={uploading}
                onclick={removeCover}
              >
                <Trash2 size={14} />
                Remove cover
              </button>
            {/if}
          </div>
        </div>

        <div class="grid min-w-0 flex-1 gap-3 sm:grid-cols-2">
          {#if book.contentHash}
            <div class="sm:col-span-2 rounded-lg border border-border bg-bg-elevated px-3 py-2">
              <p class="text-xs text-muted">Content hash</p>
              <p class="mt-0.5 break-all font-mono text-xs text-fg">{book.contentHash}</p>
            </div>
          {/if}
          <label class="block sm:col-span-2">
            <span class="text-xs text-muted">Title</span>
            <input class="input mt-1 w-full" bind:value={title} required />
          </label>
          <label class="block">
            <span class="text-xs text-muted">Author</span>
            <input class="input mt-1 w-full" bind:value={author} />
          </label>
          <label class="block">
            <span class="text-xs text-muted">Language</span>
            <input class="input mt-1 w-full" bind:value={language} />
          </label>
          <label class="block">
            <span class="text-xs text-muted">Series</span>
            <input class="input mt-1 w-full" bind:value={series} />
          </label>
          <label class="block">
            <span class="text-xs text-muted">Series index</span>
            <input
              class="input mt-1 w-full"
              type="number"
              step="any"
              min="0"
              bind:value={seriesIndex}
            />
          </label>
          <label class="block sm:col-span-2">
            <span class="text-xs text-muted">Description</span>
            <textarea class="input mt-1 min-h-24 w-full resize-y" bind:value={description}
            ></textarea>
          </label>
        </div>
      </div>

      <div class="flex justify-end gap-2">
        <button type="button" class="btn btn-ghost" onclick={() => (open = false)}>Cancel</button>
        <button type="submit" class="btn btn-primary" disabled={saving}>
          {saving ? "Saving..." : "Save metadata"}
        </button>
      </div>
    </form>
  {/if}
</div>

<style>
  .identify-panel {
    border-radius: var(--radius-card);
    background: var(--color-bg-elevated);
    padding: 1rem;
    box-shadow: inset 0 0 0 1px var(--color-border);
  }

  .identify-header {
    display: flex;
    gap: 0.75rem;
    margin-bottom: 1rem;
    color: var(--color-muted);
  }

  .identify-title {
    margin: 0;
    font-size: 0.875rem;
    font-weight: 600;
    color: var(--color-fg);
  }

  .identify-hint {
    margin: 0.25rem 0 0;
    font-size: 0.75rem;
    line-height: 1.45;
    color: var(--color-muted);
  }

  .provider-row {
    display: flex;
    flex-wrap: wrap;
    gap: 0.5rem;
    margin-top: 0.75rem;
  }

  .provider-chip {
    display: inline-flex;
    align-items: center;
    gap: 0.35rem;
    border-radius: 9999px;
    padding: 0.25rem 0.625rem;
    font-size: 0.75rem;
    background: var(--color-surface);
    box-shadow: inset 0 0 0 1px var(--color-border);
    cursor: pointer;
  }

  .identify-actions {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    justify-content: space-between;
    gap: 0.75rem;
    margin-top: 0.75rem;
  }

  .cover-check {
    display: inline-flex;
    align-items: center;
    gap: 0.4rem;
    font-size: 0.75rem;
    color: var(--color-muted);
    cursor: pointer;
  }

  .match-list {
    list-style: none;
    margin: 1rem 0 0;
    padding: 0;
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
    max-height: 18rem;
    overflow-y: auto;
  }

  .match-item {
    display: flex;
    gap: 0.75rem;
    align-items: flex-start;
    padding: 0.625rem;
    border-radius: var(--radius-card);
    background: var(--color-surface);
    box-shadow: inset 0 0 0 1px var(--color-border);
  }

  .match-cover {
    width: 3rem;
    height: 4.5rem;
    flex-shrink: 0;
    object-fit: cover;
    border-radius: 4px;
    background: var(--color-bg);
  }

  .match-cover--empty {
    box-shadow: inset 0 0 0 1px var(--color-border);
  }

  .match-body {
    min-width: 0;
    flex: 1;
  }

  .match-title {
    margin: 0;
    font-size: 0.8125rem;
    font-weight: 600;
    color: var(--color-fg);
  }

  .match-author {
    margin: 0.15rem 0 0;
    font-size: 0.75rem;
    color: var(--color-muted);
  }

  .match-meta {
    display: flex;
    flex-wrap: wrap;
    gap: 0.35rem 0.5rem;
    margin: 0.35rem 0 0;
    font-size: 0.6875rem;
    color: var(--color-subtle);
  }

  .match-source {
    font-weight: 600;
    color: var(--color-primary);
  }

  .match-desc {
    margin: 0.35rem 0 0;
    font-size: 0.6875rem;
    line-height: 1.4;
    color: var(--color-muted);
    display: -webkit-box;
    -webkit-line-clamp: 2;
    line-clamp: 2;
    -webkit-box-orient: vertical;
    overflow: hidden;
  }

  .match-actions {
    display: flex;
    flex-direction: column;
    gap: 0.35rem;
    flex-shrink: 0;
  }
</style>
