<script lang="ts">
  import {
    BookMarked,
    Library,
    LogOut,
    RefreshCw,
    Settings,
    Files,
    BookText,
    Headphones,
    Layers,
    FolderOpen,
    HardDrive,
    Star,
    ListChecks,
    BookOpen,
    PanelsTopLeft,
    Tablet,
  } from "@lucide/svelte";
  import { router, link } from "$lib/router.svelte";
  import { library } from "$lib/stores/library.svelte";
  import { libraries } from "$lib/stores/libraries.svelte";
  import { collections } from "$lib/stores/collections.svelte";
  import { sidebarPrefs } from "$lib/stores/sidebar.svelte";
  import SidebarSection from "./SidebarSection.svelte";
  import type { BookFormat } from "$lib/api/types";
  import { api } from "$lib/api/client";
  import { i18n } from "$lib/stores/i18n.svelte";
  import { auth } from "$lib/stores/auth.svelte";
  import { brand } from "$lib/brand";

  interface Props {
    collapsed?: boolean;
    onnavigate?: () => void;
  }

  let { collapsed = false, onnavigate }: Props = $props();

  let appVersion = $state<string | null>(null);

  const formatKeys: { key: string; value: BookFormat | ""; icon: typeof Files }[] = [
    { key: "nav.allBooks", value: "", icon: Library },
    { key: "nav.epub", value: "epub", icon: BookText },
    { key: "nav.pdf", value: "pdf", icon: Files },
    { key: "nav.comics", value: "comic", icon: PanelsTopLeft },
    { key: "nav.kindle", value: "kindle", icon: Tablet },
    { key: "nav.audiobooks", value: "audio", icon: Headphones },
  ];

  let formats = $derived(
    formatKeys.map((item) => ({
      label: i18n.t(item.key),
      value: item.value,
      icon: item.icon,
    })),
  );

  let visibleSections = $derived(sidebarPrefs.visibleSections());
  let readingItems = $derived(collections.readingItems());
  let shelfItems = $derived(collections.shelfItems());
  let libraryItemCount = $derived(libraries.items.length + (libraries.items.length > 1 ? 1 : 0));

  $effect(() => {
    if (!auth.canAccessApp) return;
    void collections.refresh();
    void libraries.refresh();
  });

  $effect(() => {
    void api
      .health()
      .then((health) => {
        appVersion = health.version ?? null;
      })
      .catch(() => {});
  });

  function go(path: string) {
    router.navigate(path);
    onnavigate?.();
  }

  function goHome() {
    library.search = "";
    library.clearFilters();
    go("/");
  }

  function filterBy(value: BookFormat | "") {
    library.setFormat(value);
    go("/");
  }

  function filterSeries(name: string) {
    library.setSeries(name);
    go("/");
  }

  function openCollection(id: number) {
    library.setCollection(id);
    go("/");
  }

  function selectLibrary(id: number | null) {
    library.setLibrary(id);
    go("/");
  }

  function formatActive(value: BookFormat | ""): boolean {
    return (
      router.current.name === "library" &&
      library.formatFilter === value &&
      !library.seriesFilter &&
      library.collectionFilter == null &&
      library.libraryFilter == null &&
      !library.favoritesFilter
    );
  }

  function libraryActive(id: number | null): boolean {
    return (
      router.current.name === "library" &&
      library.libraryFilter === id &&
      !library.formatFilter &&
      !library.seriesFilter &&
      library.collectionFilter == null &&
      !library.favoritesFilter &&
      !library.inProgressFilter
    );
  }

  function favoritesActive(): boolean {
    return router.current.name === "library" && library.favoritesFilter;
  }

  function continueActive(): boolean {
    return router.current.name === "library" && library.inProgressFilter;
  }

  function authorActive(name: string): boolean {
    return router.current.name === "library" && library.authorFilter === name;
  }

  function sectionBordered(section: string): boolean {
    return section !== visibleSections[0];
  }
  async function signOut() {
    await auth.logout();
    onnavigate?.();
  }
</script>

{#snippet manageCollectionsLink()}
  <a href="/collections" use:link={"/collections"} class="text-xs text-primary hover:underline"
    >{i18n.t("nav.manage")}</a
  >
{/snippet}

<nav class="flex h-full min-h-0 flex-col" aria-label={i18n.t("nav.primary")}>
  <div class="shrink-0 p-3 pb-0">
    <a
      href="/"
      use:link={"/"}
      onclick={() => goHome()}
      class="mb-4 flex items-center gap-2.5 px-2 py-1.5 text-fg
        {collapsed ? 'justify-center px-0' : ''}"
    >
      <span
        class="grid h-8 w-8 shrink-0 place-items-center rounded-lg bg-primary text-primary-fg shadow-sm"
      >
        <BookMarked size={18} />
      </span>
      {#if !collapsed}
        <span class="min-w-0 truncate font-display text-xl font-semibold tracking-tight"
          >{brand.appName}</span
        >
      {/if}
    </a>
  </div>

  <div class="min-h-0 flex-1 overflow-x-hidden overflow-y-auto px-3">
    <div class="flex flex-col gap-1">
      {#each visibleSections as section (section)}
        {#if section === "libraries" && libraries.items.length > 0}
          <SidebarSection
            id="libraries"
            title={i18n.t("nav.libraries")}
            itemCount={libraryItemCount}
            sidebarCollapsed={collapsed}
            bordered={sectionBordered(section)}
          >
            {#if libraries.items.length > 1}
              <button
                type="button"
                onclick={() => selectLibrary(null)}
                class="flex items-center justify-between gap-2 rounded-lg px-2.5 py-2 text-sm transition-colors
                  {libraryActive(null)
                  ? 'bg-surface text-fg ring-1 ring-border/70'
                  : 'text-muted hover:bg-surface-hover hover:text-fg'}"
                title={i18n.t("nav.allLibraries")}
                aria-current={libraryActive(null) ? "page" : undefined}
              >
                <span class="flex items-center gap-3 truncate">
                  <HardDrive size={18} />
                  {#if !collapsed}<span>{i18n.t("nav.allLibraries")}</span>{/if}
                </span>
              </button>
            {/if}
            {#each libraries.items as lib (lib.id)}
              {@const active = libraryActive(lib.id)}
              <button
                type="button"
                onclick={() => selectLibrary(lib.id)}
                class="flex items-center justify-between gap-2 rounded-lg px-2.5 py-2 text-sm transition-colors
              {active ? 'bg-surface text-fg' : 'text-muted hover:bg-surface-hover hover:text-fg'}"
                title={lib.mountPath}
                aria-current={active ? "page" : undefined}
              >
                <span class="flex items-center gap-3 truncate">
                  <Library size={18} />
                  {#if !collapsed}<span class="truncate">{lib.name}</span>{/if}
                </span>
                {#if !collapsed}<span class="text-xs text-subtle">{lib.bookCount}</span>{/if}
              </button>
            {/each}
          </SidebarSection>
        {:else if section === "formats"}
          <SidebarSection
            id="formats"
            title={i18n.t("nav.browse")}
            itemCount={formats.length}
            sidebarCollapsed={collapsed}
            bordered={sectionBordered(section)}
          >
            {#each formats as item (item.value)}
              <button
                type="button"
                onclick={() => filterBy(item.value)}
                class="flex items-center gap-3 rounded-lg px-2.5 py-2 text-sm transition-colors
              {formatActive(item.value)
                  ? 'bg-surface text-fg'
                  : 'text-muted hover:bg-surface-hover hover:text-fg'}"
                title={item.label}
                aria-current={formatActive(item.value) ? "page" : undefined}
              >
                <item.icon size={18} />
                {#if !collapsed}<span>{item.label}</span>{/if}
              </button>
            {/each}
          </SidebarSection>
        {:else if section === "favorites"}
          <div class={sectionBordered(section) ? "mt-3 border-t border-border pt-3" : ""}>
            <button
              type="button"
              onclick={() => {
                library.setFavorites(true);
                go("/");
              }}
              class="flex w-full items-center gap-3 rounded-lg px-2.5 py-2 text-sm transition-colors
            {favoritesActive()
                ? 'bg-surface text-fg'
                : 'text-muted hover:bg-surface-hover hover:text-fg'}"
              title={i18n.t("nav.favorites")}
              aria-current={favoritesActive() ? "page" : undefined}
            >
              <Star size={18} />
              {#if !collapsed}<span>{i18n.t("nav.favorites")}</span>{/if}
            </button>
          </div>
        {:else if section === "continue"}
          <div class={sectionBordered(section) ? "mt-3 border-t border-border pt-3" : ""}>
            <button
              type="button"
              onclick={() => {
                library.setInProgress(true);
                go("/");
              }}
              class="flex w-full items-center justify-between gap-2 rounded-lg px-2.5 py-2 text-sm transition-colors
            {continueActive()
                ? 'bg-surface text-fg'
                : 'text-muted hover:bg-surface-hover hover:text-fg'}"
              title={i18n.t("nav.continueReading")}
              aria-current={continueActive() ? "page" : undefined}
            >
              <span class="flex items-center gap-3">
                <BookOpen size={18} />
                {#if !collapsed}<span>{i18n.t("nav.continueReading")}</span>{/if}
              </span>
              {#if !collapsed && library.stats?.readingInProgress}
                <span class="text-xs text-subtle">{library.stats.readingInProgress}</span>
              {/if}
            </button>
          </div>
        {:else if section === "reading" && readingItems.length > 0 && !collapsed}
          <SidebarSection
            id="reading"
            title={i18n.t("nav.readingLists")}
            itemCount={readingItems.length}
            sidebarCollapsed={collapsed}
            bordered={sectionBordered(section)}
            headerExtra={manageCollectionsLink}
          >
            {#each readingItems as c (c.id)}
              {@const active = library.collectionFilter === c.id}
              <button
                type="button"
                onclick={() => openCollection(c.id)}
                class="flex items-center justify-between gap-2 rounded-lg px-2.5 py-1.5 text-sm
              {active ? 'bg-surface text-fg' : 'text-muted hover:bg-surface-hover hover:text-fg'}"
                aria-current={active ? "page" : undefined}
              >
                <span class="flex items-center gap-2 truncate">
                  <ListChecks size={14} />
                  {c.name}
                </span>
                <span class="text-xs text-subtle">{c.bookCount}</span>
              </button>
            {/each}
          </SidebarSection>
        {:else if section === "series" && library.seriesList.length > 0 && !collapsed}
          <SidebarSection
            id="series"
            title={i18n.t("nav.series")}
            itemCount={library.seriesList.length}
            sidebarCollapsed={collapsed}
            bordered={sectionBordered(section)}
            defaultExpanded={false}
          >
            <div class="flex max-h-32 flex-col gap-0.5 overflow-y-auto">
              {#each library.seriesList as s (s.name)}
                {@const active = library.seriesFilter === s.name}
                <button
                  type="button"
                  onclick={() => filterSeries(s.name)}
                  class="flex items-center justify-between gap-2 rounded-lg px-2.5 py-1.5 text-sm
                {active ? 'bg-surface text-fg' : 'text-muted hover:bg-surface-hover hover:text-fg'}"
                  aria-current={active ? "page" : undefined}
                >
                  <span class="truncate">{s.name}</span>
                  <span class="text-xs text-subtle">{s.count}</span>
                </button>
              {/each}
            </div>
          </SidebarSection>
          {#if library.authorList.length > 0}
            <SidebarSection
              id="authors"
              title={i18n.t("nav.authors")}
              itemCount={Math.min(library.authorList.length, 20)}
              sidebarCollapsed={collapsed}
              bordered={true}
              defaultExpanded={false}
            >
              <div class="flex max-h-32 flex-col gap-0.5 overflow-y-auto">
                {#each library.authorList.slice(0, 20) as a (a.name)}
                  {@const active = authorActive(a.name)}
                  <button
                    type="button"
                    onclick={() => {
                      library.setAuthor(a.name);
                      go("/");
                    }}
                    class="flex items-center justify-between gap-2 rounded-lg px-2.5 py-1.5 text-sm
                  {active
                      ? 'bg-surface text-fg'
                      : 'text-muted hover:bg-surface-hover hover:text-fg'}"
                    aria-current={active ? "page" : undefined}
                  >
                    <span class="truncate">{a.name}</span>
                    <span class="text-xs text-subtle">{a.count}</span>
                  </button>
                {/each}
              </div>
            </SidebarSection>
          {/if}
        {:else if section === "shelves" && shelfItems.length > 0 && !collapsed}
          <SidebarSection
            id="shelves"
            title={i18n.t("nav.shelves")}
            itemCount={shelfItems.length}
            sidebarCollapsed={collapsed}
            bordered={sectionBordered(section)}
            headerExtra={manageCollectionsLink}
          >
            {#each shelfItems as c (c.id)}
              {@const active = library.collectionFilter === c.id}
              <button
                type="button"
                onclick={() => openCollection(c.id)}
                class="flex items-center justify-between gap-2 rounded-lg px-2.5 py-1.5 text-sm
              {active ? 'bg-surface text-fg' : 'text-muted hover:bg-surface-hover hover:text-fg'}"
                aria-current={active ? "page" : undefined}
              >
                <span class="flex items-center gap-2 truncate">
                  <Layers size={14} />
                  {c.name}
                  {#if c.kind === "auto"}
                    <span class="text-[10px] uppercase text-subtle">{i18n.t("nav.shelfAuto")}</span>
                  {:else if c.kind === "smart"}
                    <span class="text-[10px] uppercase text-subtle">{i18n.t("nav.shelfSmart")}</span
                    >
                  {/if}
                </span>
                <span class="text-xs text-subtle">{c.bookCount}</span>
              </button>
            {/each}
          </SidebarSection>
        {/if}
      {/each}
    </div>
  </div>

  <div class="mt-auto flex shrink-0 flex-col gap-0.5 border-t border-border p-3 pt-3">
    <a
      href="/collections"
      use:link={"/collections"}
      onclick={() => onnavigate?.()}
      class="flex items-center gap-3 rounded-lg px-2.5 py-2 text-sm transition-colors
        {router.current.name === 'collections' || router.current.name === 'collection'
        ? 'bg-surface text-fg'
        : 'text-muted hover:bg-surface-hover hover:text-fg'}"
      title={i18n.t("nav.collections")}
      aria-current={router.current.name === "collections" || router.current.name === "collection"
        ? "page"
        : undefined}
    >
      <FolderOpen size={18} />
      {#if !collapsed}<span>{i18n.t("nav.collections")}</span>{/if}
    </a>

    <button
      type="button"
      onclick={() => library.triggerScan()}
      class="flex items-center gap-3 rounded-lg px-2.5 py-2 text-sm text-muted transition-colors hover:bg-surface-hover hover:text-fg"
      title={i18n.t("nav.rescan")}
    >
      <RefreshCw size={18} class={library.stats?.scanning ? "animate-spin" : ""} />
      {#if !collapsed}
        <span>{library.stats?.scanning ? i18n.t("nav.scanning") : i18n.t("nav.rescan")}</span>
      {/if}
    </button>

    <a
      href="/settings/library"
      use:link={"/settings/library"}
      onclick={() => onnavigate?.()}
      class="flex items-center gap-3 rounded-lg px-2.5 py-2 text-sm transition-colors
        {router.current.name === 'settings'
        ? 'bg-surface text-fg'
        : 'text-muted hover:bg-surface-hover hover:text-fg'}"
      title={i18n.t("nav.settings")}
      aria-current={router.current.name === "settings" ? "page" : undefined}
    >
      <Settings size={18} />
      {#if !collapsed}<span>{i18n.t("nav.settings")}</span>{/if}
    </a>

    {#if auth.authEnabled && auth.user}
      <button
        type="button"
        onclick={signOut}
        class="flex items-center gap-3 rounded-lg px-2.5 py-2 text-sm text-muted transition-colors hover:bg-surface-hover hover:text-fg"
        title={i18n.t("auth.signOut")}
      >
        <LogOut size={18} />
        {#if !collapsed}<span>{i18n.t("auth.signOut")}</span>{/if}
      </button>
    {/if}

    {#if !collapsed && library.stats}
      <p class="px-2.5 pt-2 text-xs text-subtle">
        {library.stats.totalBooks.toLocaleString()} items
        {#if appVersion}
          <span class="text-subtle"> · {appVersion}</span>
        {/if}
        {#if library.libraryFilter}
          {@const lib = libraries.items.find((l) => l.id === library.libraryFilter)}
          {#if lib}in {lib.name}{/if}
        {/if}
      </p>
    {/if}
  </div>
</nav>
