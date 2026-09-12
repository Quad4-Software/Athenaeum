<script lang="ts">
  import { Search } from "@lucide/svelte";
  import { i18n } from "$lib/stores/i18n.svelte";
  import {
    isAudioFormat,
    type Book,
    type MetadataMatch,
    type MetadataProvider,
  } from "$lib/api/types";
  import { providerLabel, type MetadataSearchFields } from "$lib/components/metadata/fields";

  interface Props {
    book: Book;
    providers: MetadataProvider[];
    selectedProviders: string[];
    search: MetadataSearchFields;
    matches: MetadataMatch[];
    searching: boolean;
    saving: boolean;
    applyCoverOnSave: boolean;
    sectionEl?: HTMLElement;
    onsearch: () => void;
    onuse: (match: MetadataMatch) => void;
    onapply: (match: MetadataMatch) => void;
  }

  let {
    book,
    providers,
    selectedProviders = $bindable(),
    search,
    matches,
    searching,
    saving,
    applyCoverOnSave = $bindable(),
    sectionEl = $bindable(),
    onsearch,
    onuse,
    onapply,
  }: Props = $props();

  let providerHint = $derived(
    providers.length > 0
      ? i18n.t("book.providerHint", { providers: providers.map((p) => p.label).join(", ") })
      : i18n.t("book.providerHintEmpty"),
  );

  let needsAsin = $derived(
    isAudioFormat(book.format) ||
      providers.some((p) => p.requiresAsin && selectedProviders.includes(p.id)),
  );

  function toggleProvider(id: string) {
    if (selectedProviders.includes(id)) {
      selectedProviders = selectedProviders.filter((p) => p !== id);
    } else {
      selectedProviders = [...selectedProviders, id];
    }
  }
</script>

<section class="identify-panel" bind:this={sectionEl}>
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
      <input class="input mt-1 w-full" bind:value={search.title} />
    </label>
    <label class="block">
      <span class="text-xs text-muted">Search author</span>
      <input class="input mt-1 w-full" bind:value={search.author} />
    </label>
    <label class="block">
      <span class="text-xs text-muted">ISBN</span>
      <input class="input mt-1 w-full" bind:value={search.isbn} placeholder="Optional" />
    </label>
    <label class="block">
      <span class="text-xs text-muted">DOI</span>
      <input class="input mt-1 w-full" bind:value={search.doi} placeholder="Optional" />
    </label>
    <label class="block">
      <span class="text-xs text-muted">arXiv ID</span>
      <input class="input mt-1 w-full" bind:value={search.arxivId} placeholder="Optional" />
    </label>
    <label class="block">
      <span class="text-xs text-muted">PubMed ID</span>
      <input class="input mt-1 w-full" bind:value={search.pubmedId} placeholder="Optional" />
    </label>
    {#if needsAsin}
      <label class="block sm:col-span-2">
        <span class="text-xs text-muted">ASIN (Audnexus / Audible)</span>
        <input class="input mt-1 w-full" bind:value={search.asin} placeholder="B00XXXXXXXX" />
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
      onclick={onsearch}
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
              <span class="match-source">{providerLabel(providers, match.source)}</span>
              {#if match.publishedYear}
                <span>{match.publishedYear}</span>
              {/if}
              {#if match.isbn}
                <span>ISBN {match.isbn}</span>
              {/if}
              {#if match.asin}
                <span>ASIN {match.asin}</span>
              {/if}
              {#if match.doi}
                <span>DOI {match.doi}</span>
              {/if}
              {#if match.arxivId}
                <span>arXiv {match.arxivId}</span>
              {/if}
              {#if match.pubmedId}
                <span>PMID {match.pubmedId}</span>
              {/if}
              {#if match.journal}
                <span>{match.journal}</span>
              {/if}
            </p>
            {#if match.description}
              <p class="match-desc">{match.description}</p>
            {/if}
          </div>
          <div class="match-actions">
            <button type="button" class="btn btn-ghost text-xs" onclick={() => onuse(match)}>
              Use
            </button>
            <button
              type="button"
              class="btn btn-primary text-xs"
              disabled={saving}
              onclick={() => onapply(match)}
            >
              Apply
            </button>
          </div>
        </li>
      {/each}
    </ul>
  {/if}
</section>

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
    border-radius: var(--radius-sm);
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
