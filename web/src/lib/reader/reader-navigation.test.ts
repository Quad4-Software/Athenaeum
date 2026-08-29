import { describe, expect, it } from "vitest";
import {
  adjacentSpreadStartPages,
  clampSpreadStart,
  nextSpreadStart,
  pdfPageCacheKey,
  preloadPageNumbers,
  prevSpreadStart,
  shouldCommitSpreadMount,
  spinePreloadIndices,
  visiblePageNumbers,
} from "./reader-navigation";

describe("visiblePageNumbers", () => {
  it("returns one page for single-page view", () => {
    expect(visiblePageNumbers(5, 1, 10)).toEqual([5]);
  });

  it("returns a spread capped at total pages", () => {
    expect(visiblePageNumbers(9, 2, 10)).toEqual([9, 10]);
    expect(visiblePageNumbers(10, 2, 10)).toEqual([10]);
  });

  it("returns empty when total is zero", () => {
    expect(visiblePageNumbers(1, 1, 0)).toEqual([]);
  });
});

describe("preloadPageNumbers", () => {
  it("preloads previous and next spreads", () => {
    expect(preloadPageNumbers(5, 1, 10)).toEqual([4, 6]);
  });

  it("preloads adjacent spreads for multi-page view", () => {
    expect(preloadPageNumbers(5, 2, 12)).toEqual([3, 4, 7, 8]);
  });

  it("skips preloads at the start and end", () => {
    expect(preloadPageNumbers(1, 1, 5)).toEqual([2]);
    expect(preloadPageNumbers(5, 1, 5)).toEqual([4]);
  });
});

describe("adjacentSpreadStartPages", () => {
  it("returns neighboring spread starts", () => {
    expect(adjacentSpreadStartPages(5, 2, 20)).toEqual([3, 7]);
  });
});

describe("spinePreloadIndices", () => {
  it("returns previous and next spine indices", () => {
    expect(spinePreloadIndices(3, 10)).toEqual([2, 4]);
  });

  it("returns only next at the start", () => {
    expect(spinePreloadIndices(0, 5)).toEqual([1]);
  });

  it("returns only previous at the end", () => {
    expect(spinePreloadIndices(4, 5)).toEqual([3]);
  });

  it("returns empty for invalid input", () => {
    expect(spinePreloadIndices(-1, 5)).toEqual([]);
    expect(spinePreloadIndices(0, 0)).toEqual([]);
  });
});

describe("pdfPageCacheKey", () => {
  it("includes layout parameters", () => {
    expect(pdfPageCacheKey(3, 1.2, true, 400.4, 720.6)).toBe("3|1.2|1|400|721");
  });

  it("distinguishes auto-fit from manual zoom", () => {
    expect(pdfPageCacheKey(1, 1, true, 400, 720)).not.toBe(pdfPageCacheKey(1, 1, false, 400, 720));
  });

  it("rounds fractional viewport dimensions", () => {
    expect(pdfPageCacheKey(1, 1, true, 399.5, 719.5)).toBe("1|1|1|400|720");
  });
});

describe("prevSpreadStart", () => {
  it("steps back by pages per view", () => {
    expect(prevSpreadStart(5, 2)).toBe(3);
    expect(prevSpreadStart(2, 2)).toBe(1);
    expect(prevSpreadStart(1, 1)).toBe(1);
  });
});

describe("nextSpreadStart", () => {
  it("steps forward by pages per view", () => {
    expect(nextSpreadStart(1, 2, 10)).toBe(3);
    expect(nextSpreadStart(9, 2, 10)).toBe(10);
  });

  it("stays put when already at the end", () => {
    expect(nextSpreadStart(10, 1, 10)).toBe(10);
    expect(nextSpreadStart(10, 2, 10)).toBe(10);
  });
});

describe("clampSpreadStart", () => {
  it("keeps spread start valid for multi-page view near the end", () => {
    expect(clampSpreadStart(8, 2, 10)).toBe(8);
    expect(clampSpreadStart(10, 2, 10)).toBe(9);
  });

  it("returns spread unchanged when total is zero", () => {
    expect(clampSpreadStart(4, 2, 0)).toBe(4);
  });
});

describe("shouldCommitSpreadMount", () => {
  it("commits only the active uncancelled mount", () => {
    expect(shouldCommitSpreadMount(2, 2, false)).toBe(true);
    expect(shouldCommitSpreadMount(2, 3, false)).toBe(false);
    expect(shouldCommitSpreadMount(2, 2, true)).toBe(false);
  });
});

describe("visiblePageNumbers edge cases", () => {
  it("supports three-up spreads", () => {
    expect(visiblePageNumbers(4, 3, 20)).toEqual([4, 5, 6]);
    expect(visiblePageNumbers(19, 3, 20)).toEqual([19, 20]);
  });

  it("never returns pages below one", () => {
    expect(visiblePageNumbers(1, 3, 20)).toEqual([1, 2, 3]);
  });
});

describe("preloadPageNumbers edge cases", () => {
  it("never includes currently visible pages", () => {
    const visible = new Set(visiblePageNumbers(5, 2, 12));
    for (const page of preloadPageNumbers(5, 2, 12)) {
      expect(visible.has(page)).toBe(false);
    }
  });

  it("returns empty when total is zero", () => {
    expect(preloadPageNumbers(1, 1, 0)).toEqual([]);
  });
});

describe("adjacentSpreadStartPages boundaries", () => {
  it("returns only next spread at the beginning", () => {
    expect(adjacentSpreadStartPages(1, 1, 10)).toEqual([2]);
  });

  it("returns only previous spread at the end", () => {
    expect(adjacentSpreadStartPages(10, 1, 10)).toEqual([9]);
  });
});
