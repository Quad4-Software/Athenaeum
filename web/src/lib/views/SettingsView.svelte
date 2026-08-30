<script lang="ts">
  import { router } from "$lib/router.svelte";
  import { auth } from "$lib/stores/auth.svelte";
  import { i18n } from "$lib/stores/i18n.svelte";
  import { tick } from "svelte";
  import SettingsLibraryTab from "./settings/SettingsLibraryTab.svelte";
  import SettingsProfileTab from "./settings/SettingsProfileTab.svelte";
  import SettingsApiTab from "./settings/SettingsApiTab.svelte";
  import SettingsAdminTab from "./settings/SettingsAdminTab.svelte";

  interface Props {
    tab?: string;
  }

  let { tab = "library" }: Props = $props();

  const settingsTabs = $derived([
    {
      id: "library",
      label: i18n.t("settings.tabs.library"),
      description: i18n.t("settings.tabs.libraryDesc"),
    },
    {
      id: "api",
      label: i18n.t("settings.tabs.api"),
      description: i18n.t("settings.tabs.apiDesc"),
    },
    ...(auth.authEnabled
      ? [
          {
            id: "profile",
            label: i18n.t("settings.tabs.profile"),
            description: i18n.t("settings.tabs.profileDesc"),
          },
        ]
      : []),
    ...(auth.user?.isAdmin
      ? [
          {
            id: "admin",
            label: i18n.t("settings.tabs.admin"),
            description: i18n.t("settings.tabs.adminDesc"),
          },
        ]
      : []),
  ] as { id: string; label: string; description: string }[]);

  function goTab(id: string) {
    router.navigate(`/settings/${id}`);
  }

  $effect(() => {
    const id = tab;
    void tick().then(() => {
      document
        .getElementById(`settings-tab-${id}`)
        ?.scrollIntoView({ inline: "center", block: "nearest", behavior: "smooth" });
    });
  });
</script>

<section class="mx-auto flex h-full min-h-0 w-full max-w-6xl flex-col px-3 pt-5 sm:px-6">
  <header class="shrink-0">
    <h1 class="text-2xl font-bold text-fg">{i18n.t("settings.title")}</h1>
    <p class="mt-1 text-sm text-muted">{i18n.t("settings.subtitle")}</p>
  </header>

  <nav
    class="mt-6 flex shrink-0 gap-1 overflow-x-auto border-b border-border pb-3 lg:hidden"
    aria-label={i18n.t("settings.title")}
  >
    {#each settingsTabs as t (t.id)}
      <button
        type="button"
        id={`settings-tab-${t.id}`}
        class="shrink-0 rounded-lg px-3 py-2.5 text-left transition-colors
          {tab === t.id
          ? 'bg-primary text-primary-fg'
          : 'text-muted hover:bg-surface-hover hover:text-fg'}"
        aria-current={tab === t.id ? "page" : undefined}
        onclick={() => goTab(t.id)}
      >
        <span class="block text-sm font-medium">{t.label}</span>
      </button>
    {/each}
  </nav>

  <div class="mt-6 flex min-h-0 flex-1 gap-8 max-lg:mt-4">
    <nav
      class="hidden h-full w-52 shrink-0 flex-col gap-1 overflow-y-auto border-r border-border pr-6 lg:flex"
      aria-label={i18n.t("settings.title")}
    >
      {#each settingsTabs as t (t.id)}
        <button
          type="button"
          id={`settings-tab-lg-${t.id}`}
          class="shrink-0 rounded-lg px-3 py-2.5 text-left transition-colors
            {tab === t.id
            ? 'bg-primary text-primary-fg'
            : 'text-muted hover:bg-surface-hover hover:text-fg'}"
          aria-current={tab === t.id ? "page" : undefined}
          onclick={() => goTab(t.id)}
        >
          <span class="block text-sm font-medium">{t.label}</span>
          <span class="mt-0.5 block text-xs {tab === t.id ? 'text-primary-fg/80' : 'text-subtle'}"
            >{t.description}</span
          >
        </button>
      {/each}
    </nav>

    <div class="min-h-0 min-w-0 flex-1 overflow-y-auto pb-6">
      {#if tab === "library"}
        <SettingsLibraryTab />
      {:else if tab === "profile" && auth.authEnabled}
        <SettingsProfileTab />
      {:else if tab === "admin" && auth.user?.isAdmin}
        <SettingsAdminTab />
      {:else if tab === "api"}
        <SettingsApiTab />
      {/if}
    </div>
  </div>
</section>
