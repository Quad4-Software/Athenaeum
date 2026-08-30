import { describe, expect, it } from "vitest";
import type { RenditionOptions } from "epubjs/types/rendition";
import { ensureEpubDocumentHead, epubOpenOptions, epubRenderOptions } from "./epub-options";

describe("epubOpenOptions", () => {
  it("opens extensionless API URLs as binary EPUB with credentials", () => {
    const opts = epubOpenOptions();
    expect(opts.openAs).toBe("epub");
    expect(opts.requestCredentials).toBe(true);
  });
});

describe("epubRenderOptions", () => {
  it("uses blob URLs and allows scripted content for library EPUBs", () => {
    const opts = epubRenderOptions("auto") as RenditionOptions & { method?: string };
    expect(opts.method).toBe("blobUrl");
    expect(opts.allowScriptedContent).toBe(false);
    expect(opts.spread).toBe("auto");
  });
});

describe("ensureEpubDocumentHead", () => {
  it("adds a head element when spine XHTML has only a body", () => {
    const doc = document.implementation.createDocument(
      "http://www.w3.org/1999/xhtml",
      "html",
      null,
    );
    const body = doc.createElement("body");
    doc.documentElement.appendChild(body);
    ensureEpubDocumentHead(doc);
    expect(doc.getElementsByTagName("head").length).toBe(1);
    expect(doc.documentElement.firstElementChild?.tagName.toLowerCase()).toBe("head");
  });
});
