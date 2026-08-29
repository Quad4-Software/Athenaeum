import { describe, expect, it } from "vitest";
import { EPUB_PALETTE, injectEpubContentBackground, resolvedEpubTheme } from "./epub-theme";

describe("resolvedEpubTheme", () => {
  it("keeps explicit light and sepia themes", () => {
    expect(resolvedEpubTheme("light", true)).toBe("light");
    expect(resolvedEpubTheme("sepia", true)).toBe("sepia");
  });

  it("maps dark theme to app dark mode", () => {
    expect(resolvedEpubTheme("dark", true)).toBe("dark");
    expect(resolvedEpubTheme("dark", false)).toBe("dark");
  });

  it("keeps night theme independent of app mode", () => {
    expect(resolvedEpubTheme("night", false)).toBe("night");
  });
});

describe("injectEpubContentBackground", () => {
  it("sets html and body background colors", () => {
    const doc = document.implementation.createHTMLDocument("chapter");
    injectEpubContentBackground(doc, EPUB_PALETTE.sepia.bg);
    expect(doc.documentElement.style.backgroundColor).not.toBe("");
    expect(doc.body.style.backgroundColor).not.toBe("");
  });
});
