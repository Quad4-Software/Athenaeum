<script lang="ts">
  interface Props {
    open: boolean;
    onclose: () => void;
    items: { keys: string; action: string }[];
  }

  let { open, onclose, items }: Props = $props();

  function onKey(event: KeyboardEvent) {
    if (event.key === "Escape") onclose();
  }
</script>

<svelte:window onkeydown={onKey} />

{#if open}
  <div class="fixed inset-0 z-50 grid place-items-center p-4">
    <button
      type="button"
      class="absolute inset-0 bg-overlay"
      aria-label="Close shortcuts"
      onclick={onclose}
    ></button>
    <div
      class="relative w-full max-w-sm rounded-[var(--radius-card)] border border-border bg-surface p-5 shadow-[var(--shadow)]"
      role="dialog"
      aria-modal="true"
      aria-labelledby="shortcuts-title"
    >
      <h2 id="shortcuts-title" class="text-lg font-semibold text-fg">Keyboard shortcuts</h2>
      <ul class="mt-4 space-y-2 text-sm">
        {#each items as item (item.action)}
          <li class="flex justify-between gap-4">
            <span class="text-muted">{item.action}</span>
            <kbd class="rounded border border-border bg-bg px-2 py-0.5 font-mono text-xs text-fg">
              {item.keys}
            </kbd>
          </li>
        {/each}
      </ul>
      <button type="button" class="btn btn-primary mt-5 w-full" onclick={onclose}>Close</button>
    </div>
  </div>
{/if}
