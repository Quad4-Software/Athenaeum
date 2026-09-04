<script lang="ts">
  import { Star } from "@lucide/svelte";
  import Cover from "$lib/components/Cover.svelte";
  import BookCoverProgress from "$lib/components/BookCoverProgress.svelte";
  import ContextMenu from "$lib/components/ContextMenu.svelte";
  import BookMetadataEditor from "$lib/components/BookMetadataEditor.svelte";
  import HtmlDescription from "$lib/components/HtmlDescription.svelte";
  import Skeleton from "$lib/components/Skeleton.svelte";
  import ErrorView from "$lib/views/ErrorView.svelte";
  import BookActionBar from "$lib/views/book/BookActionBar.svelte";
  import BookShareLink from "$lib/views/book/BookShareLink.svelte";
  import BookTagsSection from "$lib/views/book/BookTagsSection.svelte";
  import BookCollectionsQuickAdd from "$lib/views/book/BookCollectionsQuickAdd.svelte";
  import BookCitationSection from "$lib/views/book/BookCitationSection.svelte";
  import * as bookViewActions from "$lib/views/book/book-view-actions";
  import { api, ApiError } from "$lib/api/client";
  import { router } from "$lib/router.svelte";
  import { collections } from "$lib/stores/collections.svelte";
  import { favorites } from "$lib/stores/favorites.svelte";
  import { library } from "$lib/stores/library.svelte";
  import { bookEditorIntent } from "$lib/stores/bookEditor.svelte";
  import { confirmDialog } from "$lib/stores/confirm.svelte";
  import { toast } from "$lib/stores/toast.svelte";
  import { ui } from "$lib/stores/ui.svelte";
  import { i18n } from "$lib/stores/i18n.svelte";
  import { rememberBook } from "$lib/commands/recent";
  import { formatBytes, seriesLabel } from "$lib/utils/format";
  import { descriptionLooksLikeHtml } from "$lib/utils/sanitize-html";
  import { bookOfflineCache, type BookOfflineStatus } from "$lib/offline/book-cache";
  import { isAudioFormat, type Book, type Progress } from "$lib/api/types";
  import type { MenuItem } from "$lib/components/MenuList.svelte";

  interface Props {
    id: number;
  }

  let { id }: Props = $props();

  let book = $state<Book | null>(null);
  let progress = $state<Progress | null>(null);
  let error = $state<string | null>(null);
  let errorCode = $state<number | null>(null);
  let loading = $state(true);
  let editorOpen = $state(false);
  let editorPanel = $state<"edit" | "identify">("edit");
  let menuOpen = $state(false);
  let menuX = $state(0);
  let menuY = $state(0);
  let bibtexBusy = $state(false);

  let isFavorite = $derived(book ? favorites.isFavorite(book.id) : false);
  let progressPercent = $derived(progress?.percent ?? 0);
  let tagInput = $state("");
  let addingTag = $state(false);
  let hoverRating = $state(0);
  let shareBusy = $state(false);
  let shareUrl = $state("");
  let offlineBusy = $state(false);
  let offlineStatus = $state<BookOfflineStatus>({
    cachedBytes: 0,
    totalBytes: 0,
    complete: false,
    downloading: false,
    error: null,
  });
  let unsubOffline: (() => void) | null = null;

  $effect(() => {
    const bookId = id;
    book = null;
    progress = null;
    error = null;
    errorCode = null;
    loading = true;
    editorOpen = false;
    editorPanel = "edit";
    ui.pageTitle = "";
    Promise.all([api.getBook(bookId), api.getProgress(bookId)])
      .then(([b, p]) => {
        book = b;
        progress = p;
        ui.pageTitle = b.title;
        rememberBook({
          id: b.id,
          title: b.title,
          author: b.author,
          hasCover: b.hasCover,
          modifiedAt: b.modifiedAt,
        });
        const intent = bookEditorIntent.consume(bookId);
        if (intent) {
          editorPanel = intent;
          editorOpen = true;
        }
      })
      .catch((e) => {
        if (e instanceof ApiError) {
          errorCode = e.status;
          error = e.message;
          if (e.status !== 401) toast.error(e.message);
        } else {
          error = "Failed to load book";
          toast.error(error);
        }
      })
      .finally(() => {
        loading = false;
      });
    unsubOffline?.();
    unsubOffline = bookOfflineCache.subscribe(bookId, (status) => {
      if (bookId === id) offlineStatus = status;
    });
    return () => {
      ui.pageTitle = "";
      unsubOffline?.();
      unsubOffline = null;
    };
  });

  async function addToCollection(collectionId: number) {
    if (!book) return;
    await bookViewActions.addToCollection(book.id, collectionId);
  }

  async function toggleFavorite() {
    if (!book) return;
    await bookViewActions.toggleFavorite(book.id);
  }

  async function addTag() {
    if (!book) return;
    const name = tagInput.trim();
    if (!name) return;
    addingTag = true;
    try {
      const tags = await bookViewActions.addTag(book.id, name);
      if (tags) {
        book = { ...book, tags };
        tagInput = "";
      }
    } finally {
      addingTag = false;
    }
  }

  async function removeTag(name: string) {
    if (!book) return;
    const remaining = (book.tags ?? []).filter((t) => t !== name);
    const tags = await bookViewActions.removeTag(book.id, remaining);
    if (tags) book = { ...book, tags };
  }

  function filterByTag(name: string) {
    bookViewActions.filterByTag(name);
  }

  async function setRating(value: number) {
    if (!book) return;
    const rating = await bookViewActions.setRating(book.id, book.userRating ?? 0, value);
    if (rating !== null) book = { ...book, userRating: rating };
  }

  async function createShare() {
    if (!book) return;
    shareBusy = true;
    shareUrl = "";
    try {
      shareUrl = (await bookViewActions.createShare(book.id)) ?? "";
    } finally {
      shareBusy = false;
    }
  }

  async function copyShareUrl() {
    await bookViewActions.copyShareUrl(shareUrl);
  }

  async function copyBibTeX() {
    if (!book || bibtexBusy) return;
    bibtexBusy = true;
    try {
      await bookViewActions.copyBibTeX(book.id);
    } finally {
      bibtexBusy = false;
    }
  }

  async function downloadBibTeX() {
    if (!book || bibtexBusy) return;
    bibtexBusy = true;
    try {
      await bookViewActions.downloadBibTeX(book);
    } finally {
      bibtexBusy = false;
    }
  }

  async function toggleOffline() {
    if (!book || isAudioFormat(book.format)) return;
    offlineBusy = true;
    try {
      await bookViewActions.toggleOffline(book, offlineStatus.complete);
    } finally {
      offlineBusy = false;
    }
  }

  function openMenu(event: MouseEvent) {
    event.preventDefault();
    menuX = event.clientX;
    menuY = event.clientY;
    menuOpen = true;
  }

  function openMenuFromButton(event: MouseEvent) {
    event.preventDefault();
    const btn = event.currentTarget as HTMLButtonElement;
    const rect = btn.getBoundingClientRect();
    menuX = rect.right;
    menuY = rect.bottom;
    menuOpen = true;
  }

  function closeMenu() {
    menuOpen = false;
  }

  function openEditor(panel: "edit" | "identify") {
    closeMenu();
    editorPanel = panel;
    editorOpen = true;
  }

  async function deleteBook() {
    if (!book) return;
    closeMenu();
    const ok = await confirmDialog.ask({
      title: i18n.t("book.deleteTitle"),
      message: i18n.t("book.deleteConfirm", { title: book.title }),
      confirmLabel: i18n.t("confirm.delete"),
      cancelLabel: i18n.t("confirm.cancel"),
      danger: true,
    });
    if (!ok) return;
    await bookViewActions.deleteBook(book);
  }

  async function convertBook(target: "epub" | "pdf") {
    if (!book) return;
    closeMenu();
    await bookViewActions.convertBook(book, target);
  }

  const menuItems = $derived<MenuItem[]>(
    book
      ? bookViewActions.buildMenuItems({
          book,
          isFavorite,
          onEdit: () => openEditor("edit"),
          onIdentify: () => openEditor("identify"),
          onConvertEpub: () => void convertBook("epub"),
          onCopyBibtex: () => {
            closeMenu();
            void copyBibTeX();
          },
          onDownloadBibtex: () => {
            closeMenu();
            void downloadBibTeX();
          },
          onToggleFavorite: () => {
            closeMenu();
            void toggleFavorite();
          },
          onDelete: () => void deleteBook(),
        })
      : [],
  );

  function resumeLabel(): string | null {
    if (!book) return null;
    return bookViewActions.resumeLabel(book, progress);
  }
</script>

<section
  class="book-details relative w-full overflow-hidden"
  aria-label="Book details"
  oncontextmenu={openMenu}
>
  {#if errorCode}
    <div class="relative mx-auto w-full max-w-7xl px-3 py-5 sm:px-6">
      <ErrorView code={errorCode} message={error ?? undefined} compact />
    </div>
  {:else if book}
    {#if book.hasCover}
      <div class="book-backdrop" aria-hidden="true">
        <img src={api.coverUrl(book.id, book.modifiedAt)} alt="" />
        <div class="book-backdrop-wash"></div>
      </div>
    {/if}
    <div
      class="relative z-[1] mx-auto flex w-full max-w-7xl flex-col gap-8 px-3 py-5 sm:flex-row sm:gap-10 sm:px-6"
    >
      <div class="mx-auto w-44 shrink-0 sm:mx-0 sm:w-60 lg:w-72">
        <div
          class="relative overflow-hidden rounded-[var(--radius-card)] shadow-panel ring-1 ring-border/50"
        >
          <Cover {book} />
          <BookCoverProgress percent={progressPercent} />
        </div>
      </div>

      <div class="min-w-0 flex-1">
        <span
          class="inline-block rounded-full bg-surface/80 px-2.5 py-0.5 text-xs font-medium uppercase tracking-wide text-muted ring-1 ring-border/60"
        >
          {book.format}
        </span>
        <h1 class="font-display mt-3 text-3xl font-semibold tracking-tight text-fg sm:text-4xl">
          {book.title}
        </h1>
        {#if book.author}<p class="mt-2 text-base text-muted">{book.author}</p>{/if}
        {#if book.series}
          <button
            type="button"
            class="mt-2 text-sm text-primary hover:underline"
            onclick={() => {
              if (!book?.series) return;
              library.setSeries(book.series);
              router.navigate("/");
            }}
          >
            {seriesLabel(book.series, book.seriesIndex)}
          </button>
        {/if}
        {#if book.duplicateOf}
          <p class="mt-2 text-sm text-muted">
            {i18n.t("book.duplicateOf")}
            <a
              href={`/book/${book.duplicateOf}`}
              class="text-primary underline-offset-2 hover:underline"
            >
              book #{book.duplicateOf}
            </a>
          </p>
        {/if}

        {#if resumeLabel()}
          <p class="mt-3 text-sm text-primary">{resumeLabel()}</p>
        {/if}

        <BookActionBar
          {book}
          {isFavorite}
          resumeText={resumeLabel()}
          {shareBusy}
          {offlineBusy}
          {offlineStatus}
          {menuOpen}
          hasMenu={menuItems.length > 0}
          ontogglefavorite={() => void toggleFavorite()}
          oncreateshare={() => void createShare()}
          ontoggleoffline={() => void toggleOffline()}
          onopenmenu={openMenuFromButton}
        />

        {#if shareUrl}
          <BookShareLink
            url={shareUrl}
            oncopy={() => void copyShareUrl()}
            ondismiss={() => (shareUrl = "")}
          />
        {/if}

        <div
          class="mt-4 flex items-center gap-1"
          role="radiogroup"
          aria-label={i18n.t("book.rating")}
        >
          {#each [1, 2, 3, 4, 5] as value (value)}
            <button
              type="button"
              class="rounded p-0.5 text-subtle transition-colors hover:text-primary"
              class:text-primary={(hoverRating || book.userRating || 0) >= value}
              role="radio"
              aria-checked={book.userRating === value}
              aria-label={i18n.t("book.rateStars", { count: value })}
              onmouseenter={() => (hoverRating = value)}
              onmouseleave={() => (hoverRating = 0)}
              onclick={() => void setRating(value)}
            >
              <Star
                size={18}
                fill={(hoverRating || book.userRating || 0) >= value ? "currentColor" : "none"}
              />
            </button>
          {/each}
        </div>

        <BookTagsSection
          tags={book.tags ?? []}
          bind:tagInput
          {addingTag}
          onfilter={filterByTag}
          onremove={(name) => void removeTag(name)}
          onadd={() => void addTag()}
        />

        <BookCollectionsQuickAdd
          label={i18n.t("book.addToShelf")}
          items={collections.shelfItems().filter((c) => c.kind === "manual")}
          onadd={addToCollection}
        />

        <BookCollectionsQuickAdd
          label={i18n.t("book.addToReadingList")}
          items={collections.readingItems()}
          onadd={addToCollection}
        />

        {#if bookViewActions.hasCitation(book)}
          <BookCitationSection
            {book}
            {bibtexBusy}
            oncopybibtex={() => void copyBibTeX()}
            ondownloadbibtex={() => void downloadBibTeX()}
          />
        {/if}

        {#if book.description}
          {#if descriptionLooksLikeHtml(book.description)}
            <HtmlDescription value={book.description} />
          {:else}
            <p class="mt-6 whitespace-pre-line text-sm leading-relaxed text-muted">
              {book.description}
            </p>
          {/if}
        {/if}

        <dl class="mt-6 grid grid-cols-2 gap-x-6 gap-y-2 text-sm sm:max-w-sm">
          {#if book.language}
            <dt class="text-subtle">{i18n.t("book.language")}</dt>
            <dd class="text-fg">{book.language}</dd>
          {/if}
          <dt class="text-subtle">{i18n.t("book.size")}</dt>
          <dd class="text-fg">{formatBytes(book.fileSize)}</dd>
          <dt class="text-subtle">{i18n.t("book.added")}</dt>
          <dd class="text-fg">{new Date(book.addedAt).toLocaleDateString()}</dd>
        </dl>

        <BookMetadataEditor
          {book}
          bind:open={editorOpen}
          bind:panel={editorPanel}
          onsaved={(b) => (book = b)}
        />
      </div>
    </div>
  {:else if loading}
    <div
      class="relative mx-auto flex w-full max-w-7xl flex-col gap-8 px-3 py-5 sm:flex-row sm:px-6"
    >
      <Skeleton rounded="card" width="14rem" height="21rem" class="mx-auto shrink-0 sm:mx-0" />
      <div class="flex-1 space-y-3">
        <Skeleton width="4rem" height="1.25rem" rounded="full" />
        <Skeleton width="70%" height="2rem" />
        <Skeleton width="40%" height="1rem" />
        <div class="flex gap-2 pt-4">
          <Skeleton width="6rem" height="2.25rem" rounded="lg" />
          <Skeleton width="6rem" height="2.25rem" rounded="lg" />
        </div>
      </div>
    </div>
  {/if}
</section>

<ContextMenu
  open={menuOpen}
  x={menuX}
  y={menuY}
  title={book?.title}
  items={menuItems}
  onclose={closeMenu}
/>

<style>
  .book-details {
    isolation: isolate;
  }

  .book-backdrop {
    pointer-events: none;
    position: absolute;
    inset: 0 0 auto;
    width: 100%;
    height: min(28rem, 55vh);
    z-index: 0;
    overflow: hidden;
  }

  .book-backdrop img {
    width: 100%;
    height: 100%;
    object-fit: cover;
    object-position: center top;
    filter: blur(56px) saturate(1.2);
    transform: scale(1.2);
    opacity: 0.55;
  }

  .book-backdrop-wash {
    position: absolute;
    inset: 0;
    background:
      linear-gradient(
        to bottom,
        color-mix(in oklch, var(--color-bg) 8%, transparent) 0%,
        color-mix(in oklch, var(--color-bg) 55%, transparent) 45%,
        var(--color-bg) 100%
      ),
      linear-gradient(
        to right,
        color-mix(in oklch, var(--color-bg) 35%, transparent),
        transparent 18%,
        transparent 82%,
        color-mix(in oklch, var(--color-bg) 35%, transparent)
      );
  }

  :global([data-theme="light"]) .book-backdrop img {
    opacity: 0.38;
    filter: blur(64px) saturate(1.05) brightness(1.08);
  }

  :global([data-theme="light"]) .book-backdrop-wash {
    background:
      linear-gradient(
        to bottom,
        color-mix(in oklch, var(--color-bg) 20%, transparent) 0%,
        color-mix(in oklch, var(--color-bg) 70%, transparent) 40%,
        var(--color-bg) 100%
      ),
      linear-gradient(
        to right,
        color-mix(in oklch, var(--color-bg) 45%, transparent),
        transparent 22%,
        transparent 78%,
        color-mix(in oklch, var(--color-bg) 45%, transparent)
      );
  }
</style>
