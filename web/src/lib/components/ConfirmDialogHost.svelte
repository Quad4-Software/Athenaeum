<script lang="ts">
  import { tick } from "svelte";
  import { fade, scale } from "svelte/transition";
  import { CircleAlert } from "@lucide/svelte";
  import Button from "./Button.svelte";
  import { confirmDialog } from "$lib/stores/confirm.svelte";

  let panelEl = $state<HTMLDivElement | null>(null);
  let confirmBtn = $state<HTMLButtonElement | null>(null);

  const FOCUSABLE =
    'button:not([disabled]), [href], input:not([disabled]), select:not([disabled]), textarea:not([disabled]), [tabindex]:not([tabindex="-1"])';

  function focusables(): HTMLElement[] {
    if (!panelEl) return [];
    return Array.from(panelEl.querySelectorAll<HTMLElement>(FOCUSABLE)).filter(
      (el) => el.getClientRects().length > 0,
    );
  }

  function onKeydown(e: KeyboardEvent) {
    if (!confirmDialog.open) return;
    if (e.key === "Escape") {
      e.preventDefault();
      confirmDialog.cancel();
      return;
    }
    if (e.key !== "Tab" || !panelEl) return;
    const items = focusables();
    if (items.length === 0) return;
    const first = items[0];
    const last = items[items.length - 1];
    const active = document.activeElement;
    if (e.shiftKey) {
      if (active === first || !panelEl.contains(active)) {
        e.preventDefault();
        last.focus();
      }
    } else if (active === last || !panelEl.contains(active)) {
      e.preventDefault();
      first.focus();
    }
  }

  $effect(() => {
    if (!confirmDialog.open) return;
    const prev = document.body.style.overflow;
    document.body.style.overflow = "hidden";
    void tick().then(() => confirmBtn?.focus());
    return () => {
      document.body.style.overflow = prev;
    };
  });
</script>

<svelte:window onkeydown={onKeydown} />

{#if confirmDialog.open}
  <div class="confirm-root" role="presentation">
    <button
      type="button"
      class="confirm-backdrop"
      tabindex="-1"
      aria-label={confirmDialog.cancelLabel}
      transition:fade={{ duration: 140 }}
      onclick={() => confirmDialog.cancel()}
    ></button>
    <div
      bind:this={panelEl}
      class="confirm-panel"
      role="alertdialog"
      aria-modal="true"
      aria-labelledby="confirm-title"
      aria-describedby="confirm-message"
      transition:scale={{ duration: 160, start: 0.96 }}
    >
      <h2 id="confirm-title" class="confirm-title">
        {#if confirmDialog.danger}
          <CircleAlert size={18} class="confirm-title-icon" aria-hidden="true" />
        {/if}
        {confirmDialog.title}
      </h2>
      <p id="confirm-message" class="confirm-message">{confirmDialog.message}</p>
      <div class="confirm-actions">
        <Button variant="ghost" class="ring-1 ring-border" onclick={() => confirmDialog.cancel()}>
          {confirmDialog.cancelLabel}
        </Button>
        <button
          type="button"
          bind:this={confirmBtn}
          class="btn min-h-11 min-w-[5.5rem] {confirmDialog.danger ? 'btn-danger' : 'btn-primary'}"
          onclick={() => confirmDialog.accept()}
        >
          {confirmDialog.confirmLabel}
        </button>
      </div>
    </div>
  </div>
{/if}

<style>
  .confirm-root {
    position: fixed;
    inset: 0;
    z-index: 70;
    display: grid;
    place-items: center;
    padding: max(1rem, env(safe-area-inset-top)) max(1rem, env(safe-area-inset-right))
      max(1rem, env(safe-area-inset-bottom)) max(1rem, env(safe-area-inset-left));
  }

  .confirm-backdrop {
    position: absolute;
    inset: 0;
    border: 0;
    background: var(--overlay);
    cursor: pointer;
  }

  .confirm-panel {
    position: relative;
    width: min(100%, 24rem);
    border-radius: 0.75rem;
    border: 1px solid var(--color-border);
    background: var(--color-bg-elevated);
    box-shadow: var(--shadow);
    padding: 1.25rem;
  }

  .confirm-title {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    margin: 0;
    font-size: 1.0625rem;
    font-weight: 600;
    color: var(--color-fg);
  }

  .confirm-title :global(.confirm-title-icon) {
    flex-shrink: 0;
    color: var(--color-danger);
  }

  .confirm-message {
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
</style>
