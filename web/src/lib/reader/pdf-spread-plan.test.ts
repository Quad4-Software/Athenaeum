import { describe, expect, it } from "vitest";
import { buildPreloadPlan, buildVisibleSpreadPlan, spreadPageCacheKey } from "./pdf-spread-plan";

const layout = {
  targetScale: 1,
  autoFit: true,
  slotWidth: 512,
  viewportHeight: 800,
};

describe("spreadPageCacheKey", () => {
  it("delegates to pdf page cache key rounding", () => {
    expect(spreadPageCacheKey(7, { ...layout, slotWidth: 511.6, viewportHeight: 799.4 })).toBe(
      "7|1|1|512|799",
    );
  });
});

describe("buildVisibleSpreadPlan", () => {
  it("marks cached and uncached visible pages", () => {
    const cached = new Set([spreadPageCacheKey(2, layout), spreadPageCacheKey(3, layout)]);
    const plan = buildVisibleSpreadPlan(2, 2, 10, layout, (key) => cached.has(key));

    expect(plan).toEqual([
      { page: 2, cacheKey: spreadPageCacheKey(2, layout), fromCache: true },
      { page: 3, cacheKey: spreadPageCacheKey(3, layout), fromCache: true },
    ]);
  });

  it("returns empty plan when no pages are visible", () => {
    expect(buildVisibleSpreadPlan(1, 1, 0, layout, () => false)).toEqual([]);
  });
});

describe("buildPreloadPlan", () => {
  it("skips pages already in cache", () => {
    const cached = new Set([spreadPageCacheKey(4, layout)]);
    const plan = buildPreloadPlan(5, 1, 10, layout, (key) => cached.has(key));
    expect(plan.map((entry) => entry.page)).toEqual([6]);
  });

  it("returns empty when all adjacent pages are visible or cached", () => {
    const cached = new Set([spreadPageCacheKey(4, layout), spreadPageCacheKey(6, layout)]);
    expect(buildPreloadPlan(5, 1, 10, layout, (key) => cached.has(key))).toEqual([]);
  });
});
