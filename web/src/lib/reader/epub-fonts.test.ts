import { describe, expect, it } from "vitest";
import {
  BUILTIN_EPUB_FONTS,
  customFontCss,
  isBuiltInFontId,
  revokeCustomFontBlobUrl,
  vendoredFontCss,
  type StoredCustomFont,
} from "./epub-fonts";

describe("epub-fonts", () => {
  it("recognizes built-in font ids", () => {
    expect(isBuiltInFontId("serif")).toBe(true);
    expect(isBuiltInFontId("literata")).toBe(true);
    expect(isBuiltInFontId("custom")).toBe(false);
  });

  it("defines a book default option", () => {
    expect(BUILTIN_EPUB_FONTS[0]?.id).toBe("book");
    expect(BUILTIN_EPUB_FONTS[0]?.family).toBeUndefined();
  });

  it("includes vendored face urls for hosted fonts", () => {
    const literata = BUILTIN_EPUB_FONTS.find((f) => f.id === "literata");
    expect(literata?.face?.family).toBe("Literata Variable");
    expect(literata?.face?.url).toMatch(/literata.*\.woff2/);
  });

  it("builds custom font-face css", () => {
    const stored: StoredCustomFont = {
      fileName: "test.woff2",
      mimeType: "font/woff2",
      data: new Uint8Array([1, 2, 3]).buffer,
    };
    const css = customFontCss(stored);
    expect(css).toContain('font-family:"Athenaeum Custom"');
    expect(css).toContain('format("woff2")');
    revokeCustomFontBlobUrl();
  });

  it("builds vendored font-face css", () => {
    const css = vendoredFontCss({
      family: "Literata Variable",
      url: "/fonts/literata.woff2",
      weight: "200 900",
    });
    expect(css).toContain('font-family:"Literata Variable"');
    expect(css).toContain("/fonts/literata.woff2");
    expect(css).toContain("font-weight:200 900");
  });
});
