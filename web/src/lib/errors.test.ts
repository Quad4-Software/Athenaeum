import { describe, expect, it } from "vitest";
import { errorCodeFromSlug } from "./errors";

describe("errorCodeFromSlug", () => {
  it("maps known slugs", () => {
    expect(errorCodeFromSlug("unauthorized")).toBe(401);
    expect(errorCodeFromSlug("forbidden")).toBe(403);
    expect(errorCodeFromSlug("not-found")).toBe(404);
    expect(errorCodeFromSlug("server")).toBe(500);
    expect(errorCodeFromSlug("offline")).toBe("offline");
  });

  it("parses numeric status codes", () => {
    expect(errorCodeFromSlug("502")).toBe(502);
    expect(errorCodeFromSlug("503")).toBe(503);
  });

  it("falls back to 404 for unknown or low values", () => {
    expect(errorCodeFromSlug("unknown")).toBe(404);
    expect(errorCodeFromSlug("200")).toBe(404);
    expect(errorCodeFromSlug("")).toBe(404);
  });
});
