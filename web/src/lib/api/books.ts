import type {
  Book,
  BookPage,
  BookQueryParams,
  Chapter,
  AudiobookTrack,
  ComicManifest,
  ConvertResult,
  BookUpdate,
  MetadataProvider,
  MetadataMatch,
  MetadataSearchQuery,
  Progress,
  Bookmark,
  Highlight,
  MobiSection,
  ReadingStats,
  Tag,
  BookRating,
} from "./types";
import { request, ensureCsrf, ApiError, buildQuery, CSRF_HEADER } from "./core";
import { isDemoMode } from "$lib/demo/mode";
import { demoCoverUrlForBook } from "$lib/demo/covers";

export const booksApi = {
  listBooks: (params: BookQueryParams = {}) => request<BookPage>(`/api/books${buildQuery(params)}`),

  getBook: (id: number) => request<Book>(`/api/books/${id}`),

  deleteBook: (id: number) => request<void>(`/api/books/${id}`, { method: "DELETE" }),

  getChapters: (id: number) => request<Chapter[]>(`/api/books/${id}/chapters`),

  getAudiobookTracks: (id: number) => request<AudiobookTrack[]>(`/api/books/${id}/tracks`),

  getComicManifest: (id: number) => request<ComicManifest>(`/api/books/${id}/pages`),

  getMobiSections: (id: number) => request<MobiSection[]>(`/api/books/${id}/mobi-sections`),

  convertBook: (id: number, target: "epub" | "pdf") =>
    request<ConvertResult>(`/api/books/${id}/convert?target=${target}`, { method: "POST" }),

  updateBook: (id: number, data: BookUpdate) =>
    request<Book>(`/api/books/${id}`, {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(data),
    }),

  uploadCover: async (id: number, file: File) => {
    const csrf = await ensureCsrf();
    const form = new FormData();
    form.append("cover", file);
    const res = await fetch(`/api/books/${id}/cover`, {
      method: "PUT",
      credentials: "same-origin",
      headers: { Accept: "application/json", [CSRF_HEADER]: csrf },
      body: form,
    });
    if (!res.ok) {
      let message = res.statusText;
      try {
        const body = (await res.json()) as { error?: string };
        if (body.error) message = body.error;
      } catch {
        // ignore
      }
      throw new ApiError(res.status, message);
    }
    return (await res.json()) as Book;
  },

  deleteCover: (id: number) => request<Book>(`/api/books/${id}/cover`, { method: "DELETE" }),

  listMetadataProviders: () => request<MetadataProvider[]>("/api/metadata/providers"),

  searchMetadata: (id: number, query: MetadataSearchQuery) =>
    request<{ matches: MetadataMatch[] }>(`/api/books/${id}/metadata/search`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(query),
    }),

  applyMetadataMatch: (id: number, match: MetadataMatch, applyCover = false) =>
    request<Book>(`/api/books/${id}/metadata/apply`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ match, applyCover }),
    }),

  coverFromUrl: (id: number, url: string) =>
    request<Book>(`/api/books/${id}/cover-from-url`, {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ url }),
    }),

  getProgress: (id: number) => request<Progress>(`/api/books/${id}/progress`),

  saveProgress: (id: number, p: Pick<Progress, "location" | "percent">) =>
    request<Progress>(`/api/books/${id}/progress`, {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(p),
    }),

  readingStats: () => request<ReadingStats>("/api/stats/reading"),

  addReadingTime: (bookId: number, seconds: number) =>
    request<{ ok: boolean }>(`/api/books/${bookId}/reading-time`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ seconds }),
    }),

  listBookmarks: (bookId: number) => request<Bookmark[]>(`/api/books/${bookId}/bookmarks`),

  createBookmark: (bookId: number, location: string, label = "") =>
    request<Bookmark>(`/api/books/${bookId}/bookmarks`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ location, label }),
    }),

  deleteBookmark: (bookId: number, bookmarkId: number) =>
    request<void>(`/api/books/${bookId}/bookmarks/${bookmarkId}`, { method: "DELETE" }),

  listHighlights: (bookId: number) => request<Highlight[]>(`/api/books/${bookId}/highlights`),

  createHighlight: (bookId: number, location: string, excerpt = "", note = "", color = "yellow") =>
    request<Highlight>(`/api/books/${bookId}/highlights`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ location, excerpt, note, color }),
    }),

  deleteHighlight: (bookId: number, highlightId: number) =>
    request<void>(`/api/books/${bookId}/highlights/${highlightId}`, { method: "DELETE" }),

  listTags: () => request<Tag[]>("/api/tags"),

  createTag: (name: string) =>
    request<Tag>("/api/tags", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ name }),
    }),

  listBookTags: (bookId: number) => request<string[]>(`/api/books/${bookId}/tags`),

  setBookTags: (bookId: number, tags: string[]) =>
    request<string[]>(`/api/books/${bookId}/tags`, {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ tags }),
    }),

  addBookTag: (bookId: number, name: string) =>
    request<string[]>(`/api/books/${bookId}/tags`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ name }),
    }),

  removeBookTag: (bookId: number, tagId: number) =>
    request<{ ok: boolean }>(`/api/books/${bookId}/tags/${tagId}`, { method: "DELETE" }),

  getRating: (bookId: number) => request<BookRating>(`/api/books/${bookId}/rating`),

  setRating: (bookId: number, rating: number) =>
    request<BookRating>(`/api/books/${bookId}/rating`, {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ rating }),
    }),

  createShareLink: (bookId: number, expiresInHours = 0, maxDownloads = 0) =>
    request<{ id: number; token: string; url: string; expiresAt?: string }>(
      `/api/books/${bookId}/share`,
      {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ expiresInHours, maxDownloads }),
      },
    ),

  listShareLinks: (bookId: number) =>
    request<{ id: number; token: string; expiresAt?: string; downloadCount: number }[]>(
      `/api/books/${bookId}/share`,
    ),

  deleteShareLink: (bookId: number, shareId: number) =>
    request<{ ok: boolean }>(`/api/books/${bookId}/share/${shareId}`, { method: "DELETE" }),

  sendBook: (bookId: number, opts?: { to?: string; kindle?: boolean }) =>
    request<{ ok: string; to: string }>(`/api/books/${bookId}/send`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(opts ?? { kindle: true }),
    }),

  listOffline: () => request<{ bookIds: number[] }>("/api/offline"),

  addOffline: (bookIds: number[]) =>
    request<{ bookIds: number[] }>("/api/offline", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ bookIds }),
    }),

  removeOffline: (bookIds: number[]) =>
    request<{ bookIds: number[] }>("/api/offline", {
      method: "DELETE",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ bookIds }),
    }),

  coverUrl: (id: number, version?: string) => {
    if (isDemoMode()) return demoCoverUrlForBook(id);
    const base = `/api/books/${id}/cover`;
    return version ? `${base}?v=${encodeURIComponent(version)}` : base;
  },
  fileUrl: (id: number, track?: number) =>
    track != null ? `/api/books/${id}/file?track=${track}` : `/api/books/${id}/file`,

  comicPageUrl: (id: number, page: number) => `/api/books/${id}/pages/${page}`,
  downloadUrl: (id: number) => `/api/books/${id}/download`,
};
