<script lang="ts">
  import { Library, Plus, RefreshCw, Trash2 } from "@lucide/svelte";
  import Button from "$lib/components/Button.svelte";
  import EmptyState from "$lib/components/EmptyState.svelte";
  import Skeleton from "$lib/components/Skeleton.svelte";
  import { router } from "$lib/router.svelte";
  import { collections } from "$lib/stores/collections.svelte";
  import { library } from "$lib/stores/library.svelte";
  import { confirmDialog } from "$lib/stores/confirm.svelte";
  import { i18n } from "$lib/stores/i18n.svelte";
  import type { BookFormat, SmartQuery } from "$lib/api/types";

  let name = $state("");
  let readingName = $state("");
  let smartName = $state("");
  let smartFormat = $state<BookFormat | "">("");
  let smartAuthor = $state("");
  let smartAddedDays = $state("");
  let creating = $state(false);
  let creatingReading = $state(false);
  let creatingSmart = $state(false);

  $effect(() => {
    void collections.refresh();
  });

  function kindLabel(kind: string): string {
    if (kind === "auto") return i18n.t("collections.kindAuto");
    if (kind === "smart") return i18n.t("collections.kindSmart");
    if (kind === "reading") return i18n.t("collections.kindReading");
    return i18n.t("collections.kindManual");
  }

  async function create(event: Event) {
    event.preventDefault();
    if (!name.trim()) return;
    creating = true;
    try {
      await collections.create(name.trim());
      name = "";
    } finally {
      creating = false;
    }
  }

  async function createReading(event: Event) {
    event.preventDefault();
    if (!readingName.trim()) return;
    creatingReading = true;
    try {
      await collections.createReadingList(readingName.trim());
      readingName = "";
    } finally {
      creatingReading = false;
    }
  }

  async function createSmart(event: Event) {
    event.preventDefault();
    if (!smartName.trim()) return;
    const query: SmartQuery = {};
    if (smartFormat) query.format = smartFormat;
    if (smartAuthor.trim()) query.author = smartAuthor.trim();
    const days = parseInt(smartAddedDays, 10);
    if (days > 0) query.addedDays = days;
    if (!query.format && !query.author && !query.addedDays) return;

    creatingSmart = true;
    try {
      await collections.createSmart(smartName.trim(), query);
      smartName = "";
      smartFormat = "";
      smartAuthor = "";
      smartAddedDays = "";
    } finally {
      creatingSmart = false;
    }
  }

  async function removeCollection(id: number, collectionName: string) {
    const ok = await confirmDialog.ask({
      title: i18n.t("collections.deleteTitle"),
      message: i18n.t("collections.deleteConfirm", { name: collectionName }),
      confirmLabel: i18n.t("confirm.delete"),
      cancelLabel: i18n.t("confirm.cancel"),
      danger: true,
    });
    if (!ok) return;
    await collections.remove(id);
  }

  function open(id: number) {
    router.navigate(`/collections/${id}`);
  }
</script>

<section class="mx-auto w-full max-w-7xl px-3 py-5 sm:px-6">
  <h1 class="text-2xl font-bold text-fg">{i18n.t("collections.title")}</h1>
  <p class="mt-1 text-sm text-muted">{i18n.t("collections.subtitle")}</p>

  <form class="mt-6 flex gap-2" onsubmit={create}>
    <input
      type="text"
      placeholder={i18n.t("collections.manualPlaceholder")}
      bind:value={name}
      class="input flex-1"
    />
    <Button type="submit" loading={creating}>
      <Plus size={16} />
      {i18n.t("collections.add")}
    </Button>
  </form>

  <form class="mt-4 flex gap-2" onsubmit={createReading}>
    <input
      type="text"
      placeholder={i18n.t("collections.readingPlaceholder")}
      bind:value={readingName}
      class="input flex-1"
    />
    <Button type="submit" variant="ghost" class="ring-1 ring-border" loading={creatingReading}>
      <Plus size={16} />
      {i18n.t("collections.readingList")}
    </Button>
  </form>

  <form
    class="mt-4 space-y-3 rounded-[var(--radius-card)] border border-border bg-bg-elevated/40 p-4"
    onsubmit={createSmart}
  >
    <p class="text-sm font-medium text-fg">{i18n.t("collections.smartTitle")}</p>
    <input
      type="text"
      placeholder={i18n.t("collections.namePlaceholder")}
      bind:value={smartName}
      class="input w-full"
    />
    <div class="grid gap-2 sm:grid-cols-3">
      <select bind:value={smartFormat} class="input">
        <option value="">{i18n.t("collections.anyFormat")}</option>
        <option value="epub">EPUB</option>
        <option value="pdf">PDF</option>
        <option value="mp3">MP3</option>
        <option value="m4b">M4B</option>
      </select>
      <input
        type="text"
        placeholder={i18n.t("collections.authorPlaceholder")}
        bind:value={smartAuthor}
        class="input"
      />
      <input
        type="number"
        min="1"
        placeholder={i18n.t("collections.addedDaysPlaceholder")}
        bind:value={smartAddedDays}
        class="input"
      />
    </div>
    <Button type="submit" variant="ghost" class="ring-1 ring-border" loading={creatingSmart}>
      {i18n.t("collections.createSmart")}
    </Button>
  </form>

  {#if collections.error}
    <p class="mt-4 text-sm text-danger">{collections.error}</p>
  {:else if collections.loading}
    <div class="mt-6 space-y-2">
      {#each Array(5) as _, i (i)}
        <Skeleton height="3.5rem" rounded="lg" />
      {/each}
    </div>
  {:else if collections.items.length === 0}
    <EmptyState
      class="mt-8 rounded-[var(--radius-card)] border border-dashed border-border px-4"
      title={i18n.t("collections.emptyTitle")}
      body={i18n.t("collections.emptyBody")}
    >
      {#snippet icon(size)}
        <Library {size} />
      {/snippet}
      <div class="flex flex-wrap justify-center gap-2">
        <Button onclick={() => library.triggerScan()}>
          <RefreshCw size={16} />
          {i18n.t("collections.scanLibrary")}
        </Button>
      </div>
    </EmptyState>
  {:else}
    <ul
      class="mt-6 divide-y divide-border overflow-hidden rounded-[var(--radius-card)] border border-border"
    >
      {#each collections.items as c (c.id)}
        <li
          class="flex items-center justify-between gap-3 bg-bg-elevated/30 px-4 py-3 transition-colors hover:bg-surface/60"
        >
          <button type="button" class="min-w-0 flex-1 text-left" onclick={() => open(c.id)}>
            <div class="flex items-center gap-2">
              <p class="font-medium text-fg">{c.name}</p>
              <span
                class="rounded-full bg-surface px-2 py-0.5 text-[10px] font-medium uppercase tracking-wide text-subtle ring-1 ring-border"
              >
                {kindLabel(c.kind)}
              </span>
            </div>
            <p class="text-xs text-muted">
              {i18n.t("collections.bookCount", { count: c.bookCount })}
            </p>
          </button>
          {#if c.kind !== "auto"}
            <button
              type="button"
              class="btn btn-ghost min-h-10 min-w-10 text-danger"
              aria-label={i18n.t("collections.deleteAria", { name: c.name })}
              onclick={() => void removeCollection(c.id, c.name)}
            >
              <Trash2 size={16} />
            </button>
          {/if}
        </li>
      {/each}
    </ul>
  {/if}
</section>
