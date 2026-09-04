<script lang="ts">
  import { library } from "$lib/stores/library.svelte";
  import { auth } from "$lib/stores/auth.svelte";
  import { scan } from "$lib/stores/scan.svelte";
  import { formatBytes } from "$lib/utils/format";
</script>

<div class="rounded-[var(--radius-card)] border border-border bg-surface p-5">
  <h2 class="text-sm font-semibold text-fg">Catalog stats</h2>
  <dl class="mt-3 grid grid-cols-2 gap-4 text-center sm:grid-cols-4">
    <div>
      <dt class="text-xs text-subtle">Total</dt>
      <dd class="text-xl font-semibold text-fg">{library.stats?.totalBooks ?? 0}</dd>
    </div>
    <div>
      <dt class="text-xs text-subtle">EPUB</dt>
      <dd class="text-xl font-semibold text-fg">{library.stats?.epubCount ?? 0}</dd>
    </div>
    <div>
      <dt class="text-xs text-subtle">PDF</dt>
      <dd class="text-xl font-semibold text-fg">{library.stats?.pdfCount ?? 0}</dd>
    </div>
    <div>
      <dt class="text-xs text-subtle">Audio</dt>
      <dd class="text-xl font-semibold text-fg">{library.stats?.audioCount ?? 0}</dd>
    </div>
    <div>
      <dt class="text-xs text-subtle">Authors</dt>
      <dd class="text-xl font-semibold text-fg">{library.stats?.authorCount ?? 0}</dd>
    </div>
    <div>
      <dt class="text-xs text-subtle">Series</dt>
      <dd class="text-xl font-semibold text-fg">{library.stats?.seriesCount ?? 0}</dd>
    </div>
    <div>
      <dt class="text-xs text-subtle">Added (7d)</dt>
      <dd class="text-xl font-semibold text-fg">{library.stats?.addedLast7Days ?? 0}</dd>
    </div>
    <div>
      <dt class="text-xs text-subtle">Storage</dt>
      <dd class="text-xl font-semibold text-fg">
        {formatBytes(library.stats?.totalSizeBytes ?? 0)}
      </dd>
    </div>
    <div>
      <dt class="text-xs text-subtle">Libraries</dt>
      <dd class="text-xl font-semibold text-fg">{library.stats?.libraryCount ?? 0}</dd>
    </div>
    <div>
      <dt class="text-xs text-subtle">Collections</dt>
      <dd class="text-xl font-semibold text-fg">{library.stats?.collectionCount ?? 0}</dd>
    </div>
    {#if auth.user}
      <div>
        <dt class="text-xs text-subtle">In progress</dt>
        <dd class="text-xl font-semibold text-fg">
          {library.stats?.readingInProgress ?? 0}
        </dd>
      </div>
      <div>
        <dt class="text-xs text-subtle">Completed</dt>
        <dd class="text-xl font-semibold text-fg">
          {library.stats?.readingCompleted ?? 0}
        </dd>
      </div>
      <div>
        <dt class="text-xs text-subtle">Favorites</dt>
        <dd class="text-xl font-semibold text-fg">{library.stats?.favoriteCount ?? 0}</dd>
      </div>
    {/if}
    {#if auth.user?.isAdmin && library.stats?.userCount != null}
      <div>
        <dt class="text-xs text-subtle">Users</dt>
        <dd class="text-xl font-semibold text-fg">{library.stats.userCount}</dd>
      </div>
    {/if}
  </dl>
  {#if library.stats?.lastScanAt}
    <p class="mt-3 text-xs text-subtle">
      Last scan: {new Date(library.stats.lastScanAt).toLocaleString()}
    </p>
  {/if}
  {#if scan.status?.scanning}
    <div class="mt-3 rounded-lg border border-border bg-bg px-3 py-2 text-xs text-muted">
      <p class="font-medium text-fg">Scan in progress</p>
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
</div>
