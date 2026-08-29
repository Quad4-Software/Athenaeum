<script lang="ts">
  import { fly } from "svelte/transition";
  import { brand } from "$lib/brand";
  import { router } from "$lib/router.svelte";
  import { theme } from "$lib/stores/theme.svelte";
  import { ApiError, api } from "$lib/api/client";
  import { toast } from "$lib/stores/toast.svelte";
  import { i18n } from "$lib/stores/i18n.svelte";
  import LanguagePicker from "$lib/components/LanguagePicker.svelte";
  import ThemeToggle from "$lib/components/ThemeToggle.svelte";
  import type { InviteMeta } from "$lib/api/types";
  import { untrack } from "svelte";

  let meta = $state<InviteMeta | null>(null);
  let loading = $state(true);
  let submitting = $state(false);
  let username = $state("");
  let password = $state("");
  let guestPassword = $state<string | null>(null);
  let guestUsername = $state<string | null>(null);

  const token = $derived(router.current.params.token || "");
  const showcaseSrc = $derived(
    `${import.meta.env.BASE_URL}showcase/library-${theme.mode === "light" ? "light" : "dark"}.jpg`,
  );

  const invalidReason = $derived.by(() => {
    if (!meta || meta.valid) return null;
    switch (meta.reason) {
      case "expired":
        return i18n.t("invite.expired");
      case "revoked":
        return i18n.t("invite.revoked");
      case "accepted":
        return i18n.t("invite.accepted");
      case "not_found":
        return i18n.t("invite.notFound");
      default:
        return i18n.t("invite.invalid");
    }
  });

  $effect(() => {
    const t = token;
    untrack(() => {
      if (!t) {
        loading = false;
        meta = {
          kind: "",
          emailPresent: false,
          valid: false,
          reason: "not_found",
          pocketIdConfigured: false,
        };
        return;
      }
      void loadMeta(t);
    });
  });

  async function loadMeta(t: string) {
    loading = true;
    guestPassword = null;
    guestUsername = null;
    try {
      meta = await api.getInviteMeta(t);
    } catch (e) {
      meta = {
        kind: "",
        emailPresent: false,
        valid: false,
        reason: "not_found",
        pocketIdConfigured: false,
      };
      toast.error(e instanceof ApiError ? e.message : i18n.t("invite.loadFailed"));
    } finally {
      loading = false;
    }
  }

  async function continueWithSSO() {
    if (!token) return;
    submitting = true;
    try {
      const res = await api.acceptInvite(token, {});
      const redirect =
        typeof res.redirect === "string" && res.redirect ? res.redirect : "/auth/oidc/login";
      window.location.href = redirect;
    } catch (e) {
      toast.error(e instanceof ApiError ? e.message : i18n.t("invite.acceptFailed"));
      submitting = false;
    }
  }

  async function acceptPermanent(event: Event) {
    event.preventDefault();
    if (!token) return;
    submitting = true;
    try {
      await api.acceptInvite(token, { username: username.trim(), password });
      router.navigate("/login");
    } catch (e) {
      toast.error(e instanceof ApiError ? e.message : i18n.t("invite.acceptFailed"));
    } finally {
      submitting = false;
    }
  }

  async function acceptGuest(event: Event) {
    event.preventDefault();
    if (!token) return;
    submitting = true;
    try {
      const res = await api.acceptInvite(token, {
        username: username.trim() || undefined,
      });
      const pwd = typeof res.password === "string" ? res.password : null;
      const user = res.user as { username?: string } | undefined;
      guestPassword = pwd;
      guestUsername = user?.username ?? (username.trim() || null);
      toast.success(i18n.t("invite.guestReady"));
    } catch (e) {
      toast.error(e instanceof ApiError ? e.message : i18n.t("invite.acceptFailed"));
    } finally {
      submitting = false;
    }
  }
</script>

<section
  class="login-shell relative flex min-h-[100dvh] w-full flex-col items-center justify-center overflow-hidden px-4 py-12"
>
  <div class="login-showcase pointer-events-none absolute inset-0" aria-hidden="true">
    <img class="login-showcase-img" src={showcaseSrc} alt="" decoding="async" />
  </div>
  <div class="login-glow pointer-events-none absolute inset-0" aria-hidden="true"></div>

  <div
    class="absolute right-3 top-[max(0.75rem,env(safe-area-inset-top))] z-10 flex items-center gap-1 sm:right-4"
  >
    <LanguagePicker />
    <ThemeToggle />
  </div>

  <div
    class="relative z-[1] w-full max-w-sm rounded-[var(--radius-card)] border border-border/60 bg-bg/80 p-6 shadow-panel backdrop-blur-md sm:p-8"
    in:fly={{ y: 12, duration: 280 }}
  >
    <p class="text-center text-3xl font-semibold tracking-tight text-fg sm:text-4xl">
      {brand.appName}
    </p>
    <h1 class="mt-3 text-center text-lg font-medium text-muted">{i18n.t("invite.title")}</h1>
    <p class="mt-1 text-center text-sm text-subtle">{i18n.t("invite.subtitle")}</p>

    {#if loading}
      <p class="mt-8 text-center text-sm text-muted">{i18n.t("common.loading")}</p>
    {:else if invalidReason}
      <p
        class="mt-8 rounded-lg border border-border bg-surface/80 px-3 py-2 text-center text-sm text-muted"
        role="status"
      >
        {invalidReason}
      </p>
      <a href="/login" class="btn btn-primary mt-6 w-full text-center">{i18n.t("invite.goLogin")}</a
      >
    {:else if guestPassword}
      <div class="mt-8 space-y-3 rounded-lg border border-border bg-surface/80 p-4">
        <p class="text-sm font-medium text-fg">{i18n.t("invite.guestCredentials")}</p>
        {#if guestUsername}
          <p class="text-sm text-muted">
            {i18n.t("invite.username")}: <span class="font-mono text-fg">{guestUsername}</span>
          </p>
        {/if}
        <p class="text-sm text-muted">
          {i18n.t("invite.oneTimePassword")}:
          <span class="font-mono text-fg">{guestPassword}</span>
        </p>
        <p class="text-xs text-subtle">{i18n.t("invite.passwordOnce")}</p>
      </div>
      <a href="/login" class="btn btn-primary mt-6 w-full text-center">{i18n.t("invite.goLogin")}</a
      >
    {:else if meta?.kind === "permanent" && meta.pocketIdConfigured}
      <p class="mt-6 text-center text-sm text-muted">{i18n.t("invite.ssoHint")}</p>
      <button
        type="button"
        class="btn btn-primary mt-6 w-full"
        disabled={submitting}
        onclick={continueWithSSO}
      >
        {submitting ? i18n.t("invite.working") : i18n.t("invite.continueSso")}
      </button>
    {:else if meta?.kind === "permanent"}
      <form class="mt-8 space-y-4" onsubmit={acceptPermanent}>
        <label class="block">
          <span class="text-sm text-muted">{i18n.t("invite.username")}</span>
          <input
            type="text"
            autocomplete="username"
            bind:value={username}
            required
            minlength={2}
            class="field-input mt-1"
          />
        </label>
        <label class="block">
          <span class="text-sm text-muted">{i18n.t("invite.password")}</span>
          <input
            type="password"
            autocomplete="new-password"
            bind:value={password}
            required
            class="field-input mt-1"
          />
        </label>
        <button type="submit" class="btn btn-primary w-full" disabled={submitting}>
          {submitting ? i18n.t("invite.working") : i18n.t("invite.accept")}
        </button>
      </form>
    {:else if meta?.kind === "guest"}
      <form class="mt-8 space-y-4" onsubmit={acceptGuest}>
        <label class="block">
          <span class="text-sm text-muted">{i18n.t("invite.usernameOptional")}</span>
          <input
            type="text"
            autocomplete="username"
            bind:value={username}
            class="field-input mt-1"
          />
        </label>
        <p class="text-xs text-subtle">{i18n.t("invite.guestHint")}</p>
        <button type="submit" class="btn btn-primary w-full" disabled={submitting}>
          {submitting ? i18n.t("invite.working") : i18n.t("invite.acceptGuest")}
        </button>
      </form>
    {:else}
      <p class="mt-8 text-center text-sm text-muted">{i18n.t("invite.invalid")}</p>
    {/if}
  </div>
</section>

<style>
  .login-shell {
    background: var(--bg);
  }

  .login-showcase {
    display: grid;
    place-items: center;
    overflow: hidden;
  }

  .login-showcase-img {
    width: min(140%, 58rem);
    max-width: none;
    height: auto;
    object-fit: cover;
    opacity: 0.22;
    transform: rotate(-8deg) scale(1.18);
    filter: saturate(0.85) contrast(1.05);
    border-radius: 1rem;
    box-shadow: var(--shadow);
  }

  :global([data-theme="light"]) .login-showcase-img {
    opacity: 0.28;
  }

  .login-glow {
    background:
      radial-gradient(
        ellipse 80% 50% at 50% -10%,
        color-mix(in oklab, var(--primary) 18%, transparent),
        transparent
      ),
      radial-gradient(
        circle at 50% 120%,
        color-mix(in oklab, var(--accent) 10%, transparent),
        transparent 55%
      ),
      linear-gradient(to bottom, color-mix(in oklab, var(--bg) 35%, transparent), var(--bg) 88%);
  }
</style>
