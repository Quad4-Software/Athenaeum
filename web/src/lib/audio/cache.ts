const DB_NAME = "athenaeum-audio";
const DB_VERSION = 2;
const META_STORE = "meta";
const CHUNK_STORE = "chunks";
export const CHUNK_SIZE = 2 * 1024 * 1024;
const MAX_CACHE_BYTES = 2 * 1024 * 1024 * 1024;

export type CacheKey = string | number;

export interface AudioCacheStatus {
  cachedBytes: number;
  totalBytes: number;
  complete: boolean;
  downloading: boolean;
  error: string | null;
}

interface MetaRecord {
  cacheKey: string;
  url: string;
  fileSize: number;
  bookModifiedAt: string;
  downloaded: number;
  complete: boolean;
}

type StatusListener = (status: AudioCacheStatus) => void;

function normalizeKey(key: CacheKey): string {
  return String(key);
}

function openDB(): Promise<IDBDatabase> {
  return new Promise((resolve, reject) => {
    const req = indexedDB.open(DB_NAME, DB_VERSION);
    req.onupgradeneeded = (event) => {
      const db = req.result;
      const oldVersion = event.oldVersion;
      if (oldVersion < 2 && db.objectStoreNames.contains(META_STORE)) {
        db.deleteObjectStore(META_STORE);
      }
      if (oldVersion < 2 && db.objectStoreNames.contains(CHUNK_STORE)) {
        db.deleteObjectStore(CHUNK_STORE);
      }
      if (!db.objectStoreNames.contains(META_STORE)) {
        db.createObjectStore(META_STORE, { keyPath: "cacheKey" });
      }
      if (!db.objectStoreNames.contains(CHUNK_STORE)) {
        db.createObjectStore(CHUNK_STORE);
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

async function getMeta(cacheKey: string): Promise<MetaRecord | null> {
  const db = await openDB();
  return new Promise((resolve, reject) => {
    const tx = db.transaction(META_STORE, "readonly");
    const req = tx.objectStore(META_STORE).get(cacheKey);
    req.onsuccess = () => resolve((req.result as MetaRecord | undefined) ?? null);
    req.onerror = () => reject(req.error);
  });
}

async function putMeta(meta: MetaRecord): Promise<void> {
  const db = await openDB();
  const tx = db.transaction(META_STORE, "readwrite");
  tx.objectStore(META_STORE).put(meta);
  await txDone(tx);
}

async function putChunk(cacheKey: string, index: number, data: ArrayBuffer): Promise<void> {
  const db = await openDB();
  const tx = db.transaction(CHUNK_STORE, "readwrite");
  tx.objectStore(CHUNK_STORE).put(data, chunkKey(cacheKey, index));
  await txDone(tx);
}

async function readChunk(cacheKey: string, index: number): Promise<ArrayBuffer | null> {
  const db = await openDB();
  return new Promise((resolve, reject) => {
    const tx = db.transaction(CHUNK_STORE, "readonly");
    const req = tx.objectStore(CHUNK_STORE).get(chunkKey(cacheKey, index));
    req.onsuccess = () => resolve((req.result as ArrayBuffer | undefined) ?? null);
    req.onerror = () => reject(req.error);
  });
}

async function deleteKeyChunks(
  db: IDBDatabase,
  cacheKey: string,
  totalChunks: number,
): Promise<void> {
  const tx = db.transaction(CHUNK_STORE, "readwrite");
  const store = tx.objectStore(CHUNK_STORE);
  for (let i = 0; i < totalChunks; i++) {
    store.delete(chunkKey(cacheKey, i));
  }
  await txDone(tx);
}

function chunkKey(cacheKey: string, index: number): string {
  return `${cacheKey}#${index}`;
}

function chunkCount(fileSize: number): number {
  return Math.ceil(fileSize / CHUNK_SIZE);
}

function statusFromMeta(
  meta: MetaRecord | null,
  downloading: boolean,
  error: string | null,
): AudioCacheStatus {
  if (!meta) {
    return { cachedBytes: 0, totalBytes: 0, complete: false, downloading, error };
  }
  return {
    cachedBytes: meta.downloaded,
    totalBytes: meta.fileSize,
    complete: meta.complete,
    downloading,
    error,
  };
}

async function fetchChunk(
  url: string,
  start: number,
  end: number,
  signal: AbortSignal,
): Promise<ArrayBuffer> {
  const res = await fetch(url, {
    credentials: "same-origin",
    headers: { Range: `bytes=${start}-${end}` },
    signal,
  });
  if (res.status !== 206 && res.status !== 200) {
    throw new Error(`range request failed (${res.status})`);
  }
  return res.arrayBuffer();
}

export class AudioCacheManager {
  private jobs = new Map<string, AbortController>();
  private blobUrls = new Map<string, string>();
  private listeners = new Map<string, Set<StatusListener>>();
  private jobError = new Map<string, string | null>();

  subscribe(key: CacheKey, listener: StatusListener): () => void {
    const cacheKey = normalizeKey(key);
    let set = this.listeners.get(cacheKey);
    if (!set) {
      set = new Set();
      this.listeners.set(cacheKey, set);
    }
    set.add(listener);
    void this.getStatus(cacheKey).then(listener);
    return () => set?.delete(listener);
  }

  private emit(cacheKey: string, status: AudioCacheStatus) {
    for (const listener of this.listeners.get(cacheKey) ?? []) listener(status);
  }

  async getStatus(key: CacheKey): Promise<AudioCacheStatus> {
    const cacheKey = normalizeKey(key);
    const meta = await getMeta(cacheKey);
    return statusFromMeta(meta, this.jobs.has(cacheKey), this.jobError.get(cacheKey) ?? null);
  }

  /** Aggregate status for a book across all track keys (bookId and bookId:N). */
  async getBookStatus(bookId: number, trackCount = 1): Promise<AudioCacheStatus> {
    if (trackCount <= 1) return this.getStatus(bookId);
    let cached = 0;
    let total = 0;
    let complete = true;
    let downloading = false;
    let error: string | null = null;
    for (let i = 0; i < trackCount; i++) {
      const st = await this.getStatus(`${bookId}:${i}`);
      cached += st.cachedBytes;
      total += st.totalBytes;
      if (!st.complete) complete = false;
      if (st.downloading) downloading = true;
      if (st.error && !error) error = st.error;
    }
    return { cachedBytes: cached, totalBytes: total, complete, downloading, error };
  }

  async isPlayableOffline(key: CacheKey, url: string, bookModifiedAt: string): Promise<boolean> {
    const meta = await getMeta(normalizeKey(key));
    return !!meta?.complete && meta.url === url && meta.bookModifiedAt === bookModifiedAt;
  }

  async getBlobUrl(key: CacheKey): Promise<string | null> {
    const cacheKey = normalizeKey(key);
    const existing = this.blobUrls.get(cacheKey);
    if (existing) return existing;

    const meta = await getMeta(cacheKey);
    if (!meta?.complete) return null;

    const parts: BlobPart[] = [];
    const totalChunks = chunkCount(meta.fileSize);
    for (let i = 0; i < totalChunks; i++) {
      const buf = await readChunk(cacheKey, i);
      if (!buf) return null;
      const start = i * CHUNK_SIZE;
      const end = Math.min(start + buf.byteLength, meta.fileSize);
      parts.push(buf.slice(0, end - start));
    }
    const blob = new Blob(parts);
    const objectUrl = URL.createObjectURL(blob);
    this.blobUrls.set(cacheKey, objectUrl);
    return objectUrl;
  }

  async resolvePlaybackUrl(
    key: CacheKey,
    streamUrl: string,
    fileSize: number,
    bookModifiedAt: string,
  ): Promise<string> {
    const offline = await this.isPlayableOffline(key, streamUrl, bookModifiedAt);
    if (offline) {
      const blobUrl = await this.getBlobUrl(key);
      if (blobUrl) return blobUrl;
    }
    void fileSize;
    return streamUrl;
  }

  startPrefetch(key: CacheKey, streamUrl: string, fileSize: number, bookModifiedAt: string): void {
    if (fileSize <= 0 || fileSize > MAX_CACHE_BYTES) return;
    void this.runPrefetch(normalizeKey(key), streamUrl, fileSize, bookModifiedAt);
  }

  stopPrefetch(key: CacheKey): void {
    const cacheKey = normalizeKey(key);
    this.jobs.get(cacheKey)?.abort();
    this.jobs.delete(cacheKey);
  }

  async clear(key: CacheKey): Promise<void> {
    const cacheKey = normalizeKey(key);
    this.stopPrefetch(cacheKey);
    const blobUrl = this.blobUrls.get(cacheKey);
    if (blobUrl) URL.revokeObjectURL(blobUrl);
    this.blobUrls.delete(cacheKey);
    this.jobError.delete(cacheKey);

    const meta = await getMeta(cacheKey);
    const db = await openDB();
    if (meta) {
      await deleteKeyChunks(db, cacheKey, chunkCount(meta.fileSize));
    }
    const tx = db.transaction(META_STORE, "readwrite");
    tx.objectStore(META_STORE).delete(cacheKey);
    await txDone(tx);
    this.emit(cacheKey, statusFromMeta(null, false, null));
  }

  async clearBook(bookId: number, trackCount = 1): Promise<void> {
    await this.clear(bookId);
    for (let i = 0; i < Math.max(trackCount, 1); i++) {
      await this.clear(`${bookId}:${i}`);
    }
  }

  release(key: CacheKey): void {
    const cacheKey = normalizeKey(key);
    this.stopPrefetch(cacheKey);
    const blobUrl = this.blobUrls.get(cacheKey);
    if (blobUrl) URL.revokeObjectURL(blobUrl);
    this.blobUrls.delete(cacheKey);
  }

  releaseBook(bookId: number, trackCount = 1): void {
    this.release(bookId);
    for (let i = 0; i < Math.max(trackCount, 1); i++) {
      this.release(`${bookId}:${i}`);
    }
  }

  private async runPrefetch(
    cacheKey: string,
    streamUrl: string,
    fileSize: number,
    bookModifiedAt: string,
  ): Promise<void> {
    this.stopPrefetch(cacheKey);

    let meta = await getMeta(cacheKey);
    if (
      meta &&
      (meta.url !== streamUrl ||
        meta.bookModifiedAt !== bookModifiedAt ||
        meta.fileSize !== fileSize)
    ) {
      await this.clear(cacheKey);
      meta = null;
    }

    if (!meta) {
      meta = {
        cacheKey,
        url: streamUrl,
        fileSize,
        bookModifiedAt,
        downloaded: 0,
        complete: false,
      };
      await putMeta(meta);
    }

    if (meta.complete) {
      this.emit(cacheKey, statusFromMeta(meta, false, null));
      return;
    }

    const controller = new AbortController();
    this.jobs.set(cacheKey, controller);
    this.jobError.set(cacheKey, null);
    this.emit(cacheKey, statusFromMeta(meta, true, null));

    let downloaded = meta.downloaded;
    const totalChunks = chunkCount(fileSize);
    let chunkIndex = Math.floor(downloaded / CHUNK_SIZE);

    while (chunkIndex < totalChunks && !controller.signal.aborted) {
      const start = chunkIndex * CHUNK_SIZE;
      const end = Math.min(start + CHUNK_SIZE - 1, fileSize - 1);
      try {
        if (!navigator.onLine) {
          throw new Error("offline");
        }
        const data = await fetchChunk(streamUrl, start, end, controller.signal);
        await putChunk(cacheKey, chunkIndex, data);
        downloaded = end + 1;
        chunkIndex += 1;
        const next: MetaRecord = {
          cacheKey,
          url: streamUrl,
          fileSize,
          bookModifiedAt,
          downloaded,
          complete: downloaded >= fileSize,
        };
        await putMeta(next);
        this.emit(cacheKey, statusFromMeta(next, true, null));
        if (next.complete) break;
      } catch (err) {
        if (controller.signal.aborted) break;
        const message = err instanceof Error ? err.message : "download failed";
        this.jobError.set(cacheKey, message);
        const current = await getMeta(cacheKey);
        this.emit(cacheKey, statusFromMeta(current, false, message));
        await sleep(3000);
        if (!navigator.onLine) break;
      }
    }

    this.jobs.delete(cacheKey);
    const finalMeta = await getMeta(cacheKey);
    this.emit(cacheKey, statusFromMeta(finalMeta, false, this.jobError.get(cacheKey) ?? null));
  }
}

function sleep(ms: number): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

export const audioCache = new AudioCacheManager();
