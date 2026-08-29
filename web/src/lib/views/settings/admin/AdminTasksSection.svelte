<script lang="ts">
  import Button from "$lib/components/Button.svelte";
  import { library } from "$lib/stores/library.svelte";
  import { auth } from "$lib/stores/auth.svelte";
  import { toast } from "$lib/stores/toast.svelte";
  import { i18n } from "$lib/stores/i18n.svelte";
  import { ApiError, api } from "$lib/api/client";
  import type { IntegrityReport, MaintenanceStatus } from "$lib/api/types";
  import { scan } from "$lib/stores/scan.svelte";
  import { apiAction, apiErrorMessage } from "$lib/utils/api-action";
  import { untrack } from "svelte";

  const MAINTENANCE_POLL_MS = 2_000;

  let integrityReport = $state<IntegrityReport | null>(null);
  let maintenanceStatus = $state<MaintenanceStatus | null>(null);
  const maintenanceProgressPct = $derived(
    maintenanceStatus && maintenanceStatus.total > 0
      ? Math.min(100, (maintenanceStatus.done / maintenanceStatus.total) * 100)
      : 0,
  );
  let taskVerify = $state(false);
  let taskPrune = $state(false);
  let taskCleanupCovers = $state(false);
  let taskRegenerate = $state(false);
  let taskCleanupSeries = $state(false);
  let taskCleanupText = $state(false);
  let taskRescan = $state(false);
  let taskContentIndex = $state(false);

  $effect(() => {
    if (!auth.user?.isAdmin) return;
    untrack(() => {
      void refreshMaintenanceStatus();
    });
  });

  $effect(() => {
    if (!auth.user?.isAdmin) return;
    const id = setInterval(() => {
      if (maintenanceStatus?.running) void refreshMaintenanceStatus();
    }, MAINTENANCE_POLL_MS);
    return () => clearInterval(id);
  });

  async function refreshMaintenanceStatus() {
    try {
      const prev = maintenanceStatus;
      maintenanceStatus = await api.adminTaskStatus();
      if (prev?.running && maintenanceStatus && !maintenanceStatus.running) {
        if (maintenanceStatus.task === "regenerate_covers") {
          toast.success(
            i18n.t("admin.tasks.regenerateDoneFull", {
              updated: String(maintenanceStatus.updated),
              skipped: String(maintenanceStatus.skipped),
              failed: String(maintenanceStatus.failed),
            }),
          );
        }
        void library.refresh({ background: true });
      }
    } catch {
      maintenanceStatus = null;
    }
  }

  async function verifyIntegrity() {
    taskVerify = true;
    try {
      integrityReport = await api.adminVerifyIntegrity();
      if (integrityReport.missingCount === 0 && integrityReport.orphanCovers === 0) {
        toast.success(i18n.t("admin.tasks.verifyOk"));
      } else {
        toast.info(
          i18n.t("admin.tasks.verifyIssues", {
            missing: String(integrityReport.missingCount),
            orphan: String(integrityReport.orphanCovers),
          }),
        );
      }
    } catch (e) {
      toast.error(apiErrorMessage(e, i18n.t("admin.tasks.failed")));
    } finally {
      taskVerify = false;
    }
  }

  async function pruneMissingBooks() {
    taskPrune = true;
    const result = await apiAction(() => api.adminPruneMissing(), {
      errorFallback: i18n.t("admin.tasks.failed"),
    });
    if (result) {
      toast.success(i18n.t("admin.tasks.pruneOk", { count: String(result.removed) }));
      integrityReport = null;
      void library.refresh({ background: true });
    }
    taskPrune = false;
  }

  async function cleanupOrphanCovers() {
    taskCleanupCovers = true;
    const result = await apiAction(() => api.adminCleanupCovers(), {
      errorFallback: i18n.t("admin.tasks.failed"),
    });
    if (result) {
      toast.success(i18n.t("admin.tasks.cleanupCoversOk", { count: String(result.removed) }));
      integrityReport = null;
    }
    taskCleanupCovers = false;
  }

  async function regenerateCovers() {
    taskRegenerate = true;
    try {
      await api.adminRegenerateCovers();
      toast.success(i18n.t("admin.tasks.regenerateStarted"));
      await refreshMaintenanceStatus();
    } catch (e) {
      const msg =
        e instanceof ApiError && e.status === 409
          ? i18n.t("admin.tasks.busy")
          : apiErrorMessage(e, i18n.t("admin.tasks.failed"));
      toast.error(msg);
    } finally {
      taskRegenerate = false;
    }
  }

  async function cleanupSeriesNames() {
    taskCleanupSeries = true;
    const result = await apiAction(() => api.adminCleanupSeries(), {
      errorFallback: i18n.t("admin.tasks.failed"),
    });
    if (result) {
      toast.success(i18n.t("admin.tasks.cleanupSeriesOk", { count: String(result.updated) }));
      void library.loadSeries();
      void library.refresh({ background: true });
    }
    taskCleanupSeries = false;
  }

  async function cleanupBookText() {
    taskCleanupText = true;
    const result = await apiAction(() => api.adminCleanupText(), {
      errorFallback: i18n.t("admin.tasks.failed"),
    });
    if (result) {
      toast.success(i18n.t("admin.tasks.cleanupTextOk", { count: String(result.updated) }));
      void library.refresh({ background: true });
    }
    taskCleanupText = false;
  }

  async function rescanAllLibraries() {
    taskRescan = true;
    await apiAction(() => library.triggerScan(), {
      success: i18n.t("admin.tasks.rescanStarted"),
      errorFallback: i18n.t("admin.tasks.failed"),
    });
    taskRescan = false;
  }

  async function startContentIndex() {
    taskContentIndex = true;
    try {
      await toast.promise(() => api.adminContentIndex(), {
        loading: i18n.t("admin.tasks.contentIndexLoading"),
        success: i18n.t("admin.tasks.contentIndexStarted"),
        error: (e) => (e instanceof ApiError ? e.message : i18n.t("admin.tasks.failed")),
      });
    } catch {
      // toast.promise already surfaced the error
    } finally {
      taskContentIndex = false;
    }
  }
</script>

<div
  id="admin-tasks"
  class="scroll-mt-6 rounded-[var(--radius-card)] border border-border bg-surface p-5"
>
  <h2 class="text-sm font-semibold text-fg">{i18n.t("admin.tasks.title")}</h2>
  <p class="mt-1 text-xs text-muted">{i18n.t("admin.tasks.description")}</p>

  {#if maintenanceStatus?.running}
    <div class="mt-4 rounded-lg border border-primary/30 bg-primary/5 px-3 py-3">
      <div class="flex flex-wrap items-center justify-between gap-2">
        <p class="text-sm font-medium text-fg">
          {i18n.t("admin.tasks.regenerateRunning", {
            done: String(maintenanceStatus.done),
            total: String(maintenanceStatus.total),
          })}
        </p>
        {#if maintenanceStatus.total > 0}
          <span class="text-xs tabular-nums text-muted">
            {Math.round(maintenanceProgressPct)}%
          </span>
        {/if}
      </div>
      {#if maintenanceStatus.total > 0}
        <div class="mt-2 h-1.5 overflow-hidden rounded-full bg-border">
          <div
            class="h-full bg-primary transition-[width]"
            style:width="{maintenanceProgressPct}%"
          ></div>
        </div>
      {/if}
      {#if maintenanceStatus.currentTitle}
        <p class="mt-2 truncate text-xs text-subtle">
          {maintenanceStatus.currentTitle}
        </p>
      {/if}
      <dl class="mt-2 flex flex-wrap gap-x-4 gap-y-1 text-xs text-muted">
        <div>
          <dt class="inline">{i18n.t("admin.tasks.statusUpdated")}:</dt>
          <dd class="inline tabular-nums text-fg">{maintenanceStatus.updated}</dd>
        </div>
        <div>
          <dt class="inline">{i18n.t("admin.tasks.statusSkipped")}:</dt>
          <dd class="inline tabular-nums text-fg">{maintenanceStatus.skipped}</dd>
        </div>
        <div>
          <dt class="inline">{i18n.t("admin.tasks.statusFailed")}:</dt>
          <dd class="inline tabular-nums text-fg">{maintenanceStatus.failed}</dd>
        </div>
      </dl>
    </div>
  {:else if maintenanceStatus?.finishedAt && maintenanceStatus.task === "regenerate_covers"}
    <div class="mt-4 rounded-lg border border-border bg-bg px-3 py-3 text-xs text-muted">
      <p class="font-medium text-fg">{i18n.t("admin.tasks.statusComplete")}</p>
      <p class="mt-1">
        {i18n.t("admin.tasks.regenerateDoneFull", {
          updated: String(maintenanceStatus.updated),
          skipped: String(maintenanceStatus.skipped),
          failed: String(maintenanceStatus.failed),
        })}
      </p>
      <p class="mt-1 text-subtle">
        {new Date(maintenanceStatus.finishedAt).toLocaleString()}
      </p>
    </div>
  {/if}

  {#if scan.status?.scanning}
    <div class="mt-4 rounded-lg border border-border bg-bg px-3 py-3 text-xs text-muted">
      <p class="font-medium text-fg">{i18n.t("admin.tasks.scanRunning")}</p>
      <p class="mt-1">
        Indexed {scan.status.indexed.toLocaleString()}, skipped {scan.status.skipped.toLocaleString()}
        {#if scan.status.libraryName}
          · {scan.status.libraryName}
        {/if}
      </p>
      {#if scan.status.currentPath}
        <p class="mt-1 truncate text-subtle">{scan.status.currentPath}</p>
      {/if}
    </div>
  {/if}

  <div class="mt-4 flex flex-wrap gap-2">
    <Button
      variant="ghost"
      class="ring-1 ring-border"
      loading={taskVerify}
      onclick={verifyIntegrity}
    >
      {i18n.t("admin.tasks.verify")}
    </Button>
    <Button
      variant="ghost"
      class="ring-1 ring-border"
      loading={taskPrune}
      onclick={pruneMissingBooks}
    >
      {i18n.t("admin.tasks.pruneMissing")}
    </Button>
    <Button
      variant="ghost"
      class="ring-1 ring-border"
      loading={taskCleanupCovers}
      onclick={cleanupOrphanCovers}
    >
      {i18n.t("admin.tasks.cleanupCovers")}
    </Button>
    <Button
      variant="ghost"
      class="ring-1 ring-border"
      loading={taskRegenerate || maintenanceStatus?.running === true}
      onclick={regenerateCovers}
    >
      {i18n.t("admin.tasks.regenerateCovers")}
    </Button>
    <Button
      variant="ghost"
      class="ring-1 ring-border"
      loading={taskCleanupSeries}
      onclick={cleanupSeriesNames}
    >
      {i18n.t("admin.tasks.cleanupSeries")}
    </Button>
    <Button
      variant="ghost"
      class="ring-1 ring-border"
      loading={taskCleanupText}
      onclick={cleanupBookText}
    >
      {i18n.t("admin.tasks.cleanupText")}
    </Button>
    <Button
      variant="ghost"
      class="ring-1 ring-border"
      loading={taskRescan}
      onclick={rescanAllLibraries}
    >
      {i18n.t("admin.tasks.rescan")}
    </Button>
    <Button
      variant="ghost"
      class="ring-1 ring-border"
      loading={taskContentIndex}
      onclick={startContentIndex}
    >
      {i18n.t("admin.tasks.contentIndex")}
    </Button>
  </div>

  {#if integrityReport && (integrityReport.missingCount > 0 || integrityReport.orphanCovers > 0)}
    <div class="mt-4 rounded-lg border border-border p-3 text-xs text-muted">
      <p>
        {i18n.t("admin.tasks.verifyIssues", {
          missing: String(integrityReport.missingCount),
          orphan: String(integrityReport.orphanCovers),
        })}
      </p>
      {#if integrityReport.missingFiles.length > 0}
        <p class="mt-2 font-medium text-fg">{i18n.t("admin.tasks.missingSample")}</p>
        <ul class="mt-1 space-y-1 font-mono">
          {#each integrityReport.missingFiles as item (item.id)}
            <li class="truncate" title={item.relPath}>{item.title} — {item.relPath}</li>
          {/each}
        </ul>
      {/if}
    </div>
  {/if}
</div>
