<script lang="ts">
  import { fly } from "svelte/transition";
  import { brand } from "$lib/brand";
  import { router } from "$lib/router.svelte";
  import { auth } from "$lib/stores/auth.svelte";
  import { theme } from "$lib/stores/theme.svelte";
  import { ApiError } from "$lib/api/client";
  import { isLoginChallenge } from "$lib/api/types";
  import { safeReturnPath } from "$lib/auth-redirect";
  import { i18n } from "$lib/stores/i18n.svelte";
  import { scorePassword } from "$lib/utils/password-strength";
  import AltchaWidget from "$lib/components/AltchaWidget.svelte";
  import LanguagePicker from "$lib/components/LanguagePicker.svelte";
  import PasswordStrength from "$lib/components/PasswordStrength.svelte";
  import ThemeToggle from "$lib/components/ThemeToggle.svelte";

  interface Props {
    onSuccess?: () => void;
  }

  let { onSuccess }: Props = $props();

  let username = $state("");
  let password = $state("");
  let totpCode = $state("");
  let totpToken = $state("");
  let mode = $state<"login" | "totp" | "register">("login");
  let altchaPayload = $state("");
  let error = $state<string | null>(null);
  let loading = $state(false);
  let usernameEl = $state<HTMLInputElement | null>(null);

  const showLocal = $derived(auth.methods?.loginLocal !== false);
  const showOidc = $derived(auth.methods?.loginOidc === true);
  const allowRegistration = $derived(!!auth.methods?.allowRegistration);
  const oidcLabel = $derived(auth.methods?.oidcButtonText || i18n.t("auth.signInWithSso"));
  const altchaRequired = $derived(!!(auth.altcha?.enabled && auth.altcha.protectLogin));
  const showcaseSrc = $derived(
    `${import.meta.env.BASE_URL}showcase/library-${theme.mode === "light" ? "light" : "dark"}.jpg`,
  );
  const passwordOk = $derived(scorePassword(password, auth.passwordPolicy).valid);

  const reasonMessage = $derived.by(() => {
    if (typeof window === "undefined") return null;
    const reason = new URLSearchParams(window.location.search).get("reason");
    switch (reason) {
      case "session_expired":
        return i18n.t("auth.sessionExpired");
      case "logged_out":
        return i18n.t("auth.loggedOut");
      case "required":
        return i18n.t("auth.signInRequired");
      default:
        return null;
    }
  });

  function oidcErrorMessage(code: string): string {
    const key = code.trim().toLowerCase().replace(/\s+/g, "_");
    switch (key) {
      case "access_denied":
        return i18n.t("auth.oidcAccessDenied");
      case "invalid_state":
      case "state_mismatch":
      case "nonce_mismatch":
        return i18n.t("auth.oidcStateInvalid");
      case "config_error":
      case "provider_error":
        return i18n.t("auth.oidcProviderError");
      case "account_error":
        return i18n.t("auth.oidcAccountError");
      case "session_error":
        return i18n.t("auth.oidcSessionError");
      default:
        return i18n.t("auth.oidcFailed");
    }
  }

  function loginRedirectPath(): string {
    if (typeof window === "undefined") return "/";
    const next = new URLSearchParams(window.location.search).get("next");
    return safeReturnPath(next) ?? "/";
  }

  $effect(() => {
    if (auth.loading || typeof window === "undefined") return;
    const params = new URLSearchParams(window.location.search);
    const oidcError = params.get("oidc_error");
    if (oidcError) {
      error = oidcErrorMessage(oidcError);
      router.navigate("/login", true);
      return;
    }
    if (params.get("oidc") === "1") {
      router.navigate(loginRedirectPath(), true);
      return;
    }
    if (auth.methods?.oidcAutoLaunch && showOidc && !showLocal && mode === "login") {
      window.location.href = "/auth/oidc/login";
    }
  });

  $effect(() => {
    if (!showLocal || auth.loading || mode !== "login") return;
    queueMicrotask(() => usernameEl?.focus());
  });

  async function submit(event: Event) {
    event.preventDefault();
    if (mode === "totp") {
      await submitTotp();
      return;
    }
    if (mode === "register") {
      await submitRegister();
      return;
    }
    if (altchaRequired && !altchaPayload) {
      error = i18n.t("auth.altchaRequired");
      return;
    }
    loading = true;
    error = null;
    try {
      const result = await auth.login(username, password, altchaPayload || undefined);
      if (isLoginChallenge(result)) {
        totpToken = result.totpToken;
        totpCode = "";
        mode = "totp";
        return;
      }
      onSuccess?.();
      router.navigate(loginRedirectPath());
    } catch (e) {
      error = e instanceof ApiError ? e.message : i18n.t("auth.loginFailed");
    } finally {
      loading = false;
    }
  }

  async function submitTotp() {
    loading = true;
    error = null;
    try {
      await auth.verifyTotp(totpToken, totpCode);
      onSuccess?.();
      router.navigate(loginRedirectPath());
    } catch (e) {
      error = e instanceof ApiError ? e.message : i18n.t("auth.totpInvalid");
    } finally {
      loading = false;
    }
  }

  async function submitRegister() {
    if (!passwordOk) {
      error = i18n.t("auth.passwordWeak");
      return;
    }
    if (altchaRequired && !altchaPayload) {
      error = i18n.t("auth.altchaRequired");
      return;
    }
    loading = true;
    error = null;
    try {
      await auth.registerPublic(username, password, altchaPayload || undefined);
      onSuccess?.();
      router.navigate(loginRedirectPath());
    } catch (e) {
      error = e instanceof ApiError ? e.message : i18n.t("auth.registerFailed");
    } finally {
      loading = false;
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
    <h1 class="mt-3 text-center text-lg font-medium text-muted">
      {#if mode === "totp"}
        {i18n.t("auth.totpTitle")}
      {:else if mode === "register"}
        {i18n.t("auth.createAccount")}
      {:else}
        {i18n.t("auth.signIn")}
      {/if}
    </h1>
    <p class="mt-1 text-center text-sm text-subtle">
      {#if mode === "totp"}
        {i18n.t("auth.totpHint")}
      {:else if mode === "register"}
        {i18n.t("auth.registerHint")}
      {:else if showLocal && showOidc}
        {i18n.t("auth.signInLocalAndSso")}
      {:else if showOidc}
        {i18n.t("auth.signInSso")}
      {:else}
        {i18n.t("auth.signInLocal")}
      {/if}
    </p>

    {#if reasonMessage && mode === "login"}
      <p
        class="mt-6 rounded-lg border border-border bg-surface/80 px-3 py-2 text-center text-sm text-muted"
        role="status"
      >
        {reasonMessage}
      </p>
    {/if}

    {#if showOidc && mode === "login"}
      <a href="/auth/oidc/login" class="btn btn-primary mt-8 w-full text-center">
        {oidcLabel}
      </a>
    {/if}

    {#if showLocal || mode === "totp"}
      <form class="mt-8 space-y-4" onsubmit={submit}>
        {#if mode === "login" && showOidc}
          <p class="text-center text-xs uppercase tracking-wide text-subtle">
            {i18n.t("auth.orLocal")}
          </p>
        {/if}

        {#if mode === "totp"}
          <label class="block">
            <span class="text-sm text-muted">{i18n.t("auth.totpCode")}</span>
            <input
              type="text"
              inputmode="numeric"
              autocomplete="one-time-code"
              bind:value={totpCode}
              required
              maxlength={8}
              class="input mt-1"
            />
          </label>
        {:else}
          <label class="block">
            <span class="text-sm text-muted">{i18n.t("auth.username")}</span>
            <input
              bind:this={usernameEl}
              type="text"
              autocomplete="username"
              bind:value={username}
              required
              class="input mt-1"
            />
          </label>
          <label class="block">
            <span class="text-sm text-muted">{i18n.t("auth.password")}</span>
            <input
              type="password"
              autocomplete={mode === "register" ? "new-password" : "current-password"}
              bind:value={password}
              required
              class="input mt-1"
            />
            {#if mode === "register"}
              <PasswordStrength {password} policy={auth.passwordPolicy} />
            {/if}
          </label>
          {#if altchaRequired && auth.altcha}
            <AltchaWidget config={auth.altcha} bind:value={altchaPayload} required />
          {/if}
        {/if}

        {#if error}
          <p class="text-sm text-danger" role="alert">{error}</p>
        {/if}
        <button
          type="submit"
          class="btn btn-primary w-full"
          disabled={loading ||
            (mode === "register" && !passwordOk) ||
            (mode !== "totp" && altchaRequired && !altchaPayload)}
        >
          {#if loading}
            {i18n.t("auth.signingIn")}
          {:else if mode === "totp"}
            {i18n.t("auth.verifyCode")}
          {:else if mode === "register"}
            {i18n.t("auth.createAccount")}
          {:else}
            {i18n.t("auth.signIn")}
          {/if}
        </button>
      </form>

      {#if mode === "totp"}
        <button
          type="button"
          class="mt-4 w-full text-center text-sm text-muted underline"
          onclick={() => {
            mode = "login";
            totpToken = "";
            totpCode = "";
            error = null;
          }}
        >
          {i18n.t("auth.backToSignIn")}
        </button>
      {:else if allowRegistration}
        <button
          type="button"
          class="mt-4 w-full text-center text-sm text-muted underline"
          onclick={() => {
            mode = mode === "register" ? "login" : "register";
            error = null;
          }}
        >
          {mode === "register" ? i18n.t("auth.backToSignIn") : i18n.t("auth.needAccount")}
        </button>
      {/if}
    {:else if error}
      <p class="mt-4 text-center text-sm text-danger" role="alert">{error}</p>
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
