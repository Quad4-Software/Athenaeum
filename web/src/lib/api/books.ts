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
import { opURL } from "./op";

export const booksApi = {
  listBooks: (params: BookQueryParams = {}) =>
    request<BookPage>(`${opURL("GET__api_books")}${buildQuery(params)}`),

  getBook: (id: number) => request<Book>(opURL("GET__api_books__id", { id })),

  deleteBook: (id: number) => request<void>(`/api/books/${id}`, { method: "DELETE" }),

  getChapters: (id: number) => request<Chapter[]>(opURL("GET__api_books__id__chapters", { id })),

  getAudiobookTracks: (id: number) => request<AudiobookTrack[]>(`/api/books/${id}/tracks`),

  getComicManifest: (id: number) => request<ComicManifest>(`/api/books/${id}/pages`),

  getMobiSections: (id: number) => request<MobiSection[]>(`/api/books/${id}/mobi-sections`),

  convertBook: (id: number, target: "epub" | "pdf") =>
    request<ConvertResult>(`/api/books/${id}/convert?target=${target}`, { method: "POST" }),

  updateBook: (id: number, data: BookUpdate) =>
    request<Book>(opURL("PUT__api_books__id", { id }), {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(data),
    }),

  uploadCover: async (id: number, file: File) => {
    const csrf = await ensureCsrf();
    const form = new FormData();
    form.append("cover", file);
    const res = await fetch(opURL("GET__api_books__id__cover", { id }), {
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

  deleteCover: (id: number) =>
    request<Book>(opURL("GET__api_books__id__cover", { id }), { method: "DELETE" }),

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

  getBibTeX: async (id: number) => {
    const res = await fetch(`/api/books/${id}/bibtex`, {
      credentials: "same-origin",
      headers: { Accept: "application/x-bibtex, text/plain" },
    });
    if (!res.ok) {
      throw new ApiError(res.status, res.statusText || "BibTeX export failed");
    }
    return await res.text();
  },

  importBibTeX: (bibtex: string) =>
    request<{
      matched: number;
      updated: number;
      unmatched: number;
      skipped: number;
      unmatchedTitles?: string[];
    }>("/api/library/bibtex/import", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ bibtex }),
    }),

  coverFromUrl: (id: number, url: string) =>
    request<Book>(`/api/books/${id}/cover-from-url`, {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ url }),
    }),

  getProgress: (id: number) => request<Progress>(opURL("GET__api_books__id__progress", { id })),

  saveProgress: (id: number, p: Pick<Progress, "location" | "percent">) =>
    request<Progress>(opURL("PUT__api_books__id__progress", { id }), {
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

  listTags: () => request<Tag[]>(opURL("GET__api_tags")),

  createTag: (name: string) =>
    request<Tag>(opURL("POST__api_tags"), {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ name }),
    }),

  listBookTags: (bookId: number) =>
    request<string[]>(opURL("GET__api_books__id__tags", { id: bookId })),

  setBookTags: (bookId: number, tags: string[]) =>
    request<string[]>(opURL("PUT__api_books__id__tags", { id: bookId }), {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ tags }),
    }),

  addBookTag: (bookId: number, name: string) =>
    request<string[]>(opURL("POST__api_books__id__tags", { id: bookId }), {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ name }),
    }),

  removeBookTag: (bookId: number, tagId: number) =>
    request<{ ok: boolean }>(opURL("DELETE__api_books__id__tags__tagId", { id: bookId, tagId }), {
      method: "DELETE",
    }),

  getRating: (bookId: number) =>
    request<BookRating>(opURL("GET__api_books__id__rating", { id: bookId })),

  setRating: (bookId: number, rating: number) =>
    request<BookRating>(opURL("PUT__api_books__id__rating", { id: bookId }), {
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
    const base = opURL("GET__api_books__id__cover", { id });
    return version ? `${base}?v=${encodeURIComponent(version)}` : base;
  },
  fileUrl: (id: number, track?: number) => {
    const base = opURL("GET__api_books__id__file", { id });
    return track != null ? `${base}?track=${track}` : base;
  },

  comicPageUrl: (id: number, page: number) => `/api/books/${id}/pages/${page}`,
  downloadUrl: (id: number) => opURL("GET__api_books__id__download", { id }),
};
