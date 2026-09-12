import { describe, expect, it } from "vitest";
import { COVER_SIZE_MIN_PX, coverColumnMinPx, isCoverSize } from "./cover-size.svelte";

describe("isCoverSize", () => {
  it("accepts the three sizes and rejects anything else", () => {
    expect(isCoverSize("s")).toBe(true);
    expect(isCoverSize("m")).toBe(true);
    expect(isCoverSize("l")).toBe(true);
    expect(isCoverSize("")).toBe(false);
    expect(isCoverSize("xl")).toBe(false);
  });
});

describe("coverColumnMinPx", () => {
  it("maps each size to its base width", () => {
    expect(coverColumnMinPx("s", false)).toBe(COVER_SIZE_MIN_PX.s);
    expect(coverColumnMinPx("m", false)).toBe(COVER_SIZE_MIN_PX.m);
    expect(coverColumnMinPx("l", false)).toBe(COVER_SIZE_MIN_PX.l);
  });

  it("shrinks the column under compact density", () => {
    expect(coverColumnMinPx("m", true)).toBeLessThan(COVER_SIZE_MIN_PX.m);
    expect(coverColumnMinPx("s", true)).toBeGreaterThan(0);
  });
});
