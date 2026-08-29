const DB_NAME = "athenaeum-books";
const DB_VERSION = 1;
const FILE_STORE = "files";
const PAGE_STORE = "pages";
const MAX_FILE_BYTES = 512 * 1024 * 1024;

export interface BookOfflineStatus {
  cachedBytes: number;
  totalBytes: number;
  complete: boolean;
  downloading: boolean;
  error: string | null;
  pagesComplete?: boolean;
}

interface FileMeta {
  bookId: number;
  url: string;
  fileSize: number;
  bookModifiedAt: string;
  downloaded: number;
  complete: boolean;
  mimeType: string;
}

type StatusListener = (status: BookOfflineStatus) => void;

function openDB(): Promise<IDBDatabase> {
  return new Promise((resolve, reject) => {
    const req = indexedDB.open(DB_NAME, DB_VERSION);
    req.onupgradeneeded = () => {
      const db = req.result;
      if (!db.objectStoreNames.contains(FILE_STORE)) {
        db.createObjectStore(FILE_STORE, { keyPath: "bookId" });
      }
      if (!db.objectStoreNames.contains(PAGE_STORE)) {
        db.createObjectStore(PAGE_STORE);
      }
    };
    req.onsuccess = () => resolve(req.result);
    req.onerror = () => reject(req.error ?? new Error("indexedDB open failed"));
  });
}

function txDone(tx: IDBTransaction): Promise<void> {
  return new Promise((resolve, reject) => {
    tx.oncomplete = () => resolve();
    tx.onerror = () => reject(tx.error ?? new Error("indexedDB transaction failed"));
    tx.onabort = () => reject(tx.error ?? new Error("indexedDB transaction aborted"));
  });
}

async function getMeta(bookId: number): Promise<FileMeta | null> {
  const db = await openDB();
  return new Promise((resolve, reject) => {
    const tx = db.transaction(FILE_STORE, "readonly");
    const req = tx.objectStore(FILE_STORE).get(bookId);
    req.onsuccess = () => resolve((req.result as FileMeta | undefined) ?? null);
    req.onerror = () => reject(req.error);
  });
}

async function putMeta(meta: FileMeta): Promise<void> {
  const db = await openDB();
  const tx = db.transaction(FILE_STORE, "readwrite");
  tx.objectStore(FILE_STORE).put(meta);
  await txDone(tx);
}

async function putFileBlob(bookId: number, blob: Blob): Promise<void> {
  const db = await openDB();
  const tx = db.transaction(PAGE_STORE, "readwrite");
  tx.objectStore(PAGE_STORE).put(blob, `file:${bookId}`);
  await txDone(tx);
}

async function getFileBlob(bookId: number): Promise<Blob | null> {
  const db = await openDB();
  return new Promise((resolve, reject) => {
    const tx = db.transaction(PAGE_STORE, "readonly");
    const req = tx.objectStore(PAGE_STORE).get(`file:${bookId}`);
    req.onsuccess = () => resolve((req.result as Blob | undefined) ?? null);
    req.onerror = () => reject(req.error);
  });
}

async function putPageBlob(bookId: number, page: number, blob: Blob): Promise<void> {
  const db = await openDB();
  const tx = db.transaction(PAGE_STORE, "readwrite");
  tx.objectStore(PAGE_STORE).put(blob, `page:${bookId}:${page}`);
  await txDone(tx);
}

async function getPageBlob(bookId: number, page: number): Promise<Blob | null> {
  const db = await openDB();
  return new Promise((resolve, reject) => {
    const tx = db.transaction(PAGE_STORE, "readonly");
    const req = tx.objectStore(PAGE_STORE).get(`page:${bookId}:${page}`);
    req.onsuccess = () => resolve((req.result as Blob | undefined) ?? null);
    req.onerror = () => reject(req.error);
  });
}

function statusFromMeta(
  meta: FileMeta | null,
  downloading: boolean,
  error: string | null,
  pagesComplete = false,
): BookOfflineStatus {
  if (!meta) {
    return { cachedBytes: 0, totalBytes: 0, complete: false, downloading, error, pagesComplete };
  }
  return {
    cachedBytes: meta.downloaded,
    totalBytes: meta.fileSize,
    complete: meta.complete,
    downloading,
    error,
    pagesComplete: meta.complete && pagesComplete,
  };
}

export class BookOfflineCache {
  private jobs = new Map<number, AbortController>();
  private blobUrls = new Map<number, string>();
  private pageUrls = new Map<string, string>();
  private listeners = new Map<number, Set<StatusListener>>();
  private jobError = new Map<number, string | null>();
  private pagesDone = new Map<number, boolean>();

  subscribe(bookId: number, listener: StatusListener): () => void {
    let set = this.listeners.get(bookId);
    if (!set) {
      set = new Set();
      this.listeners.set(bookId, set);
    }
    set.add(listener);
    void this.getStatus(bookId).then(listener);
    return () => set?.delete(listener);
  }

  private emit(bookId: number, status: BookOfflineStatus) {
    for (const listener of this.listeners.get(bookId) ?? []) listener(status);
  }

  async getStatus(bookId: number): Promise<BookOfflineStatus> {
    const meta = await getMeta(bookId);
    return statusFromMeta(
      meta,
      this.jobs.has(bookId),
      this.jobError.get(bookId) ?? null,
      this.pagesDone.get(bookId) ?? false,
    );
  }

  async isComplete(bookId: number, url: string, bookModifiedAt: string): Promise<boolean> {
    const meta = await getMeta(bookId);
    return !!meta?.complete && meta.url === url && meta.bookModifiedAt === bookModifiedAt;
  }

  async getBlobUrl(bookId: number): Promise<string | null> {
    const existing = this.blobUrls.get(bookId);
    if (existing) return existing;
    const meta = await getMeta(bookId);
    if (!meta?.complete) return null;
    const blob = await getFileBlob(bookId);
    if (!blob) return null;
    const objectUrl = URL.createObjectURL(blob);
    this.blobUrls.set(bookId, objectUrl);
    return objectUrl;
  }

  async resolveFileUrl(
    bookId: number,
    streamUrl: string,
    fileSize: number,
    bookModifiedAt: string,
  ): Promise<string> {
    if (await this.isComplete(bookId, streamUrl, bookModifiedAt)) {
      const blobUrl = await this.getBlobUrl(bookId);
      if (blobUrl) return blobUrl;
    }
    void fileSize;
    return streamUrl;
  }

  async resolvePageUrl(bookId: number, page: number, onlineUrl: string): Promise<string> {
    const key = `${bookId}:${page}`;
    const cached = this.pageUrls.get(key);
    if (cached) return cached;
    const blob = await getPageBlob(bookId, page);
    if (blob) {
      const url = URL.createObjectURL(blob);
      this.pageUrls.set(key, url);
      return url;
    }
    return onlineUrl;
  }

  startDownload(
    bookId: number,
    streamUrl: string,
    fileSize: number,
    bookModifiedAt: string,
    mimeType = "application/octet-stream",
    comicPages?: { total: number; pageUrl: (page: number) => string },
  ): void {
    if (fileSize <= 0 || fileSize > MAX_FILE_BYTES) return;
    void this.runDownload(bookId, streamUrl, fileSize, bookModifiedAt, mimeType, comicPages);
  }

  stop(bookId: number): void {
    this.jobs.get(bookId)?.abort();
    this.jobs.delete(bookId);
  }

  async clear(bookId: number): Promise<void> {
    this.stop(bookId);
    const fileUrl = this.blobUrls.get(bookId);
    if (fileUrl) URL.revokeObjectURL(fileUrl);
    this.blobUrls.delete(bookId);
    for (const [key, url] of this.pageUrls) {
      if (key.startsWith(`${bookId}:`)) {
        URL.revokeObjectURL(url);
        this.pageUrls.delete(key);
      }
    }
    this.jobError.delete(bookId);
    this.pagesDone.delete(bookId);

    const db = await openDB();
    const meta = await getMeta(bookId);
    const tx = db.transaction([FILE_STORE, PAGE_STORE], "readwrite");
    tx.objectStore(FILE_STORE).delete(bookId);
    tx.objectStore(PAGE_STORE).delete(`file:${bookId}`);
    if (meta) {
      // best-effort page cleanup for large comics
    }
    await txDone(tx);
    this.emit(bookId, statusFromMeta(null, false, null));
  }

  private async runDownload(
    bookId: number,
    streamUrl: string,
    fileSize: number,
    bookModifiedAt: string,
    mimeType: string,
    comicPages?: { total: number; pageUrl: (page: number) => string },
  ): Promise<void> {
    this.stop(bookId);
    let meta = await getMeta(bookId);
    if (
      meta &&
      (meta.url !== streamUrl ||
        meta.bookModifiedAt !== bookModifiedAt ||
        meta.fileSize !== fileSize)
    ) {
      await this.clear(bookId);
      meta = null;
    }

    const controller = new AbortController();
    this.jobs.set(bookId, controller);
    this.jobError.set(bookId, null);

    if (!meta?.complete) {
      meta = {
        bookId,
        url: streamUrl,
        fileSize,
        bookModifiedAt,
        downloaded: 0,
        complete: false,
        mimeType,
      };
      await putMeta(meta);
      this.emit(bookId, statusFromMeta(meta, true, null));

      try {
        const res = await fetch(streamUrl, {
          credentials: "same-origin",
          signal: controller.signal,
        });
        if (!res.ok) throw new Error(`download failed (${res.status})`);
        const blob = await res.blob();
        await putFileBlob(bookId, blob);
        meta = {
          ...meta,
          downloaded: blob.size || fileSize,
          complete: true,
          mimeType: blob.type || mimeType,
        };
        await putMeta(meta);
        this.emit(bookId, statusFromMeta(meta, true, null));
      } catch (err) {
        if (!controller.signal.aborted) {
          const message = err instanceof Error ? err.message : "download failed";
          this.jobError.set(bookId, message);
          this.emit(bookId, statusFromMeta(await getMeta(bookId), false, message));
        }
        this.jobs.delete(bookId);
        return;
      }
    }

    if (comicPages && comicPages.total > 0 && !controller.signal.aborted) {
      try {
        for (let page = 0; page < comicPages.total; page++) {
          if (controller.signal.aborted) break;
          const existing = await getPageBlob(bookId, page);
          if (existing) continue;
          const res = await fetch(comicPages.pageUrl(page), {
            credentials: "same-origin",
            signal: controller.signal,
          });
          if (!res.ok) continue;
          const blob = await res.blob();
          await putPageBlob(bookId, page, blob);
        }
        this.pagesDone.set(bookId, true);
      } catch {
        // pages are best-effort
      }
    } else if (!comicPages) {
      this.pagesDone.set(bookId, true);
    }

    this.jobs.delete(bookId);
    const finalMeta = await getMeta(bookId);
    this.emit(
      bookId,
      statusFromMeta(
        finalMeta,
        false,
        this.jobError.get(bookId) ?? null,
        this.pagesDone.get(bookId) ?? false,
      ),
    );
  }
}

export const bookOfflineCache = new BookOfflineCache();
