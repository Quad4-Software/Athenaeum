<script lang="ts">
  import { brand } from "$lib/brand";
  import { toast } from "$lib/stores/toast.svelte";
  import { i18n } from "$lib/stores/i18n.svelte";
  import { ApiError, api } from "$lib/api/client";
  import { apiAction } from "$lib/utils/api-action";

  let restoreFile = $state<File | null>(null);
  let restoring = $state(false);
  let configImportText = $state("");

  function downloadBackup() {
    window.location.href = "/api/admin/backup";
  }

  async function restoreBackup(e: SubmitEvent) {
    e.preventDefault();
    if (!restoreFile) return;
    restoring = true;
    try {
      const res = await toast.promise(() => api.restoreBackup(restoreFile!), {
        loading: i18n.t("admin.backup.restoring"),
        success: (r) => r.message,
        error: (err) =>
          err instanceof ApiError ? err.message : i18n.t("admin.backup.restoreFailed"),
      });
      restoreFile = null;
      void res;
    } catch {
      // toast.promise already surfaced the error
    } finally {
      restoring = false;
    }
  }

  async function exportConfigFile() {
    const cfg = await apiAction(() => api.exportConfig(), {
      errorFallback: i18n.t("admin.backup.exportFailed"),
    });
    if (!cfg) return;
    const blob = new Blob([JSON.stringify(cfg, null, 2)], { type: "application/json" });
    const url = URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = url;
    a.download = brand.configExportName;
    a.click();
    URL.revokeObjectURL(url);
  }

  async function importConfigFile(e: SubmitEvent) {
    e.preventDefault();
    if (!configImportText.trim()) return;
    let body: Record<string, unknown>;
    try {
      body = JSON.parse(configImportText) as Record<string, unknown>;
    } catch {
      toast.error(i18n.t("admin.backup.importFailed"));
      return;
    }
    const ok = await apiAction(() => api.importConfig(body), {
      success: i18n.t("admin.backup.importOk"),
      errorFallback: i18n.t("admin.backup.importFailed"),
    });
    if (ok !== undefined) configImportText = "";
  }
</script>

<div
  id="admin-backup"
  class="scroll-mt-6 rounded-[var(--radius-card)] border border-border bg-surface p-5"
>
  <h2 class="text-sm font-semibold text-fg">{i18n.t("admin.backup.title")}</h2>
  <p class="mt-1 text-xs text-muted">{i18n.t("admin.backup.description")}</p>
  <div class="mt-4 flex flex-wrap gap-2">
    <button type="button" class="btn btn-primary text-sm" onclick={downloadBackup}>
      {i18n.t("admin.backup.download")}
    </button>
    <button
      type="button"
      class="btn btn-ghost text-sm ring-1 ring-border"
      onclick={exportConfigFile}
    >
      {i18n.t("admin.backup.exportConfig")}
    </button>
  </div>
  <form class="mt-4 space-y-3" onsubmit={restoreBackup}>
    <p class="text-sm font-medium text-fg">{i18n.t("admin.backup.restore")}</p>
    <input
      type="file"
      accept=".zip,application/zip"
      class="field-input"
      onchange={(e) => {
        const input = e.currentTarget;
        restoreFile = input.files?.[0] ?? null;
      }}
    />
    <div class="pt-1">
      <button
        type="submit"
        class="btn btn-ghost text-sm ring-1 ring-border"
        disabled={!restoreFile || restoring}
      >
        {restoring ? i18n.t("admin.backup.restoring") : i18n.t("admin.backup.restoreSubmit")}
      </button>
    </div>
  </form>
  <form class="mt-6 space-y-3 border-t border-border pt-6" onsubmit={importConfigFile}>
    <p class="text-sm font-medium text-fg">{i18n.t("admin.backup.importConfig")}</p>
    <textarea
      bind:value={configImportText}
      class="field-input font-mono text-xs"
      placeholder={i18n.t("admin.backup.importPlaceholder")}></textarea>
    <div class="pt-1">
      <button
        type="submit"
        class="btn btn-ghost text-sm ring-1 ring-border"
        disabled={!configImportText.trim()}
      >
        {i18n.t("admin.backup.importSubmit")}
      </button>
    </div>
  </form>
</div>
