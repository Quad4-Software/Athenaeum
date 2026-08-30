<script lang="ts">
  import { tick } from "svelte";
  import { fade, fly } from "svelte/transition";
  import { X } from "@lucide/svelte";
  import Sidebar from "$lib/components/Sidebar.svelte";
  import Topbar from "$lib/components/Topbar.svelte";
  import ToastHost from "$lib/components/ToastHost.svelte";
  import ConfirmDialogHost from "$lib/components/ConfirmDialogHost.svelte";
  import Skeleton from "$lib/components/Skeleton.svelte";
  import LibraryView from "$lib/views/LibraryView.svelte";
  import LoginView from "$lib/views/LoginView.svelte";
  import InviteView from "$lib/views/InviteView.svelte";
  import ErrorView from "$lib/views/ErrorView.svelte";
  import SetupView from "$lib/views/SetupView.svelte";
  import ConnectivityBanner from "$lib/components/ConnectivityBanner.svelte";
  import PwaUpdateBanner from "$lib/components/PwaUpdateBanner.svelte";
  import AudioMiniPlayer from "$lib/components/AudioMiniPlayer.svelte";
  import NarratorBar from "$lib/components/NarratorBar.svelte";
  import CommandPalette from "$lib/components/CommandPalette.svelte";
  import { handleShellKeydown } from "$lib/commands/dispatch";
  import BottomNav from "$lib/components/BottomNav.svelte";
  import { audioPlayer } from "$lib/stores/audioPlayer.svelte";
  import { narrator } from "$lib/stores/narrator.svelte";
  import { router } from "$lib/router.svelte";
  import { library } from "$lib/stores/library.svelte";
  import { scan } from "$lib/stores/scan.svelte";
  import { metadataMatch } from "$lib/stores/metadataMatch.svelte";
  import { auth } from "$lib/stores/auth.svelte";
  import { favorites } from "$lib/stores/favorites.svelte";
  import { collections } from "$lib/stores/collections.svelte";
  import { libraries } from "$lib/stores/libraries.svelte";
  import { ui } from "$lib/stores/ui.svelte";
  import { i18n } from "$lib/stores/i18n.svelte";
  import {
    loginGuardTarget,
    sanitizeLoginLocation,
    shouldRedirectFromLogin,
    shouldRedirectToLogin,
  } from "$lib/auth-redirect";
  import { errorCodeFromSlug } from "$lib/errors";
  import { captureException } from "$lib/telemetry/sentry";

  let mobileNavPanel = $state<HTMLDivElement | null>(null);
  let mobileNavCloseBtn = $state<HTMLButtonElement | null>(null);

  const FOCUSABLE =
    'a[href], button:not([disabled]), input:not([disabled]), select:not([disabled]), textarea:not([disabled]), [tabindex]:not([tabindex="-1"])';

  function focusableIn(root: HTMLElement): HTMLElement[] {
    return Array.from(root.querySelectorAll<HTMLElement>(FOCUSABLE)).filter(
      (el) => el.getClientRects().length > 0,
    );
  }

  function onMobileNavKeydown(e: KeyboardEvent) {
    if (!ui.mobileNavOpen) return;
    if (e.key === "Escape") {
      e.preventDefault();
      ui.closeMobileNav();
      return;
    }
    if (e.key !== "Tab" || !mobileNavPanel) return;
    const focusables = focusableIn(mobileNavPanel);
    if (focusables.length === 0) return;
    const first = focusables[0];
    const last = focusables[focusables.length - 1];
    const active = document.activeElement;
    if (e.shiftKey) {
      if (active === first || !mobileNavPanel.contains(active)) {
        e.preventDefault();
        last.focus();
      }
    } else if (active === last || !mobileNavPanel.contains(active)) {
      e.preventDefault();
      first.focus();
    }
  }

  function onCrash(error: unknown) {
    captureException(error);
  }

  function reloadPage() {
    window.location.reload();
  }

  $effect(() => {
    if (!ui.mobileNavOpen) return;
    const prev = document.body.style.overflow;
    document.body.style.overflow = "hidden";
    void tick().then(() => mobileNavCloseBtn?.focus());
    return () => {
      document.body.style.overflow = prev;
    };
  });

  $effect(() => {
    void auth.init();
    void i18n.init();
  });

  $effect(() => {
    if (!auth.loading && auth.setupNeeded && router.current.name !== "setup") {
      router.navigate("/setup", true);
    }
    if (
      shouldRedirectToLogin({
        loading: auth.loading,
        needsLogin: auth.needsLogin,
        routeName: router.current.name,
      })
    ) {
      router.navigate(loginGuardTarget(router.current.path), true);
    }
    if (
      shouldRedirectFromLogin({
        loading: auth.loading,
        needsLogin: auth.needsLogin,
        setupNeeded: auth.setupNeeded,
        routeName: router.current.name,
      })
    ) {
      router.navigate("/", true);
    }
    if (!auth.loading && !auth.setupNeeded && router.current.name === "setup") {
      router.navigate("/", true);
    }
  });

  $effect(() => {
    if (typeof window === "undefined" || router.current.name !== "login") return;
    const cleaned = sanitizeLoginLocation(router.appPathname(), window.location.search);
    if (cleaned) router.navigate(cleaned, true);
  });

  $effect(() => {
    if (auth.canAccessApp) {
      void library.refresh();
      void favorites.refresh();
      void collections.refresh();
      void libraries.refresh();
      void scan.checkActive(() => {
        void library.refresh({ background: true });
      });
      void metadataMatch.checkActive(() => {
        void library.refresh({ background: true });
      });
    }
  });

  $effect(() => {
    const route = router.current;
    if (route.name === "reader") library.releaseMemory();
    if (route.name === "collection") {
      const id = Number(route.params.id);
      if (Number.isFinite(id)) library.setCollection(id);
    }
    if (route.name === "settings" && !route.params.tab) {
      router.navigate("/settings/library", true);
    }
  });

  let route = $derived(router.current);
  let bookId = $derived(Number(route.params.id));
  // Keep in sync with AudioMiniPlayer / NarratorBar showBar
  let audioMiniPad = $derived((audioPlayer.active && !audioPlayer.expanded) || narrator.showBar);

  $effect(() => {
    const root = document.documentElement;
    if (audioMiniPad) {
      root.style.setProperty("--mini-player-height", "4.5rem");
    } else {
      root.style.removeProperty("--mini-player-height");
    }
    return () => root.style.removeProperty("--mini-player-height");
  });
</script>

<svelte:window
  onkeydown={(e) => {
    onMobileNavKeydown(e);
    handleShellKeydown(e);
  }}
/>

<svelte:boundary onerror={onCrash}>
  {#if route.name === "reader"}
    {#if auth.loading || auth.needsLogin}
      <div class="grid h-[100dvh] place-items-center text-sm text-muted">
        {i18n.t("common.loading")}
      </div>
    {:else}
      {#await import("$lib/views/ReaderView.svelte")}
        <div class="grid h-[100dvh] place-items-center text-sm text-muted">
          {i18n.t("common.loading")}
        </div>
      {:then { default: ReaderView }}
        <ReaderView id={bookId} />
      {:catch}
        <ErrorView
          title={i18n.t("error.chunkTitle")}
          message={i18n.t("error.chunkMessage")}
          onRetry={reloadPage}
        />
      {/await}
    {/if}
  {:else if route.name === "setup"}
    <SetupView />
  {:else if route.name === "login"}
    <LoginView />
  {:else if route.name === "invite"}
    <InviteView />
  {:else if route.name === "error"}
    <ErrorView code={errorCodeFromSlug(route.params.code || "not-found")} />
  {:else if auth.loading || auth.needsLogin}
    <div class="flex h-[100dvh] flex-col gap-4 p-6">
      <Skeleton height="2.5rem" width="12rem" rounded="lg" />
      <div class="grid flex-1 grid-cols-2 gap-4 sm:grid-cols-4 lg:grid-cols-6">
        {#each Array(12) as _, i (i)}
          <div class="space-y-2">
            <Skeleton rounded="card" height="0" class="aspect-[2/3] w-full" />
            <Skeleton height="0.875rem" width="80%" />
          </div>
        {/each}
      </div>
    </div>
  {:else}
    <a
      href="#main-content"
      class="sr-only focus:not-sr-only focus:absolute focus:left-4 focus:top-4 focus:z-50 focus:rounded-lg focus:bg-primary focus:px-3 focus:py-2 focus:text-sm focus:text-primary-fg"
    >
      {i18n.t("a11y.skipToContent")}
    </a>
    <div class="flex h-[100dvh] overflow-hidden">
      <aside
        class="hidden h-full min-h-0 shrink-0 overflow-hidden border-r border-border bg-bg-elevated transition-[width] duration-200 md:block
        {ui.sidebarCollapsed ? 'w-16' : 'w-60'}"
      >
        <Sidebar collapsed={ui.sidebarCollapsed} />
      </aside>

      <div class="flex min-h-0 min-w-0 flex-1 flex-col">
        <ConnectivityBanner />
        <PwaUpdateBanner />
        <Topbar bookTitle={ui.pageTitle} />
        <main
          id="main-content"
          tabindex="-1"
          class="min-h-0 flex-1 outline-none pb-bottom-chrome
            {route.name === 'settings' ? 'flex flex-col overflow-hidden' : 'overflow-y-auto'}"
        >
          {#if route.name === "library" || route.name === "collection"}
            <LibraryView />
          {:else if route.name === "book"}
            {#await import("$lib/views/BookView.svelte")}
              <div class="grid place-items-center p-8 text-sm text-muted">
                {i18n.t("common.loading")}
              </div>
            {:then { default: BookView }}
              <BookView id={bookId} />
            {:catch}
              <ErrorView
                title={i18n.t("error.chunkTitle")}
                message={i18n.t("error.chunkMessage")}
                onRetry={reloadPage}
              />
            {/await}
          {:else if route.name === "settings"}
            {#await import("$lib/views/SettingsView.svelte")}
              <div class="grid place-items-center p-8 text-sm text-muted">
                {i18n.t("common.loading")}
              </div>
            {:then { default: SettingsView }}
              <SettingsView tab={route.params.tab || "library"} />
            {:catch}
              <ErrorView
                title={i18n.t("error.chunkTitle")}
                message={i18n.t("error.chunkMessage")}
                onRetry={reloadPage}
              />
            {/await}
          {:else if route.name === "collections"}
            {#await import("$lib/views/CollectionsView.svelte")}
              <div class="grid place-items-center p-8 text-sm text-muted">
                {i18n.t("common.loading")}
              </div>
            {:then { default: CollectionsView }}
              <CollectionsView />
            {:catch}
              <ErrorView
                title={i18n.t("error.chunkTitle")}
                message={i18n.t("error.chunkMessage")}
                onRetry={reloadPage}
              />
            {/await}
          {:else if route.name === "notfound"}
            <ErrorView code={404} />
          {:else}
            <ErrorView code={404} />
          {/if}
        </main>
      </div>
    </div>

    {#if ui.mobileNavOpen}
      <div class="fixed inset-0 z-50 md:hidden">
        <button
          type="button"
          tabindex="-1"
          aria-label={i18n.t("topbar.closeNav")}
          class="absolute inset-0 bg-overlay"
          transition:fade={{ duration: 150 }}
          onclick={() => ui.closeMobileNav()}
        ></button>
        <div
          id="mobile-nav"
          bind:this={mobileNavPanel}
          role="dialog"
          aria-modal="true"
          aria-label={i18n.t("a11y.navigation")}
          class="absolute left-0 top-0 flex h-full min-h-0 w-64 flex-col overflow-hidden border-r border-border bg-bg-elevated"
          transition:fly={{ x: -260, duration: 200 }}
        >
          <button
            type="button"
            class="btn btn-ghost absolute right-2 top-2 z-10"
            bind:this={mobileNavCloseBtn}
            aria-label={i18n.t("topbar.closeNav")}
            onclick={() => ui.closeMobileNav()}
          >
            <X size={18} />
          </button>
          <Sidebar onnavigate={() => ui.closeMobileNav()} />
        </div>
      </div>
    {/if}

    <BottomNav />
  {/if}

  {#if audioPlayer.active}
    <AudioMiniPlayer />
  {/if}
  {#if narrator.showBar}
    <NarratorBar />
  {/if}

  <ConfirmDialogHost />
  <ToastHost />
  <CommandPalette />

  {#snippet failed(_error, reset)}
    <ErrorView
      title={i18n.t("error.crashTitle")}
      message={i18n.t("error.crashMessage")}
      onRetry={reset}
    />
  {/snippet}
</svelte:boundary>
