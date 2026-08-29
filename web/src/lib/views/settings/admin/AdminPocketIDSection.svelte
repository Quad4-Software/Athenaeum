<script lang="ts">
  import Button from "$lib/components/Button.svelte";
  import Skeleton from "$lib/components/Skeleton.svelte";
  import { auth } from "$lib/stores/auth.svelte";
  import { i18n } from "$lib/stores/i18n.svelte";
  import { api } from "$lib/api/client";
  import type { PocketIDSettingsPublic } from "$lib/api/types";
  import { apiAction } from "$lib/utils/api-action";
  import { untrack } from "svelte";

  let cfg = $state<PocketIDSettingsPublic | null>(null);
  let loading = $state(false);
  let saving = $state(false);
  let testing = $state(false);
  let applying = $state(false);
  let enabled = $state(false);
  let baseUrl = $state("");
  let apiKey = $state("");
  let defaultGroupIds = $state("");

  $effect(() => {
    if (!auth.user?.isAdmin) return;
    untrack(() => {
      void load();
    });
  });

  async function load() {
    loading = true;
    const next = await apiAction(() => api.getPocketID(), {
      errorFallback: i18n.t("settings.pocketidLoadFailed"),
    });
    if (next) {
      cfg = next;
      enabled = next.enabled;
      baseUrl = next.baseUrl;
      defaultGroupIds = (next.defaultGroupIds ?? []).join(", ");
      apiKey = "";
    }
    loading = false;
  }

  function parseGroupIds(raw: string): string[] {
    return raw
      .split(",")
      .map((s) => s.trim())
      .filter(Boolean);
  }

  async function save(event: Event) {
    event.preventDefault();
    saving = true;
    const next = await apiAction(
      () =>
        api.savePocketID({
          enabled,
          baseUrl: baseUrl.trim(),
          apiKey: apiKey || undefined,
          defaultGroupIds: parseGroupIds(defaultGroupIds),
        }),
      {
        success: i18n.t("settings.pocketidSaved"),
        errorFallback: i18n.t("settings.pocketidSaveFailed"),
      },
    );
    if (next) {
      cfg = next;
      apiKey = "";
      defaultGroupIds = (next.defaultGroupIds ?? []).join(", ");
    }
    saving = false;
  }

  async function test() {
    testing = true;
    await apiAction(() => api.testPocketID(), {
      success: i18n.t("settings.pocketidTestOk"),
      errorFallback: i18n.t("settings.pocketidTestFailed"),
    });
    testing = false;
  }

  async function applyOidc() {
    applying = true;
    await apiAction(() => api.applyPocketIDOIDC(), {
      success: i18n.t("settings.pocketidApplyOk"),
      errorFallback: i18n.t("settings.pocketidApplyFailed"),
    });
    applying = false;
  }
</script>

<div
  id="admin-pocketid"
  class="scroll-mt-6 rounded-[var(--radius-card)] border border-border bg-surface p-5"
>
  <h2 class="text-sm font-semibold text-fg">{i18n.t("settings.adminSections.pocketid")}</h2>
  <p class="mt-1 text-xs text-muted">{i18n.t("settings.pocketidHint")}</p>
  {#if loading && !cfg}
    <div class="mt-3">
      <Skeleton height="8rem" rounded="lg" />
    </div>
  {:else}
    <form class="mt-4 space-y-3" onsubmit={save}>
      <label class="flex items-center gap-2 text-sm text-fg">
        <input type="checkbox" bind:checked={enabled} />
        {i18n.t("settings.pocketidEnabled")}
      </label>
      <input
        type="url"
        class="field-input"
        placeholder={i18n.t("settings.pocketidBaseUrl")}
        bind:value={baseUrl}
      />
      <input
        type="password"
        class="field-input"
        placeholder={cfg?.apiKeySet
          ? i18n.t("settings.pocketidApiKeyKeep")
          : i18n.t("settings.pocketidApiKey")}
        bind:value={apiKey}
        autocomplete="new-password"
      />
      <input
        type="text"
        class="field-input"
        placeholder={i18n.t("settings.pocketidDefaultGroupIds")}
        bind:value={defaultGroupIds}
      />
      <div class="flex flex-wrap gap-2 pt-1">
        <Button type="submit" size="sm" loading={saving}>{i18n.t("settings.pocketidSave")}</Button>
        <Button type="button" size="sm" variant="ghost" loading={testing} onclick={test}>
          {i18n.t("settings.pocketidTest")}
        </Button>
        <Button type="button" size="sm" variant="ghost" loading={applying} onclick={applyOidc}>
          {i18n.t("settings.pocketidApplyOidc")}
        </Button>
      </div>
    </form>
  {/if}
</div>
