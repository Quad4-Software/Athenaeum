<script lang="ts">
  import { BookOpen, Search } from "@lucide/svelte";
  import EmptyState from "$lib/components/EmptyState.svelte";
  import Popover from "$lib/components/Popover.svelte";
  import MenuList from "$lib/components/MenuList.svelte";
  import type { ReaderChapter, ReaderSearchHit } from "$lib/reader/reader-search";
  import { i18n } from "$lib/stores/i18n.svelte";

  interface Props {
    chapters?: ReaderChapter[];
    onChapterSelect?: (chapter: ReaderChapter) => void;
    onSearch?: (query: string) => Promise<ReaderSearchHit[]>;
    onSearchSelect?: (hit: ReaderSearchHit) => void;
  }

  let { chapters = [], onChapterSelect, onSearch, onSearchSelect }: Props = $props();

  let chaptersOpen = $state(false);
  let searchOpen = $state(false);
  let searchQuery = $state("");
  let searchBusy = $state(false);
  let searchResults = $state<ReaderSearchHit[]>([]);
  let searchEmpty = $state(false);
  let searchFailed = $state(false);

  async function runSearch() {
    const q = searchQuery.trim();
    if (!q || !onSearch) return;
    searchBusy = true;
    searchEmpty = false;
    searchFailed = false;
    try {
      searchResults = await onSearch(q);
      searchEmpty = searchResults.length === 0;
    } catch {
      searchFailed = true;
      searchResults = [];
    } finally {
      searchBusy = false;
    }
  }

  function selectChapter(chapter: ReaderChapter) {
    onChapterSelect?.(chapter);
    chaptersOpen = false;
  }

  function selectHit(hit: ReaderSearchHit) {
    onSearchSelect?.(hit);
    searchOpen = false;
  }
</script>

{#if chapters.length > 0}
  <Popover bind:open={chaptersOpen} placement="bottom" align="start" minWidth={260}>
    {#snippet trigger(toggle)}
      <button
        type="button"
        class="btn btn-ghost text-xs"
        class:ring-1={chaptersOpen}
        class:ring-border={chaptersOpen}
        aria-expanded={chaptersOpen}
        onclick={toggle}
      >
        <BookOpen size={14} />
        <span class="hidden sm:inline">{i18n.t("reader.chapters")}</span>
      </button>
    {/snippet}
    <MenuList
      title={i18n.t("reader.chapters")}
      items={chapters.map((chapter) => ({
        id: chapter.id,
        label: `${chapter.depth ? "\u2003".repeat(chapter.depth) : ""}${chapter.label}`,
        hint: chapter.hint,
        onclick: () => selectChapter(chapter),
      }))}
    />
  </Popover>
{/if}

{#if onSearch}
  <Popover bind:open={searchOpen} placement="bottom" align="start" minWidth={320}>
    {#snippet trigger(toggle)}
      <button
        type="button"
        class="btn btn-ghost text-xs"
        class:ring-1={searchOpen}
        class:ring-border={searchOpen}
        aria-expanded={searchOpen}
        onclick={toggle}
      >
        <Search size={14} />
        <span class="hidden sm:inline">{i18n.t("reader.search")}</span>
      </button>
    {/snippet}
    <form
      class="space-y-2 p-1"
      onsubmit={(e) => {
        e.preventDefault();
        void runSearch();
      }}
    >
      <input
        type="search"
        class="field-input text-sm"
        placeholder={i18n.t("reader.searchPlaceholder")}
        bind:value={searchQuery}
      />
      <button type="submit" class="btn btn-primary w-full text-xs" disabled={searchBusy}>
        {searchBusy ? i18n.t("reader.loading") : i18n.t("reader.search")}
      </button>
      {#if searchEmpty}
        <EmptyState
          size="sm"
          class="py-3"
          title={i18n.t("reader.searchNoResultsTitle")}
          body={i18n.t("reader.searchNoResultsBody")}
        >
          {#snippet icon(size)}
            <Search {size} />
          {/snippet}
        </EmptyState>
      {:else if searchFailed}
        <EmptyState size="sm" class="py-3" title={i18n.t("reader.searchFailed")}>
          {#snippet icon(size)}
            <Search {size} />
          {/snippet}
        </EmptyState>
      {/if}
      {#if searchResults.length > 0}
        <ul class="max-h-56 space-y-1 overflow-y-auto border-t border-border pt-2">
          {#each searchResults as hit (hit.id)}
            <li>
              <button
                type="button"
                class="w-full rounded-md px-2 py-1.5 text-left text-xs hover:bg-surface-hover"
                onclick={() => selectHit(hit)}
              >
                <span class="font-medium text-fg">{hit.label}</span>
                <span class="mt-0.5 block line-clamp-2 text-muted">{hit.excerpt}</span>
              </button>
            </li>
          {/each}
        </ul>
      {/if}
    </form>
  </Popover>
{/if}
