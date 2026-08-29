import { describe, expect, it } from "vitest";
import { PdfPageCache, readerViewportFillColor } from "./pdf-page-cache";

function createRenderedPage(pageNum: number): HTMLDivElement {
  const wrap = document.createElement("div");
  wrap.className = "pdf-page";
  wrap.dataset.page = String(pageNum);
  wrap.style.width = "400px";
  wrap.style.height = "560px";

  const canvas = document.createElement("canvas");
  canvas.width = 100;
  canvas.height = 140;
  canvas.style.width = "400px";
  canvas.style.height = "560px";

  const layer = document.createElement("div");
  layer.className = "pdf-text-layer";
  layer.innerHTML = `<span>page ${pageNum}</span>`;

  wrap.appendChild(canvas);
  wrap.appendChild(layer);
  return wrap;
}

describe("PdfPageCache", () => {
  it("returns null for missing keys", () => {
    const cache = new PdfPageCache();
    expect(cache.take("missing")).toBeNull();
    expect(cache.has("missing")).toBe(false);
  });

  it("stores page wraps by reference", () => {
    const cache = new PdfPageCache();
    const wrap = createRenderedPage(2);

    cache.set("page-2", wrap);
    const taken = cache.take("page-2");

    expect(taken).toBe(wrap);
    expect(taken?.className).toBe("pdf-page");
    expect(taken?.dataset.page).toBe("2");
    expect(taken?.querySelector(".pdf-text-layer")?.innerHTML).toContain("page 2");
  });

  it("refreshes an existing cache entry", () => {
    const cache = new PdfPageCache();
    cache.set("page-1", createRenderedPage(1));
    cache.set("page-1", createRenderedPage(11));

    const taken = cache.take("page-1");
    expect(taken?.querySelector(".pdf-text-layer")?.innerHTML).toContain("page 11");
  });

  it("takes cached entries and removes them from the cache", () => {
    const cache = new PdfPageCache();
    const wrap = createRenderedPage(7);
    cache.set("page-7", wrap);

    const taken = cache.take("page-7");
    expect(taken).toBe(wrap);
    expect(cache.has("page-7")).toBe(false);
    expect(cache.take("page-7")).toBeNull();
  });

  it("evicts oldest entries when over capacity", () => {
    const cache = new PdfPageCache(2);
    const first = createRenderedPage(1);
    cache.set("a", first);
    cache.set("b", createRenderedPage(2));
    cache.set("c", createRenderedPage(3));

    expect(cache.has("a")).toBe(false);
    expect(first.querySelector("canvas")?.width).toBe(0);
    expect(cache.has("b")).toBe(true);
    expect(cache.has("c")).toBe(true);
  });

  it("clears all entries and releases canvas memory", () => {
    const cache = new PdfPageCache();
    const wrap = createRenderedPage(1);
    cache.set("a", wrap);
    cache.clear();
    expect(cache.has("a")).toBe(false);
    expect(wrap.querySelector("canvas")?.width).toBe(0);
  });
});

describe("readerViewportFillColor", () => {
  it("returns css variable or fallback", () => {
    expect(readerViewportFillColor()).toMatch(/^#|rgb|oklch/);
  });
});
