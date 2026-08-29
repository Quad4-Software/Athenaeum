<script lang="ts">
  import { X } from "@lucide/svelte";
  import { library } from "$lib/stores/library.svelte";
  import { collections } from "$lib/stores/collections.svelte";
  import { libraries } from "$lib/stores/libraries.svelte";
  import { i18n } from "$lib/stores/i18n.svelte";

  function clearAll() {
    library.clearFilters();
  }

  let chips = $derived.by(() => {
    const out: { label: string; clear: () => void }[] = [];
    if (library.search) {
      out.push({
        label: i18n.t("library.filters.search", { query: library.search }),
        clear: () => library.setSearch(""),
      });
    }
    if (library.inProgressFilter) {
      out.push({
        label: i18n.t("library.filters.continueReading"),
        clear: () => library.setInProgress(false),
      });
    }
    if (library.favoritesFilter) {
      out.push({
        label: i18n.t("library.filters.favorites"),
        clear: () => library.setFavorites(false),
      });
    }
    if (library.authorFilter) {
      out.push({
        label: i18n.t("library.filters.authorPrefix", { name: library.authorFilter }),
        clear: () => library.setAuthor(""),
      });
    }
    if (library.seriesFilter) {
      out.push({
        label: i18n.t("library.filters.seriesPrefix", { name: library.seriesFilter }),
        clear: () => library.setSeries(""),
      });
    }
    if (library.tagFilter) {
      out.push({
        label: `#${library.tagFilter}`,
        clear: () => library.setTag(""),
      });
    }
    if (library.formatFilter) {
      out.push({
        label: library.formatFilter.toUpperCase(),
        clear: () => library.setFormat(""),
      });
    }
    if (library.libraryFilter != null) {
      const lib = libraries.items.find((l) => l.id === library.libraryFilter);
      out.push({
        label: lib?.name ?? i18n.t("library.filters.libraryFallback"),
        clear: () => library.setLibrary(null),
      });
    }
    if (library.collectionFilter != null) {
      const col = collections.items.find((c) => c.id === library.collectionFilter);
      out.push({
        label: col?.name ?? i18n.t("library.filters.collectionFallback"),
        clear: () => library.clearFilters(),
      });
    }
    return out;
  });
</script>

{#if chips.length > 0}
  <div class="mb-3 flex flex-wrap items-center gap-2">
    {#each chips as chip (chip.label)}
      <button
        type="button"
        class="inline-flex items-center gap-1 rounded-full border border-border bg-surface px-2.5 py-1 text-xs text-fg hover:bg-surface-hover"
        onclick={chip.clear}
      >
        {chip.label}
        <X size={12} />
      </button>
    {/each}
    {#if chips.length > 1}
      <button type="button" class="text-xs text-primary hover:underline" onclick={clearAll}>
        {i18n.t("library.filters.clearAll")}
      </button>
    {/if}
  </div>
{/if}
