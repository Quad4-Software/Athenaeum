<script lang="ts">
  import {
    BookOpen,
    CloudOff,
    Download,
    FileOutput,
    Headphones,
    Link2,
    MoreVertical,
    Pencil,
    Plus,
    ScanSearch,
    Star,
    Tag,
    Trash2,
    X,
  } from "@lucide/svelte";
  import Cover from "$lib/components/Cover.svelte";
  import BookCoverProgress from "$lib/components/BookCoverProgress.svelte";
  import ContextMenu from "$lib/components/ContextMenu.svelte";
  import BookMetadataEditor from "$lib/components/BookMetadataEditor.svelte";
  import HtmlDescription from "$lib/components/HtmlDescription.svelte";
  import Button from "$lib/components/Button.svelte";
  import Skeleton from "$lib/components/Skeleton.svelte";
  import ErrorView from "$lib/views/ErrorView.svelte";
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
  import { can } from "$lib/permissions";
  import { formatBytes, seriesLabel } from "$lib/utils/format";
  import { descriptionLooksLikeHtml } from "$lib/utils/sanitize-html";
  import { parseAudioLocation } from "$lib/audio/progress";
  import { bookOfflineCache, type BookOfflineStatus } from "$lib/offline/book-cache";
  import {
    isAudioFormat,
    isComicFormat,
    isMobiFormat,
    type Book,
    type Progress,
  } from "$lib/api/types";
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
    try {
      await api.addToCollection(collectionId, book.id);
      toast.success("Added to list");
      void collections.refresh();
    } catch (e) {
      toast.error(e instanceof ApiError ? e.message : "Failed to add");
    }
  }

  async function toggleFavorite() {
    if (!book) return;
    await favorites.toggle(book.id);
  }

  async function addTag() {
    if (!book) return;
    const name = tagInput.trim();
    if (!name) return;
    addingTag = true;
    try {
      const tags = await api.addBookTag(book.id, name);
      book = { ...book, tags };
      tagInput = "";
    } catch (e) {
      toast.error(e instanceof ApiError ? e.message : "Failed to add tag");
    } finally {
      addingTag = false;
    }
  }

  async function removeTag(name: string) {
    if (!book) return;
    const remaining = (book.tags ?? []).filter((t) => t !== name);
    try {
      const tags = await api.setBookTags(book.id, remaining);
      book = { ...book, tags };
    } catch (e) {
      toast.error(e instanceof ApiError ? e.message : "Failed to remove tag");
    }
  }

  function filterByTag(name: string) {
    library.setTag(name);
    router.navigate("/");
  }

  async function setRating(value: number) {
    if (!book) return;
    const next = book.userRating === value ? 0 : value;
    try {
      const rating = await api.setRating(book.id, next);
      book = { ...book, userRating: rating.rating };
    } catch (e) {
      toast.error(e instanceof ApiError ? e.message : "Failed to save rating");
    }
  }

  async function createShare() {
    if (!book) return;
    shareBusy = true;
    shareUrl = "";
    try {
      const link = await api.createShareLink(book.id, 168);
      shareUrl = link.url;
      try {
        await navigator.clipboard.writeText(link.url);
        toast.success(i18n.t("book.shareCopied"));
      } catch {
        toast.success(i18n.t("book.shareCreated"));
      }
    } catch (e) {
      toast.error(e instanceof ApiError ? e.message : i18n.t("book.shareFailed"));
    } finally {
      shareBusy = false;
    }
  }

  async function copyShareUrl() {
    if (!shareUrl) return;
    try {
      await navigator.clipboard.writeText(shareUrl);
      toast.success(i18n.t("book.shareCopied"));
    } catch {
      toast.info(i18n.t("book.shareCopyManual"));
    }
  }

  async function toggleOffline() {
    if (!book || isAudioFormat(book.format)) return;
    offlineBusy = true;
    try {
      if (offlineStatus.complete) {
        await bookOfflineCache.clear(book.id);
        await api.removeOffline([book.id]);
        toast.info(i18n.t("book.offlineCleared"));
      } else {
        await api.addOffline([book.id]);
        const url = api.fileUrl(book.id);
        if (isComicFormat(book.format)) {
          const manifest = await api.getComicManifest(book.id);
          bookOfflineCache.startDownload(
            book.id,
            url,
            book.fileSize,
            book.modifiedAt,
            "application/octet-stream",
            {
              total: manifest.total,
              pageUrl: (page) => api.comicPageUrl(book!.id, page),
            },
          );
        } else {
          bookOfflineCache.startDownload(book.id, url, book.fileSize, book.modifiedAt);
        }
        toast.success(i18n.t("book.offlineStarted"));
      }
    } catch (e) {
      toast.error(e instanceof ApiError ? e.message : i18n.t("book.offlineFailed"));
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
    try {
      await api.deleteBook(book.id);
      toast.success(i18n.t("book.deleted"));
      void library.refresh();
      router.navigate("/");
    } catch (e) {
      toast.error(e instanceof ApiError ? e.message : i18n.t("book.deleteFailed"));
    }
  }

  async function convertBook(target: "epub" | "pdf") {
    if (!book) return;
    closeMenu();
    try {
      const res = await toast.promise(() => api.convertBook(book!.id, target), {
        loading: `Converting to ${target}…`,
        success: (r) => r.message || `Converted to ${target}`,
        error: (e) => (e instanceof ApiError ? e.message : "Conversion failed"),
      });
      if (res.bookId) {
        void library.refresh();
        router.navigate(`/book/${res.bookId}`);
      }
    } catch {
      // toast.promise already surfaced the error
    }
  }

  const menuItems = $derived<MenuItem[]>(
    book
      ? [
          ...(can("edit_metadata")
            ? [
                {
                  id: "edit",
                  label: i18n.t("book.editMetadata"),
                  icon: Pencil,
                  onclick: () => openEditor("edit"),
                },
                {
                  id: "identify",
                  label: i18n.t("book.identify"),
                  icon: ScanSearch,
                  onclick: () => openEditor("identify"),
                },
                ...(isMobiFormat(book.format) || book.format === "pdf"
                  ? [
                      {
                        id: "convert-epub",
                        label: i18n.t("book.convertEpub"),
                        icon: FileOutput,
                        onclick: () => void convertBook("epub"),
                      },
                    ]
                  : []),
              ]
            : []),
          {
            id: "favorite",
            label: isFavorite ? i18n.t("book.removeFavorite") : i18n.t("book.addFavorite"),
            icon: Star,
            active: isFavorite,
            onclick: () => {
              closeMenu();
              void toggleFavorite();
            },
          },
          ...(can("delete_books")
            ? [
                {
                  id: "delete",
                  label: i18n.t("book.delete"),
                  icon: Trash2,
                  danger: true,
                  separator: true,
                  onclick: () => void deleteBook(),
                },
              ]
            : []),
        ]
      : [],
  );

  function resumeLabel(): string | null {
    if (!book || !progress || progress.percent <= 0) return null;
    if (isAudioFormat(book.format)) {
      const loc = parseAudioLocation(progress.location);
      if (loc.seconds <= 0 && loc.trackIndex <= 0) return null;
      const m = Math.floor(loc.seconds / 60);
      const time = `${m}:${String(Math.floor(loc.seconds % 60)).padStart(2, "0")}`;
      if (loc.trackIndex > 0) {
        return i18n.t("book.resumeTrackAt", { track: String(loc.trackIndex + 1), time });
      }
      return i18n.t("book.resumeAt", { time });
    }
    if (book.format === "pdf") {
      const page = Number(progress.location) || 0;
      if (page > 1) return i18n.t("book.resumePage", { page });
      return null;
    }
    return null;
  }
</script>

<section
  class="mx-auto w-full max-w-7xl px-3 py-5 sm:px-6"
  aria-label="Book details"
  oncontextmenu={openMenu}
>
  {#if errorCode}
    <ErrorView code={errorCode} message={error ?? undefined} compact />
  {:else if book}
    <div class="flex flex-col gap-8 sm:flex-row">
      <div class="mx-auto w-40 shrink-0 sm:mx-0 sm:w-52">
        <div class="relative overflow-hidden rounded-[var(--radius-card)] ring-1 ring-border">
          <Cover {book} />
          <BookCoverProgress percent={progressPercent} />
        </div>
      </div>

      <div class="min-w-0 flex-1">
        <span
          class="inline-block rounded-full bg-surface px-2.5 py-0.5 text-xs font-medium uppercase tracking-wide text-muted"
        >
          {book.format}
        </span>
        <h1 class="mt-3 text-2xl font-bold text-fg sm:text-3xl">{book.title}</h1>
        {#if book.author}<p class="mt-1 text-muted">{book.author}</p>{/if}
        {#if book.series}
          <p class="mt-1 text-sm text-subtle">{seriesLabel(book.series, book.seriesIndex)}</p>
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

        <div class="book-actions mt-6 flex flex-wrap gap-2 sm:gap-3">
          <Button
            class="min-h-11 flex-1 sm:flex-none"
            onclick={() => router.navigate(`/read/${book?.id}`)}
          >
            {#if isAudioFormat(book.format)}
              <Headphones size={16} /> {i18n.t("book.listen")}
            {:else if resumeLabel()}
              <BookOpen size={16} /> {i18n.t("book.resume")}
            {:else}
              <BookOpen size={16} /> {i18n.t("book.read")}
            {/if}
          </Button>
          <Button variant="ghost" class="min-h-11 ring-1 ring-border" onclick={toggleFavorite}>
            <Star size={16} fill={isFavorite ? "currentColor" : "none"} />
            {isFavorite ? i18n.t("book.favorited") : i18n.t("book.favorite")}
          </Button>
          <a
            class="btn btn-ghost min-h-11 ring-1 ring-border"
            href={api.downloadUrl(book.id)}
            download
          >
            <Download size={16} />
            {i18n.t("book.download")}
          </a>
          {#if !isAudioFormat(book.format)}
            <Button
              variant="ghost"
              class="min-h-11 ring-1 ring-border"
              loading={offlineBusy || offlineStatus.downloading}
              onclick={() => void toggleOffline()}
            >
              <CloudOff size={16} />
              {#if offlineStatus.complete}
                {i18n.t("book.offlineReady")}
              {:else if offlineStatus.downloading}
                {i18n.t("book.offlineCaching", {
                  pct: String(
                    Math.round(
                      (offlineStatus.cachedBytes / Math.max(offlineStatus.totalBytes, 1)) * 100,
                    ),
                  ),
                })}
              {:else}
                {i18n.t("book.saveOffline")}
              {/if}
            </Button>
          {/if}
          <Button
            variant="ghost"
            class="min-h-11 ring-1 ring-border"
            loading={shareBusy}
            onclick={() => void createShare()}
          >
            <Link2 size={16} />
            {i18n.t("book.share")}
          </Button>
          {#if menuItems.length > 0}
            <button
              type="button"
              class="btn btn-ghost min-h-11 min-w-11 ring-1 ring-border"
              aria-label={i18n.t("book.moreActions")}
              aria-haspopup="menu"
              aria-expanded={menuOpen}
              onclick={openMenuFromButton}
            >
              <MoreVertical size={18} />
            </button>
          {/if}
        </div>

        {#if shareUrl}
          <div
            class="mt-3 flex flex-wrap items-center gap-2 rounded-lg border border-border bg-surface px-3 py-2"
          >
            <input
              class="field-input min-w-0 flex-1 text-xs"
              type="text"
              readonly
              value={shareUrl}
              onclick={(e) => {
                const el = e.currentTarget;
                if (el instanceof HTMLInputElement) el.select();
              }}
            />
            <Button
              variant="ghost"
              size="sm"
              class="ring-1 ring-border"
              onclick={() => void copyShareUrl()}
            >
              {i18n.t("book.copyLink")}
            </Button>
            <button
              type="button"
              class="btn btn-ghost text-xs"
              aria-label={i18n.t("book.dismissShare")}
              onclick={() => (shareUrl = "")}
            >
              <X size={14} />
            </button>
          </div>
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

        <div class="mt-4">
          <p class="mb-1.5 flex items-center gap-1.5 text-sm text-muted">
            <Tag size={14} />
            {i18n.t("book.tags")}
          </p>
          <div class="flex flex-wrap items-center gap-2">
            {#each book.tags ?? [] as tagName (tagName)}
              <span
                class="inline-flex items-center gap-1 rounded-full border border-border bg-surface px-2.5 py-1 text-xs text-fg"
              >
                <button type="button" class="hover:underline" onclick={() => filterByTag(tagName)}>
                  {tagName}
                </button>
                <button
                  type="button"
                  class="text-subtle hover:text-danger"
                  aria-label={i18n.t("book.removeTag", { name: tagName })}
                  onclick={() => void removeTag(tagName)}
                >
                  <X size={12} />
                </button>
              </span>
            {/each}
            <form
              class="inline-flex items-center gap-1"
              onsubmit={(e) => {
                e.preventDefault();
                void addTag();
              }}
            >
              <input
                type="text"
                class="field-input h-8 w-28 text-xs"
                placeholder={i18n.t("book.addTagPlaceholder")}
                bind:value={tagInput}
                disabled={addingTag}
              />
              <button
                type="submit"
                class="btn btn-ghost min-h-8 px-2 text-xs ring-1 ring-border"
                disabled={addingTag || !tagInput.trim()}
              >
                <Plus size={14} />
              </button>
            </form>
          </div>
        </div>

        {#if collections.shelfItems().filter((c) => c.kind === "manual").length > 0}
          <div class="mt-4">
            <p class="text-sm text-muted">{i18n.t("book.addToShelf")}</p>
            <div class="mt-1 flex flex-wrap gap-2">
              {#each collections.shelfItems().filter((c) => c.kind === "manual") as c (c.id)}
                <button
                  type="button"
                  class="btn btn-ghost min-h-10 ring-1 ring-border text-xs"
                  onclick={() => addToCollection(c.id)}
                >
                  <Plus size={14} />
                  {c.name}
                </button>
              {/each}
            </div>
          </div>
        {/if}

        {#if collections.readingItems().length > 0}
          <div class="mt-4">
            <p class="text-sm text-muted">{i18n.t("book.addToReadingList")}</p>
            <div class="mt-1 flex flex-wrap gap-2">
              {#each collections.readingItems() as c (c.id)}
                <button
                  type="button"
                  class="btn btn-ghost min-h-10 ring-1 ring-border text-xs"
                  onclick={() => addToCollection(c.id)}
                >
                  <Plus size={14} />
                  {c.name}
                </button>
              {/each}
            </div>
          </div>
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
    <div class="flex flex-col gap-8 sm:flex-row">
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
  .book-actions {
    position: sticky;
    bottom: calc(var(--bottom-chrome) + 0.5rem);
    z-index: 5;
    margin-inline: -0.25rem;
    padding: 0.5rem 0.25rem;
    border-radius: 0.75rem;
    background: color-mix(in oklch, var(--color-bg) 88%, transparent);
    backdrop-filter: blur(10px);
  }

  @media (min-width: 640px) {
    .book-actions {
      position: static;
      margin-inline: 0;
      padding: 0;
      background: transparent;
      backdrop-filter: none;
    }
  }
</style>
