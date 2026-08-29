import { api, ApiError } from "$lib/api/client";
import { toast } from "$lib/stores/toast.svelte";

const CHUNK_SIZE = 2 * 1024 * 1024;
const MAX_CONCURRENT = 3;

export type UploadStatus = "queued" | "uploading" | "done" | "error";

export interface UploadJob {
  id: string;
  libraryId: number;
  file: File;
  relPath: string;
  status: UploadStatus;
  progress: number;
  error?: string;
  bookId?: number;
}

class UploadQueue {
  #jobs = $state<UploadJob[]>([]);
  #active = 0;

  get jobs() {
    return this.#jobs;
  }

  enqueue(libraryId: number, file: File, relPath?: string) {
    const job: UploadJob = {
      id: crypto.randomUUID(),
      libraryId,
      file,
      relPath: relPath ?? file.name,
      status: "queued",
      progress: 0,
    };
    this.#jobs = [...this.#jobs, job];
    this.#pump();
    return job.id;
  }

  remove(localId: string) {
    const job = this.#jobs.find((j) => j.id === localId);
    if (job?.status === "uploading") return;
    this.#jobs = this.#jobs.filter((j) => j.id !== localId);
  }

  async #pump() {
    while (this.#active < MAX_CONCURRENT) {
      const next = this.#jobs.find((j) => j.status === "queued");
      if (!next) return;
      this.#active++;
      void this.#run(next).finally(() => {
        this.#active--;
        this.#pump();
      });
    }
  }

  async #run(job: UploadJob) {
    this.#update(job.id, { status: "uploading", progress: 0, error: undefined });
    try {
      let session = await api.createUpload(job.libraryId, job.relPath, job.file.size);
      let offset = session.offset;
      while (offset < job.file.size) {
        const end = Math.min(offset + CHUNK_SIZE, job.file.size) - 1;
        const chunk = job.file.slice(offset, end + 1);
        session = await api.uploadChunk(
          job.libraryId,
          session.id,
          chunk,
          offset,
          end,
          job.file.size,
        );
        offset = session.offset;
        this.#update(job.id, {
          progress: job.file.size > 0 ? offset / job.file.size : 1,
        });
      }
      this.#update(job.id, {
        status: "done",
        progress: 1,
        bookId: session.bookId,
      });
      toast.success(`Uploaded ${job.file.name}`);
    } catch (e) {
      const msg = e instanceof ApiError ? e.message : "upload failed";
      this.#update(job.id, { status: "error", error: msg });
      toast.error(`Upload failed: ${job.file.name}`);
    }
  }

  #update(localId: string, patch: Partial<UploadJob>) {
    this.#jobs = this.#jobs.map((j) => (j.id === localId ? { ...j, ...patch } : j));
  }
}

export const uploads = new UploadQueue();
