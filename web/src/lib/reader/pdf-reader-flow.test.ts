import { describe, expect, it } from "vitest";
import { PdfPageCache } from "./pdf-page-cache";
import { buildPreloadPlan, buildVisibleSpreadPlan } from "./pdf-spread-plan";
import { pdfPageCacheKey, preloadPageNumbers, visiblePageNumbers } from "./reader-navigation";

function createRenderedPage(pageNum: number): HTMLDivElement {
  const wrap = document.createElement("div");
  wrap.className = "pdf-page";
  wrap.dataset.page = String(pageNum);
  const canvas = document.createElement("canvas");
  canvas.width = 80;
  canvas.height = 120;
  wrap.appendChild(canvas);
  return wrap;
}

const layout = {
  targetScale: 1,
  autoFit: true,
  slotWidth: 400,
  viewportHeight: 720,
};

function cacheKey(page: number): string {
  return pdfPageCacheKey(
    page,
    layout.targetScale,
    layout.autoFit,
    layout.slotWidth,
    layout.viewportHeight,
  );
}

describe("pdf reader cache navigation flow", () => {
  it("serves preloaded next page from cache on navigation", () => {
    const cache = new PdfPageCache();
    const spreadPage = 1;
    const pagesPerView = 1;
    const total = 10;

    for (const page of preloadPageNumbers(spreadPage, pagesPerView, total)) {
      cache.set(cacheKey(page), createRenderedPage(page));
    }

    const nextSpread = 2;
    const plan = buildVisibleSpreadPlan(nextSpread, pagesPerView, total, layout, (key) =>
      cache.has(key),
    );

    expect(plan).toEqual([{ page: 2, cacheKey: cacheKey(2), fromCache: true }]);
    const restored = cache.take(plan[0].cacheKey);
    expect(restored?.dataset.page).toBe("2");
    expect(restored?.querySelector("canvas")?.width).toBe(80);
  });

  it("serves previous page from cache when navigating back", () => {
    const cache = new PdfPageCache();
    cache.set(cacheKey(1), createRenderedPage(1));
    cache.set(cacheKey(2), createRenderedPage(2));

    const backPlan = buildVisibleSpreadPlan(1, 1, 10, layout, (key) => cache.has(key));
    expect(backPlan[0].fromCache).toBe(true);
    expect(cache.take(backPlan[0].cacheKey)?.dataset.page).toBe("1");
  });

  it("cache misses when zoom changes layout key", () => {
    const cache = new PdfPageCache();
    cache.set(cacheKey(5), createRenderedPage(5));

    const zoomedLayout = { ...layout, targetScale: 1.4 };
    const plan = buildVisibleSpreadPlan(5, 1, 10, zoomedLayout, (key) => cache.has(key));

    expect(plan[0].fromCache).toBe(false);
    expect(plan[0].cacheKey).not.toBe(cacheKey(5));
  });

  it("does not preload pages already cached", () => {
    const cache = new PdfPageCache();
    cache.set(cacheKey(2), createRenderedPage(2));

    const preload = buildPreloadPlan(1, 1, 10, layout, (key) => cache.has(key));
    expect(preload.some((entry) => entry.page === 2)).toBe(false);
    expect(preload.some((entry) => entry.page === 1)).toBe(false);
    expect(preload.some((entry) => entry.page === 3)).toBe(false);
  });

  it("preloads uncached adjacent pages only", () => {
    const cache = new PdfPageCache();
    const preload = buildPreloadPlan(5, 1, 10, layout, (key) => cache.has(key));
    expect(preload.map((entry) => entry.page)).toEqual([4, 6]);
  });

  it("visible spread uses all cache hits for two-up mode", () => {
    const cache = new PdfPageCache();
    cache.set(cacheKey(5), createRenderedPage(5));
    cache.set(cacheKey(6), createRenderedPage(6));

    const plan = buildVisibleSpreadPlan(5, 2, 12, layout, (key) => cache.has(key));
    expect(plan).toEqual([
      { page: 5, cacheKey: cacheKey(5), fromCache: true },
      { page: 6, cacheKey: cacheKey(6), fromCache: true },
    ]);
    expect(visiblePageNumbers(5, 2, 12)).toEqual([5, 6]);
  });
});
