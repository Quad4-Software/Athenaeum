import { readFileSync } from "node:fs";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";
import { describe, expect, it } from "vitest";
import { epubRenderOptions } from "./epub-options";

const here = dirname(fileURLToPath(import.meta.url));

const READER_CSS_IMPORTS = [
  { svelte: "PdfReader.svelte", css: "PdfReader.css" },
  { svelte: "EpubReader.svelte", css: "EpubReader.css" },
  { svelte: "AudioReader.svelte", css: "AudioReader.css" },
] as const;

describe("reader stylesheet wiring", () => {
  for (const { svelte, css } of READER_CSS_IMPORTS) {
    it(`${svelte} imports ${css} (vitePreprocess ignores style src=)`, () => {
      const source = readFileSync(join(here, svelte), "utf8");
      expect(source).toContain(`import "$lib/reader/${css}"`);
      expect(source).not.toMatch(/<style\s+src=/);
    });
  }

  it("PdfReader.css hides and positions the text layer over the canvas", () => {
    const css = readFileSync(join(here, "PdfReader.css"), "utf8");
    expect(css).toMatch(/\.pdf-page\s+\.pdf-text-layer/);
    expect(css).toMatch(/\.pdf-page\s+\.textLayer/);
    expect(css).toMatch(/position:\s*absolute/);
    expect(css).toMatch(/color:\s*transparent/);
    expect(css).toMatch(/--font-height/);
  });

  it("EpubReader.css styles the mount surface", () => {
    const css = readFileSync(join(here, "EpubReader.css"), "utf8");
    expect(css).toMatch(/\.epub-surface\s*\{/);
    expect(css).toMatch(/\.epub-mount\s*\{/);
  });

  it("createPdfPageWrap assigns textLayer and pdf-text-layer classes", () => {
    const source = readFileSync(join(here, "pdf-page-render.ts"), "utf8");
    expect(source).toMatch(/className\s*=\s*["']textLayer pdf-text-layer["']/);
  });
});

describe("epubRenderOptions sandbox", () => {
  it("enables allow-scripts via allowScriptedContent for epubjs blob iframes", () => {
    const opts = epubRenderOptions("none");
    expect(opts.allowScriptedContent).toBe(true);
    expect((opts as { method?: string }).method).toBe("blobUrl");
  });
});
