<script lang="ts">
  import Button from "$lib/components/Button.svelte";
  import Skeleton from "$lib/components/Skeleton.svelte";
  import { auth } from "$lib/stores/auth.svelte";
  import { i18n } from "$lib/stores/i18n.svelte";
  import { api } from "$lib/api/client";
  import type { SMTPSettingsPublic } from "$lib/api/types";
  import { apiAction } from "$lib/utils/api-action";
  import { untrack } from "svelte";

  let cfg = $state<SMTPSettingsPublic | null>(null);
  let loading = $state(false);
  let saving = $state(false);
  let password = $state("");
  let enabled = $state(false);
  let host = $state("");
  let port = $state(587);
  let username = $state("");
  let fromAddr = $state("");
  let useTls = $state(true);

  $effect(() => {
    if (!auth.user?.isAdmin) return;
    untrack(() => {
      void load();
    });
  });

  async function load() {
    loading = true;
    const next = await apiAction(() => api.getSMTP(), {
      errorFallback: i18n.t("settings.smtpLoadFailed"),
    });
    if (next) {
      cfg = next;
      enabled = next.enabled;
      host = next.host;
      port = next.port || 587;
      username = next.username;
      fromAddr = next.fromAddr;
      useTls = next.useTls;
      password = "";
    }
    loading = false;
  }

  async function save(event: Event) {
    event.preventDefault();
    saving = true;
    const next = await apiAction(
      () =>
        api.saveSMTP({
          enabled,
          host,
          port,
          username,
          password: password || undefined,
          fromAddr,
          useTls,
        }),
      {
        success: i18n.t("settings.smtpSaved"),
        errorFallback: i18n.t("settings.smtpSaveFailed"),
      },
    );
    if (next) {
      cfg = next;
      password = "";
    }
    saving = false;
  }
</script>

<div
  id="admin-smtp"
  class="scroll-mt-6 rounded-[var(--radius-card)] border border-border bg-surface p-5"
>
  <h2 class="text-sm font-semibold text-fg">{i18n.t("settings.adminSections.smtp")}</h2>
  <p class="mt-1 text-xs text-muted">{i18n.t("settings.smtpHint")}</p>
  {#if loading && !cfg}
    <div class="mt-3">
      <Skeleton height="8rem" rounded="lg" />
    </div>
  {:else}
    <form class="mt-4 space-y-3" onsubmit={save}>
      <label class="flex items-center gap-2 text-sm text-fg">
        <input type="checkbox" bind:checked={enabled} />
        {i18n.t("settings.smtpEnabled")}
      </label>
      <input
        type="text"
        class="field-input"
        placeholder={i18n.t("settings.smtpHost")}
        bind:value={host}
      />
      <input
        type="number"
        class="field-input"
        placeholder={i18n.t("settings.smtpPort")}
        bind:value={port}
        min="1"
        max="65535"
      />
      <input
        type="text"
        class="field-input"
        placeholder={i18n.t("settings.smtpUsername")}
        bind:value={username}
        autocomplete="off"
      />
      <input
        type="password"
        class="field-input"
        placeholder={cfg?.passwordSet
          ? i18n.t("settings.smtpPasswordKeep")
          : i18n.t("settings.smtpPassword")}
        bind:value={password}
        autocomplete="new-password"
      />
      <input
        type="email"
        class="field-input"
        placeholder={i18n.t("settings.smtpFrom")}
        bind:value={fromAddr}
      />
      <label class="flex items-center gap-2 text-sm text-fg">
        <input type="checkbox" bind:checked={useTls} />
        {i18n.t("settings.smtpTls")}
      </label>
      <div class="pt-1">
        <Button type="submit" size="sm" loading={saving}>{i18n.t("settings.smtpSave")}</Button>
      </div>
    </form>
  {/if}
</div>
