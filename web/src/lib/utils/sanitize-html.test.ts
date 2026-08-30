import { describe, expect, it } from "vitest";
import { descriptionLooksLikeHtml, sanitizeHtml, sanitizeReaderHtml } from "./sanitize-html";

describe("sanitizeHtml", () => {
  it("strips script tags", () => {
    const out = sanitizeHtml("<p>Hello</p><script>alert(1)</script>");
    expect(out).toContain("Hello");
    expect(out).not.toContain("script");
  });

  it("allows basic formatting", () => {
    const out = sanitizeHtml("<p>Two men possess <strong>vital</strong> information.</p>");
    expect(out).toBe("<p>Two men possess <strong>vital</strong> information.</p>");
  });

  it("blocks javascript hrefs", () => {
    const out = sanitizeHtml('<a href="javascript:alert(1)">x</a>');
    expect(out).not.toContain("javascript:");
  });

  it("blocks spaced javascript hrefs", () => {
    const out = sanitizeHtml('<a href=" javaScript:alert(1)">x</a>');
    expect(out).not.toMatch(/javascript:/i);
  });

  it("strips event handlers on allowed tags", () => {
    const out = sanitizeHtml('<p onclick="alert(1)">hi</p>');
    expect(out).toBe("<p>hi</p>");
  });

  it("drops script even when nested under broken markup", () => {
    const out = sanitizeHtml("<p><A</p><script>alert(1)</script>");
    expect(out.toLowerCase()).not.toContain("<script");
  });

  it("removes script tags rather than unwrapping their contents as html", () => {
    const out = sanitizeHtml("<p>hi</p><script>alert(1)</script>");
    expect(out.toLowerCase()).not.toContain("<script");
    expect(out).toContain("hi");
  });

  it("keeps safe http links", () => {
    const out = sanitizeHtml('<a href="https://example.com">x</a>');
    expect(out).toContain('href="https://example.com"');
  });
});

describe("sanitizeReaderHtml", () => {
  it("strips script from mobi-like body html", () => {
    const out = sanitizeReaderHtml(
      "<p>Chapter</p><script>alert(1)</script><img src=x onerror=alert(1)>",
    );
    expect(out.toLowerCase()).not.toContain("script");
    expect(out.toLowerCase()).not.toContain("onerror");
    expect(out).toContain("Chapter");
  });

  it("blocks javascript img src", () => {
    const out = sanitizeReaderHtml('<img src="javascript:alert(1)">');
    expect(out.toLowerCase()).not.toContain("javascript:");
  });
});

describe("descriptionLooksLikeHtml", () => {
  it("detects html descriptions", () => {
    expect(descriptionLooksLikeHtml("<p>Hi</p>")).toBe(true);
    expect(descriptionLooksLikeHtml("Plain text")).toBe(false);
  });
});
