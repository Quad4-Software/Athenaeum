import { untrack } from "svelte";
import { api } from "$lib/api/client";
import { toast } from "$lib/stores/toast.svelte";
import { scan } from "$lib/stores/scan.svelte";
import type { Book, BookFormat, LibraryStats, SortKey } from "$lib/api/types";

const PAGE_SIZE = 60;
const MAX_RETAINED_BOOKS = 240;
import { storageKey } from "$lib/brand/storage";
import { PersistedState, useDebounce } from "runed";

const SORT_KEY = storageKey("sort");
const SEARCH_DEBOUNCE_MS = 250;

function isSortKey(value: string): value is SortKey {
  return (
    value === "recent" ||
    value === "oldest" ||
    value === "title" ||
    value === "author" ||
    value === "progress"
  );
}

class LibraryStore {
  books = $state<Book[]>([]);
  total = $state(0);
  stats = $state<LibraryStats | null>(null);
  seriesList = $state<{ name: string; count: number }[]>([]);
  authorList = $state<{ name: string; count: number }[]>([]);

  #sortState = new PersistedState<SortKey>(SORT_KEY, "recent", {
    serializer: {
      // Stored as a bare sort key string, not JSON.
      serialize: (value) => value,
      deserialize: (value) => (isSortKey(value) ? value : "recent"),
    },
  });

  search = $state("");
  formatFilter = $state<BookFormat | "">("");
  seriesFilter = $state("");
  authorFilter = $state("");
  collectionFilter = $state<number | null>(null);
  libraryFilter = $state<number | null>(null);
  favoritesFilter = $state(false);
  inProgressFilter = $state(false);
  tagFilter = $state("");

  loading = $state(false);
  error = $state<string | null>(null);

  private offset = 0;
  private scheduleSearchRefresh = useDebounce(
    () => void this.refresh({ facets: false }),
    SEARCH_DEBOUNCE_MS,
  );

  get sort(): SortKey {
    return this.#sortState.current;
  }

  set sort(sort: SortKey) {
    this.#sortState.current = sort;
  }

  get hasMore(): boolean {
    return this.books.length < this.total;
  }

  releaseMemory() {
    if (this.books.length <= PAGE_SIZE) return;
    this.books = this.books.slice(0, PAGE_SIZE);
    this.offset = PAGE_SIZE;
  }

  get hasActiveFilters(): boolean {
    return (
      !!this.search ||
      !!this.formatFilter ||
      !!this.seriesFilter ||
      !!this.authorFilter ||
      this.collectionFilter != null ||
      this.libraryFilter != null ||
      this.favoritesFilter ||
      this.inProgressFilter ||
      !!this.tagFilter
    );
  }

  async refresh(opts?: { background?: boolean; facets?: boolean }) {
    this.offset = 0;
    await this.load(true, opts?.background ?? false);
    void this.loadStats();
    if (opts?.facets ?? true) {
      void this.loadSeries();
      void this.loadAuthors();
    }
  }

  async loadMore() {
    if (this.loading || !this.hasMore) return;
    await this.load(false);
  }

  private queryParams() {
    const sort = this.inProgressFilter ? "progress" : this.sort;
    return {
      search: this.search,
      sort,
      format: this.formatFilter,
      series: this.seriesFilter || undefined,
      author: this.authorFilter || undefined,
      library: this.libraryFilter ?? undefined,
      collection: this.collectionFilter ?? undefined,
      favorites: this.favoritesFilter || undefined,
      inProgress: this.inProgressFilter || undefined,
      tag: this.tagFilter || undefined,
      limit: PAGE_SIZE,
      offset: this.offset,
    };
  }

  private async load(reset: boolean, background = false) {
    if (!background || this.books.length === 0) {
      this.loading = true;
    }
    this.error = null;
    try {
      const page = await api.listBooks(this.queryParams());
      this.total = page.total;
      const merged = reset ? page.items : [...this.books, ...page.items];
      this.books = merged.length > MAX_RETAINED_BOOKS ? merged.slice(-MAX_RETAINED_BOOKS) : merged;
      this.offset = reset ? page.items.length : this.offset + page.items.length;
    } catch (err) {
      this.error = err instanceof Error ? err.message : "Failed to load library";
      toast.error(this.error);
    } finally {
      this.loading = false;
    }
  }

  async loadStats() {
    try {
      this.stats = await api.stats(this.libraryFilter ?? undefined);
      if (this.stats?.scanning) {
        scan.startPolling(() => {
          toast.success("Library scan finished");
          void this.refresh({ background: true });
        });
      }
    } catch {
      // stats are non-critical
    }
  }

  async loadSeries() {
    try {
      this.seriesList = await api.listSeries(this.libraryFilter ?? undefined);
    } catch {
      this.seriesList = [];
    }
  }

  async loadAuthors() {
    try {
      this.authorList = await api.listAuthors(this.libraryFilter ?? undefined);
    } catch {
      this.authorList = [];
    }
  }

  setSearch(value: string) {
    this.search = value;
    // debounced() touches runed $state internally; untrack so callers inside
    // $effect are not invalidated when the timer context mutates.
    untrack(() => void this.scheduleSearchRefresh().catch(() => undefined));
  }

  setSort(sort: SortKey) {
    this.sort = sort;
    void this.refresh({ facets: false });
  }

  setLibrary(id: number | null) {
    this.libraryFilter = id;
    this.seriesFilter = "";
    this.authorFilter = "";
    this.collectionFilter = null;
    this.favoritesFilter = false;
    this.inProgressFilter = false;
    this.formatFilter = "";
    this.tagFilter = "";
    void this.refresh();
  }

  setFormat(format: BookFormat | "") {
    this.formatFilter = format;
    this.seriesFilter = "";
    this.authorFilter = "";
    this.collectionFilter = null;
    this.favoritesFilter = false;
    this.inProgressFilter = false;
    this.tagFilter = "";
    void this.refresh({ facets: false });
  }

  setSeries(name: string) {
    this.seriesFilter = name;
    this.formatFilter = "";
    this.authorFilter = "";
    this.collectionFilter = null;
    this.favoritesFilter = false;
    this.inProgressFilter = false;
    this.tagFilter = "";
    void this.refresh({ facets: false });
  }

  setAuthor(name: string) {
    this.authorFilter = name;
    this.formatFilter = "";
    this.seriesFilter = "";
    this.collectionFilter = null;
    this.favoritesFilter = false;
    this.inProgressFilter = false;
    this.tagFilter = "";
    void this.refresh({ facets: false });
  }

  setCollection(id: number) {
    this.collectionFilter = id;
    this.seriesFilter = "";
    this.authorFilter = "";
    this.formatFilter = "";
    this.favoritesFilter = false;
    this.inProgressFilter = false;
    this.tagFilter = "";
    void this.refresh({ facets: false });
  }

  setFavorites(enabled = true) {
    this.favoritesFilter = enabled;
    this.seriesFilter = "";
    this.authorFilter = "";
    this.collectionFilter = null;
    this.formatFilter = "";
    this.libraryFilter = null;
    this.inProgressFilter = false;
    this.tagFilter = "";
    void this.refresh({ facets: false });
  }

  setInProgress(enabled = true) {
    this.inProgressFilter = enabled;
    this.seriesFilter = "";
    this.authorFilter = "";
    this.collectionFilter = null;
    this.formatFilter = "";
    this.libraryFilter = null;
    this.favoritesFilter = false;
    this.tagFilter = "";
    void this.refresh({ facets: false });
  }

  setTag(name: string) {
    this.tagFilter = name;
    this.seriesFilter = "";
    this.authorFilter = "";
    this.collectionFilter = null;
    this.formatFilter = "";
    this.favoritesFilter = false;
    this.inProgressFilter = false;
    void this.refresh({ facets: false });
  }

  clearFilters() {
    this.formatFilter = "";
    this.seriesFilter = "";
    this.authorFilter = "";
    this.collectionFilter = null;
    this.libraryFilter = null;
    this.favoritesFilter = false;
    this.inProgressFilter = false;
    this.tagFilter = "";
    void this.refresh({ facets: false });
  }

  async triggerScan(libraryId?: number) {
    await api.scan(libraryId ?? this.libraryFilter ?? undefined);
    toast.info("Library scan started");
    scan.startPolling(() => {
      toast.success("Library scan finished");
      void this.refresh({ background: true });
    });
  }
}

export const library = new LibraryStore();
