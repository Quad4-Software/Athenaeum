<script lang="ts">
  import Button from "$lib/components/Button.svelte";
  import Skeleton from "$lib/components/Skeleton.svelte";
  import { auth } from "$lib/stores/auth.svelte";
  import { toast } from "$lib/stores/toast.svelte";
  import { i18n } from "$lib/stores/i18n.svelte";
  import { getTTSAdmin, saveTTSAdmin, testTTSAdmin } from "$lib/narrator/kokoro";
  import type { TTSSettingsPublic } from "$lib/narrator/types";
  import { apiAction, apiErrorMessage } from "$lib/utils/api-action";
  import { untrack } from "svelte";

  let cfg = $state<TTSSettingsPublic | null>(null);
  let loading = $state(false);
  let saving = $state(false);
  let testing = $state(false);
  let apiKey = $state("");
  let enabled = $state(false);
  let baseUrl = $state("");
  let defaultVoice = $state("af_heart");
  let timeoutSec = $state(60);

  $effect(() => {
    if (!auth.user?.isAdmin) return;
    untrack(() => {
      void load();
    });
  });

  async function load() {
    loading = true;
    const next = await apiAction(() => getTTSAdmin(), {
      errorFallback: i18n.t("settings.ttsLoadFailed"),
    });
    if (next) {
      cfg = next;
      enabled = next.enabled;
      baseUrl = next.baseUrl;
      defaultVoice = next.defaultVoice || "af_heart";
      timeoutSec = next.timeoutSec || 60;
      apiKey = "";
    }
    loading = false;
  }

  async function save(event: Event) {
    event.preventDefault();
    saving = true;
    const next = await apiAction(
      () =>
        saveTTSAdmin({
          enabled,
          baseUrl: baseUrl.trim(),
          defaultVoice: defaultVoice.trim() || "af_heart",
          apiKey: apiKey || undefined,
          timeoutSec,
        }),
      {
        success: i18n.t("settings.ttsSaved"),
        errorFallback: i18n.t("settings.ttsSaveFailed"),
      },
    );
    if (next) {
      cfg = next;
      apiKey = "";
    }
    saving = false;
  }

  async function testConnection() {
    testing = true;
    try {
      const res = await testTTSAdmin();
      if (res.ok) toast.success(res.message || i18n.t("settings.ttsTestOk"));
      else toast.error(res.message || i18n.t("settings.ttsTestFailed"));
    } catch (e) {
      toast.error(apiErrorMessage(e, i18n.t("settings.ttsTestFailed")));
    } finally {
      testing = false;
    }
  }
</script>

<div
  id="admin-tts"
  class="scroll-mt-6 rounded-[var(--radius-card)] border border-border bg-surface p-5"
>
  <h2 class="text-sm font-semibold text-fg">{i18n.t("settings.adminSections.tts")}</h2>
  <p class="mt-1 text-xs text-muted">{i18n.t("settings.ttsHint")}</p>
  {#if loading && !cfg}
    <div class="mt-3">
      <Skeleton height="8rem" rounded="lg" />
    </div>
  {:else}
    <form class="mt-4 space-y-3" onsubmit={save}>
      <label class="flex items-center gap-2 text-sm text-fg">
        <input type="checkbox" bind:checked={enabled} />
        {i18n.t("settings.ttsEnabled")}
      </label>
      <input
        type="url"
        class="field-input"
        placeholder={i18n.t("settings.ttsBaseUrl")}
        bind:value={baseUrl}
        autocomplete="off"
      />
      <input
        type="text"
        class="field-input"
        placeholder={i18n.t("settings.ttsDefaultVoice")}
        bind:value={defaultVoice}
        autocomplete="off"
      />
      <input
        type="password"
        class="field-input"
        placeholder={cfg?.apiKeySet
          ? i18n.t("settings.ttsApiKeyKeep")
          : i18n.t("settings.ttsApiKey")}
        bind:value={apiKey}
        autocomplete="new-password"
      />
      <label class="block text-sm text-fg">
        <span class="mb-1 block text-xs text-muted">{i18n.t("settings.ttsTimeout")}</span>
        <input type="number" class="field-input" bind:value={timeoutSec} min="5" max="300" />
      </label>
      <div class="flex flex-wrap gap-2">
        <Button type="submit" disabled={saving} loading={saving}>
          {i18n.t("settings.ttsSave")}
        </Button>
        <Button
          type="button"
          variant="ghost"
          disabled={testing || !enabled}
          loading={testing}
          onclick={testConnection}
        >
          {i18n.t("settings.ttsTest")}
        </Button>
      </div>
    </form>
  {/if}
</div>
