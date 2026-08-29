<script lang="ts">
  import { Check } from "@lucide/svelte";

  interface Props {
    percent: number;
  }

  let { percent }: Props = $props();

  let hasProgress = $derived(percent > 0.005);
  let completed = $derived(percent >= 0.95);
  let progressPct = $derived(Math.min(100, Math.round(percent * 100)));
</script>

{#if hasProgress}
  <div class="pointer-events-none absolute inset-x-0 bottom-0 h-1 bg-black/25" aria-hidden="true">
    <div
      class="h-full transition-[width] duration-300 {completed ? 'bg-success' : 'bg-primary'}"
      style:width="{progressPct}%"
    ></div>
  </div>
  {#if completed}
    <span
      class="pointer-events-none absolute left-1.5 top-1.5 inline-flex items-center gap-0.5 rounded-md bg-success/90 px-1.5 py-0.5 text-[10px] font-semibold text-white shadow-sm backdrop-blur-sm"
    >
      <Check size={10} strokeWidth={3} />
      Done
    </span>
  {:else}
    <span
      class="pointer-events-none absolute left-1.5 top-1.5 rounded-md bg-black/55 px-1.5 py-0.5 text-[10px] font-semibold tabular-nums text-white backdrop-blur-sm"
    >
      {progressPct}%
    </span>
  {/if}
{/if}
