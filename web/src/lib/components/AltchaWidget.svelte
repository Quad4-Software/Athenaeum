<script lang="ts">
  import type { AltchaPublic } from "$lib/api/types";
  import { i18n } from "$lib/stores/i18n.svelte";
  import { theme } from "$lib/stores/theme.svelte";
  import type { Configuration } from "altcha/types";
  import type {} from "altcha/types/svelte";

  interface Props {
    config: AltchaPublic;
    value?: string;
    required?: boolean;
  }

  // Bindable payload starts empty until the widget verifies.
  // eslint-disable-next-line no-useless-assignment -- $bindable initial value for parents
  let { config, value = $bindable(""), required = false }: Props = $props();

  const language = $derived.by(() => {
    const configured = config.widget.language?.trim();
    if (configured) return configured;
    const locale = i18n.locale?.trim();
    if (locale && locale.length >= 2) return locale.slice(0, 2);
    return undefined;
  });

  const widgetTheme = $derived.by(() => {
    const configured = config.widget.theme?.trim();
    if (!configured || configured === "auto") return theme.mode;
    return configured;
  });

  const configuration = $derived.by(() => {
    const opts: { hideFooter?: boolean; hideLogo?: boolean } = {};
    if (config.widget.hideFooter) opts.hideFooter = true;
    if (config.widget.hideLogo) opts.hideLogo = true;
    return Object.keys(opts).length > 0 ? JSON.stringify(opts) : undefined;
  });

  const auto = $derived((config.widget.auto || undefined) as Configuration["auto"] | undefined);
  const display = $derived(
    (config.widget.display || undefined) as Configuration["display"] | undefined,
  );
  const type = $derived((config.widget.type || undefined) as Configuration["type"] | undefined);

  const workers = $derived(
    config.widget.workers && config.widget.workers > 0 ? config.widget.workers : undefined,
  );

  const altchaReady = $derived.by(() => {
    const lang = language;
    return (async () => {
      await import("altcha");
      if (lang) await import("altcha/i18n");
    })();
  });

  function onStateChange(event: CustomEvent<{ payload?: string; state: string }>) {
    const { state, payload } = event.detail;
    if (state === "verified" && payload) {
      value = payload;
    } else {
      value = "";
    }
  }
</script>

{#await altchaReady then}
  <div class="altcha-wrap w-full">
    <altcha-widget
      challenge={config.challengeUrl || undefined}
      {auto}
      {display}
      {type}
      name={config.widget.name || undefined}
      theme={widgetTheme}
      {language}
      {configuration}
      {workers}
      aria-required={required || undefined}
      onstatechange={onStateChange}
    ></altcha-widget>
  </div>
{/await}

<style>
  .altcha-wrap {
    margin-top: 0.5rem;
    --altcha-max-width: 100%;
    --altcha-border-color: var(--border);
    --altcha-border-radius: var(--radius-sm);
    --altcha-color-base: var(--surface);
    --altcha-color-base-content: var(--fg);
    --altcha-color-primary: var(--primary);
    --altcha-color-primary-content: var(--primary-fg);
    --altcha-color-error: var(--danger);
    --altcha-color-success: var(--success);
    --altcha-color-neutral: var(--bg-elevated);
    --altcha-color-neutral-content: var(--fg-muted);
  }

  .altcha-wrap :global(altcha-widget) {
    width: 100%;
  }
</style>
