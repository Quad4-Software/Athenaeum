import { api, ApiError } from "$lib/api/client";
import { apiAction } from "$lib/utils/api-action";
import { toast } from "$lib/stores/toast.svelte";
import type { LibraryCreateInput, LibraryMount, LibraryS3Input } from "$lib/api/types";

class LibrariesStore {
  items = $state<LibraryMount[]>([]);
  loading = $state(false);
  error = $state<string | null>(null);

  async refresh() {
    this.loading = true;
    this.error = null;
    try {
      this.items = await api.listLibraries();
    } catch (err) {
      if (err instanceof ApiError && err.status === 401) {
        this.items = [];
        return;
      }
      this.error = err instanceof Error ? err.message : "Failed to load libraries";
      toast.error(this.error);
      this.items = [];
    } finally {
      this.loading = false;
    }
  }

  async create(input: LibraryCreateInput) {
    const lib = await apiAction(() => api.createLibrary(input), {
      errorFallback: "Failed to add library",
    });
    if (!lib) return undefined;
    this.items = [...this.items, lib].sort((a, b) => a.sortOrder - b.sortOrder);
    toast.success(`Library "${lib.name}" added`);
    return lib;
  }

  async update(id: number, input: LibraryCreateInput) {
    const lib = await api.updateLibrary(id, input);
    this.items = this.items.map((x) => (x.id === id ? lib : x));
    return lib;
  }

  async testS3(s3: LibraryS3Input) {
    return api.testS3(s3);
  }

  async remove(id: number) {
    await api.deleteLibrary(id);
    this.items = this.items.filter((x) => x.id !== id);
  }

  async reorder(ids: number[]) {
    await api.reorderLibraries(ids);
    const byId = new Map(this.items.map((x) => [x.id, x]));
    this.items = ids.map((id, i) => ({ ...byId.get(id)!, sortOrder: i })).filter(Boolean);
  }

  move(id: number, dir: -1 | 1) {
    const idx = this.items.findIndex((x) => x.id === id);
    const next = idx + dir;
    if (idx < 0 || next < 0 || next >= this.items.length) return;
    const ids = this.items.map((x) => x.id);
    [ids[idx], ids[next]] = [ids[next], ids[idx]];
    void this.reorder(ids);
  }
}

export const libraries = new LibrariesStore();
