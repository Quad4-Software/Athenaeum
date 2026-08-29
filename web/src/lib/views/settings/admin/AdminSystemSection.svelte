<script lang="ts">
  import Skeleton from "$lib/components/Skeleton.svelte";
  import { brand } from "$lib/brand";
  import { auth } from "$lib/stores/auth.svelte";
  import { api } from "$lib/api/client";
  import { formatBytes } from "$lib/utils/format";
  import { apiAction } from "$lib/utils/api-action";
  import type { SystemStats } from "$lib/api/types";
  import { untrack } from "svelte";

  let systemStats = $state<SystemStats | null>(null);
  let systemLoading = $state(false);

  $effect(() => {
    if (!auth.user?.isAdmin) return;
    untrack(() => {
      void loadSystemStats();
    });
  });

  async function loadSystemStats() {
    const initial = !systemStats;
    if (initial) systemLoading = true;
    const stats = await apiAction(() => api.getSystemStats(), {
      errorFallback: "Failed to load system stats",
    });
    if (stats) systemStats = stats;
    if (initial) systemLoading = false;
  }
</script>

<div
  id="admin-system"
  class="scroll-mt-6 rounded-[var(--radius-card)] border border-border bg-surface p-5"
>
  <div class="flex flex-wrap items-center justify-between gap-2">
    <h2 class="text-sm font-semibold text-fg">System</h2>
    <button
      type="button"
      class="btn btn-ghost text-xs ring-1 ring-border"
      onclick={() => loadSystemStats()}
    >
      Refresh
    </button>
  </div>
  {#if systemStats}
    <p class="mt-2 text-xs text-muted">
      {brand.appName}
      {systemStats.version}
      {#if systemStats.webVersion && systemStats.webVersion !== systemStats.version}
        (web {systemStats.webVersion})
      {/if}
    </p>
  {/if}
  {#if systemLoading && !systemStats}
    <div class="mt-3 grid gap-3 sm:grid-cols-3">
      {#each Array(3) as _, i (i)}
        <Skeleton height="5rem" rounded="lg" />
      {/each}
    </div>
  {:else if systemStats}
    <div class="mt-3 grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
      <div class="rounded-lg border border-border p-3">
        <p class="text-xs text-muted">CPU</p>
        <p class="mt-1 text-2xl font-semibold tabular-nums text-fg">
          {systemStats.cpuPercent.toFixed(1)}%
        </p>
      </div>
      <div class="rounded-lg border border-border p-3">
        <p class="text-xs text-muted">Memory</p>
        <p class="mt-1 text-2xl font-semibold tabular-nums text-fg">
          {systemStats.memPercent.toFixed(1)}%
        </p>
        <p class="mt-1 text-xs text-subtle">
          {formatBytes(systemStats.memUsed)} / {formatBytes(systemStats.memTotal)}
        </p>
      </div>
      {#each systemStats.disks as disk (disk.path)}
        <div class="rounded-lg border border-border p-3">
          <p class="truncate text-xs text-muted" title={disk.path}>{disk.path}</p>
          <p class="mt-1 text-2xl font-semibold tabular-nums text-fg">
            {disk.percent.toFixed(1)}%
          </p>
          <p class="mt-1 text-xs text-subtle">
            {formatBytes(disk.used)} / {formatBytes(disk.total)}
          </p>
        </div>
      {/each}
    </div>
  {/if}
</div>
