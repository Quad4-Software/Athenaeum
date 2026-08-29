<script lang="ts">
  import { Minus, Plus, Type } from "@lucide/svelte";
  import Popover from "$lib/components/Popover.svelte";
  import { i18n } from "$lib/stores/i18n.svelte";
  import { BUILTIN_EPUB_FONTS, type EpubFontId, type StoredCustomFont } from "$lib/reader/epub-fonts";
  import type { ReaderTheme } from "$lib/reader/epub-theme";
  import type { EpubSpreadMode } from "$lib/reader/epub-reader";

  interface Props {
    open: boolean;
    fontPct: number;
    readerTheme: ReaderTheme;
    fontId: EpubFontId;
    customFont: StoredCustomFont | null;
    lineHeight: number;
    marginPx: number;
    spreadMode: EpubSpreadMode;
    onSmallerFont: () => void;
    onLargerFont: () => void;
    onFontChange: (event: Event) => void;
    onFontUpload: (event: Event) => void;
    onRemoveCustomFont: () => void;
    onSpreadMode: (mode: EpubSpreadMode) => void;
  }

  let {
    open = $bindable(false),
    fontPct,
    readerTheme = $bindable(),
    fontId,
    customFont,
    lineHeight = $bindable(),
    marginPx = $bindable(),
    spreadMode,
    onSmallerFont,
    onLargerFont,
    onFontChange,
    onFontUpload,
    onRemoveCustomFont,
    onSpreadMode,
  }: Props = $props();

  let fontInput = $state<HTMLInputElement>();
</script>

<input
  type="file"
  accept=".ttf,.otf,.woff,.woff2,font/ttf,font/otf,font/woff,font/woff2"
  class="sr-only"
  bind:this={fontInput}
  onchange={onFontUpload}
/>

<Popover bind:open placement="bottom" align="end" minWidth={272}>
  {#snippet trigger(toggle)}
    <button
      type="button"
      class="btn btn-ghost text-xs"
      class:ring-1={open}
      class:ring-border={open}
      aria-expanded={open}
      aria-label={i18n.t("reader.display")}
      onclick={toggle}
    >
      <Type size={16} />
      <span class="tabular-nums">{fontPct}%</span>
    </button>
  {/snippet}

  <div class="space-y-4 p-1">
    <div>
      <p class="mb-1.5 text-xs font-medium text-muted">{i18n.t("reader.textSize")}</p>
      <div class="flex items-center gap-2">
        <button
          type="button"
          class="btn btn-ghost flex-1 ring-1 ring-border"
          aria-label="Smaller text"
          onclick={onSmallerFont}
        >
          <Minus size={16} />
        </button>
        <span class="w-12 text-center text-sm tabular-nums text-fg">{fontPct}%</span>
        <button
          type="button"
          class="btn btn-ghost flex-1 ring-1 ring-border"
          aria-label="Larger text"
          onclick={onLargerFont}
        >
          <Plus size={16} />
        </button>
      </div>
    </div>

    <label class="block">
      <span class="mb-1.5 block text-xs font-medium text-muted">{i18n.t("reader.theme")}</span>
      <select class="field-input w-full text-sm" bind:value={readerTheme}>
        <option value="light">{i18n.t("reader.themeLight")}</option>
        <option value="dark">{i18n.t("reader.themeDark")}</option>
        <option value="sepia">{i18n.t("reader.themeSepia")}</option>
        <option value="night">{i18n.t("reader.themeNight")}</option>
      </select>
    </label>

    <div>
      <span class="mb-1.5 block text-xs font-medium text-muted">{i18n.t("reader.font")}</span>
      <select class="field-input w-full text-sm" value={fontId} onchange={onFontChange}>
        {#each BUILTIN_EPUB_FONTS as font (font.id)}
          <option value={font.id}>{i18n.t(font.labelKey)}</option>
        {/each}
        <option value="custom" disabled={!customFont}>
          {customFont
            ? `${i18n.t("reader.fontCustom")}: ${customFont.fileName}`
            : i18n.t("reader.fontCustom")}
        </option>
      </select>
      <div class="mt-2 flex flex-wrap gap-2">
        <button
          type="button"
          class="btn btn-ghost text-xs ring-1 ring-border"
          onclick={() => fontInput?.click()}
        >
          {i18n.t("reader.fontUpload")}
        </button>
        {#if customFont}
          <button
            type="button"
            class="btn btn-ghost text-xs ring-1 ring-border"
            onclick={onRemoveCustomFont}
          >
            {i18n.t("reader.fontClear")}
          </button>
        {/if}
      </div>
    </div>

    <label class="block">
      <span class="mb-1.5 flex items-center justify-between text-xs font-medium text-muted">
        <span>{i18n.t("reader.lineSpacing")}</span>
        <span class="tabular-nums">{lineHeight.toFixed(1)}</span>
      </span>
      <input
        type="range"
        min="1.2"
        max="2.2"
        step="0.1"
        bind:value={lineHeight}
        class="w-full"
      />
    </label>

    <label class="block">
      <span class="mb-1.5 flex items-center justify-between text-xs font-medium text-muted">
        <span>{i18n.t("reader.margins")}</span>
        <span class="tabular-nums">{marginPx}px</span>
      </span>
      <input type="range" min="8" max="64" step="4" bind:value={marginPx} class="w-full" />
    </label>

    <label class="block">
      <span class="mb-1.5 block text-xs font-medium text-muted">{i18n.t("reader.spread")}</span>
      <select
        class="field-input w-full text-sm"
        value={spreadMode}
        onchange={(e) => onSpreadMode(e.currentTarget.value as EpubSpreadMode)}
      >
        <option value="single">{i18n.t("reader.spreadSingle")}</option>
        <option value="auto">{i18n.t("reader.spreadAuto")}</option>
        <option value="always">{i18n.t("reader.spreadDouble")}</option>
      </select>
    </label>
  </div>
</Popover>
