<script lang="ts">
  import {
    ChevronDown,
    ChevronUp,
    FolderOpen,
    Moon,
    Pencil,
    Plus,
    Sun,
    Trash2,
  } from "@lucide/svelte";
  import FolderBrowser from "$lib/components/FolderBrowser.svelte";
  import Button from "$lib/components/Button.svelte";
  import Skeleton from "$lib/components/Skeleton.svelte";
  import UploadPanel from "$lib/components/UploadPanel.svelte";
  import { listAppThemes } from "$lib/brand";
  import { theme } from "$lib/stores/theme.svelte";
  import { library } from "$lib/stores/library.svelte";
  import { libraries } from "$lib/stores/libraries.svelte";
  import { auth } from "$lib/stores/auth.svelte";
  import { toast } from "$lib/stores/toast.svelte";
  import { sidebarPrefs, SIDEBAR_SECTION_LABELS } from "$lib/stores/sidebar.svelte";
  import { ApiError } from "$lib/api/client";
  import { formatBytes } from "$lib/utils/format";
  import type { LibraryS3Input, SidebarSectionId } from "$lib/api/types";
  import { scan } from "$lib/stores/scan.svelte";

  type BackendKind = "local" | "s3";

  let libName = $state("");
  let libPath = $state("");
  let libBackend = $state<BackendKind>("local");
  let libS3 = $state(emptyS3());
  let libMsg = $state<string | null>(null);
  let libCreating = $state(false);
  let editingId = $state<number | null>(null);
  let editName = $state("");
  let editPath = $state("");
  let editBackend = $state<BackendKind>("local");
  let editS3 = $state(emptyS3());
  let confirmDeleteId = $state<number | null>(null);
  let libSaving = $state(false);
  let browseOpen = $state(false);
  let browseTarget = $state<"add" | "edit">("add");
  let scanningAll = $state(false);
  let testingS3 = $state(false);

  function emptyS3(): LibraryS3Input {
    return {
      endpoint: "",
      region: "us-east-1",
      bucket: "",
      prefix: "",
      accessKey: "",
      secretKey: "",
      usePathStyle: true,
      tls: true,
    };
  }

  $effect(() => {
    void libraries.refresh();
  });

  async function addLibrary(event: Event) {
    event.preventDefault();
    libMsg = null;
    if (!libName.trim()) return;
    if (libBackend === "local" && !libPath.trim()) return;
    if (libBackend === "s3" && (!libS3.bucket.trim() || !libS3.endpoint.trim())) return;
    libCreating = true;
    try {
      await libraries.create({
        name: libName.trim(),
        backend: libBackend,
        mountPath: libBackend === "local" ? libPath.trim() : "",
        s3: libBackend === "s3" ? { ...libS3 } : undefined,
      });
      libName = "";
      libPath = "";
      libS3 = emptyS3();
      void library.refresh();
    } catch (e) {
      libMsg = e instanceof ApiError ? e.message : "Failed to add library";
      toast.error(libMsg);
    } finally {
      libCreating = false;
    }
  }

  async function removeLibrary(id: number) {
    libMsg = null;
    try {
      await libraries.remove(id);
      if (library.libraryFilter === id) library.setLibrary(null);
      confirmDeleteId = null;
      toast.info("Library removed");
      void library.refresh();
    } catch (e) {
      libMsg = e instanceof ApiError ? e.message : "Failed to remove library";
      toast.error(libMsg);
    }
  }

  function startEdit(lib: (typeof libraries.items)[number]) {
    editingId = lib.id;
    editName = lib.name;
    editPath = lib.mountPath;
    editBackend = lib.backend === "s3" ? "s3" : "local";
    editS3 = emptyS3();
    if (lib.s3) {
      editS3 = {
        endpoint: lib.s3.endpoint,
        region: lib.s3.region || "us-east-1",
        bucket: lib.s3.bucket,
        prefix: lib.s3.prefix || "",
        accessKey: lib.s3.accessKey,
        secretKey: "",
        usePathStyle: lib.s3.usePathStyle,
        tls: lib.s3.tls,
      };
    }
    confirmDeleteId = null;
  }

  function cancelEdit() {
    editingId = null;
    editName = "";
    editPath = "";
    editBackend = "local";
    editS3 = emptyS3();
  }

  async function saveEdit(event: Event) {
    event.preventDefault();
    if (editingId == null || !editName.trim()) return;
    if (editBackend === "local" && !editPath.trim()) return;
    if (editBackend === "s3" && (!editS3.bucket.trim() || !editS3.endpoint.trim())) return;
    libSaving = true;
    libMsg = null;
    try {
      await libraries.update(editingId, {
        name: editName.trim(),
        backend: editBackend,
        mountPath: editBackend === "local" ? editPath.trim() : "",
        s3: editBackend === "s3" ? { ...editS3 } : undefined,
      });
      cancelEdit();
      toast.success("Library updated");
      void library.refresh();
    } catch (e) {
      libMsg = e instanceof ApiError ? e.message : "Failed to update library";
      toast.error(libMsg);
    } finally {
      libSaving = false;
    }
  }

  async function testS3Connection(cfg: LibraryS3Input) {
    testingS3 = true;
    try {
      await libraries.testS3(cfg);
      toast.success("S3 connection OK");
    } catch (e) {
      const msg = e instanceof ApiError ? e.message : "S3 test failed";
      toast.error(msg);
    } finally {
      testingS3 = false;
    }
  }

  async function scanLibrary(id: number) {
    try {
      await library.triggerScan(id);
    } catch (e) {
      libMsg = e instanceof ApiError ? e.message : "Scan failed";
      toast.error(libMsg);
    }
  }

  function openBrowse(target: "add" | "edit") {
    browseTarget = target;
    browseOpen = true;
  }

  function onBrowseSelect(path: string) {
    if (browseTarget === "add") libPath = path;
    else editPath = path;
  }

  async function scanAll() {
    scanningAll = true;
    try {
      await library.triggerScan();
    } finally {
      scanningAll = false;
    }
  }
</script>

<FolderBrowser
  bind:open={browseOpen}
  value={browseTarget === "add" ? libPath : editPath}
  onselect={onBrowseSelect}
/>

<div class="space-y-6">
  <div class="rounded-[var(--radius-card)] border border-border bg-surface p-5">
    <h2 class="text-sm font-semibold text-fg">Appearance</h2>
    <p class="mt-1 text-sm text-muted">Choose your preferred color theme.</p>
    <div class="mt-3 flex flex-wrap gap-2">
      {#each listAppThemes() as appTheme (appTheme.id)}
        <button
          class="btn ring-1 ring-border {theme.preference === appTheme.id ||
          (theme.preference === 'system' && theme.activeThemeId === appTheme.id)
            ? 'bg-primary text-primary-fg'
            : 'btn-ghost'}"
          onclick={() => theme.set(appTheme.id)}
        >
          {#if appTheme.id === "light"}
            <Sun size={16} />
          {:else if appTheme.id === "dark"}
            <Moon size={16} />
          {/if}
          {appTheme.label}
        </button>
      {/each}
      <button
        class="btn ring-1 ring-border {theme.preference === 'system'
          ? 'bg-primary text-primary-fg'
          : 'btn-ghost'}"
        onclick={() => theme.set("system")}
      >
        System
      </button>
    </div>
  </div>

  <div class="rounded-[var(--radius-card)] border border-border bg-surface p-5">
    <h2 class="text-sm font-semibold text-fg">Library mounts</h2>
    <p class="mt-1 text-sm text-muted">
      Add local folders or MinIO-compatible S3 buckets. Each mount is scanned independently. S3
      mounts use periodic or manual scan (no filesystem watch).
    </p>

    {#if libraries.loading}
      <div class="mt-3 space-y-2">
        {#each Array(3) as _, i (i)}
          <Skeleton height="4rem" rounded="lg" />
        {/each}
      </div>
    {:else if libraries.items.length > 0}
      <ul class="mt-4 divide-y divide-border rounded-lg border border-border">
        {#each libraries.items as lib (lib.id)}
          <li class="px-3 py-3">
            {#if editingId === lib.id}
              <form class="space-y-3" onsubmit={saveEdit}>
                <p class="text-sm font-medium text-fg">Edit library</p>
                <input type="text" bind:value={editName} required class="field-input" />
                <div class="flex flex-wrap gap-2">
                  <button
                    type="button"
                    class="btn text-xs ring-1 ring-border {editBackend === 'local'
                      ? 'bg-primary text-primary-fg'
                      : 'btn-ghost'}"
                    onclick={() => (editBackend = "local")}
                  >
                    Local
                  </button>
                  <button
                    type="button"
                    class="btn text-xs ring-1 ring-border {editBackend === 's3'
                      ? 'bg-primary text-primary-fg'
                      : 'btn-ghost'}"
                    onclick={() => (editBackend = "s3")}
                  >
                    S3
                  </button>
                </div>
                {#if editBackend === "local"}
                  <div class="flex gap-2">
                    <input
                      type="text"
                      bind:value={editPath}
                      required
                      class="field-input font-mono flex-1"
                    />
                    <Button
                      type="button"
                      variant="ghost"
                      class="ring-1 ring-border"
                      onclick={() => openBrowse("edit")}
                    >
                      <FolderOpen size={14} />
                    </Button>
                  </div>
                {:else}
                  <div class="grid gap-2 sm:grid-cols-2">
                    <input
                      type="text"
                      placeholder="Endpoint (minio:9000)"
                      bind:value={editS3.endpoint}
                      class="field-input font-mono sm:col-span-2"
                      required
                    />
                    <input
                      type="text"
                      placeholder="Bucket"
                      bind:value={editS3.bucket}
                      class="field-input"
                      required
                    />
                    <input
                      type="text"
                      placeholder="Prefix (optional)"
                      bind:value={editS3.prefix}
                      class="field-input font-mono"
                    />
                    <input
                      type="text"
                      placeholder="Region"
                      bind:value={editS3.region}
                      class="field-input"
                    />
                    <input
                      type="text"
                      placeholder="Access key"
                      bind:value={editS3.accessKey}
                      class="field-input font-mono"
                      required
                    />
                    <input
                      type="password"
                      placeholder="Secret key (leave blank to keep)"
                      bind:value={editS3.secretKey}
                      class="field-input font-mono sm:col-span-2"
                    />
                    <label class="flex items-center gap-2 text-xs text-muted">
                      <input type="checkbox" bind:checked={editS3.usePathStyle} />
                      Path-style
                    </label>
                    <label class="flex items-center gap-2 text-xs text-muted">
                      <input type="checkbox" bind:checked={editS3.tls} />
                      TLS
                    </label>
                  </div>
                  <Button
                    type="button"
                    size="sm"
                    variant="ghost"
                    class="ring-1 ring-border"
                    loading={testingS3}
                    onclick={() => testS3Connection(editS3)}
                  >
                    Test connection
                  </Button>
                {/if}
                <div class="flex flex-wrap gap-2">
                  <Button type="submit" size="sm" loading={libSaving}>Save</Button>
                  <button
                    type="button"
                    class="btn btn-ghost text-xs ring-1 ring-border"
                    onclick={cancelEdit}
                  >
                    Cancel
                  </button>
                </div>
              </form>
            {:else if confirmDeleteId === lib.id}
              <div class="rounded-lg border border-danger/30 bg-danger/5 p-3">
                <p class="text-sm font-medium text-fg">Delete "{lib.name}"?</p>
                <p class="mt-1 text-xs text-muted">
                  Removes this mount and all {lib.bookCount} indexed books from the catalog. Files on
                  disk are not deleted.
                </p>
                <div class="mt-3 flex flex-wrap gap-2">
                  <button
                    type="button"
                    class="btn btn-primary text-xs !bg-danger hover:!bg-danger"
                    onclick={() => removeLibrary(lib.id)}
                  >
                    Delete
                  </button>
                  <button
                    type="button"
                    class="btn btn-ghost text-xs ring-1 ring-border"
                    onclick={() => (confirmDeleteId = null)}
                  >
                    Cancel
                  </button>
                </div>
              </div>
            {:else}
              <div class="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
                <div class="min-w-0">
                  <div class="flex flex-wrap items-center gap-2">
                    <p class="font-medium text-fg">{lib.name}</p>
                    {#if lib.id === 1}
                      <span
                        class="rounded-full bg-surface px-2 py-0.5 text-[0.65rem] font-medium uppercase tracking-wide text-muted ring-1 ring-border"
                      >
                        Default
                      </span>
                    {/if}
                    <span
                      class="rounded-full bg-surface px-2 py-0.5 text-[0.65rem] font-medium uppercase tracking-wide text-muted ring-1 ring-border"
                    >
                      {lib.backend === "s3" ? "S3" : "Local"}
                    </span>
                  </div>
                  <p class="truncate text-xs text-muted">{lib.mountPath}</p>
                  <p class="text-xs text-subtle">{lib.bookCount} books</p>
                </div>
                <div class="flex shrink-0 flex-wrap gap-1">
                  <button
                    type="button"
                    class="btn btn-ghost text-xs ring-1 ring-border"
                    aria-label="Move up"
                    onclick={() => libraries.move(lib.id, -1)}
                  >
                    <ChevronUp size={14} />
                  </button>
                  <button
                    type="button"
                    class="btn btn-ghost text-xs ring-1 ring-border"
                    aria-label="Move down"
                    onclick={() => libraries.move(lib.id, 1)}
                  >
                    <ChevronDown size={14} />
                  </button>
                  <button
                    type="button"
                    class="btn btn-ghost text-xs ring-1 ring-border"
                    onclick={() => scanLibrary(lib.id)}
                  >
                    Scan
                  </button>
                  <button
                    type="button"
                    class="btn btn-ghost text-xs ring-1 ring-border"
                    onclick={() => startEdit(lib)}
                  >
                    <Pencil size={14} />
                  </button>
                  <button
                    type="button"
                    class="btn btn-ghost text-xs text-danger"
                    aria-label="Remove library"
                    onclick={() => {
                      confirmDeleteId = lib.id;
                      editingId = null;
                    }}
                  >
                    <Trash2 size={14} />
                  </button>
                </div>
              </div>
            {/if}
          </li>
        {/each}
      </ul>
    {/if}

    <form class="mt-4 space-y-3 border-t border-border pt-4" onsubmit={addLibrary}>
      <p class="text-sm font-medium text-fg">Add library</p>
      <input
        type="text"
        placeholder="Name (e.g. Audiobooks)"
        bind:value={libName}
        class="field-input"
      />
      <div class="flex flex-wrap gap-2">
        <button
          type="button"
          class="btn text-xs ring-1 ring-border {libBackend === 'local'
            ? 'bg-primary text-primary-fg'
            : 'btn-ghost'}"
          onclick={() => (libBackend = "local")}
        >
          Local
        </button>
        <button
          type="button"
          class="btn text-xs ring-1 ring-border {libBackend === 's3'
            ? 'bg-primary text-primary-fg'
            : 'btn-ghost'}"
          onclick={() => (libBackend = "s3")}
        >
          S3
        </button>
      </div>
      {#if libBackend === "local"}
        <div class="flex gap-2">
          <input
            type="text"
            placeholder="Mount path (e.g. /mnt/media/books)"
            bind:value={libPath}
            class="field-input font-mono flex-1"
          />
          <Button
            type="button"
            variant="ghost"
            class="ring-1 ring-border"
            onclick={() => openBrowse("add")}
          >
            <FolderOpen size={14} />
          </Button>
        </div>
      {:else}
        <div class="grid gap-2 sm:grid-cols-2">
          <input
            type="text"
            placeholder="Endpoint (minio:9000)"
            bind:value={libS3.endpoint}
            class="field-input font-mono sm:col-span-2"
          />
          <input type="text" placeholder="Bucket" bind:value={libS3.bucket} class="field-input" />
          <input
            type="text"
            placeholder="Prefix (optional)"
            bind:value={libS3.prefix}
            class="field-input font-mono"
          />
          <input type="text" placeholder="Region" bind:value={libS3.region} class="field-input" />
          <input
            type="text"
            placeholder="Access key"
            bind:value={libS3.accessKey}
            class="field-input font-mono"
          />
          <input
            type="password"
            placeholder="Secret key"
            bind:value={libS3.secretKey}
            class="field-input font-mono sm:col-span-2"
          />
          <label class="flex items-center gap-2 text-xs text-muted">
            <input type="checkbox" bind:checked={libS3.usePathStyle} />
            Path-style
          </label>
          <label class="flex items-center gap-2 text-xs text-muted">
            <input type="checkbox" bind:checked={libS3.tls} />
            TLS
          </label>
        </div>
        <Button
          type="button"
          size="sm"
          variant="ghost"
          class="ring-1 ring-border"
          loading={testingS3}
          onclick={() => testS3Connection(libS3)}
        >
          Test connection
        </Button>
      {/if}
      <div class="pt-1">
        <Button type="submit" loading={libCreating}>
          <Plus size={16} /> Add mount
        </Button>
      </div>
      {#if libMsg}<p class="text-xs text-muted">{libMsg}</p>{/if}
      {#if libraries.error}<p class="text-xs text-danger">{libraries.error}</p>{/if}
    </form>

    <div class="mt-4 flex flex-wrap gap-2">
      <Button loading={scanningAll} onclick={scanAll}>Rescan all libraries</Button>
    </div>

    <div class="mt-4">
      <UploadPanel />
    </div>
  </div>

  <div class="rounded-[var(--radius-card)] border border-border bg-surface p-5">
    <h2 class="text-sm font-semibold text-fg">Sidebar layout</h2>
    <p class="mt-1 text-sm text-muted">Show, hide, and reorder sidebar sections.</p>
    <ul class="mt-3 space-y-2">
      {#each sidebarPrefs.order as section (section)}
        <li
          class="flex items-center justify-between gap-2 rounded-lg border border-border px-3 py-2"
        >
          <label class="flex items-center gap-2 text-sm text-fg">
            <input
              type="checkbox"
              checked={!sidebarPrefs.isHidden(section)}
              onchange={() => sidebarPrefs.toggleSection(section)}
            />
            {SIDEBAR_SECTION_LABELS[section as SidebarSectionId]}
          </label>
          <div class="flex gap-1">
            <button
              type="button"
              class="btn btn-ghost text-xs ring-1 ring-border"
              aria-label="Move section up"
              onclick={() => sidebarPrefs.moveSection(section, -1)}
            >
              <ChevronUp size={14} />
            </button>
            <button
              type="button"
              class="btn btn-ghost text-xs ring-1 ring-border"
              aria-label="Move section down"
              onclick={() => sidebarPrefs.moveSection(section, 1)}
            >
              <ChevronDown size={14} />
            </button>
          </div>
        </li>
      {/each}
    </ul>
    <button
      type="button"
      class="btn btn-ghost mt-3 text-xs ring-1 ring-border"
      onclick={() => sidebarPrefs.reset()}
    >
      Reset sidebar
    </button>
  </div>

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
</div>
