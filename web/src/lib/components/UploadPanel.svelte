<script lang="ts">
  import { Upload } from "@lucide/svelte";
  import { libraries } from "$lib/stores/libraries.svelte";
  import { uploads } from "$lib/stores/uploads.svelte";
  import { formatBytes } from "$lib/utils/format";

  let libraryId = $state(1);
  let relPath = $state("");
  let dragOver = $state(false);

  $effect(() => {
    if (libraries.items.length > 0 && !libraries.items.some((l) => l.id === libraryId)) {
      libraryId = libraries.items[0].id;
    }
  });

  function pickFiles(files: FileList | null) {
    if (!files?.length) return;
    for (const file of files) {
      const path = relPath.trim() || file.name;
      uploads.enqueue(libraryId, file, path);
      relPath = "";
    }
  }

  function onDrop(event: DragEvent) {
    event.preventDefault();
    dragOver = false;
    pickFiles(event.dataTransfer?.files ?? null);
  }
</script>

<section class="rounded-[var(--radius-card)] border border-border bg-surface p-5">
  <h2 class="text-sm font-semibold text-fg">Upload books</h2>
  <p class="mt-1 text-sm text-muted">
    Resumable chunked uploads into a library mount. Supported: EPUB, PDF, and audio formats.
  </p>

  <div class="mt-3 flex flex-wrap gap-2">
    <label class="block text-xs text-muted">
      Library
      <select class="input mt-1 min-w-[10rem]" bind:value={libraryId}>
        {#each libraries.items as lib (lib.id)}
          <option value={lib.id}>{lib.name}</option>
        {/each}
      </select>
    </label>
    <label class="block min-w-[12rem] flex-1 text-xs text-muted">
      Relative path (optional)
      <input
        type="text"
        class="input mt-1 w-full font-mono"
        placeholder="subdir/book.epub"
        bind:value={relPath}
      />
    </label>
  </div>

  <div
    class="upload-drop"
    class:upload-drop--active={dragOver}
    role="button"
    tabindex="0"
    ondragover={(e) => {
      e.preventDefault();
      dragOver = true;
    }}
    ondragleave={() => (dragOver = false)}
    ondrop={onDrop}
  >
    <Upload size={22} class="text-muted" />
    <p class="text-sm text-fg">Drop files here or choose from disk</p>
    <label class="btn btn-ghost mt-2 text-xs ring-1 ring-border cursor-pointer">
      Browse
      <input
        type="file"
        multiple
        class="sr-only"
        onchange={(e) => pickFiles(e.currentTarget.files)}
      />
    </label>
  </div>

  {#if uploads.jobs.length > 0}
    <ul class="mt-4 space-y-2">
      {#each uploads.jobs as job (job.id)}
        <li class="rounded-lg border border-border px-3 py-2">
          <div class="flex flex-wrap items-center justify-between gap-2">
            <div class="min-w-0">
              <p class="truncate text-sm font-medium text-fg">{job.relPath}</p>
              <p class="text-xs text-muted">{formatBytes(job.file.size)}</p>
            </div>
            <div class="flex items-center gap-2">
              {#if job.status === "done" && job.bookId}
                <a href={`/book/${job.bookId}`} class="text-xs text-primary">View</a>
              {/if}
              {#if job.status !== "uploading"}
                <button
                  type="button"
                  class="text-xs text-muted"
                  onclick={() => uploads.remove(job.id)}
                >
                  Dismiss
                </button>
              {/if}
            </div>
          </div>
          <div class="mt-2 h-1.5 overflow-hidden rounded-full bg-bg">
            <div
              class="h-full bg-primary transition-all"
              class:!bg-danger={job.status === "error"}
              style:width={`${Math.round(job.progress * 100)}%`}
            ></div>
          </div>
          <p class="mt-1 text-xs text-subtle">
            {#if job.status === "queued"}Queued{/if}
            {#if job.status === "uploading"}{Math.round(job.progress * 100)}%{/if}
            {#if job.status === "done"}Complete{/if}
            {#if job.status === "error"}{job.error}{/if}
          </p>
        </li>
      {/each}
    </ul>
  {/if}
</section>

<style>
  .upload-drop {
    margin-top: 0.75rem;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 0.25rem;
    padding: 1.5rem;
    border: 1px dashed var(--color-border);
    border-radius: var(--radius-card);
    text-align: center;
    transition:
      border-color 120ms ease,
      background-color 120ms ease;
  }

  .upload-drop--active {
    border-color: var(--color-primary);
    background: color-mix(in oklch, var(--color-primary) 8%, transparent);
  }
</style>
