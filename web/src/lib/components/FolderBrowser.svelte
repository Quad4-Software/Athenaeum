<script lang="ts">
  import { ChevronUp, Folder, X } from "@lucide/svelte";
  import { api, ApiError } from "$lib/api/client";
  import Button from "$lib/components/Button.svelte";
  import EmptyState from "$lib/components/EmptyState.svelte";
  import Skeleton from "$lib/components/Skeleton.svelte";
  import { i18n } from "$lib/stores/i18n.svelte";
  import type { FSBrowseResult } from "$lib/api/types";

  interface Props {
    open?: boolean;
    value?: string;
    onselect?: (path: string) => void;
    onclose?: () => void;
  }

  let { open = $bindable(false), value = "", onselect, onclose }: Props = $props();

  let currentPath = $state("");
  let parentPath = $state("");
  let entries = $state<FSBrowseResult["entries"]>([]);
  let loading = $state(false);
  let error = $state<string | null>(null);

  $effect(() => {
    if (!open) return;
    void loadBrowse(value || "");
  });

  async function loadBrowse(path: string) {
    loading = true;
    error = null;
    try {
      const res = await api.browseFS(path);
      currentPath = res.path ?? "";
      parentPath = res.parent ?? "";
      entries = res.entries;
    } catch (e) {
      error = e instanceof ApiError ? e.message : "Failed to browse folders";
      entries = [];
    } finally {
      loading = false;
    }
  }

  function close() {
    open = false;
    onclose?.();
  }

  function selectPath(path: string) {
    onselect?.(path);
    close();
  }
</script>

{#if open}
  <div class="browser-backdrop" role="presentation" onclick={close}></div>
  <div class="browser-panel" role="dialog" aria-modal="true" aria-label="Choose folder">
    <div class="browser-header">
      <div class="min-w-0">
        <p class="browser-title">Choose folder</p>
        <p class="browser-path">{currentPath || "Starting locations"}</p>
      </div>
      <button type="button" class="browser-close" aria-label="Close" onclick={close}>
        <X size={18} />
      </button>
    </div>

    <div class="browser-body">
      {#if loading}
        <div class="browser-loading">
          {#each Array(6) as _, i (i)}
            <Skeleton height="2.25rem" rounded="lg" />
          {/each}
        </div>
      {:else if error}
        <p class="browser-error">{error}</p>
      {:else}
        {#if parentPath}
          <button type="button" class="browser-row" onclick={() => loadBrowse(parentPath)}>
            <ChevronUp size={16} />
            <span>Parent folder</span>
          </button>
        {/if}
        {#each entries as entry (entry.path)}
          <button type="button" class="browser-row" onclick={() => loadBrowse(entry.path)}>
            <Folder size={16} />
            <span class="truncate">{entry.name}</span>
            <span class="browser-row-path">{entry.path}</span>
          </button>
        {/each}
        {#if entries.length === 0 && !parentPath}
          <EmptyState
            size="sm"
            title={i18n.t("folderBrowser.emptyTitle")}
            body={i18n.t("folderBrowser.emptyBody")}
          >
            {#snippet icon(size)}
              <Folder {size} />
            {/snippet}
          </EmptyState>
        {/if}
      {/if}
    </div>

    <div class="browser-footer">
      <Button variant="ghost" class="ring-1 ring-border" onclick={close}>Cancel</Button>
      {#if currentPath}
        <Button onclick={() => selectPath(currentPath)}>Select this folder</Button>
      {/if}
    </div>
  </div>
{/if}

<style>
  .browser-backdrop {
    position: fixed;
    inset: 0;
    z-index: 60;
    background: var(--color-overlay);
  }

  .browser-panel {
    position: fixed;
    z-index: 70;
    left: 50%;
    top: 50%;
    transform: translate(-50%, -50%);
    display: flex;
    width: min(32rem, calc(100vw - 2rem));
    max-height: min(28rem, calc(100vh - 2rem));
    flex-direction: column;
    border-radius: var(--radius-card);
    background: var(--color-surface);
    box-shadow: var(--shadow);
    border: 1px solid var(--color-border);
  }

  .browser-header {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: 0.75rem;
    padding: 1rem 1rem 0.75rem;
    border-bottom: 1px solid var(--color-border);
  }

  .browser-title {
    margin: 0;
    font-size: 0.9375rem;
    font-weight: 600;
    color: var(--color-fg);
  }

  .browser-path {
    margin: 0.25rem 0 0;
    font-size: 0.75rem;
    color: var(--color-muted);
    word-break: break-all;
  }

  .browser-close {
    border: 0;
    background: none;
    color: var(--color-muted);
    cursor: pointer;
    padding: 0.25rem;
  }

  .browser-body {
    flex: 1;
    min-height: 0;
    overflow-y: auto;
    padding: 0.5rem;
  }

  .browser-loading {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
  }

  .browser-row {
    display: grid;
    grid-template-columns: auto 1fr;
    gap: 0.5rem 0.75rem;
    width: 100%;
    padding: 0.625rem 0.75rem;
    border: 0;
    border-radius: 0.5rem;
    background: transparent;
    text-align: left;
    color: var(--color-fg);
    cursor: pointer;
  }

  .browser-row:hover {
    background: var(--color-surface-hover);
  }

  .browser-row-path {
    grid-column: 2;
    font-size: 0.6875rem;
    color: var(--color-subtle);
    word-break: break-all;
  }

  .browser-error {
    margin: 0.5rem;
    font-size: 0.8125rem;
    color: var(--color-muted);
  }

  .browser-footer {
    display: flex;
    justify-content: flex-end;
    gap: 0.5rem;
    padding: 0.75rem 1rem 1rem;
    border-top: 1px solid var(--color-border);
  }
</style>
