import { describe, expect, it } from "vitest";
import * as fc from "fast-check";
import { formatBytes, seriesLabel } from "./format";
import { scorePassword } from "./password-strength";
import { descriptionLooksLikeHtml, sanitizeHtml } from "./sanitize-html";

describe("property: formatBytes", () => {
  it("never throws and returns a string", () => {
    fc.assert(
      fc.property(fc.double(), (n) => {
        const out = formatBytes(n);
        expect(typeof out).toBe("string");
        expect(out.length).toBeGreaterThan(0);
      }),
      { numRuns: 200 },
    );
  });

  it("formats finite non-negative values without dash", () => {
    fc.assert(
      fc.property(fc.integer({ min: 0, max: Number.MAX_SAFE_INTEGER }), (n) => {
        const out = formatBytes(n);
        expect(out).not.toBe("-");
        expect(out).toMatch(/ B$| KB$| MB$| GB$| TB$/);
      }),
      { numRuns: 150 },
    );
  });
});

describe("property: seriesLabel", () => {
  it("empty series always yields empty label", () => {
    fc.assert(
      fc.property(fc.option(fc.integer(), { nil: undefined }), (index) => {
        expect(seriesLabel(undefined, index)).toBe("");
        expect(seriesLabel("", index)).toBe("");
      }),
      { numRuns: 100 },
    );
  });

  it("non-empty series includes the name", () => {
    fc.assert(
      fc.property(
        fc.string({ minLength: 1 }),
        fc.option(fc.integer(), { nil: undefined }),
        (series, index) => {
          const out = seriesLabel(series, index);
          expect(out.startsWith(series)).toBe(true);
        },
      ),
      { numRuns: 100 },
    );
  });
});

describe("property: scorePassword", () => {
  it("is deterministic", () => {
    fc.assert(
      fc.property(fc.string(), (pw) => {
        const a = scorePassword(pw);
        const b = scorePassword(pw);
        expect(a).toEqual(b);
      }),
      { numRuns: 200 },
    );
  });

  it("short passwords are never valid", () => {
    fc.assert(
      fc.property(fc.string({ maxLength: 7 }), (pw) => {
        expect(scorePassword(pw).valid).toBe(false);
      }),
      { numRuns: 100 },
    );
  });
});

describe("property: sanitizeHtml", () => {
  it("never emits script tags or javascript urls for any input", () => {
    fc.assert(
      fc.property(fc.string(), (input) => {
        const out = sanitizeHtml(input).toLowerCase();
        expect(out).not.toContain("<script");
        expect(out).not.toMatch(/javascript:/);
      }),
      { numRuns: 200 },
    );
  });

  it("drops script payloads even when nested in broken markup", () => {
    fc.assert(
      fc.property(fc.string(), (input) => {
        const out = sanitizeHtml(`<p>${input}</p><script>alert(1)</script>`).toLowerCase();
        expect(out).not.toContain("<script");
      }),
      { numRuns: 100 },
    );
  });

  it("descriptionLooksLikeHtml matches angle-bracket tags", () => {
    fc.assert(
      fc.property(fc.constantFrom("p", "div", "span", "strong", "em"), (tag) => {
        expect(descriptionLooksLikeHtml(`<${tag}>`)).toBe(true);
      }),
      { numRuns: 20 },
    );
  });
});
