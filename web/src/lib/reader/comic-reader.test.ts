import { describe, expect, it } from "vitest";
import {
  comicFitClass,
  comicSpreadPages,
  comicSpreadStart,
  nextComicPage,
  prevComicPage,
} from "./comic-reader";

describe("comicFitClass", () => {
  it("maps fit modes to sizing classes", () => {
    expect(comicFitClass("contain")).toContain("object-contain");
    expect(comicFitClass("width")).toContain("w-full");
    expect(comicFitClass("height")).toContain("h-full");
  });
});

describe("comicSpreadStart", () => {
  it("keeps the cover page alone", () => {
    expect(comicSpreadStart(0)).toBe(0);
  });

  it("pairs pages after the cover", () => {
    expect(comicSpreadStart(1)).toBe(1);
    expect(comicSpreadStart(2)).toBe(1);
    expect(comicSpreadStart(3)).toBe(3);
    expect(comicSpreadStart(4)).toBe(3);
    expect(comicSpreadStart(5)).toBe(5);
  });
});

describe("comicSpreadPages", () => {
  it("returns a single page when spread mode is off", () => {
    expect(comicSpreadPages(2, 10, false, true)).toEqual([2]);
  });

  it("returns a single page on narrow viewports", () => {
    expect(comicSpreadPages(2, 10, true, false)).toEqual([2]);
  });

  it("returns the cover alone", () => {
    expect(comicSpreadPages(0, 10, true, true)).toEqual([0]);
  });

  it("pairs pages after the cover", () => {
    expect(comicSpreadPages(1, 10, true, true)).toEqual([1, 2]);
    expect(comicSpreadPages(2, 10, true, true)).toEqual([1, 2]);
    expect(comicSpreadPages(3, 10, true, true)).toEqual([3, 4]);
  });

  it("clamps to the last page when the final spread is unpaired", () => {
    expect(comicSpreadPages(3, 4, true, true)).toEqual([3]);
  });

  it("returns empty for no pages", () => {
    expect(comicSpreadPages(0, 0, true, true)).toEqual([]);
  });
});

describe("nextComicPage", () => {
  it("advances by one page in single mode", () => {
    expect(nextComicPage(0, 10, false, true)).toBe(1);
  });

  it("advances by the spread size in spread mode", () => {
    expect(nextComicPage(0, 10, true, true)).toBe(1);
    expect(nextComicPage(1, 10, true, true)).toBe(3);
    expect(nextComicPage(3, 10, true, true)).toBe(5);
  });

  it("stops at the last page", () => {
    expect(nextComicPage(9, 10, true, true)).toBe(9);
    expect(nextComicPage(3, 4, true, true)).toBe(3);
  });
});

describe("prevComicPage", () => {
  it("steps back by one page in single mode", () => {
    expect(prevComicPage(2, 10, false, true)).toBe(1);
  });

  it("steps back to the previous spread start", () => {
    expect(prevComicPage(3, 10, true, true)).toBe(1);
    expect(prevComicPage(1, 10, true, true)).toBe(0);
  });

  it("stays at the cover", () => {
    expect(prevComicPage(0, 10, true, true)).toBe(0);
  });
});
