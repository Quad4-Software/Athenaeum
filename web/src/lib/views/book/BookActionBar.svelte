<script lang="ts">
  import { BookOpen, Download, Headphones, Link2, MoreVertical, Star } from "@lucide/svelte";
  import Button from "$lib/components/Button.svelte";
  import { api } from "$lib/api/client";
  import { router } from "$lib/router.svelte";
  import { i18n } from "$lib/stores/i18n.svelte";
  import { isAudioFormat, type Book } from "$lib/api/types";
  import type { BookOfflineStatus } from "$lib/offline/book-cache";
  import BookOfflineButton from "./BookOfflineButton.svelte";

  interface Props {
    book: Book;
    isFavorite: boolean;
    resumeText: string | null;
    shareBusy: boolean;
    offlineBusy: boolean;
    offlineStatus: BookOfflineStatus;
    menuOpen: boolean;
    hasMenu: boolean;
    ontogglefavorite: () => void;
    oncreateshare: () => void;
    ontoggleoffline: () => void;
    onopenmenu: (event: MouseEvent) => void;
  }

  let {
    book,
    isFavorite,
    resumeText,
    shareBusy,
    offlineBusy,
    offlineStatus,
    menuOpen,
    hasMenu,
    ontogglefavorite,
    oncreateshare,
    ontoggleoffline,
    onopenmenu,
  }: Props = $props();
</script>

<div class="book-actions mt-6 flex flex-wrap gap-2 sm:gap-3">
  <Button class="min-h-11 flex-1 sm:flex-none" onclick={() => router.navigate(`/read/${book.id}`)}>
    {#if isAudioFormat(book.format)}
      <Headphones size={16} /> {i18n.t("book.listen")}
    {:else if resumeText}
      <BookOpen size={16} /> {i18n.t("book.resume")}
    {:else}
      <BookOpen size={16} /> {i18n.t("book.read")}
    {/if}
  </Button>
  <Button variant="ghost" class="min-h-11 ring-1 ring-border" onclick={ontogglefavorite}>
    <Star size={16} fill={isFavorite ? "currentColor" : "none"} />
    {isFavorite ? i18n.t("book.favorited") : i18n.t("book.favorite")}
  </Button>
  <a class="btn btn-ghost min-h-11 ring-1 ring-border" href={api.downloadUrl(book.id)} download>
    <Download size={16} />
    {i18n.t("book.download")}
  </a>
  {#if !isAudioFormat(book.format)}
    <BookOfflineButton busy={offlineBusy} status={offlineStatus} onclick={ontoggleoffline} />
  {/if}
  <Button
    variant="ghost"
    class="min-h-11 ring-1 ring-border"
    loading={shareBusy}
    onclick={oncreateshare}
  >
    <Link2 size={16} />
    {i18n.t("book.share")}
  </Button>
  {#if hasMenu}
    <button
      type="button"
      class="btn btn-ghost min-h-11 min-w-11 ring-1 ring-border"
      aria-label={i18n.t("book.moreActions")}
      aria-haspopup="menu"
      aria-expanded={menuOpen}
      onclick={onopenmenu}
    >
      <MoreVertical size={18} />
    </button>
  {/if}
</div>

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
