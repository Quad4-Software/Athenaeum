import { describe, expect, it, vi } from "vitest";
import { detectBrowserLocale } from "$lib/i18n/core";
import { loadBundledLocales } from "$lib/i18n";

describe("loadBundledLocales", () => {
  it("loads en bundle", async () => {
    const locales = await loadBundledLocales();
    expect(locales.has("en")).toBe(true);
    expect(locales.get("en")?.messages["nav.allBooks"]).toBeTruthy();
  });
});

describe("detectBrowserLocale", () => {
  it("picks base language from navigator", () => {
    vi.stubGlobal("navigator", { language: "de-DE", languages: ["de-DE"] });
    expect(detectBrowserLocale(["en", "de"])).toBe("de");
    vi.unstubAllGlobals();
  });
});
