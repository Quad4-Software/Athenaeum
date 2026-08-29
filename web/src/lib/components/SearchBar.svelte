<script lang="ts">
  import { Search, X } from "@lucide/svelte";
  import { library } from "$lib/stores/library.svelte";
  import { i18n } from "$lib/stores/i18n.svelte";

  let inputEl = $state<HTMLInputElement | null>(null);
  let value = $derived(library.search);

  function update(next: string) {
    library.setSearch(next);
  }

  function onWindowKeydown(event: KeyboardEvent) {
    if (event.defaultPrevented || event.altKey) return;
    const meta = event.metaKey || event.ctrlKey;
    if (!meta || event.key.toLowerCase() !== "k") return;
    event.preventDefault();
    inputEl?.focus();
    inputEl?.select();
  }
</script>

<svelte:window onkeydown={onWindowKeydown} />

<div class="relative w-full max-w-md">
  <span class="pointer-events-none absolute inset-y-0 left-3 flex items-center text-subtle">
    <Search size={16} />
  </span>
  <input
    bind:this={inputEl}
    type="search"
    placeholder={i18n.t("search.placeholder")}
    aria-keyshortcuts="Control+K Meta+K"
    aria-label={i18n.t("search.placeholder")}
    {value}
    oninput={(e) => update(e.currentTarget.value)}
    class="input h-10 w-full pl-9 {value ? 'pr-9' : 'pr-14'}"
  />
  {#if value}
    <button
      type="button"
      aria-label={i18n.t("search.clear")}
      class="absolute inset-y-0 right-2 flex items-center text-subtle hover:text-fg"
      onclick={() => update("")}
    >
      <X size={16} />
    </button>
  {:else}
    <kbd
      class="pointer-events-none absolute inset-y-0 right-2 hidden items-center sm:flex"
      aria-hidden="true"
    >
      <span
        class="self-center rounded border border-border bg-bg-elevated px-1.5 py-0.5 text-[10px] font-medium text-subtle"
      >
        {i18n.t("search.shortcut")}
      </span>
    </kbd>
  {/if}
</div>
