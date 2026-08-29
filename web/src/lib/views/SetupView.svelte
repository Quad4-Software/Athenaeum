<script lang="ts">
  import { fly } from "svelte/transition";
  import { brand } from "$lib/brand";
  import { router } from "$lib/router.svelte";
  import { auth } from "$lib/stores/auth.svelte";
  import { theme } from "$lib/stores/theme.svelte";
  import { ApiError } from "$lib/api/client";
  import PasswordStrength from "$lib/components/PasswordStrength.svelte";
  import { scorePassword } from "$lib/utils/password-strength";
  import { i18n } from "$lib/stores/i18n.svelte";
  import AltchaWidget from "$lib/components/AltchaWidget.svelte";
  import LanguagePicker from "$lib/components/LanguagePicker.svelte";
  import ThemeToggle from "$lib/components/ThemeToggle.svelte";

  let username = $state("");
  let password = $state("");
  let altchaPayload = $state("");
  let error = $state<string | null>(null);
  let loading = $state(false);
  let usernameEl = $state<HTMLInputElement | null>(null);

  const altchaRequired = $derived(!!(auth.altcha?.enabled && auth.altcha.protectSetup));
  const passwordOk = $derived(scorePassword(password, auth.passwordPolicy).valid);
  const showcaseSrc = $derived(
    `${import.meta.env.BASE_URL}showcase/library-${theme.mode === "light" ? "light" : "dark"}.jpg`,
  );

  $effect(() => {
    queueMicrotask(() => usernameEl?.focus());
  });

  async function submit(event: Event) {
    event.preventDefault();
    if (!passwordOk) {
      error = i18n.t("setup.passwordWeak");
      return;
    }
    if (altchaRequired && !altchaPayload) {
      error = i18n.t("setup.altchaRequired");
      return;
    }
    loading = true;
    error = null;
    try {
      await auth.setup(username.trim(), password, altchaPayload || undefined);
      router.navigate("/");
    } catch (e) {
      error = e instanceof ApiError ? e.message : i18n.t("setup.setupFailed");
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
    <h1 class="mt-3 text-center text-lg font-medium text-muted">{i18n.t("setup.welcome")}</h1>
    <p class="mt-1 text-center text-sm text-subtle">
      {i18n.t("setup.description")}
    </p>

    <form class="mt-8 space-y-4" onsubmit={submit}>
      <label class="block">
        <span class="text-sm text-muted">{i18n.t("setup.adminUsername")}</span>
        <input
          bind:this={usernameEl}
          type="text"
          autocomplete="username"
          bind:value={username}
          required
          minlength="2"
          class="input mt-1"
        />
      </label>
      <label class="block">
        <span class="text-sm text-muted">{i18n.t("auth.password")}</span>
        <input
          type="password"
          autocomplete="new-password"
          bind:value={password}
          required
          class="input mt-1"
        />
        <PasswordStrength {password} policy={auth.passwordPolicy} />
      </label>
      {#if altchaRequired && auth.altcha}
        <AltchaWidget config={auth.altcha} bind:value={altchaPayload} required />
      {/if}
      {#if error}
        <p class="text-sm text-danger" role="alert">{error}</p>
      {/if}
      <button
        type="submit"
        class="btn btn-primary w-full"
        disabled={loading || !passwordOk || (altchaRequired && !altchaPayload)}
      >
        {loading ? i18n.t("setup.creating") : i18n.t("setup.submit")}
      </button>
    </form>
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
