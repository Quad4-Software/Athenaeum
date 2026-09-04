import { Copy, FileOutput, FileText, Pencil, ScanSearch, Star, Trash2 } from "@lucide/svelte";
import { api, ApiError } from "$lib/api/client";
import { router } from "$lib/router.svelte";
import { collections } from "$lib/stores/collections.svelte";
import { favorites } from "$lib/stores/favorites.svelte";
import { library } from "$lib/stores/library.svelte";
import { toast } from "$lib/stores/toast.svelte";
import { i18n } from "$lib/stores/i18n.svelte";
import { can } from "$lib/permissions";
import { parseAudioLocation } from "$lib/audio/progress";
import { bookOfflineCache } from "$lib/offline/book-cache";
import {
  isAudioFormat,
  isComicFormat,
  isMobiFormat,
  type Book,
  type Progress,
} from "$lib/api/types";
import type { MenuItem } from "$lib/components/MenuList.svelte";

export async function addToCollection(bookId: number, collectionId: number): Promise<void> {
  try {
    await api.addToCollection(collectionId, bookId);
    toast.success("Added to list");
    void collections.refresh();
  } catch (e) {
    toast.error(e instanceof ApiError ? e.message : "Failed to add");
  }
}

export async function toggleFavorite(bookId: number): Promise<void> {
  await favorites.toggle(bookId);
}

export async function addTag(bookId: number, name: string): Promise<string[] | null> {
  try {
    return await api.addBookTag(bookId, name);
  } catch (e) {
    toast.error(e instanceof ApiError ? e.message : "Failed to add tag");
    return null;
  }
}

export async function removeTag(bookId: number, remaining: string[]): Promise<string[] | null> {
  try {
    return await api.setBookTags(bookId, remaining);
  } catch (e) {
    toast.error(e instanceof ApiError ? e.message : "Failed to remove tag");
    return null;
  }
}

export function filterByTag(name: string): void {
  library.setTag(name);
  router.navigate("/");
}

export async function setRating(
  bookId: number,
  currentRating: number,
  value: number,
): Promise<number | null> {
  const next = currentRating === value ? 0 : value;
  try {
    const rating = await api.setRating(bookId, next);
    return rating.rating;
  } catch (e) {
    toast.error(e instanceof ApiError ? e.message : "Failed to save rating");
    return null;
  }
}

export async function createShare(bookId: number): Promise<string | null> {
  try {
    const link = await api.createShareLink(bookId, 168);
    try {
      await navigator.clipboard.writeText(link.url);
      toast.success(i18n.t("book.shareCopied"));
    } catch {
      toast.success(i18n.t("book.shareCreated"));
    }
    return link.url;
  } catch (e) {
    toast.error(e instanceof ApiError ? e.message : i18n.t("book.shareFailed"));
    return null;
  }
}

export async function copyShareUrl(shareUrl: string): Promise<void> {
  if (!shareUrl) return;
  try {
    await navigator.clipboard.writeText(shareUrl);
    toast.success(i18n.t("book.shareCopied"));
  } catch {
    toast.info(i18n.t("book.shareCopyManual"));
  }
}

export function hasCitation(b: Book): boolean {
  return Boolean(b.doi || b.arxivId || b.pubmedId || b.journal || b.publishedYear);
}

export function citationVolumeLine(b: Book): string {
  const parts: string[] = [];
  if (b.volume) parts.push(`${i18n.t("book.volume")} ${b.volume}`);
  if (b.issue) parts.push(`${i18n.t("book.issue")} ${b.issue}`);
  if (b.pages) parts.push(`${i18n.t("book.pages")} ${b.pages}`);
  return parts.join(", ");
}

export async function fetchBibTeX(bookId: number): Promise<string | null> {
  try {
    return await api.getBibTeX(bookId);
  } catch (e) {
    toast.error(e instanceof ApiError ? e.message : i18n.t("book.bibtexFailed"));
    return null;
  }
}

export async function copyBibTeX(bookId: number): Promise<void> {
  try {
    const text = await fetchBibTeX(bookId);
    if (!text) return;
    await navigator.clipboard.writeText(text);
    toast.success(i18n.t("book.bibtexCopied"));
  } catch {
    toast.error(i18n.t("book.bibtexFailed"));
  }
}

export async function downloadBibTeX(book: Book): Promise<void> {
  const text = await fetchBibTeX(book.id);
  if (!text) return;
  const blob = new Blob([text], { type: "application/x-bibtex" });
  const url = URL.createObjectURL(blob);
  const a = document.createElement("a");
  a.href = url;
  a.download = `${book.title.replace(/[^\w.-]+/g, "_").slice(0, 80) || "citation"}.bib`;
  a.click();
  URL.revokeObjectURL(url);
}

export async function toggleOffline(book: Book, isComplete: boolean): Promise<void> {
  try {
    if (isComplete) {
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
            pageUrl: (page) => api.comicPageUrl(book.id, page),
          },
        );
      } else {
        bookOfflineCache.startDownload(book.id, url, book.fileSize, book.modifiedAt);
      }
      toast.success(i18n.t("book.offlineStarted"));
    }
  } catch (e) {
    toast.error(e instanceof ApiError ? e.message : i18n.t("book.offlineFailed"));
  }
}

export async function deleteBook(book: Book): Promise<void> {
  try {
    await api.deleteBook(book.id);
    toast.success(i18n.t("book.deleted"));
    void library.refresh();
    router.navigate("/");
  } catch (e) {
    toast.error(e instanceof ApiError ? e.message : i18n.t("book.deleteFailed"));
  }
}

export async function convertBook(book: Book, target: "epub" | "pdf"): Promise<void> {
  try {
    const res = await toast.promise(() => api.convertBook(book.id, target), {
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

export function resumeLabel(book: Book, progress: Progress | null): string | null {
  if (!progress || progress.percent <= 0) return null;
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

export interface BuildMenuItemsParams {
  book: Book;
  isFavorite: boolean;
  onEdit: () => void;
  onIdentify: () => void;
  onConvertEpub: () => void;
  onCopyBibtex: () => void;
  onDownloadBibtex: () => void;
  onToggleFavorite: () => void;
  onDelete: () => void;
}

export function buildMenuItems(params: BuildMenuItemsParams): MenuItem[] {
  const {
    book,
    isFavorite,
    onEdit,
    onIdentify,
    onConvertEpub,
    onCopyBibtex,
    onDownloadBibtex,
    onToggleFavorite,
    onDelete,
  } = params;
  return [
    ...(can("edit_metadata")
      ? [
          {
            id: "edit",
            label: i18n.t("book.editMetadata"),
            icon: Pencil,
            onclick: onEdit,
          },
          {
            id: "identify",
            label: i18n.t("book.identify"),
            icon: ScanSearch,
            onclick: onIdentify,
          },
          ...(isMobiFormat(book.format) || book.format === "pdf"
            ? [
                {
                  id: "convert-epub",
                  label: i18n.t("book.convertEpub"),
                  icon: FileOutput,
                  onclick: onConvertEpub,
                },
              ]
            : []),
        ]
      : []),
    ...(hasCitation(book)
      ? [
          {
            id: "copy-bibtex",
            label: i18n.t("book.copyBibtex"),
            icon: Copy,
            onclick: onCopyBibtex,
          },
          {
            id: "download-bibtex",
            label: i18n.t("book.downloadBibtex"),
            icon: FileText,
            onclick: onDownloadBibtex,
          },
        ]
      : []),
    {
      id: "favorite",
      label: isFavorite ? i18n.t("book.removeFavorite") : i18n.t("book.addFavorite"),
      icon: Star,
      active: isFavorite,
      onclick: onToggleFavorite,
    },
    ...(can("delete_books")
      ? [
          {
            id: "delete",
            label: i18n.t("book.delete"),
            icon: Trash2,
            danger: true,
            separator: true,
            onclick: onDelete,
          },
        ]
      : []),
  ];
}
