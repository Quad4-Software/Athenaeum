<script lang="ts">
  import { fly } from "svelte/transition";
  import { CheckCircle2, CircleAlert, Info, Loader2, X } from "@lucide/svelte";
  import { toast } from "$lib/stores/toast.svelte";
  import type { ToastKind } from "$lib/stores/toast.svelte";
  import { i18n } from "$lib/stores/i18n.svelte";
  import type { Component } from "svelte";

  const kindClass: Record<ToastKind, string> = {
    success: "border-success/40 bg-success/10 text-fg",
    error: "border-danger/40 bg-danger/10 text-fg",
    info: "border-border bg-bg-elevated text-fg",
    loading: "border-border bg-bg-elevated text-fg",
  };

  const kindIcon: Record<ToastKind, Component<{ size?: number; class?: string }>> = {
    success: CheckCircle2,
    error: CircleAlert,
    info: Info,
    loading: Loader2,
  };

  const kindIconClass: Record<ToastKind, string> = {
    success: "toast-icon toast-icon--success",
    error: "toast-icon toast-icon--error",
    info: "toast-icon toast-icon--info",
    loading: "toast-icon toast-icon--loading",
  };
</script>

<div
  class="pointer-events-none fixed right-4 z-50 flex w-full max-w-sm flex-col gap-2 px-4 bottom-[calc(var(--bottom-chrome)+0.75rem)] sm:px-0"
  aria-live="polite"
>
  {#each toast.items as item (item.id)}
    {@const KindIcon = kindIcon[item.kind]}
    <div
      class="pointer-events-auto flex flex-col gap-2 rounded-lg border px-3 py-2.5 text-sm shadow-[var(--shadow)] {kindClass[
        item.kind
      ]}"
      role={item.kind === "error" ? "alert" : "status"}
      transition:fly={{ y: 16, duration: 220 }}
    >
      <div class="flex items-start gap-2.5">
        <KindIcon
          size={16}
          class="{kindIconClass[item.kind]}{item.kind === 'loading' ? ' animate-spin' : ''}"
        />
        <p class="min-w-0 flex-1 leading-snug pt-px">{item.message}</p>
        <button
          type="button"
          class="btn btn-ghost shrink-0 self-start p-1"
          aria-label={i18n.t("common.dismiss")}
          onclick={() => toast.dismiss(item.id)}
        >
          <X size={14} />
        </button>
      </div>
      {#if item.kind === "loading" && item.progress !== undefined}
        <div
          class="toast-progress h-1.5 w-full overflow-hidden rounded-full bg-border/60"
          role="progressbar"
          aria-valuemin={0}
          aria-valuemax={100}
          aria-valuenow={Math.round(item.progress * 100)}
        >
          <div
            class="h-full rounded-full bg-accent transition-[width] duration-200"
            style:width="{Math.round(item.progress * 100)}%"
          ></div>
        </div>
      {:else if item.kind === "loading"}
        <div class="toast-progress h-1 w-full overflow-hidden rounded-full bg-border/60">
          <div class="toast-indeterminate h-full rounded-full bg-accent/80"></div>
        </div>
      {/if}
    </div>
  {/each}
</div>

<style>
  :global(.toast-icon) {
    flex-shrink: 0;
    margin-top: 1px;
  }

  :global(.toast-icon--success) {
    color: var(--color-success);
  }

  :global(.toast-icon--error) {
    color: var(--color-danger);
  }

  :global(.toast-icon--info) {
    color: var(--color-muted);
  }

  :global(.toast-icon--loading) {
    color: var(--color-accent, var(--color-muted));
  }

  .toast-indeterminate {
    width: 40%;
    animation: toast-indeterminate 1.2s ease-in-out infinite;
  }

  @keyframes toast-indeterminate {
    0% {
      transform: translateX(-120%);
    }
    100% {
      transform: translateX(320%);
    }
  }
</style>
