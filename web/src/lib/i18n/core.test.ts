import { describe, expect, it, vi, beforeEach, afterEach } from "vitest";
import {
  clearMissingKeys,
  detectBrowserLocale,
  flattenMessages,
  getMissingKeys,
  interpolate,
  parseLocaleJSON,
  setMissingKeyHandler,
  translate,
} from "./core";

describe("flattenMessages", () => {
  it("flattens nested objects", () => {
    const out = flattenMessages({
      $name: "English",
      nav: { allBooks: "All books" },
    });
    expect(out).toEqual({ "nav.allBooks": "All books" });
  });

  it("rejects non-string leaves", () => {
    expect(flattenMessages({ count: 1 } as Record<string, unknown>)).toBeNull();
  });
});

describe("parseLocaleJSON", () => {
  it("parses valid locale files", () => {
    const parsed = parseLocaleJSON({
      $name: "Deutsch",
      greet: "Hallo",
    });
    expect(parsed?.name).toBe("Deutsch");
    expect(parsed?.messages.greet).toBe("Hallo");
  });

  it("returns null for invalid input", () => {
    expect(parseLocaleJSON(null)).toBeNull();
    expect(parseLocaleJSON([])).toBeNull();
    expect(parseLocaleJSON({ bad: 1 })).toBeNull();
  });
});

describe("translate", () => {
  const msgs = { "nav.allBooks": "All books" };

  beforeEach(() => {
    clearMissingKeys();
    setMissingKeyHandler(null);
  });

  afterEach(() => {
    clearMissingKeys();
    setMissingKeyHandler(null);
  });

  it("returns message for key", () => {
    expect(translate(msgs, "nav.allBooks")).toBe("All books");
    expect(getMissingKeys()).toEqual([]);
  });

  it("falls back to key when missing and records absent", () => {
    expect(translate(msgs, "missing.key")).toBe("missing.key");
    expect(getMissingKeys()).toEqual(["missing.key"]);
  });

  it("uses fallback catalog and records fallback miss", () => {
    expect(translate({}, "nav.allBooks", undefined, msgs)).toBe("All books");
    expect(getMissingKeys()).toEqual(["nav.allBooks"]);
  });

  it("invokes missing key handler once per key", () => {
    const handler = vi.fn();
    setMissingKeyHandler(handler);
    translate(msgs, "a.missing");
    translate(msgs, "a.missing");
    expect(handler).toHaveBeenCalledTimes(1);
    expect(handler).toHaveBeenCalledWith("a.missing", "absent");
  });
});

describe("interpolate", () => {
  it("replaces placeholders", () => {
    expect(interpolate("Hello {name}", { name: "Ada" })).toBe("Hello Ada");
  });
});

describe("detectBrowserLocale", () => {
  it("prefers exact match", () => {
    expect(detectBrowserLocale(["en", "de"])).toBe("en");
  });
});
