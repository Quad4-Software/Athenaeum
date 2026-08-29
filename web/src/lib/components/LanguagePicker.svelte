<script lang="ts">
  import { Check, Globe, Search } from "@lucide/svelte";
  import Popover from "./Popover.svelte";
  import IconButton from "./IconButton.svelte";
  import { i18n } from "$lib/stores/i18n.svelte";

  let open = $state(false);
  let query = $state("");
  let searchEl = $state<HTMLInputElement | null>(null);

  /** Native endonyms for bundled locales. */
  const nativeNames: Record<string, string> = {
    en: "English",
    de: "Deutsch",
    es: "Español",
    fr: "Français",
    pt: "Português",
    "pt-BR": "Português (Brasil)",
    it: "Italiano",
    nl: "Nederlands",
    pl: "Polski",
    ru: "Русский",
    ja: "日本語",
    zh: "中文",
    "zh-CN": "简体中文",
    ko: "한국어",
    sv: "Svenska",
    tr: "Türkçe",
  };

  function displayName(code: string, fallback: string): string {
    return nativeNames[code] ?? fallback;
  }

  let filtered = $derived.by(() => {
    const q = query.trim().toLowerCase();
    const list = [...i18n.locales].sort((a, b) =>
      displayName(a.code, a.name).localeCompare(displayName(b.code, b.name), i18n.locale),
    );
    if (!q) return list;
    return list.filter((loc) => {
      const native = displayName(loc.code, loc.name).toLowerCase();
      return (
        loc.code.toLowerCase().includes(q) ||
        loc.name.toLowerCase().includes(q) ||
        native.includes(q)
      );
    });
  });

  const currentLabel = $derived(
    displayName(i18n.locale, i18n.locales.find((l) => l.code === i18n.locale)?.name ?? i18n.locale),
  );

  function pick(code: string) {
    void i18n.setLocale(code);
    open = false;
    query = "";
  }

  function onMenuClose() {
    query = "";
  }
</script>

<Popover bind:open align="end" minWidth={260} onclose={onMenuClose}>
  {#snippet trigger(toggle)}
    <IconButton
      ariaLabel={i18n.t("language.select")}
      title={`${i18n.t("language.label")}: ${currentLabel}`}
      onclick={toggle}
    >
      <span class="lang-trigger">
        <Globe size={18} />
        <span class="lang-trigger-code">{i18n.locale.toUpperCase()}</span>
      </span>
    </IconButton>
  {/snippet}

  <div class="lang-menu" role="menu" aria-label={i18n.t("language.label")}>
    <p class="lang-title" id="lang-menu-title">{i18n.t("language.label")}</p>
    <div class="lang-search">
      <Search size={14} class="lang-search-icon" />
      <input
        bind:this={searchEl}
        class="lang-search-input"
        type="search"
        value={query}
        oninput={(e) => (query = e.currentTarget.value)}
        placeholder={i18n.t("language.search")}
        aria-label={i18n.t("language.search")}
        autocomplete="off"
        spellcheck="false"
      />
    </div>
    <ul class="lang-list" role="none">
      {#each filtered as loc (loc.code)}
        <li role="none">
          <button
            type="button"
            role="menuitemradio"
            class="lang-item"
            class:lang-item--active={i18n.locale === loc.code}
            aria-checked={i18n.locale === loc.code}
            onclick={() => pick(loc.code)}
          >
            <span class="lang-item-main">
              <span class="lang-item-name">{displayName(loc.code, loc.name)}</span>
              <span class="lang-item-code">{loc.code}</span>
            </span>
            {#if i18n.locale === loc.code}
              <Check size={14} class="lang-item-check" />
            {/if}
          </button>
        </li>
      {:else}
        <li class="lang-empty" role="presentation">{i18n.t("language.noMatches")}</li>
      {/each}
    </ul>
  </div>
</Popover>

<style>
  .lang-trigger {
    display: inline-flex;
    align-items: center;
    gap: 0.25rem;
  }

  .lang-trigger-code {
    font-size: 0.625rem;
    font-weight: 700;
    letter-spacing: 0.04em;
    line-height: 1;
    color: var(--color-muted);
  }

  .lang-menu {
    display: flex;
    flex-direction: column;
    gap: 0.35rem;
    max-height: min(22rem, 70vh);
  }

  .lang-title {
    margin: 0;
    padding: 0.15rem 0.35rem 0;
    font-size: 0.6875rem;
    font-weight: 600;
    letter-spacing: 0.04em;
    text-transform: uppercase;
    color: var(--color-subtle);
  }

  .lang-search {
    display: flex;
    align-items: center;
    gap: 0.4rem;
    margin: 0 0.1rem;
    padding: 0.35rem 0.5rem;
    border: 1px solid var(--color-border);
    border-radius: 0.5rem;
    background: var(--color-surface);
  }

  .lang-search :global(.lang-search-icon) {
    flex-shrink: 0;
    color: var(--color-muted);
  }

  .lang-search-input {
    flex: 1;
    min-width: 0;
    border: 0;
    padding: 0;
    background: transparent;
    color: var(--color-fg);
    font-size: 0.8125rem;
    outline: none;
  }

  .lang-search-input::-webkit-search-cancel-button {
    appearance: none;
  }

  .lang-list {
    margin: 0;
    padding: 0;
    list-style: none;
    overflow-y: auto;
    overscroll-behavior: contain;
  }

  .lang-item {
    display: flex;
    width: 100%;
    align-items: center;
    gap: 0.5rem;
    border: 0;
    border-radius: 0.5rem;
    padding: 0.5rem 0.55rem;
    text-align: left;
    color: var(--color-fg);
    background: transparent;
    cursor: pointer;
    transition: background-color 100ms ease;
  }

  .lang-item:hover {
    background: var(--color-surface-hover);
  }

  .lang-item--active {
    color: var(--color-primary);
    background: color-mix(in oklch, var(--color-primary) 10%, transparent);
  }

  .lang-item-main {
    display: flex;
    flex: 1;
    min-width: 0;
    flex-direction: column;
    gap: 0.1rem;
  }

  .lang-item-name {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    font-size: 0.875rem;
    font-weight: 500;
  }

  .lang-item-code {
    font-size: 0.6875rem;
    color: var(--color-muted);
    font-variant-numeric: tabular-nums;
    letter-spacing: 0.02em;
    text-transform: uppercase;
  }

  .lang-item--active .lang-item-code {
    color: color-mix(in oklch, var(--color-primary) 70%, var(--color-muted));
  }

  .lang-item :global(.lang-item-check) {
    flex-shrink: 0;
    color: var(--color-primary);
  }

  .lang-empty {
    padding: 0.75rem 0.5rem;
    font-size: 0.8125rem;
    color: var(--color-muted);
    text-align: center;
  }
</style>
