<script lang="ts">
  import Cover from "./Cover.svelte";
  import BookCoverProgress from "./BookCoverProgress.svelte";
  import ContextMenu from "./ContextMenu.svelte";
  import { link } from "$lib/router.svelte";
  import { api, ApiError } from "$lib/api/client";
  import { favorites } from "$lib/stores/favorites.svelte";
  import { library } from "$lib/stores/library.svelte";
  import { density } from "$lib/stores/density.svelte";
  import { bookEditorIntent } from "$lib/stores/bookEditor.svelte";
  import { confirmDialog } from "$lib/stores/confirm.svelte";
  import { toast } from "$lib/stores/toast.svelte";
  import { i18n } from "$lib/stores/i18n.svelte";
  import { can } from "$lib/permissions";
  import { Star, Check, MoreVertical, Pencil, ScanSearch, Trash2 } from "@lucide/svelte";
  import type { Book } from "$lib/api/types";
  import type { MenuItem } from "./MenuList.svelte";

  interface Props {
    book: Book;
    selectMode?: boolean;
    selected?: boolean;
    onToggleSelect?: () => void;
  }

  let { book, selectMode = false, selected = false, onToggleSelect }: Props = $props();

  let isFavorite = $derived(favorites.isFavorite(book.id));
  let progress = $derived(book.progressPercent ?? 0);
  let menuOpen = $state(false);
  let menuX = $state(0);
  let menuY = $state(0);

  async function toggleFavorite(event?: MouseEvent) {
    event?.preventDefault();
    event?.stopPropagation();
    await favorites.toggle(book.id);
  }

  function openMenu(event: MouseEvent) {
    event.preventDefault();
    event.stopPropagation();
    menuX = event.clientX;
    menuY = event.clientY;
    menuOpen = true;
  }

  function openMenuFromButton(event: MouseEvent) {
    event.preventDefault();
    event.stopPropagation();
    const btn = event.currentTarget as HTMLButtonElement;
    const rect = btn.getBoundingClientRect();
    menuX = rect.left;
    menuY = rect.bottom;
    menuOpen = true;
  }

  function closeMenu() {
    menuOpen = false;
  }

  async function deleteBook() {
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
    } catch (e) {
      toast.error(e instanceof ApiError ? e.message : i18n.t("book.deleteFailed"));
    }
  }

  const menuItems = $derived<MenuItem[]>([
    ...(can("edit_metadata")
      ? [
          {
            id: "edit",
            label: i18n.t("book.editMetadata"),
            icon: Pencil,
            onclick: () => {
              closeMenu();
              bookEditorIntent.open(book.id, "edit");
            },
          },
          {
            id: "identify",
            label: i18n.t("book.identify"),
            icon: ScanSearch,
            onclick: () => {
              closeMenu();
              bookEditorIntent.open(book.id, "identify");
            },
          },
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
  ]);
</script>

<div class="group relative flex flex-col gap-2 book-card" role="group" oncontextmenu={openMenu}>
  {#if selectMode}
    <button
      type="button"
      class="select-toggle"
      class:select-toggle--active={selected}
      aria-label={selected ? i18n.t("book.deselectBook") : i18n.t("book.selectBook")}
      onclick={(e) => {
        e.preventDefault();
        e.stopPropagation();
        onToggleSelect?.();
      }}
    >
      <Check size={14} class={selected ? "" : "opacity-0"} />
    </button>
  {:else}
    {#if menuItems.length > 0}
      <button
        type="button"
        class="more-btn"
        aria-label={i18n.t("book.moreActions")}
        onclick={openMenuFromButton}
      >
        <MoreVertical size={15} />
      </button>
    {/if}
    <button
      type="button"
      class="favorite-btn"
      class:favorite-btn--active={isFavorite}
      aria-label={isFavorite ? i18n.t("book.removeFavorite") : i18n.t("book.addFavorite")}
      onclick={toggleFavorite}
    >
      <Star size={15} fill={isFavorite ? "currentColor" : "none"} />
    </button>
  {/if}
  <a
    href={`/book/${book.id}`}
    use:link={`/book/${book.id}`}
    class="flex flex-col gap-2 outline-none"
    onclick={(e) => {
      if (selectMode) {
        e.preventDefault();
        onToggleSelect?.();
      }
    }}
  >
    <div
      class="transition-transform duration-150 group-hover:-translate-y-1 group-focus-visible:-translate-y-1"
    >
      <div
        class="relative overflow-hidden rounded-[var(--radius-card)] shadow-sm ring-1 ring-border/60 transition group-hover:shadow-md group-hover:ring-border-strong"
      >
        <Cover {book} />
        <BookCoverProgress percent={progress} />
        {#if book.format}
          <span class="format-badge">{book.format}</span>
        {/if}
      </div>
    </div>
    <div class="min-w-0 px-0.5">
      <p
        class="truncate font-medium text-fg {density.value === 'compact' ? 'text-xs' : 'text-sm'}"
        title={book.title}
      >
        {book.title}
        {#if book.duplicateOf}
          <span
            class="ml-1 rounded bg-surface px-1.5 py-0.5 text-[0.65rem] font-medium uppercase tracking-wide text-muted ring-1 ring-border"
          >
            {i18n.t("book.duplicateBadge")}
          </span>
        {/if}
      </p>
      {#if book.author}
        <p class="truncate text-xs text-muted" title={book.author}>{book.author}</p>
      {/if}
    </div>
  </a>
</div>

<ContextMenu
  open={menuOpen}
  x={menuX}
  y={menuY}
  title={book.title}
  items={menuItems}
  onclose={closeMenu}
/>

<style>
  .book-card {
    content-visibility: auto;
    contain-intrinsic-size: auto 280px;
  }

  .favorite-btn,
  .more-btn {
    position: absolute;
    z-index: 2;
    display: grid;
    place-items: center;
    width: 2rem;
    height: 2rem;
    border: 0;
    border-radius: 9999px;
    background: color-mix(in oklch, var(--color-bg) 78%, transparent);
    color: var(--color-muted);
    box-shadow: 0 1px 2px rgb(0 0 0 / 0.12);
    cursor: pointer;
    transition:
      opacity 120ms ease,
      color 120ms ease,
      background-color 120ms ease;
    backdrop-filter: blur(6px);
  }

  .favorite-btn {
    top: 0.35rem;
    right: 0.35rem;
  }

  .more-btn {
    top: 0.35rem;
    left: 0.35rem;
  }

  .favorite-btn:hover,
  .more-btn:hover {
    color: var(--color-fg);
    background: var(--color-surface);
  }

  .favorite-btn--active {
    color: var(--color-primary);
  }

  .format-badge {
    position: absolute;
    bottom: 0.4rem;
    left: 0.4rem;
    z-index: 2;
    max-width: calc(100% - 0.8rem);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    padding: 0.15rem 0.4rem;
    border-radius: 9999px;
    border: 1px solid color-mix(in oklch, var(--color-border) 70%, transparent);
    background: color-mix(in oklch, var(--color-bg) 72%, transparent);
    color: var(--color-fg);
    font-size: 0.625rem;
    font-weight: 600;
    letter-spacing: 0.04em;
    text-transform: uppercase;
    backdrop-filter: blur(6px);
    box-shadow: 0 1px 2px rgb(0 0 0 / 0.1);
    pointer-events: none;
  }

  @media (hover: hover) and (pointer: fine) {
    .favorite-btn,
    .more-btn,
    .format-badge {
      opacity: 0;
    }

    .group:hover .favorite-btn,
    .group:focus-within .favorite-btn,
    .favorite-btn--active,
    .group:hover .more-btn,
    .group:focus-within .more-btn,
    .group:hover .format-badge,
    .group:focus-within .format-badge {
      opacity: 1;
    }
  }

  @media (hover: none), (pointer: coarse) {
    .favorite-btn,
    .more-btn {
      opacity: 0.95;
      width: 2.25rem;
      height: 2.25rem;
    }

    .favorite-btn--active {
      opacity: 1;
    }

    .format-badge {
      opacity: 1;
    }

    .select-toggle {
      width: 2.25rem;
      height: 2.25rem;
    }
  }

  .select-toggle {
    position: absolute;
    top: 0.35rem;
    right: 0.35rem;
    z-index: 2;
    display: grid;
    place-items: center;
    width: 2rem;
    height: 2rem;
    border: 1px solid var(--color-border);
    border-radius: 0.375rem;
    background: var(--color-bg);
    color: var(--color-fg);
    cursor: pointer;
  }

  .select-toggle--active {
    border-color: var(--color-primary);
    background: color-mix(in oklch, var(--color-primary) 18%, var(--color-bg));
    color: var(--color-primary);
  }
</style>
