import { api, ApiError } from "$lib/api/client";
import { toast } from "$lib/stores/toast.svelte";
import type { Collection, SmartQuery } from "$lib/api/types";

class CollectionsStore {
  items = $state<Collection[]>([]);
  loading = $state(false);
  error = $state<string | null>(null);

  shelfItems(): Collection[] {
    return this.items.filter((c) => c.kind !== "reading");
  }

  readingItems(): Collection[] {
    return this.items.filter((c) => c.kind === "reading");
  }

  async refresh() {
    this.loading = true;
    this.error = null;
    try {
      this.items = await api.listCollections();
    } catch (err) {
      if (err instanceof ApiError && err.status === 401) {
        this.items = [];
        return;
      }
      this.error = err instanceof Error ? err.message : "Failed to load collections";
      toast.error(this.error);
      this.items = [];
    } finally {
      this.loading = false;
    }
  }

  async createReadingList(name: string, description = "") {
    const c = await api.createCollection(name, description, "reading");
    this.items = [...this.items, c].sort((a, b) => a.name.localeCompare(b.name));
    toast.success(`Created reading list "${c.name}"`);
    return c;
  }

  async create(name: string, description = "") {
    const c = await api.createCollection(name, description, "manual");
    this.items = [...this.items, c].sort((a, b) => a.name.localeCompare(b.name));
    toast.success(`Created shelf "${c.name}"`);
    return c;
  }

  async createSmart(name: string, query: SmartQuery, description = "") {
    const c = await api.createCollection(name, description, "smart", query);
    this.items = [...this.items, c].sort((a, b) => a.name.localeCompare(b.name));
    toast.success(`Created smart shelf "${c.name}"`);
    return c;
  }

  async remove(id: number) {
    const c = this.items.find((x) => x.id === id);
    if (c?.kind === "auto") return;
    await api.deleteCollection(id);
    this.items = this.items.filter((c) => c.id !== id);
    toast.info("Shelf deleted");
  }
}

export const collections = new CollectionsStore();
