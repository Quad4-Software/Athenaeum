<script lang="ts">
  import { LayoutGrid, PanelLeftClose, PanelLeftOpen, Rows3 } from "@lucide/svelte";
  import SearchBar from "./SearchBar.svelte";
  import ThemeToggle from "./ThemeToggle.svelte";
  import LanguagePicker from "./LanguagePicker.svelte";
  import SortMenu from "./SortMenu.svelte";
  import Breadcrumbs from "./Breadcrumbs.svelte";
  import { ui } from "$lib/stores/ui.svelte";
  import { density } from "$lib/stores/density.svelte";
  import { router } from "$lib/router.svelte";
  import { breadcrumbsFor } from "$lib/navigation.svelte";
  import { i18n } from "$lib/stores/i18n.svelte";

  interface Props {
    bookTitle?: string;
  }

  let { bookTitle = "" }: Props = $props();

  let crumbs = $derived(breadcrumbsFor(router.current, bookTitle));
  let showBrowseTools = $derived(
    router.current.name === "library" || router.current.name === "collection",
  );
</script>

<header
  class="sticky top-0 z-20 flex flex-col gap-2 border-b border-border bg-bg/80 px-2 py-2 backdrop-blur pt-[max(0.5rem,env(safe-area-inset-top))] sm:px-4"
>
  <div class="flex min-h-11 items-center gap-1.5 sm:gap-2">
    <button
      type="button"
      class="btn btn-ghost btn-icon hidden md:inline-flex"
      aria-label={ui.sidebarCollapsed
        ? i18n.t("topbar.expandSidebar")
        : i18n.t("topbar.collapseSidebar")}
      onclick={() => ui.toggleSidebar()}
    >
      {#if ui.sidebarCollapsed}
        <PanelLeftOpen size={18} />
      {:else}
        <PanelLeftClose size={18} />
      {/if}
    </button>

    <div class="min-w-0 flex-1">
      <SearchBar />
    </div>

    <div class="hidden items-center gap-1.5 md:inline-flex sm:gap-2">
      {#if showBrowseTools}
        <SortMenu />
        <button
          type="button"
          class="btn btn-ghost btn-icon"
          aria-label={density.value === "compact"
            ? i18n.t("commands.densityComfortable")
            : i18n.t("commands.densityCompact")}
          title={density.value === "compact"
            ? i18n.t("commands.densityComfortable")
            : i18n.t("commands.densityCompact")}
          onclick={() => density.toggle()}
        >
          {#if density.value === "compact"}
            <Rows3 size={18} />
          {:else}
            <LayoutGrid size={18} />
          {/if}
        </button>
      {/if}
      <LanguagePicker />
      <ThemeToggle />
    </div>
  </div>

  {#if crumbs.length > 1}
    <Breadcrumbs items={crumbs} />
  {/if}
</header>
