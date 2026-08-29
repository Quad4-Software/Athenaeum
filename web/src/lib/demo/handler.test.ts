import { describe, expect, it } from "vitest";
import { handleDemoRequest } from "./handler";
import { DEMO_CATALOG } from "./catalog";
import { demoCoverDataUrl } from "./covers";

describe("demo handler", () => {
  it("lists generated books", async () => {
    const res = await handleDemoRequest("/api/books?limit=50");
    expect(res).not.toBeNull();
    const body = await res!.json();
    expect(body.total).toBe(DEMO_CATALOG.length);
    expect(body.items[0].hasCover).toBe(true);
  });

  it("returns auth setup suitable for public demo", async () => {
    const res = await handleDemoRequest("/api/auth/setup");
    const body = await res!.json();
    expect(body.needed).toBe(false);
    expect(body.authEnabled).toBe(false);
  });

  it("returns metadata matches for a book", async () => {
    const res = await handleDemoRequest("/api/books/1/metadata/search", {
      method: "POST",
      body: JSON.stringify({ title: "The Ember Protocol" }),
    });
    const body = await res!.json();
    expect(body.matches.length).toBeGreaterThan(0);
    expect(body.matches[0].coverUrl).toContain("covers.openlibrary.org");
  });

  it("builds svg cover data urls as fallback", () => {
    const url = demoCoverDataUrl("The Ember Protocol", "Mira Kade");
    expect(url.startsWith("data:image/svg+xml")).toBe(true);
  });

  it("uses Open Library covers for catalog entries", async () => {
    const { demoCoverUrlForBook } = await import("./covers");
    expect(demoCoverUrlForBook(1)).toContain("covers.openlibrary.org");
  });
});
