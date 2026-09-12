import { describe, expect, it } from "vitest";
import { LETTER_FILTER_OPTIONS, nextLetterFilter } from "./letter-filter";

describe("LETTER_FILTER_OPTIONS", () => {
  it("lists A-Z then the non-alpha bucket", () => {
    expect(LETTER_FILTER_OPTIONS).toHaveLength(27);
    expect(LETTER_FILTER_OPTIONS[0]).toBe("A");
    expect(LETTER_FILTER_OPTIONS[25]).toBe("Z");
    expect(LETTER_FILTER_OPTIONS[26]).toBe("#");
  });
});

describe("nextLetterFilter", () => {
  it("selects a letter when none is active", () => {
    expect(nextLetterFilter("", "M")).toBe("M");
  });

  it("switches to a different letter", () => {
    expect(nextLetterFilter("A", "Z")).toBe("Z");
  });

  it("clears when the active letter is picked again", () => {
    expect(nextLetterFilter("Q", "Q")).toBe("");
    expect(nextLetterFilter("#", "#")).toBe("");
  });
});
