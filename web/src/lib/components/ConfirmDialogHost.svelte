<script lang="ts">
  import { AlertDialog } from "bits-ui";
  import { CircleAlert } from "@lucide/svelte";
  import Button from "./Button.svelte";
  import { confirmDialog } from "$lib/stores/confirm.svelte";

  let confirmBtn = $state<HTMLButtonElement | null>(null);

  function onOpenChange(next: boolean) {
    if (!next) confirmDialog.cancel();
  }

  function onOpenAutoFocus(event: Event) {
    event.preventDefault();
    confirmBtn?.focus();
  }
</script>

<AlertDialog.Root open={confirmDialog.open} {onOpenChange}>
  <AlertDialog.Portal>
    <AlertDialog.Overlay class="confirm-backdrop" />
    <AlertDialog.Content class="confirm-panel" interactOutsideBehavior="close" {onOpenAutoFocus}>
      <AlertDialog.Title class="confirm-title">
        {#if confirmDialog.danger}
          <CircleAlert size={18} class="confirm-title-icon" aria-hidden="true" />
        {/if}
        {confirmDialog.title}
      </AlertDialog.Title>
      <AlertDialog.Description class="confirm-message">
        {confirmDialog.message}
      </AlertDialog.Description>
      <div class="confirm-actions">
        <AlertDialog.Cancel>
          {#snippet child({ props })}
            <Button {...props} variant="ghost" class="ring-1 ring-border">
              {confirmDialog.cancelLabel}
            </Button>
          {/snippet}
        </AlertDialog.Cancel>
        <AlertDialog.Action
          bind:ref={confirmBtn}
          class="btn min-h-11 min-w-[5.5rem] {confirmDialog.danger ? 'btn-danger' : 'btn-primary'}"
          onclick={() => confirmDialog.accept()}
        >
          {confirmDialog.confirmLabel}
        </AlertDialog.Action>
      </div>
    </AlertDialog.Content>
  </AlertDialog.Portal>
</AlertDialog.Root>

<style>
  :global(.confirm-backdrop) {
    position: fixed;
    inset: 0;
    z-index: 70;
    border: 0;
    background: var(--overlay);
    animation: confirm-fade 140ms ease-out;
  }

  :global(.confirm-panel) {
    position: fixed;
    z-index: 70;
    left: 50%;
    top: 50%;
    transform: translate(-50%, -50%);
    width: min(100% - 2rem, 24rem);
    border-radius: 0.75rem;
    border: 1px solid var(--color-border);
    background: var(--color-bg-elevated);
    box-shadow: var(--shadow);
    padding: 1.25rem;
    outline: none;
    animation: confirm-in 160ms ease-out;
  }

  :global(.confirm-title) {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    margin: 0;
    font-size: 1.0625rem;
    font-weight: 600;
    color: var(--color-fg);
  }

  :global(.confirm-title .confirm-title-icon) {
    flex-shrink: 0;
    color: var(--color-danger);
  }

  :global(.confirm-message) {
    margin: 0.5rem 0 0;
    font-size: 0.875rem;
    line-height: 1.5;
    color: var(--color-muted);
  }

  .confirm-actions {
    display: flex;
    flex-wrap: wrap;
    justify-content: flex-end;
    gap: 0.5rem;
    margin-top: 1.25rem;
  }

  @keyframes confirm-fade {
    from {
      opacity: 0;
    }
    to {
      opacity: 1;
    }
  }

  @keyframes confirm-in {
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
