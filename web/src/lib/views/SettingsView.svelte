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

<section class="mx-auto w-full max-w-6xl px-3 py-5 sm:px-6">
  <h1 class="text-2xl font-bold text-fg">{i18n.t("settings.title")}</h1>
  <p class="mt-1 text-sm text-muted">{i18n.t("settings.subtitle")}</p>

  <div class="mt-6 flex flex-col gap-8 lg:flex-row lg:items-start">
    <nav
      class="sticky top-0 z-10 -mx-3 flex shrink-0 gap-1 overflow-x-auto border-b border-border bg-bg/95 px-3 pb-3 backdrop-blur lg:top-0 lg:mx-0 lg:max-h-[calc(100dvh-8rem)] lg:w-52 lg:flex-col lg:items-stretch lg:gap-1 lg:overflow-y-auto lg:overflow-x-visible lg:border-b-0 lg:border-r lg:bg-transparent lg:px-0 lg:pb-0 lg:pr-6 lg:backdrop-blur-none"
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
          <span
            class="mt-0.5 hidden text-xs lg:block {tab === t.id
              ? 'text-primary-fg/80'
              : 'text-subtle'}">{t.description}</span
          >
        </button>
      {/each}
    </nav>

    <div class="min-w-0 flex-1 pb-2">
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
