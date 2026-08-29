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
  <h1 class="font-display text-3xl font-semibold tracking-tight text-fg">{i18n.t("collections.title")}</h1>
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
    <ul class="mt-6 grid grid-cols-2 gap-3 sm:grid-cols-3 sm:gap-4 lg:grid-cols-4">
      {#each collections.items as c (c.id)}
        <li class="group relative">
          <button
            type="button"
            class="flex h-full w-full flex-col overflow-hidden rounded-[var(--radius-card)] border border-border bg-bg-elevated/50 text-left shadow-sm transition hover:-translate-y-0.5 hover:border-border-strong hover:shadow-md"
            onclick={() => open(c.id)}
          >
            <div
              class="flex aspect-[4/3] items-end bg-gradient-to-br from-surface-hover to-bg p-3"
            >
              <span
                class="rounded-full bg-bg/80 px-2 py-0.5 text-[10px] font-medium uppercase tracking-wide text-subtle ring-1 ring-border"
              >
                {kindLabel(c.kind)}
              </span>
            </div>
            <div class="space-y-1 p-3">
              <p class="font-display truncate text-base font-semibold text-fg">{c.name}</p>
              <p class="text-xs text-muted">
                {i18n.t("collections.bookCount", { count: c.bookCount })}
              </p>
            </div>
          </button>
          {#if c.kind !== "auto"}
            <button
              type="button"
              class="btn btn-ghost absolute right-2 top-2 min-h-9 min-w-9 bg-bg/80 text-danger opacity-0 shadow-sm ring-1 ring-border backdrop-blur group-hover:opacity-100 group-focus-within:opacity-100"
              aria-label={i18n.t("collections.deleteAria", { name: c.name })}
              onclick={(e) => {
                e.stopPropagation();
                void removeCollection(c.id, c.name);
              }}
            >
              <Trash2 size={16} />
            </button>
          {/if}
        </li>
      {/each}
    </ul>
  {/if}
</section>
