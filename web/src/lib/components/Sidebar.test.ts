import { describe, expect, it, vi, beforeEach } from "vitest";
import { render, screen } from "@testing-library/svelte";
import Sidebar from "./Sidebar.svelte";

vi.mock("$lib/router.svelte", () => ({
  router: { current: { name: "library", params: {} }, navigate: vi.fn() },
  link: () => ({ destroy: () => {} }),
}));

vi.mock("$lib/stores/library.svelte", () => ({
  library: {
    search: "",
    formatFilter: "",
    seriesFilter: "",
    authorFilter: "",
    collectionFilter: null,
    libraryFilter: null,
    favoritesFilter: false,
    inProgressFilter: false,
    seriesList: [{ name: "Dune", count: 3 }],
    authorList: [{ name: "Herbert", count: 2 }],
    stats: null,
    clearFilters: vi.fn(),
    setFormat: vi.fn(),
    setFavorites: vi.fn(),
    setInProgress: vi.fn(),
    triggerScan: vi.fn(),
  },
}));

vi.mock("$lib/stores/libraries.svelte", () => ({
  libraries: { items: [], refresh: vi.fn() },
}));

vi.mock("$lib/stores/collections.svelte", () => ({
  collections: {
    readingItems: () => [{ id: 1, name: "Currently reading", bookCount: 2 }],
    shelfItems: () => [{ id: 2, name: "Sci-fi", bookCount: 5, kind: "manual" }],
    refresh: vi.fn(),
  },
}));

vi.mock("$lib/stores/favorites.svelte", () => ({
  favorites: { refresh: vi.fn() },
}));

vi.mock("$lib/api/client", () => ({
  api: { health: vi.fn().mockResolvedValue({ status: "ok", version: "1.0.0" }) },
}));

vi.mock("$lib/stores/sidebar.svelte", () => ({
  sidebarPrefs: {
    visibleSections: () => ["formats", "reading", "series", "shelves"],
    isSectionExpanded: () => true,
    toggleSectionExpanded: vi.fn(),
  },
}));

vi.mock("$lib/stores/i18n.svelte", () => ({
  i18n: {
    t: (key: string) =>
      (
        ({
          "nav.allBooks": "All books",
          "nav.primary": "Primary",
          "app.title": "Athenaeum",
        }) as Record<string, string>
      )[key] ?? key,
  },
}));

describe("Sidebar", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("hides shelves and reading lists when collapsed", () => {
    render(Sidebar, { props: { collapsed: true } });
    expect(screen.queryByText("Sci-fi")).not.toBeInTheDocument();
    expect(screen.queryByText("Currently reading")).not.toBeInTheDocument();
    expect(screen.queryByText("Dune")).not.toBeInTheDocument();
    expect(screen.getByTitle("All books")).toBeInTheDocument();
  });

  it("shows shelves when expanded", () => {
    render(Sidebar, { props: { collapsed: false } });
    expect(screen.getByText("Sci-fi")).toBeInTheDocument();
    expect(screen.getByText("Currently reading")).toBeInTheDocument();
  });
});
