<script lang="ts">
  import { BookOpen, RefreshCw, Search, Star } from "@lucide/svelte";
  import { SvelteSet } from "svelte/reactivity";
  import BookGrid from "$lib/components/BookGrid.svelte";
  import Button from "$lib/components/Button.svelte";
  import EmptyState from "$lib/components/EmptyState.svelte";
  import FilterChips from "$lib/components/FilterChips.svelte";
  import QuickFilters from "$lib/components/QuickFilters.svelte";
  import { api, ApiError } from "$lib/api/client";
  import { library } from "$lib/stores/library.svelte";
  import { metadataMatch } from "$lib/stores/metadataMatch.svelte";
  import { scan } from "$lib/stores/scan.svelte";
  import { confirmDialog } from "$lib/stores/confirm.svelte";
  import { toast } from "$lib/stores/toast.svelte";
  import { i18n } from "$lib/stores/i18n.svelte";
  import { can } from "$lib/permissions";

  let initialLoading = $derived(library.loading && library.books.length === 0);
  let selectMode = $state(false);
  let selected = new SvelteSet<number>();
  let matching = $state(false);

  let jobActive = $derived(scan.status?.scanning || metadataMatch.status?.running);
  let jobLabel = $derived.by(() => {
    if (scan.status?.scanning) {
      const n = scan.status.indexed;
      return i18n.t("library.scanning", { count: n > 0 ? ` (${n})` : "" });
    }
    if (metadataMatch.status?.running) {
      const st = metadataMatch.status;
      return i18n.t("library.matchingProgress", { done: st.done, total: st.total });
    }
    return "";
  });

  function toggleSelectMode() {
    selectMode = !selectMode;
    if (!selectMode) selected.clear();
  }

  function toggleBook(id: number) {
    if (selected.has(id)) selected.delete(id);
    else selected.add(id);
  }

  function selectAllVisible() {
    for (const book of library.books) selected.add(book.id);
  }

  async function runMetadataMatch(all: boolean) {
    if (!can("edit_metadata")) return;
    const bookIds = all ? [] : [...selected];
    if (!all && bookIds.length === 0) {
      toast.error(i18n.t("library.selectAtLeastOne"));
      return;
    }
    const ok = await confirmDialog.ask({
      title: all
        ? i18n.t("library.matchAllConfirmTitle")
        : i18n.t("library.matchSelectedConfirmTitle"),
      message: all
        ? i18n.t("library.matchAllConfirm")
        : i18n.t("library.matchSelectedConfirm", { count: bookIds.length }),
      confirmLabel: i18n.t("library.matchConfirm"),
      cancelLabel: i18n.t("confirm.cancel"),
    });
    if (!ok) return;
    matching = true;
    try {
      await api.startMetadataMatch({
        bookIds: all ? undefined : bookIds,
        libraryId: library.libraryFilter ?? undefined,
        applyCover: true,
      });
      toast.info(all ? i18n.t("library.matchStartedAll") : i18n.t("library.matchStarted"));
      selectMode = false;
      selected.clear();
      metadataMatch.startPolling(() => {
        toast.success(i18n.t("library.matchFinished"));
        void library.refresh({ background: true });
      });
    } catch (e) {
      toast.error(e instanceof ApiError ? e.message : i18n.t("library.matchFailed"));
    } finally {
      matching = false;
    }
  }
</script>

<section class="mx-auto w-full max-w-[1600px] px-3 py-5 sm:px-6">
  {#if jobActive && jobLabel}
    <div
      class="mb-4 flex items-center gap-2 rounded-lg border border-border bg-bg-elevated px-3 py-2 text-sm text-muted"
      role="status"
    >
      <RefreshCw size={14} class="animate-spin text-primary" />
      <span>{jobLabel}</span>
    </div>
  {/if}

  <QuickFilters />

  {#if can("edit_metadata")}
    <div class="mb-4 hidden flex-wrap items-center gap-2 md:flex">
      <button
        type="button"
        class="btn btn-ghost min-h-10 text-xs ring-1 ring-border"
        class:text-primary={selectMode}
        onclick={toggleSelectMode}
      >
        {selectMode ? i18n.t("library.cancelSelection") : i18n.t("library.selectBooks")}
      </button>
      {#if selectMode}
        <button type="button" class="btn btn-ghost min-h-10 text-xs" onclick={selectAllVisible}>
          {i18n.t("library.selectVisible")}
        </button>
        <button
          type="button"
          class="btn btn-primary min-h-10 text-xs"
          disabled={matching || selected.size === 0}
          onclick={() => void runMetadataMatch(false)}
        >
          <Search size={14} />
          {i18n.t("library.matchSelected", { count: selected.size })}
        </button>
      {:else}
        <button
          type="button"
          class="btn btn-ghost min-h-10 text-xs ring-1 ring-border"
          disabled={matching}
          onclick={() => void runMetadataMatch(true)}
        >
          <Search size={14} />
          {i18n.t("library.matchAll")}
        </button>
      {/if}
    </div>
  {/if}

  <FilterChips />
  {#if library.error}
    <div
      class="flex flex-col items-start gap-3 rounded-lg border border-danger/40 bg-danger/10 p-4 text-sm text-fg"
      role="alert"
    >
      <p>{i18n.t("library.loadError")}</p>
      <Button onclick={() => library.refresh()}>{i18n.t("library.retry")}</Button>
    </div>
  {:else if library.books.length === 0 && !library.loading}
    <EmptyState
      title={library.search
        ? i18n.t("library.emptySearchTitle")
        : library.inProgressFilter
          ? i18n.t("library.emptyInProgressTitle")
          : library.favoritesFilter
            ? i18n.t("library.emptyFavoritesTitle")
            : i18n.t("library.emptyTitle")}
      body={library.search
        ? i18n.t("library.emptySearchBody")
        : library.inProgressFilter
          ? i18n.t("library.emptyInProgress")
          : library.favoritesFilter
            ? i18n.t("library.emptyFavorites")
            : i18n.t("library.emptyBody")}
    >
      {#snippet icon(size)}
        {#if library.search}
          <Search {size} />
        {:else if library.favoritesFilter}
          <Star {size} />
        {:else}
          <BookOpen {size} />
        {/if}
      {/snippet}
      {#if library.search}
        <Button onclick={() => library.setSearch("")}>{i18n.t("library.clearSearch")}</Button>
      {:else if !library.favoritesFilter && !library.inProgressFilter}
        <div class="flex flex-wrap justify-center gap-2">
          <Button onclick={() => library.triggerScan()}>
            <RefreshCw size={16} />
            {i18n.t("library.scanLibrary")}
          </Button>
          <a href="/settings/library" class="btn btn-ghost ring-1 ring-border"
            >{i18n.t("library.librarySettings")}</a
          >
        </div>
      {/if}
    </EmptyState>
  {:else}
    <BookGrid
      books={library.books}
      hasMore={library.hasMore}
      loading={library.loading}
      {initialLoading}
      {selectMode}
      {selected}
      onToggleSelect={toggleBook}
      onLoadMore={() => library.loadMore()}
    />
  {/if}
</section>
