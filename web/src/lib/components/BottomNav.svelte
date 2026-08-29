<script lang="ts">
  import { BookOpen, Compass, Library, Settings, Layers } from "@lucide/svelte";
  import { router } from "$lib/router.svelte";
  import { library } from "$lib/stores/library.svelte";
  import { ui } from "$lib/stores/ui.svelte";
  import { i18n } from "$lib/stores/i18n.svelte";
  import { runCommand } from "$lib/commands/registry";
  import type { CommandId } from "$lib/commands/types";

  type TabId = "library" | "continue" | "collections" | "browse" | "settings";

  const tabs: {
    id: TabId;
    labelKey: string;
    icon: typeof Library;
  }[] = [
    { id: "library", labelKey: "nav.library", icon: Library },
    { id: "continue", labelKey: "nav.continue", icon: BookOpen },
    { id: "collections", labelKey: "nav.collections", icon: Layers },
    { id: "browse", labelKey: "nav.browse", icon: Compass },
    { id: "settings", labelKey: "nav.settings", icon: Settings },
  ];

  let activeTab = $derived.by((): TabId | null => {
    if (ui.mobileNavOpen) return "browse";
    const name = router.current.name;
    if (name === "settings") return "settings";
    if (name === "collections" || name === "collection") return "collections";
    if (name === "library") {
      if (library.inProgressFilter) return "continue";
      return "library";
    }
    return null;
  });

  $effect(() => {
    const root = document.documentElement;
    const mq = window.matchMedia("(max-width: 767px)");
    function apply() {
      if (mq.matches) root.style.setProperty("--nav-bar-height", "3.75rem");
      else root.style.removeProperty("--nav-bar-height");
    }
    apply();
    mq.addEventListener("change", apply);
    return () => {
      mq.removeEventListener("change", apply);
      root.style.removeProperty("--nav-bar-height");
    };
  });

  const tabCommands: Partial<Record<TabId, CommandId>> = {
    library: "nav.library",
    continue: "nav.continue",
    collections: "nav.collections",
    settings: "nav.settings",
  };

  function goBrowse(event: MouseEvent) {
    if (ui.mobileNavOpen) {
      ui.closeMobileNav();
      return;
    }
    ui.openMobileNav(event.currentTarget as HTMLElement);
  }

  function onTab(id: TabId, event: MouseEvent) {
    if (id === "browse") {
      goBrowse(event);
      return;
    }
    const command = tabCommands[id];
    if (command) void runCommand(command);
  }
</script>

<nav
  class="bottom-nav fixed inset-x-0 bottom-0 z-40 border-t border-border bg-bg/90 backdrop-blur md:hidden"
  aria-label={i18n.t("a11y.navigation")}
>
  <div class="bottom-nav-inner">
    {#each tabs as tab (tab.id)}
      {@const active = activeTab === tab.id}
      <button
        type="button"
        class="bottom-nav-tab"
        class:bottom-nav-tab--active={active}
        aria-current={active ? "page" : undefined}
        aria-expanded={tab.id === "browse" ? ui.mobileNavOpen : undefined}
        aria-controls={tab.id === "browse" ? "mobile-nav" : undefined}
        onclick={(e) => onTab(tab.id, e)}
      >
        <tab.icon size={20} strokeWidth={active ? 2.25 : 1.75} />
        <span>{i18n.t(tab.labelKey)}</span>
      </button>
    {/each}
  </div>
</nav>

<style>
  .bottom-nav {
    padding-bottom: env(safe-area-inset-bottom, 0px);
  }

  .bottom-nav-inner {
    display: grid;
    grid-template-columns: repeat(5, minmax(0, 1fr));
    height: var(--nav-bar-height, 3.75rem);
  }

  .bottom-nav-tab {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 0.125rem;
    min-width: 0;
    min-height: 2.75rem;
    margin: 0.25rem 0.125rem;
    padding: 0.25rem 0.125rem;
    border: 0;
    border-radius: 0.5rem;
    background: none;
    color: var(--color-muted);
    font-size: 0.625rem;
    font-weight: 500;
    line-height: 1.1;
    cursor: pointer;
    transition:
      color 150ms ease,
      background-color 150ms ease;
  }

  .bottom-nav-tab span {
    max-width: 100%;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .bottom-nav-tab--active {
    color: var(--color-primary);
    background: color-mix(in oklch, var(--color-primary) 12%, transparent);
    font-weight: 600;
  }
</style>
