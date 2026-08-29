import { api } from "$lib/api/client";
import { toast } from "$lib/stores/toast.svelte";
import { SvelteSet } from "svelte/reactivity";

class FavoritesStore {
  ids = $state<SvelteSet<number>>(new SvelteSet());
  loading = $state(false);

  async refresh() {
    this.loading = true;
    try {
      const res = await api.listFavorites();
      this.ids = new SvelteSet(res.ids);
    } catch {
      this.ids = new SvelteSet();
    } finally {
      this.loading = false;
    }
  }

  isFavorite(bookId: number): boolean {
    return this.ids.has(bookId);
  }

  async toggle(bookId: number): Promise<boolean> {
    const next = !this.isFavorite(bookId);
    try {
      await api.setFavorite(bookId, next);
      const ids = new SvelteSet(this.ids);
      if (next) ids.add(bookId);
      else ids.delete(bookId);
      this.ids = ids;
      toast.success(next ? "Added to favorites" : "Removed from favorites");
      return next;
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Failed to update favorite");
      return this.isFavorite(bookId);
    }
  }
}

export const favorites = new FavoritesStore();
