import { describe, expect, it } from "vitest";
import { nextSearchHistory } from "./search-history";

describe("nextSearchHistory", () => {
  it("prepends the query", () => {
    expect(nextSearchHistory(["b"], "a")).toEqual(["a", "b"]);
  });

  it("dedupes and moves a repeat to the front", () => {
    expect(nextSearchHistory(["a", "b", "c"], "b")).toEqual(["b", "a", "c"]);
  });

  it("trims whitespace and skips blank queries", () => {
    expect(nextSearchHistory([], "  dune  ")).toEqual(["dune"]);
    expect(nextSearchHistory(["a"], "   ")).toEqual(["a"]);
    expect(nextSearchHistory(["a"], "")).toEqual(["a"]);
  });

  it("caps the list at the max", () => {
    const full = Array.from({ length: 10 }, (_, i) => `q${i}`);
    const next = nextSearchHistory(full, "new");
    expect(next).toHaveLength(10);
    expect(next[0]).toBe("new");
    expect(next).not.toContain("q9");
  });
});
