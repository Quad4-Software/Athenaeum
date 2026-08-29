import { describe, expect, it, vi, beforeEach, afterEach } from "vitest";
import { cleanup, render, screen, within } from "@testing-library/svelte";
import userEvent from "@testing-library/user-event";
import LanguagePicker from "./LanguagePicker.svelte";

vi.mock("$lib/stores/i18n.svelte", () => ({
  i18n: {
    locale: "en",
    locales: [
      { code: "en", name: "English", source: "bundled" },
      { code: "de", name: "Deutsch", source: "custom" },
      { code: "ja", name: "日本語", source: "bundled" },
    ],
    t: (key: string) =>
      (
        ({
          "language.select": "Select language",
          "language.label": "Language",
          "language.search": "Search languages",
          "language.noMatches": "No matching languages",
        }) as Record<string, string>
      )[key] ?? key,
    setLocale: vi.fn(),
  },
}));

describe("LanguagePicker", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  afterEach(() => {
    cleanup();
  });

  async function openMenu() {
    const user = userEvent.setup();
    render(LanguagePicker);
    await user.click(screen.getByRole("button", { name: "Select language" }));
    const menus = screen.getAllByRole("menu");
    const menu = menus[menus.length - 1];
    return { user, menu };
  }

  it("opens language menu with search and native names", async () => {
    const { menu } = await openMenu();
    expect(within(menu).getByRole("searchbox", { name: "Search languages" })).toBeInTheDocument();
    expect(within(menu).getByText("English")).toBeInTheDocument();
    expect(within(menu).getByText("Deutsch")).toBeInTheDocument();
    expect(within(menu).getByText("日本語")).toBeInTheDocument();
  });

  it("filters languages by query", async () => {
    const { user, menu } = await openMenu();
    await user.type(within(menu).getByRole("searchbox", { name: "Search languages" }), "deu");
    expect(within(menu).getByText("Deutsch")).toBeInTheDocument();
    expect(within(menu).queryByText("日本語")).not.toBeInTheDocument();
  });
});
