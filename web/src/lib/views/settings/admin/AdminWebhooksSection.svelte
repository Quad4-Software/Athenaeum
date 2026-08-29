<script lang="ts">
  import Button from "$lib/components/Button.svelte";
  import Skeleton from "$lib/components/Skeleton.svelte";
  import EmptyState from "$lib/components/EmptyState.svelte";
  import { Webhook as WebhookIcon } from "@lucide/svelte";
  import { auth } from "$lib/stores/auth.svelte";
  import { toast } from "$lib/stores/toast.svelte";
  import { confirmDialog } from "$lib/stores/confirm.svelte";
  import { i18n } from "$lib/stores/i18n.svelte";
  import { api } from "$lib/api/client";
  import type { Webhook, WebhookDelivery, WebhookEvent } from "$lib/api/types";
  import { apiAction } from "$lib/utils/api-action";
  import { untrack } from "svelte";
  import { SvelteSet } from "svelte/reactivity";

  const WEBHOOK_EVENTS: WebhookEvent[] = [
    "user.create",
    "user.delete",
    "invite.created",
    "invite.accepted",
    "book.upload",
    "library.scan.complete",
  ];

  let webhooks = $state<Webhook[]>([]);
  let loading = $state(false);
  let creating = $state(false);
  let url = $state("");
  let secret = $state("");
  let enabled = $state(true);
  let selectedEvents = new SvelteSet<string>();
  let expandedId = $state<number | null>(null);
  let deliveries = $state<WebhookDelivery[]>([]);
  let deliveriesLoading = $state(false);

  $effect(() => {
    if (!auth.user?.isAdmin) return;
    untrack(() => {
      void load();
    });
  });

  async function load() {
    loading = true;
    const list = await apiAction(() => api.listWebhooks(), {
      errorFallback: i18n.t("settings.webhooksLoadFailed"),
    });
    if (list) webhooks = list;
    loading = false;
  }

  function toggleEvent(ev: string) {
    if (selectedEvents.has(ev)) selectedEvents.delete(ev);
    else selectedEvents.add(ev);
  }

  async function create(event: Event) {
    event.preventDefault();
    if (!url.trim() || selectedEvents.size === 0) {
      toast.error(i18n.t("settings.webhooksCreateNeed"));
      return;
    }
    creating = true;
    const created = await apiAction(
      () =>
        api.createWebhook({
          url: url.trim(),
          secret: secret.trim() || undefined,
          events: [...selectedEvents],
          enabled,
        }),
      {
        success: i18n.t("settings.webhooksCreated"),
        errorFallback: i18n.t("settings.webhooksCreateFailed"),
      },
    );
    if (created) {
      url = "";
      secret = "";
      selectedEvents.clear();
      enabled = true;
      await load();
    }
    creating = false;
  }

  async function toggleEnabled(wh: Webhook) {
    const updated = await apiAction(() => api.updateWebhook(wh.id, { enabled: !wh.enabled }), {
      success: wh.enabled
        ? i18n.t("settings.webhooksDisabled")
        : i18n.t("settings.webhooksEnabledToast"),
      errorFallback: i18n.t("settings.webhooksUpdateFailed"),
    });
    if (updated) await load();
  }

  async function test(id: number) {
    await apiAction(() => api.testWebhook(id), {
      success: i18n.t("settings.webhooksTestOk"),
      errorFallback: i18n.t("settings.webhooksTestFailed"),
    });
  }

  async function remove(id: number) {
    const ok = await confirmDialog.ask({
      title: i18n.t("settings.webhooksDeleteTitle"),
      message: i18n.t("settings.webhooksDelete"),
      confirmLabel: i18n.t("confirm.delete"),
      cancelLabel: i18n.t("confirm.cancel"),
      danger: true,
    });
    if (!ok) return;
    const deleted = await apiAction(() => api.deleteWebhook(id), {
      success: i18n.t("settings.webhooksDeleted"),
      errorFallback: i18n.t("settings.webhooksDeleteFailed"),
    });
    if (deleted !== undefined) {
      if (expandedId === id) {
        expandedId = null;
        deliveries = [];
      }
      await load();
    }
  }

  async function toggleDeliveries(id: number) {
    if (expandedId === id) {
      expandedId = null;
      deliveries = [];
      return;
    }
    expandedId = id;
    deliveriesLoading = true;
    const list = await apiAction(() => api.listWebhookDeliveries(id), {
      errorFallback: i18n.t("settings.webhooksDeliveriesFailed"),
    });
    if (list) deliveries = list;
    else expandedId = null;
    deliveriesLoading = false;
  }
</script>

<div
  id="admin-webhooks"
  class="scroll-mt-6 rounded-[var(--radius-card)] border border-border bg-surface p-5"
>
  <h2 class="text-sm font-semibold text-fg">{i18n.t("settings.adminSections.webhooks")}</h2>
  <p class="mt-1 text-xs text-muted">{i18n.t("settings.webhooksHint")}</p>

  <form class="mt-4 space-y-3 border-b border-border pb-4" onsubmit={create}>
    <input
      type="url"
      class="field-input"
      placeholder={i18n.t("settings.webhooksUrl")}
      bind:value={url}
      required
    />
    <input
      type="text"
      class="field-input"
      placeholder={i18n.t("settings.webhooksSecret")}
      bind:value={secret}
      autocomplete="off"
    />
    <p class="text-xs font-medium text-muted">{i18n.t("settings.webhooksEvents")}</p>
    <ul class="space-y-1">
      {#each WEBHOOK_EVENTS as ev (ev)}
        <li>
          <label class="flex items-center gap-2 text-sm text-fg">
            <input
              type="checkbox"
              checked={selectedEvents.has(ev)}
              onchange={() => toggleEvent(ev)}
            />
            {ev}
          </label>
        </li>
      {/each}
    </ul>
    <label class="flex items-center gap-2 text-sm text-fg">
      <input type="checkbox" bind:checked={enabled} />
      {i18n.t("settings.webhooksEnabled")}
    </label>
    <div class="pt-1">
      <Button type="submit" size="sm" loading={creating}>{i18n.t("settings.webhooksCreate")}</Button
      >
    </div>
  </form>

  {#if loading && webhooks.length === 0}
    <div class="mt-4">
      <Skeleton height="4rem" rounded="lg" />
    </div>
  {:else if webhooks.length === 0}
    <div class="mt-4">
      <EmptyState
        size="sm"
        title={i18n.t("settings.webhooksEmptyTitle")}
        body={i18n.t("settings.webhooksEmptyBody")}
      >
        {#snippet icon(size)}
          <WebhookIcon {size} />
        {/snippet}
      </EmptyState>
    </div>
  {:else}
    <ul class="mt-4 divide-y divide-border rounded-lg border border-border">
      {#each webhooks as wh (wh.id)}
        <li class="px-3 py-2">
          <div class="flex flex-wrap items-center justify-between gap-2">
            <div class="min-w-0">
              <p class="truncate text-sm font-medium text-fg">{wh.url}</p>
              <p class="text-xs text-subtle">
                {wh.events.join(", ")}
                · {wh.enabled ? i18n.t("settings.webhooksOn") : i18n.t("settings.webhooksOff")}
                {#if wh.secretSet}
                  · {i18n.t("settings.webhooksSecretSet")}
                {/if}
              </p>
            </div>
            <div class="flex shrink-0 flex-wrap gap-1">
              <button
                type="button"
                class="btn btn-ghost text-xs ring-1 ring-border"
                onclick={() => toggleEnabled(wh)}
              >
                {wh.enabled
                  ? i18n.t("settings.webhooksDisable")
                  : i18n.t("settings.webhooksEnable")}
              </button>
              <button
                type="button"
                class="btn btn-ghost text-xs ring-1 ring-border"
                onclick={() => test(wh.id)}
              >
                {i18n.t("settings.webhooksTest")}
              </button>
              <button
                type="button"
                class="btn btn-ghost text-xs ring-1 ring-border"
                onclick={() => toggleDeliveries(wh.id)}
              >
                {expandedId === wh.id
                  ? i18n.t("settings.webhooksHideDeliveries")
                  : i18n.t("settings.webhooksShowDeliveries")}
              </button>
              <button
                type="button"
                class="btn btn-ghost text-xs text-danger"
                onclick={() => remove(wh.id)}
              >
                {i18n.t("settings.webhooksDeleteBtn")}
              </button>
            </div>
          </div>
          {#if expandedId === wh.id}
            <div class="mt-2 rounded-lg border border-border bg-elevated p-2">
              {#if deliveriesLoading}
                <p class="text-xs text-muted">{i18n.t("common.loading")}</p>
              {:else if deliveries.length === 0}
                <p class="text-xs text-muted">{i18n.t("settings.webhooksNoDeliveries")}</p>
              {:else}
                <ul class="space-y-1 text-xs text-muted">
                  {#each deliveries as d (d.id)}
                    <li
                      class="flex flex-wrap justify-between gap-2 border-b border-border/50 py-1 last:border-0"
                    >
                      <span>
                        {d.event}
                        · {d.success
                          ? i18n.t("settings.webhooksSuccess")
                          : i18n.t("settings.webhooksFailed")}
                        · HTTP {d.statusCode}
                        · {d.attempts}
                        {i18n.t("settings.webhooksAttempts")}
                      </span>
                      <span>{new Date(d.createdAt).toLocaleString()}</span>
                      {#if d.lastError}
                        <span class="w-full text-danger">{d.lastError}</span>
                      {/if}
                    </li>
                  {/each}
                </ul>
              {/if}
            </div>
          {/if}
        </li>
      {/each}
    </ul>
  {/if}
</div>
