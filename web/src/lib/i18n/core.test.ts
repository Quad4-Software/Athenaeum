import { describe, expect, it } from "vitest";
import {
  detectBrowserLocale,
  flattenMessages,
  interpolate,
  parseLocaleJSON,
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

  it("returns message for key", () => {
    expect(translate(msgs, "nav.allBooks")).toBe("All books");
  });

  it("falls back to key when missing", () => {
    expect(translate(msgs, "missing.key")).toBe("missing.key");
  });

  it("uses fallback catalog", () => {
    expect(translate({}, "nav.allBooks", undefined, msgs)).toBe("All books");
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
