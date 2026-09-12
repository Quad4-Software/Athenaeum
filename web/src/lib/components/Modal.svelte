<script lang="ts">
  import type { Snippet } from "svelte";
  import { Dialog } from "bits-ui";

  interface Props {
    open?: boolean;
    title?: string;
    label?: string;
    class?: string;
    children: Snippet;
    onclose?: () => void;
  }

  let {
    open = $bindable(false),
    title,
    label,
    class: className = "",
    children,
    onclose,
  }: Props = $props();

  function onOpenChange(next: boolean) {
    if (!next) onclose?.();
  }
</script>

<Dialog.Root bind:open {onOpenChange}>
  <Dialog.Portal>
    <Dialog.Overlay class="modal-overlay" />
    <Dialog.Content class="modal-panel {className}" aria-label={label}>
      {#if title}
        <Dialog.Title class="modal-title">{title}</Dialog.Title>
      {/if}
      {@render children()}
    </Dialog.Content>
  </Dialog.Portal>
</Dialog.Root>

<style>
  :global(.modal-overlay) {
    position: fixed;
    inset: 0;
    z-index: 60;
    background: var(--color-overlay);
    animation: modal-fade 140ms ease-out;
  }

  :global(.modal-panel) {
    position: fixed;
    z-index: 70;
    left: 50%;
    top: 50%;
    transform: translate(-50%, -50%);
    width: min(32rem, calc(100vw - 2rem));
    max-height: min(32rem, calc(100vh - 2rem));
    border-radius: var(--radius-card);
    background: var(--color-surface);
    color: var(--color-fg);
    box-shadow: var(--shadow);
    border: 1px solid var(--color-border);
    outline: none;
    animation: modal-in 160ms ease-out;
  }

  :global(.modal-title) {
    margin: 0;
    font-size: 1.0625rem;
    font-weight: 600;
    color: var(--color-fg);
  }

  @keyframes modal-fade {
    from {
      opacity: 0;
    }
    to {
      opacity: 1;
    }
  }

  @keyframes modal-in {
    from {
      opacity: 0;
      transform: translate(-50%, -50%) scale(0.96);
    }
    to {
      opacity: 1;
      transform: translate(-50%, -50%) scale(1);
    }
  }
</style>
