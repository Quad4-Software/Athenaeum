import { describe, expect, it } from "vitest";
import { apiOp, apiOperations, apiPath } from "./paths";

describe("generated api paths", () => {
  it("includes health and openapi operations", () => {
    expect(apiOperations.length).toBeGreaterThan(20);
    expect(apiOp("GET__api_health").path).toBe("/api/health");
    expect(apiPath("/api/books/{id}", { id: 42 })).toBe("/api/books/42");
  });

  it("rejects missing path params", () => {
    expect(() => apiPath("/api/books/{id}", {})).toThrow(/missing path param/);
  });
});
