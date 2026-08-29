import type {
  LibraryMount,
  LibraryCreateInput,
  LibraryS3Input,
  LibraryStats,
  FSBrowseResult,
  MetadataAutoMatchRequest,
  MetadataMatchStatus,
  SeriesInfo,
  AuthorInfo,
  ScanStatus,
  UploadSession,
} from "./types";
import { request, ensureCsrf, ApiError, CSRF_HEADER } from "./core";

export const librariesApi = {
  stats: (library?: number) =>
    request<LibraryStats>(`/api/library/stats${library != null ? `?library=${library}` : ""}`),

  scanStatus: () => request<ScanStatus>("/api/library/scan/status"),

  startMetadataMatch: (body: MetadataAutoMatchRequest) =>
    request<{ ok: boolean }>("/api/library/metadata/match", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(body),
    }),

  metadataMatchStatus: () => request<MetadataMatchStatus>("/api/library/metadata/match/status"),

  cleanupSeriesNames: () =>
    request<{ updated: number }>("/api/library/series/cleanup", { method: "POST" }),

  scan: (library?: number) =>
    request<{ started: boolean }>(
      library != null ? `/api/libraries/${library}/scan` : "/api/library/scan",
      { method: "POST" },
    ),

  listSeries: (library?: number) =>
    request<SeriesInfo[]>(`/api/series${library != null ? `?library=${library}` : ""}`),

  listAuthors: (library?: number) =>
    request<AuthorInfo[]>(`/api/authors${library != null ? `?library=${library}` : ""}`),

  listLibraries: () => request<LibraryMount[]>("/api/libraries"),

  createLibrary: (input: LibraryCreateInput) =>
    request<LibraryMount>("/api/libraries", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(input),
    }),

  updateLibrary: (id: number, input: LibraryCreateInput) =>
    request<LibraryMount>(`/api/libraries/${id}`, {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(input),
    }),

  testS3: (s3: LibraryS3Input) =>
    request<{ ok: boolean }>("/api/libraries/test-s3", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(s3),
    }),

  deleteLibrary: (id: number) => request<void>(`/api/libraries/${id}`, { method: "DELETE" }),

  reorderLibraries: (ids: number[]) =>
    request<{ ok: boolean }>("/api/libraries/reorder", {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ ids }),
    }),

  createUpload: (libraryId: number, relPath: string, totalSize: number) =>
    request<UploadSession>(`/api/libraries/${libraryId}/uploads`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ relPath, totalSize }),
    }),

  getUpload: (libraryId: number, uploadId: string) =>
    request<UploadSession>(`/api/libraries/${libraryId}/uploads/${uploadId}`),

  cancelUpload: (libraryId: number, uploadId: string) =>
    request<void>(`/api/libraries/${libraryId}/uploads/${uploadId}`, { method: "DELETE" }),

  uploadChunk: async (
    libraryId: number,
    uploadId: string,
    chunk: Blob,
    start: number,
    end: number,
    total: number,
  ): Promise<UploadSession> => {
    const csrf = await ensureCsrf();
    const res = await fetch(`/api/libraries/${libraryId}/uploads/${uploadId}`, {
      method: "PATCH",
      credentials: "same-origin",
      headers: {
        "Content-Type": "application/octet-stream",
        "Content-Range": `bytes ${start}-${end}/${total}`,
        [CSRF_HEADER]: csrf,
      },
      body: chunk,
    });
    if (!res.ok) {
      let msg = res.statusText;
      try {
        const body = await res.json();
        if (body?.error) msg = body.error;
      } catch {
        /* ignore */
      }
      throw new ApiError(res.status, msg);
    }
    return res.json() as Promise<UploadSession>;
  },

  browseFS: (path = "") =>
    request<FSBrowseResult>(`/api/fs/browse${path ? `?path=${encodeURIComponent(path)}` : ""}`),
};
