<script lang="ts">
  import { Check, Minus, Plus, Type } from "@lucide/svelte";
  import Popover from "$lib/components/Popover.svelte";
  import { i18n } from "$lib/stores/i18n.svelte";
  import {
    BUILTIN_EPUB_FONTS,
    type EpubFontId,
    type StoredCustomFont,
  } from "$lib/reader/epub-fonts";
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
    onFontChange: (fontId: string) => void;
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

<Popover bind:open placement="bottom" align="end" minWidth={288}>
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
      <ul class="font-preview-list" role="listbox" aria-label={i18n.t("reader.font")}>
        {#each BUILTIN_EPUB_FONTS as font (font.id)}
          <li role="none">
            <button
              type="button"
              role="option"
              class="font-preview-option"
              class:font-preview-option--active={fontId === font.id}
              aria-selected={fontId === font.id}
              style:font-family={font.family}
              onclick={() => onFontChange(font.id)}
            >
              <span class="font-preview-text">
                <span class="font-preview-name">{i18n.t(font.labelKey)}</span>
                {#if font.sample}
                  <span class="font-preview-sample">{font.sample}</span>
                {/if}
              </span>
              {#if fontId === font.id}
                <Check size={14} class="font-preview-check" />
              {/if}
            </button>
          </li>
        {/each}
        <li role="none">
          <button
            type="button"
            role="option"
            class="font-preview-option"
            class:font-preview-option--active={fontId === "custom"}
            aria-selected={fontId === "custom"}
            disabled={!customFont}
            onclick={() => onFontChange("custom")}
          >
            <span class="font-preview-text">
              <span class="font-preview-name">
                {customFont
                  ? `${i18n.t("reader.fontCustom")}: ${customFont.fileName}`
                  : i18n.t("reader.fontCustom")}
              </span>
            </span>
            {#if fontId === "custom"}
              <Check size={14} class="font-preview-check" />
            {/if}
          </button>
        </li>
      </ul>
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
      <input type="range" min="1.2" max="2.2" step="0.1" bind:value={lineHeight} class="w-full" />
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

<style>
  .font-preview-list {
    margin: 0;
    padding: 0.25rem;
    list-style: none;
    max-height: 14rem;
    overflow-y: auto;
    border-radius: 0.5rem;
    border: 1px solid var(--color-border);
    background: var(--color-bg);
  }

  .font-preview-option {
    display: flex;
    width: 100%;
    align-items: center;
    gap: 0.5rem;
    border: 0;
    border-radius: 0.5rem;
    padding: 0.45rem 0.5rem;
    color: var(--color-fg);
    background: transparent;
    cursor: pointer;
    transition: background-color 100ms ease;
  }

  .font-preview-option:hover:not(:disabled) {
    background: var(--color-surface-hover);
  }

  .font-preview-option:disabled {
    opacity: 0.45;
    cursor: default;
  }

  .font-preview-option--active {
    color: var(--color-primary);
    background: color-mix(in oklch, var(--color-primary) 10%, transparent);
  }

  .font-preview-text {
    display: flex;
    min-width: 0;
    flex: 1;
    flex-direction: column;
    align-items: flex-start;
    gap: 0.05rem;
    text-align: left;
  }

  .font-preview-name {
    font-size: 0.8125rem;
    font-weight: 600;
    line-height: 1.25;
  }

  .font-preview-sample {
    font-size: 0.75rem;
    line-height: 1.3;
    color: var(--color-muted);
    font-weight: 400;
  }

  .font-preview-option :global(.font-preview-check) {
    flex-shrink: 0;
    color: var(--color-primary);
  }
</style>
