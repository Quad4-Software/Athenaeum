<script lang="ts">
  import { Plus } from "@lucide/svelte";
  import type { Collection } from "$lib/api/types";

  interface Props {
    label: string;
    items: Collection[];
    onadd: (collectionId: number) => void;
  }

  let { label, items, onadd }: Props = $props();
</script>

{#if items.length > 0}
  <div class="mt-4">
    <p class="text-sm text-muted">{label}</p>
    <div class="mt-1 flex flex-wrap gap-2">
      {#each items as c (c.id)}
        <button
          type="button"
          class="btn btn-ghost min-h-10 ring-1 ring-border text-xs"
          onclick={() => onadd(c.id)}
        >
          <Plus size={14} />
          {c.name}
        </button>
      {/each}
    </div>
  </div>
{/if}
