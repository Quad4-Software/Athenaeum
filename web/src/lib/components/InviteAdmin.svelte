<script lang="ts">
  import { onMount } from "svelte";
  import { Copy, Trash2, Mail } from "@lucide/svelte";
  import { api, ApiError } from "$lib/api/client";
  import EmptyState from "$lib/components/EmptyState.svelte";
  import { toast } from "$lib/stores/toast.svelte";
  import { i18n } from "$lib/stores/i18n.svelte";
  import type { Invite } from "$lib/api/types";

  interface Props {
    pocketIdEnabled?: boolean;
  }

  let { pocketIdEnabled }: Props = $props();

  let invites = $state<Invite[]>([]);
  let loading = $state(false);
  let creating = $state(false);
  let kind = $state<"permanent" | "guest">("permanent");
  let email = $state("");
  let username = $state("");
  let expiresInHours = $state(72);
  let guestExpiresInHours = $state(24);
  let provisionPocketId = $state(false);
  let pocketFromApi = $state(false);
  let lastSetupUrl = $state<string | null>(null);
  let lastInviteUrl = $state<string | null>(null);

  const pocketEnabled = $derived(pocketIdEnabled ?? pocketFromApi);

  async function load() {
    loading = true;
    try {
      invites = await api.listInvites();
    } catch (e) {
      toast.error(e instanceof ApiError ? e.message : i18n.t("admin.invites.loadFailed"));
    } finally {
      loading = false;
    }
  }

  async function resolvePocket() {
    if (pocketIdEnabled !== undefined) return;
    try {
      const cfg = await api.getPocketID();
      pocketFromApi = cfg.enabled;
    } catch {
      pocketFromApi = false;
    }
  }

  onMount(() => {
    void resolvePocket();
    void load();
  });

  async function create(event: Event) {
    event.preventDefault();
    creating = true;
    lastSetupUrl = null;
    lastInviteUrl = null;
    try {
      const result = await api.createInvite({
        kind,
        email: email.trim() || undefined,
        username: username.trim() || undefined,
        expiresInHours,
        guestExpiresInHours: kind === "guest" ? guestExpiresInHours : undefined,
        provisionPocketId: pocketEnabled && provisionPocketId,
      });
      lastInviteUrl = window.location.origin + result.url;
      lastSetupUrl = result.pocketIdSetupUrl || null;
      email = "";
      username = "";
      toast.success(i18n.t("admin.invites.created"));
      await load();
    } catch (e) {
      toast.error(e instanceof ApiError ? e.message : i18n.t("admin.invites.createFailed"));
    } finally {
      creating = false;
    }
  }

  async function revoke(id: number) {
    try {
      await api.revokeInvite(id);
      toast.success(i18n.t("admin.invites.revoked"));
      await load();
    } catch (e) {
      toast.error(e instanceof ApiError ? e.message : i18n.t("admin.invites.revokeFailed"));
    }
  }

  function inviteLink(inv: Invite) {
    return `${window.location.origin}/invite/${inv.token}`;
  }

  async function copyLink(url: string) {
    try {
      await navigator.clipboard.writeText(url);
      toast.success(i18n.t("admin.invites.linkCopied"));
    } catch {
      toast.error(i18n.t("admin.invites.copyFailed"));
    }
  }
</script>

<div class="space-y-4 border-t border-border pt-6">
  <p class="text-sm font-medium text-fg">{i18n.t("admin.invites.title")}</p>
  <p class="text-xs text-muted">{i18n.t("admin.invites.hint")}</p>

  <form class="space-y-3" onsubmit={create}>
    <label class="block text-xs text-muted">
      {i18n.t("admin.invites.kind")}
      <select bind:value={kind} class="field-input mt-1">
        <option value="permanent">{i18n.t("admin.invites.kindPermanent")}</option>
        <option value="guest">{i18n.t("admin.invites.kindGuest")}</option>
      </select>
    </label>
    <input
      type="email"
      placeholder={i18n.t("admin.invites.email")}
      bind:value={email}
      class="field-input"
    />
    <input
      type="text"
      placeholder={i18n.t("admin.invites.usernameOptional")}
      bind:value={username}
      class="field-input"
    />
    <label class="block text-xs text-muted">
      {i18n.t("admin.invites.expiresInHours")}
      <input
        type="number"
        min="1"
        max="8760"
        bind:value={expiresInHours}
        class="field-input mt-1"
      />
    </label>
    {#if kind === "guest"}
      <label class="block text-xs text-muted">
        {i18n.t("admin.invites.guestExpiresInHours")}
        <input
          type="number"
          min="1"
          max="8760"
          bind:value={guestExpiresInHours}
          class="field-input mt-1"
        />
      </label>
    {/if}
    {#if pocketEnabled}
      <label class="flex items-center gap-2 text-sm text-fg">
        <input type="checkbox" bind:checked={provisionPocketId} />
        {i18n.t("admin.invites.provisionPocketId")}
      </label>
    {/if}
    <div class="pt-1">
      <button type="submit" class="btn btn-primary" disabled={creating}>
        {creating ? i18n.t("admin.invites.creating") : i18n.t("admin.invites.create")}
      </button>
    </div>
  </form>

  {#if lastInviteUrl}
    <div class="rounded-lg border border-border bg-elevated p-3 text-sm">
      <p class="text-xs text-muted">{i18n.t("admin.invites.inviteUrl")}</p>
      <p class="mt-1 break-all font-mono text-xs text-fg">{lastInviteUrl}</p>
      <button
        type="button"
        class="btn btn-ghost mt-2 text-xs ring-1 ring-border"
        onclick={() => copyLink(lastInviteUrl!)}
      >
        <Copy size={12} />
        {i18n.t("admin.invites.copyLink")}
      </button>
      {#if lastSetupUrl}
        <p class="mt-3 text-xs text-muted">{i18n.t("admin.invites.pocketSetupUrl")}</p>
        <p class="mt-1 break-all font-mono text-xs text-fg">{lastSetupUrl}</p>
        <button
          type="button"
          class="btn btn-ghost mt-2 text-xs ring-1 ring-border"
          onclick={() => copyLink(lastSetupUrl!)}
        >
          <Copy size={12} />
          {i18n.t("admin.invites.copySetupUrl")}
        </button>
      {/if}
    </div>
  {/if}

  {#if loading}
    <p class="text-xs text-muted">{i18n.t("common.loading")}</p>
  {:else if invites.length === 0}
    <EmptyState
      size="sm"
      title={i18n.t("admin.invites.noneTitle")}
      body={i18n.t("admin.invites.noneBody")}
    >
      {#snippet icon(size)}
        <Mail {size} />
      {/snippet}
    </EmptyState>
  {:else}
    <ul class="divide-y divide-border rounded-lg border border-border">
      {#each invites as inv (inv.id)}
        <li class="flex flex-wrap items-center justify-between gap-2 px-3 py-2">
          <div>
            <p class="text-sm text-fg">
              {inv.kind}
              {#if inv.email}<span class="text-muted"> · {inv.email}</span>{/if}
            </p>
            <p class="text-xs text-subtle">
              {i18n.t("admin.invites.status")}: {inv.status}
              {#if inv.expiresAt}
                · {new Date(inv.expiresAt).toLocaleString()}
              {/if}
            </p>
          </div>
          <div class="flex gap-1">
            {#if inv.status === "pending"}
              <button
                type="button"
                class="btn btn-ghost text-xs ring-1 ring-border"
                onclick={() => copyLink(inviteLink(inv))}
              >
                <Copy size={12} />
                {i18n.t("admin.invites.copyLink")}
              </button>
              <button
                type="button"
                class="btn btn-ghost text-xs ring-1 ring-border"
                onclick={() => revoke(inv.id)}
              >
                <Trash2 size={12} />
                {i18n.t("admin.invites.revoke")}
              </button>
            {/if}
          </div>
        </li>
      {/each}
    </ul>
  {/if}
</div>
