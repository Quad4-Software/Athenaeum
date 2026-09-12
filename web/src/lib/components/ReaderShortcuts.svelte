<script lang="ts">
  import { Dialog } from "bits-ui";
  import Modal from "$lib/components/Modal.svelte";

  interface Props {
    open: boolean;
    onclose: () => void;
    items: { keys: string; action: string }[];
  }

  let { open, onclose, items }: Props = $props();
</script>

<Modal {open} {onclose} label="Keyboard shortcuts" class="shortcuts-panel">
  <Dialog.Title class="text-lg font-semibold text-fg">Keyboard shortcuts</Dialog.Title>
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
  <Dialog.Close type="button" class="btn btn-primary mt-5 w-full">Close</Dialog.Close>
</Modal>

<style>
  :global(.modal-panel.shortcuts-panel) {
    width: min(24rem, calc(100vw - 2rem));
    padding: 1.25rem;
  }
</style>
