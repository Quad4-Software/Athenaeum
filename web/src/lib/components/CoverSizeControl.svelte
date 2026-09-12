<script lang="ts">
  import { COVER_SIZES, coverSize } from "$lib/stores/cover-size.svelte";
  import type { CoverSize } from "$lib/stores/cover-size.svelte";
  import { i18n } from "$lib/stores/i18n.svelte";

  const labels: Record<CoverSize, { text: string; key: string }> = {
    s: { text: "S", key: "library.coverSizeSmall" },
    m: { text: "M", key: "library.coverSizeMedium" },
    l: { text: "L", key: "library.coverSizeLarge" },
  };
</script>

<div
  class="flex items-center gap-0.5 rounded-md border border-border bg-surface p-0.5"
  role="group"
  aria-label={i18n.t("library.coverSize")}
>
  {#each COVER_SIZES as size (size)}
    {@const active = coverSize.value === size}
    <button
      type="button"
      class="inline-flex h-6 min-w-7 items-center justify-center rounded-xs px-1.5 text-[0.7rem] font-semibold {active
        ? 'bg-primary text-primary-fg'
        : 'text-muted hover:bg-surface-hover hover:text-fg'}"
      aria-pressed={active}
      aria-label={i18n.t(labels[size].key)}
      title={i18n.t(labels[size].key)}
      onclick={() => coverSize.set(size)}
    >
      {labels[size].text}
    </button>
  {/each}
</div>
