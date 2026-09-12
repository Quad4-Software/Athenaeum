import { describe, expect, it } from "vitest";
import { confirmNameMatches } from "./confirm-name";

describe("confirmNameMatches", () => {
  it("requires an exact case-sensitive match", () => {
    expect(confirmNameMatches("Audiobooks", "Audiobooks")).toBe(true);
    expect(confirmNameMatches("audiobooks", "Audiobooks")).toBe(false);
  });

  it("does not trim surrounding whitespace", () => {
    expect(confirmNameMatches(" Audiobooks ", "Audiobooks")).toBe(false);
  });

  it("rejects empty and partial input", () => {
    expect(confirmNameMatches("", "Audiobooks")).toBe(false);
    expect(confirmNameMatches("Audio", "Audiobooks")).toBe(false);
  });
});
