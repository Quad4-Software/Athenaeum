<script lang="ts">
  import { ChevronRight } from "@lucide/svelte";
  import { link } from "$lib/router.svelte";

  export interface Crumb {
    label: string;
    href?: string;
  }

  interface Props {
    items: Crumb[];
  }

  let { items }: Props = $props();
</script>

<nav aria-label="Breadcrumb" class="flex min-w-0 items-center gap-1 text-sm text-muted">
  {#each items as item, i (item.label + i)}
    {#if i > 0}
      <ChevronRight size={14} class="shrink-0 text-subtle" aria-hidden="true" />
    {/if}
    {#if item.href && i < items.length - 1}
      <a href={item.href} use:link={item.href} class="truncate hover:text-fg">{item.label}</a>
    {:else}
      <span class="truncate text-fg" aria-current={i === items.length - 1 ? "page" : undefined}>
        {item.label}
      </span>
    {/if}
  {/each}
</nav>
