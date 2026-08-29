<script lang="ts">
  import { ChevronRight } from "@lucide/svelte";
  import { slide } from "svelte/transition";
  import { sidebarPrefs } from "$lib/stores/sidebar.svelte";
  import type { Snippet } from "svelte";

  interface Props {
    id: string;
    title: string;
    itemCount: number;
    sidebarCollapsed?: boolean;
    bordered?: boolean;
    defaultExpanded?: boolean;
    headerExtra?: Snippet;
    children: Snippet;
  }

  let {
    id,
    title,
    itemCount,
    sidebarCollapsed = false,
    bordered = true,
    defaultExpanded = true,
    headerExtra,
    children,
  }: Props = $props();

  let collapsible = $derived(itemCount > 1);
  let expanded = $derived(
    sidebarCollapsed ? false : sidebarPrefs.isSectionExpanded(id, defaultExpanded),
  );

  function toggle() {
    if (!collapsible || sidebarCollapsed) return;
    sidebarPrefs.toggleSectionExpanded(id);
  }
</script>

{#if sidebarCollapsed}
  <div class="flex flex-col gap-0.5 overflow-hidden">
    {@render children()}
  </div>
{:else if !collapsible}
  <div class="flex flex-col gap-0.5 {bordered ? 'mt-3 border-t border-border pt-3' : ''}">
    {#if itemCount > 0}
      <div class="mb-1 flex items-center justify-between px-2.5">
        <p class="text-xs font-medium uppercase tracking-wide text-subtle">{title}</p>
        {#if headerExtra}{@render headerExtra()}{/if}
      </div>
    {/if}
    {@render children()}
  </div>
{:else}
  <div class="flex flex-col gap-0.5 {bordered ? 'mt-3 border-t border-border pt-3' : ''}">
    <button
      type="button"
      class="mb-0.5 flex w-full items-center justify-between gap-2 rounded-lg px-2.5 py-1 text-left transition-colors hover:bg-surface-hover"
      aria-expanded={expanded}
      onclick={toggle}
    >
      <span class="flex min-w-0 items-center gap-1.5">
        <ChevronRight
          size={14}
          class="shrink-0 text-subtle transition-transform duration-200 {expanded
            ? 'rotate-90'
            : ''}"
        />
        <span class="truncate text-xs font-medium uppercase tracking-wide text-subtle">{title}</span
        >
        <span class="text-[10px] tabular-nums text-subtle">{itemCount}</span>
      </span>
      {#if headerExtra}{@render headerExtra()}{/if}
    </button>
    {#if expanded}
      <div class="flex flex-col gap-0.5" transition:slide={{ duration: 150 }}>
        {@render children()}
      </div>
    {/if}
  </div>
{/if}
