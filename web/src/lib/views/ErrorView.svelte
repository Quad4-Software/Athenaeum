<script lang="ts">
  import { router } from "$lib/router.svelte";
  import { auth } from "$lib/stores/auth.svelte";
  import { i18n } from "$lib/stores/i18n.svelte";
  import Button from "$lib/components/Button.svelte";
  import { loginUrl, safeReturnPath } from "$lib/auth-redirect";

  interface Props {
    code?: number | "offline";
    title?: string;
    message?: string;
    compact?: boolean;
    onRetry?: () => void;
  }

  let { code = 404, title, message, compact = false, onRetry }: Props = $props();

  const customCopy = $derived(Boolean(title && message));

  const resolved = $derived.by(() => {
    if (title && message) {
      return { title, message };
    }
    switch (code) {
      case 401:
        return {
          title: i18n.t("error.unauthorizedTitle"),
          message: i18n.t("error.unauthorizedMessage"),
        };
      case 403:
        return {
          title: i18n.t("error.forbiddenTitle"),
          message: i18n.t("error.forbiddenMessage"),
        };
      case 404:
        return {
          title: i18n.t("error.notFoundTitle"),
          message: i18n.t("error.notFoundMessage"),
        };
      case 500:
      case 502:
      case 503:
        return {
          title: i18n.t("error.serverTitle"),
          message: i18n.t("error.serverMessage"),
        };
      case "offline":
        return {
          title: i18n.t("error.offlineTitle"),
          message: i18n.t("error.offlineMessage"),
        };
      default:
        return {
          title: i18n.t("error.genericTitle"),
          message: i18n.t("error.genericMessage"),
        };
    }
  });

  function goLogin() {
    const returnTo =
      typeof window !== "undefined" ? safeReturnPath(window.location.pathname) : null;
    router.navigate(loginUrl("required", returnTo));
  }

  function retry() {
    if (onRetry) {
      onRetry();
      return;
    }
    if (code === "offline") {
      window.location.reload();
    }
  }
</script>

<section
  class="flex flex-col items-center justify-center px-4 text-center
    {compact ? 'py-10' : 'min-h-[60vh] py-16'}"
  aria-live="polite"
>
  {#if !compact && !customCopy}
    <p class="text-6xl font-bold tabular-nums text-subtle">{code === "offline" ? "..." : code}</p>
  {/if}
  <h1 class="mt-2 text-xl font-semibold text-fg sm:text-2xl">{resolved.title}</h1>
  <p class="mt-2 max-w-md text-sm text-muted">{resolved.message}</p>
  <div class="mt-6 flex flex-wrap justify-center gap-2">
    {#if code === 401 || (code === 403 && auth.authEnabled)}
      <Button onclick={goLogin}>{i18n.t("error.signIn")}</Button>
    {/if}
    <Button variant="ghost" onclick={() => router.navigate("/")}>
      {i18n.t("app.goToLibrary")}
    </Button>
    {#if onRetry || code === "offline"}
      <Button variant="ghost" onclick={retry}>
        {i18n.t("error.retry")}
      </Button>
    {/if}
  </div>
</section>
