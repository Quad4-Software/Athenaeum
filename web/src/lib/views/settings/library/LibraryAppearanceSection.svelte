<script lang="ts">
  import { onMount } from "svelte";
  import { Moon, Sun } from "@lucide/svelte";
  import FontSelect from "$lib/components/FontSelect.svelte";
  import { listAppThemes, UI_FONT_PRESETS } from "$lib/brand";
  import { loadUiFontCss } from "$lib/brand/load-ui-font";
  import { theme } from "$lib/stores/theme.svelte";
  import { typography } from "$lib/stores/typography.svelte";
  import { i18n } from "$lib/stores/i18n.svelte";
  import type { UiFontId } from "$lib/brand/fonts";
  import type { FontOption } from "$lib/components/FontSelect.svelte";

  let fontOptions = $derived.by(() => {
    void i18n.locale;
    return UI_FONT_PRESETS.map((preset): FontOption => ({
      id: preset.id,
      label: preset.id === "system" ? i18n.t("theme.system") : preset.label,
      sample: preset.sample,
      family: preset.family,
    }));
  });

  onMount(() => {
    for (const preset of UI_FONT_PRESETS) {
      void loadUiFontCss(preset.id);
    }
  });
</script>

<div class="rounded-[var(--radius-card)] border border-border bg-surface p-5">
  <h2 class="text-sm font-semibold text-fg">{i18n.t("settings.appearance")}</h2>
  <p class="mt-1 text-sm text-muted">{i18n.t("settings.appearanceHint")}</p>
  <div class="mt-3 flex flex-wrap gap-2">
    {#each listAppThemes() as appTheme (appTheme.id)}
      <button
        class="btn ring-1 ring-border {theme.preference === appTheme.id ||
        (theme.preference === 'system' && theme.activeThemeId === appTheme.id)
          ? 'bg-primary text-primary-fg'
          : 'btn-ghost'}"
        onclick={() => theme.set(appTheme.id)}
      >
        {#if appTheme.id === "light"}
          <Sun size={16} />
        {:else if appTheme.id === "dark"}
          <Moon size={16} />
        {/if}
        {appTheme.id === "light"
          ? i18n.t("theme.light")
          : appTheme.id === "dark"
            ? i18n.t("theme.dark")
            : appTheme.label}
      </button>
    {/each}
    <button
      class="btn ring-1 ring-border {theme.preference === 'system'
        ? 'bg-primary text-primary-fg'
        : 'btn-ghost'}"
      onclick={() => theme.set("system")}
    >
      {i18n.t("theme.system")}
    </button>
  </div>
  <div class="mt-4 max-w-md">
    <FontSelect
      label={i18n.t("settings.interfaceFont")}
      value={typography.id}
      options={fontOptions}
      onchange={(id) => typography.set(id as UiFontId)}
    />
  </div>
</div>
