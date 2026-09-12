<script lang="ts">
  import { library } from "$lib/stores/library.svelte";
  import { LETTER_FILTER_OPTIONS } from "$lib/utils/letter-filter";
  import { i18n } from "$lib/stores/i18n.svelte";
</script>

<div
  class="flex flex-wrap items-center gap-1"
  role="group"
  aria-label={i18n.t("library.letterFilter")}
>
  {#each LETTER_FILTER_OPTIONS as letter (letter)}
    {@const active = library.letterFilter === letter}
    <button
      type="button"
      class="inline-flex h-6 min-w-6 items-center justify-center rounded-xs border px-1 text-[0.7rem] font-medium {active
        ? 'border-primary bg-primary text-primary-fg'
        : 'border-border bg-surface text-muted hover:bg-surface-hover hover:text-fg'}"
      aria-pressed={active}
      title={letter === "#" ? i18n.t("library.letterNonAlpha") : undefined}
      onclick={() => library.setLetter(letter)}
    >
      {letter}
    </button>
  {/each}
</div>
