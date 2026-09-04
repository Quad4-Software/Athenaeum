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
import { opURL } from "./op";

export const librariesApi = {
  stats: (library?: number) =>
    request<LibraryStats>(
      `${opURL("GET__api_library_stats")}${library != null ? `?library=${library}` : ""}`,
    ),

  scanStatus: () => request<ScanStatus>(opURL("GET__api_library_scan_status")),

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
      library != null
        ? opURL("POST__api_libraries__id__scan", { id: library })
        : opURL("POST__api_library_scan"),
      { method: "POST" },
    ),

  listSeries: (library?: number) =>
    request<SeriesInfo[]>(
      `${opURL("GET__api_series")}${library != null ? `?library=${library}` : ""}`,
    ),

  listAuthors: (library?: number) =>
    request<AuthorInfo[]>(
      `${opURL("GET__api_authors")}${library != null ? `?library=${library}` : ""}`,
    ),

  listLibraries: () => request<LibraryMount[]>(opURL("GET__api_libraries")),

  createLibrary: (input: LibraryCreateInput) =>
    request<LibraryMount>(opURL("POST__api_libraries"), {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(input),
    }),

  updateLibrary: (id: number, input: LibraryCreateInput) =>
    request<LibraryMount>(opURL("PUT__api_libraries__id", { id }), {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(input),
    }),

  testS3: (s3: LibraryS3Input) =>
    request<{ ok: boolean }>(opURL("POST__api_libraries_test_s3"), {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(s3),
    }),

  deleteLibrary: (id: number) =>
    request<void>(opURL("DELETE__api_libraries__id", { id }), { method: "DELETE" }),

  reorderLibraries: (ids: number[]) =>
    request<{ ok: boolean }>(opURL("PUT__api_libraries_reorder"), {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ ids }),
    }),

  createUpload: (libraryId: number, relPath: string, totalSize: number) =>
    request<UploadSession>(opURL("POST__api_libraries__id__uploads", { id: libraryId }), {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ relPath, totalSize }),
    }),

  getUpload: (libraryId: number, uploadId: string) =>
    request<UploadSession>(
      opURL("GET__api_libraries__id__uploads__uploadId", { id: libraryId, uploadId }),
    ),

  cancelUpload: (libraryId: number, uploadId: string) =>
    request<void>(
      opURL("DELETE__api_libraries__id__uploads__uploadId", { id: libraryId, uploadId }),
      { method: "DELETE" },
    ),

  uploadChunk: async (
    libraryId: number,
    uploadId: string,
    chunk: Blob,
    start: number,
    end: number,
    total: number,
  ): Promise<UploadSession> => {
    const csrf = await ensureCsrf();
    const res = await fetch(
      opURL("PATCH__api_libraries__id__uploads__uploadId", { id: libraryId, uploadId }),
      {
        method: "PATCH",
        credentials: "same-origin",
        headers: {
          "Content-Type": "application/octet-stream",
          "Content-Range": `bytes ${start}-${end}/${total}`,
          [CSRF_HEADER]: csrf,
        },
        body: chunk,
      },
    );
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
