<script lang="ts">
  import { onMount } from "svelte";
  import { Copy, Trash2, Clock, UserRound } from "@lucide/svelte";
  import { api, ApiError } from "$lib/api/client";
  import EmptyState from "$lib/components/EmptyState.svelte";
  import { toast } from "$lib/stores/toast.svelte";
  import { i18n } from "$lib/stores/i18n.svelte";
  import type { User } from "$lib/api/types";
  import { SvelteSet } from "svelte/reactivity";

  let guests = $state<User[]>([]);
  let expiring = $state<User[]>([]);
  let loading = $state(false);
  let selected = new SvelteSet<number>();
  let extendHours = $state(24);

  async function load() {
    loading = true;
    try {
      const [all, soon] = await Promise.all([api.listGuests(), api.listGuests(72)]);
      guests = all;
      expiring = soon;
    } catch (e) {
      toast.error(e instanceof ApiError ? e.message : i18n.t("admin.guests.loadFailed"));
    } finally {
      loading = false;
    }
  }

  onMount(() => {
    void load();
  });

  function toggle(id: number) {
    if (selected.has(id)) selected.delete(id);
    else selected.add(id);
  }

  function toggleAll() {
    if (selected.size === guests.length) {
      selected.clear();
    } else {
      selected.clear();
      for (const g of guests) selected.add(g.id);
    }
  }

  async function bulkRevoke() {
    if (selected.size === 0) return;
    try {
      const { deleted } = await api.bulkDeleteGuests([...selected]);
      toast.success(i18n.t("admin.guests.revoked", { count: String(deleted) }));
      selected.clear();
      await load();
    } catch (e) {
      toast.error(e instanceof ApiError ? e.message : i18n.t("admin.guests.revokeFailed"));
    }
  }

  async function extendGuest(id: number) {
    try {
      await api.extendGuest(id, extendHours);
      toast.success(i18n.t("admin.guests.extended"));
      await load();
    } catch (e) {
      toast.error(e instanceof ApiError ? e.message : i18n.t("admin.guests.extendFailed"));
    }
  }

  function inviteLink(username: string) {
    const url = new URL("/login", window.location.origin);
    url.searchParams.set("username", username);
    return url.toString();
  }

  async function copyInvite(username: string) {
    try {
      await navigator.clipboard.writeText(inviteLink(username));
      toast.success(i18n.t("admin.guests.inviteCopied"));
    } catch {
      toast.error(i18n.t("admin.guests.copyFailed"));
    }
  }

  function formatExpiry(iso?: string) {
    if (!iso) return i18n.t("admin.guests.noExpiry");
    return new Date(iso).toLocaleString();
  }
</script>

<div class="space-y-4 border-t border-border pt-6">
  <p class="text-sm font-medium text-fg">{i18n.t("admin.guests.title")}</p>

  {#if expiring.length > 0}
    <div class="rounded-lg border border-amber-500/40 bg-amber-500/5 p-3">
      <p class="text-xs font-medium text-fg">{i18n.t("admin.guests.expiringSoon")}</p>
      <ul class="mt-2 space-y-1 text-xs text-muted">
        {#each expiring as g (g.id)}
          <li class="flex flex-wrap items-center justify-between gap-2">
            <span>{g.username}</span>
            <span>{formatExpiry(g.expiresAt)}</span>
            <button
              type="button"
              class="btn btn-ghost text-xs ring-1 ring-border"
              onclick={() => extendGuest(g.id)}
            >
              <Clock size={12} />
              {i18n.t("admin.guests.extend")}
            </button>
          </li>
        {/each}
      </ul>
    </div>
  {/if}

  {#if loading}
    <p class="text-xs text-muted">{i18n.t("common.loading")}</p>
  {:else if guests.length === 0}
    <EmptyState
      size="sm"
      title={i18n.t("admin.guests.noneTitle")}
      body={i18n.t("admin.guests.noneBody")}
    >
      {#snippet icon(size)}
        <UserRound {size} />
      {/snippet}
    </EmptyState>
  {:else}
    <div class="flex flex-wrap items-end gap-3">
      <label class="text-xs text-muted">
        {i18n.t("admin.guests.extendBy")}
        <input
          type="number"
          min="1"
          max="8760"
          bind:value={extendHours}
          class="field-input mt-1 w-24"
        />
      </label>
      <button
        type="button"
        class="btn btn-ghost text-xs ring-1 ring-border"
        disabled={selected.size === 0}
        onclick={bulkRevoke}
      >
        <Trash2 size={14} />
        {i18n.t("admin.guests.bulkRevoke")} ({selected.size})
      </button>
    </div>
    <ul class="divide-y divide-border rounded-lg border border-border">
      <li class="flex items-center gap-2 px-3 py-2 text-xs text-muted">
        <input type="checkbox" checked={selected.size === guests.length} onchange={toggleAll} />
        {i18n.t("admin.guests.selectAll")}
      </li>
      {#each guests as g (g.id)}
        <li class="flex flex-wrap items-center justify-between gap-2 px-3 py-2">
          <label class="flex items-center gap-2 text-sm text-fg">
            <input type="checkbox" checked={selected.has(g.id)} onchange={() => toggle(g.id)} />
            {g.username}
          </label>
          <span class="text-xs text-muted">{formatExpiry(g.expiresAt)}</span>
          <div class="flex gap-1">
            <button
              type="button"
              class="btn btn-ghost text-xs ring-1 ring-border"
              onclick={() => copyInvite(g.username)}
            >
              <Copy size={12} />
              {i18n.t("admin.guests.copyInvite")}
            </button>
            <button
              type="button"
              class="btn btn-ghost text-xs ring-1 ring-border"
              onclick={() => extendGuest(g.id)}
            >
              <Clock size={12} />
              {i18n.t("admin.guests.extend")}
            </button>
          </div>
        </li>
      {/each}
    </ul>
  {/if}
</div>
