<script lang="ts">
  import { Copy, Key, Trash2 } from "@lucide/svelte";
  import Button from "$lib/components/Button.svelte";
  import EmptyState from "$lib/components/EmptyState.svelte";
  import Skeleton from "$lib/components/Skeleton.svelte";
  import { brand } from "$lib/brand";
  import { auth } from "$lib/stores/auth.svelte";
  import { toast } from "$lib/stores/toast.svelte";
  import { confirmDialog } from "$lib/stores/confirm.svelte";
  import { i18n } from "$lib/stores/i18n.svelte";
  import { ApiError, api } from "$lib/api/client";
  import type { APIKey, APIKeyCreated, APIDoc } from "$lib/api/types";
  import { untrack } from "svelte";

  let apiDocs = $state<APIDoc | null>(null);
  let apiDocsLoading = $state(false);
  let apiKeys = $state<APIKey[]>([]);
  let apiKeysLoading = $state(false);
  let apiKeyName = $state("");
  let apiKeyCreating = $state(false);
  let newAPIKey = $state<APIKeyCreated | null>(null);

  $effect(() => {
    untrack(() => {
      void loadAPIDocs();
      if (auth.user) void loadAPIKeys();
    });
  });

  async function loadAPIDocs() {
    apiDocsLoading = true;
    try {
      apiDocs = await api.getAPIDocs();
    } catch (e) {
      toast.error(e instanceof ApiError ? e.message : "Failed to load API docs");
    } finally {
      apiDocsLoading = false;
    }
  }

  async function loadAPIKeys() {
    apiKeysLoading = true;
    try {
      apiKeys = await api.listAPIKeys();
    } catch (e) {
      toast.error(e instanceof ApiError ? e.message : "Failed to load API keys");
    } finally {
      apiKeysLoading = false;
    }
  }

  async function createAPIKey(event: Event) {
    event.preventDefault();
    const name = apiKeyName.trim();
    if (!name) return;
    apiKeyCreating = true;
    try {
      newAPIKey = await api.createAPIKey(name);
      apiKeyName = "";
      toast.success("API key created");
      void loadAPIKeys();
    } catch (e) {
      toast.error(e instanceof ApiError ? e.message : "Failed to create API key");
    } finally {
      apiKeyCreating = false;
    }
  }

  async function revokeAPIKey(key: APIKey) {
    const ok = await confirmDialog.ask({
      title: i18n.t("settings.revokeApiKeyTitle"),
      message: i18n.t("settings.revokeApiKey", { name: key.name, prefix: key.prefix }),
      confirmLabel: i18n.t("settings.revoke"),
      cancelLabel: i18n.t("confirm.cancel"),
      danger: true,
    });
    if (!ok) return;
    try {
      await api.deleteAPIKey(key.id);
      if (newAPIKey?.id === key.id) newAPIKey = null;
      toast.success("API key revoked");
      void loadAPIKeys();
    } catch (e) {
      toast.error(e instanceof ApiError ? e.message : "Failed to revoke API key");
    }
  }

  async function copyAPIKey(value: string) {
    try {
      await navigator.clipboard.writeText(value);
      toast.success("Copied to clipboard");
    } catch {
      toast.error("Could not copy to clipboard");
    }
  }
</script>

<div class="space-y-6">
  {#if auth.user}
    <div class="rounded-[var(--radius-card)] border border-border bg-surface p-5">
      <h2 class="text-sm font-semibold text-fg">API keys</h2>
      <p class="mt-1 text-sm text-muted">
        Create keys for scripts and integrations. Pass them as
        <code class="text-fg">Authorization: Bearer {brand.apiKeyPrefix}…</code> or
        <code class="text-fg">X-API-Key: {brand.apiKeyPrefix}…</code>. The full secret is shown only
        once.
      </p>

      {#if newAPIKey}
        <div class="mt-4 rounded-lg border border-primary/40 bg-primary/5 p-3">
          <p class="text-sm font-medium text-fg">New key: {newAPIKey.name}</p>
          <p class="mt-1 text-xs text-muted">
            Copy this key now. You will not be able to see it again.
          </p>
          <div class="mt-2 flex flex-wrap items-center gap-2">
            <code
              class="min-w-0 flex-1 break-all rounded bg-bg px-2 py-1 text-xs text-fg ring-1 ring-border"
            >
              {newAPIKey.key}
            </code>
            <button
              type="button"
              class="btn btn-ghost text-xs ring-1 ring-border"
              onclick={() => copyAPIKey(newAPIKey!.key)}
            >
              <Copy size={14} /> Copy
            </button>
          </div>
          <button
            type="button"
            class="btn btn-ghost mt-2 text-xs ring-1 ring-border"
            onclick={() => (newAPIKey = null)}
          >
            Dismiss
          </button>
        </div>
      {/if}

      <form class="mt-4 flex flex-wrap items-center gap-2" onsubmit={createAPIKey}>
        <input
          type="text"
          bind:value={apiKeyName}
          required
          maxlength="128"
          class="field-input min-w-[12rem] flex-1"
          placeholder="Key name (e.g. sync script)"
        />
        <Button type="submit" size="sm" loading={apiKeyCreating} disabled={!apiKeyName.trim()}>
          <Key size={14} /> Create key
        </Button>
      </form>

      {#if apiKeysLoading}
        <Skeleton height="4rem" rounded="lg" class="mt-4" />
      {:else if apiKeys.length === 0}
        <EmptyState
          size="sm"
          class="mt-2"
          title={i18n.t("settings.apiKeysEmptyTitle")}
          body={i18n.t("settings.apiKeysEmptyBody")}
        >
          {#snippet icon(size)}
            <Key {size} />
          {/snippet}
        </EmptyState>
      {:else}
        <ul class="mt-4 divide-y divide-border rounded-lg border border-border">
          {#each apiKeys as key (key.id)}
            <li class="flex flex-wrap items-start justify-between gap-2 px-3 py-2">
              <div class="min-w-0">
                <p class="text-sm font-medium text-fg">{key.name}</p>
                <p class="text-xs text-muted">
                  <code>{key.prefix}…</code>
                  · created {new Date(key.createdAt).toLocaleDateString()}
                  {#if key.lastUsedAt}
                    · last used {new Date(key.lastUsedAt).toLocaleString()}
                  {/if}
                </p>
              </div>
              <button
                type="button"
                class="btn btn-ghost text-xs text-danger"
                onclick={() => revokeAPIKey(key)}
              >
                <Trash2 size={14} /> Revoke
              </button>
            </li>
          {/each}
        </ul>
      {/if}
    </div>
  {:else if auth.authEnabled}
    <div class="rounded-[var(--radius-card)] border border-border bg-surface p-5">
      <p class="text-sm text-muted">Sign in to create and manage API keys.</p>
    </div>
  {/if}

  <div class="rounded-[var(--radius-card)] border border-border bg-surface p-5">
    <h2 class="text-sm font-semibold text-fg">API reference</h2>
    <p class="mt-1 text-sm text-muted">
      Interactive OpenAPI docs:
      <a href="/docs" class="text-primary hover:underline" target="_blank" rel="noopener noreferrer"
        >/docs</a
      >
      · machine-readable
      <a
        href="/api/openapi.json"
        class="text-primary hover:underline"
        target="_blank"
        rel="noopener noreferrer">/api/openapi.json</a
      >
      and
      <a
        href="/api/docs"
        class="text-primary hover:underline"
        target="_blank"
        rel="noopener noreferrer">/api/docs</a
      >.
    </p>

    {#if apiDocsLoading}
      <Skeleton height="12rem" rounded="lg" class="mt-4" />
    {:else if apiDocs}
      <div class="mt-4 space-y-4">
        <div>
          <p class="text-xs font-medium uppercase tracking-wide text-subtle">Authentication</p>
          <ul class="mt-2 space-y-1 text-sm text-muted">
            {#each apiDocs.auth as line (line)}
              <li>{line}</li>
            {/each}
          </ul>
        </div>

        {#each apiDocs.sections as section (section.title)}
          <div class="border-t border-border pt-4">
            <p class="text-sm font-medium text-fg">{section.title}</p>
            <div class="mt-2 overflow-x-auto">
              <table class="w-full min-w-[32rem] text-left text-xs">
                <thead>
                  <tr class="text-subtle">
                    <th class="pb-2 pr-3 font-medium">Method</th>
                    <th class="pb-2 pr-3 font-medium">Path</th>
                    <th class="pb-2 font-medium">Summary</th>
                  </tr>
                </thead>
                <tbody class="divide-y divide-border">
                  {#each section.endpoints as ep (`${ep.method}-${ep.path}`)}
                    <tr>
                      <td class="py-2 pr-3 align-top font-mono text-primary">{ep.method}</td>
                      <td class="py-2 pr-3 align-top font-mono text-fg">{ep.path}</td>
                      <td class="py-2 align-top text-muted">
                        {ep.summary}
                        {#if ep.auth}
                          <span class="block text-subtle">Auth: {ep.auth}</span>
                        {/if}
                        {#if ep.query}
                          <span class="block text-subtle">Query: {ep.query}</span>
                        {/if}
                        {#if ep.body}
                          <span class="block text-subtle">Body: {ep.body}</span>
                        {/if}
                      </td>
                    </tr>
                  {/each}
                </tbody>
              </table>
            </div>
          </div>
        {/each}
      </div>
    {/if}
  </div>
</div>
