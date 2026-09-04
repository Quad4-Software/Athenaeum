<script lang="ts">
  import { ChevronDown, ChevronUp, FolderOpen, Pencil, Plus, Trash2 } from "@lucide/svelte";
  import FolderBrowser from "$lib/components/FolderBrowser.svelte";
  import Button from "$lib/components/Button.svelte";
  import Skeleton from "$lib/components/Skeleton.svelte";
  import { library } from "$lib/stores/library.svelte";
  import { libraries } from "$lib/stores/libraries.svelte";
  import { toast } from "$lib/stores/toast.svelte";
  import { ApiError } from "$lib/api/client";
  import type { LibraryS3Input } from "$lib/api/types";
  import LibraryUploadsSection from "./LibraryUploadsSection.svelte";

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

<div class="rounded-[var(--radius-card)] border border-border bg-surface p-5">
  <h2 class="text-sm font-semibold text-fg">Library mounts</h2>
  <p class="mt-1 text-sm text-muted">
    Add local folders or MinIO-compatible S3 buckets. Each mount is scanned independently. S3 mounts
    use periodic or manual scan (no filesystem watch).
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
                Removes this mount and all {lib.bookCount} indexed books from the catalog. Files on disk
                are not deleted.
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
    <LibraryUploadsSection />
  </div>
</div>
