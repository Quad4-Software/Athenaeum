<script lang="ts">
  import type { Snippet } from "svelte";

  type Size = "sm" | "md";

  interface Props {
    title: string;
    body?: string;
    icon?: Snippet<[number]>;
    size?: Size;
    class?: string;
    children?: Snippet;
  }

  let { title, body, icon, size = "md", class: className = "", children }: Props = $props();

  let isSm = $derived(size === "sm");
  let iconSize = $derived(isSm ? 18 : 28);
</script>

<div
  class="flex flex-col items-center justify-center text-center {isSm
    ? 'gap-3 py-8'
    : 'gap-5 py-24'} {className}"
>
  {#if icon}
    <div
      class="grid place-items-center bg-surface text-muted ring-1 ring-border {isSm
        ? 'h-10 w-10 rounded-xl'
        : 'h-16 w-16 rounded-2xl'}"
    >
      {@render icon(iconSize)}
    </div>
  {/if}
  <div class="space-y-1.5">
    <h2 class="font-semibold text-fg {isSm ? 'text-sm' : 'text-lg'}">{title}</h2>
    {#if body}
      <p class="mx-auto max-w-sm text-muted {isSm ? 'text-xs' : 'text-sm'}">{body}</p>
    {/if}
  </div>
  {#if children}
    {@render children()}
  {/if}
</div>
