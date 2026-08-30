import { describe, expect, it } from "vitest";
import { EPUB_SECTION_CSP, hardenEpubDocument } from "./epub-harden";

describe("hardenEpubDocument", () => {
  it("strips scripts and injects CSP", () => {
    const doc = new DOMParser().parseFromString(
      `<html><head><link rel="stylesheet" href="styles.css"><meta http-equiv="refresh" content="0;url=https://evil"></head><body><p>Hi</p><script>alert(1)</script><svg><circle r="1"/></svg><a href="javascript:alert(1)">x</a><a href="https://example.com">y</a></body></html>`,
      "text/html",
    );
    hardenEpubDocument(doc);
    expect(doc.querySelector("script")).toBeNull();
    expect(doc.querySelector('meta[http-equiv="refresh"]')).toBeNull();
    expect(doc.querySelector('link[rel="stylesheet"]')).not.toBeNull();
    expect(doc.querySelector("svg circle")).not.toBeNull();
    expect(doc.querySelector('meta[http-equiv="Content-Security-Policy"]')?.getAttribute("content")).toBe(
      EPUB_SECTION_CSP,
    );
    expect(doc.querySelector('a[href="javascript:alert(1)"]')).toBeNull();
    const external = doc.querySelector('a[href="https://example.com"]');
    expect(external?.getAttribute("rel")).toContain("noopener");
    expect(doc.body.textContent).toContain("Hi");
  });
});
