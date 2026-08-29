import { describe, expect, it } from "vitest";
import {
  BUILTIN_EPUB_FONTS,
  customFontCss,
  isBuiltInFontId,
  revokeCustomFontBlobUrl,
  type StoredCustomFont,
} from "./epub-fonts";

describe("epub-fonts", () => {
  it("recognizes built-in font ids", () => {
    expect(isBuiltInFontId("serif")).toBe(true);
    expect(isBuiltInFontId("custom")).toBe(false);
  });

  it("defines a book default option", () => {
    expect(BUILTIN_EPUB_FONTS[0]?.id).toBe("book");
    expect(BUILTIN_EPUB_FONTS[0]?.family).toBeUndefined();
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
});
