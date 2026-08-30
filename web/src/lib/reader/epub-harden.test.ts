import { describe, expect, it } from "vitest";
import { hardenEpubDocument } from "./epub-harden";

describe("hardenEpubDocument", () => {
  it("strips scripts and keeps base, styles, and safe anchors", () => {
    const doc = new DOMParser().parseFromString(
      `<html><head><base href="https://example.test/OEBPS/chap.xhtml"><link rel="stylesheet" href="styles.css"><link rel="canonical" href="https://example.test/OEBPS/chap.xhtml"><meta http-equiv="refresh" content="0;url=https://evil"></head><body><p>Hi</p><script>alert(1)</script><svg><circle r="1"/></svg><a href="javascript:alert(1)">x</a><a href="https://example.com">y</a><a href="../next.xhtml">z</a></body></html>`,
      "text/html",
    );
    hardenEpubDocument(doc);
    expect(doc.querySelector("script")).toBeNull();
    expect(doc.querySelector('meta[http-equiv="refresh"]')).toBeNull();
    expect(doc.querySelector("base")?.getAttribute("href")).toContain("OEBPS/chap.xhtml");
    expect(doc.querySelector('link[rel="stylesheet"]')).not.toBeNull();
    expect(doc.querySelector('link[rel="canonical"]')).not.toBeNull();
    expect(doc.querySelector("svg circle")).not.toBeNull();
    expect(doc.querySelector('meta[http-equiv="Content-Security-Policy"]')).toBeNull();
    expect(doc.querySelector('a[href="javascript:alert(1)"]')).toBeNull();
    const external = doc.querySelector('a[href="https://example.com"]');
    expect(external?.getAttribute("rel")).toContain("noopener");
    expect(doc.querySelector('a[href="../next.xhtml"]')).not.toBeNull();
    expect(doc.body.textContent).toContain("Hi");
  });

  it("strips on* handlers without removing the element", () => {
    const doc = new DOMParser().parseFromString(
      `<html><head></head><body><p onclick="alert(1)" onload="evil()">ok</p></body></html>`,
      "text/html",
    );
    hardenEpubDocument(doc);
    const p = doc.querySelector("p");
    expect(p?.getAttribute("onclick")).toBeNull();
    expect(p?.getAttribute("onload")).toBeNull();
    expect(p?.textContent).toBe("ok");
  });
});
