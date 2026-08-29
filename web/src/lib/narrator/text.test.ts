import { describe, expect, it } from "vitest";
import {
  UTTERANCE_HARD_MAX,
  UTTERANCE_MAX_CHARS,
  buildUtteranceQueue,
  normalizeNarratorText,
  paragraphsFromDocument,
  splitIntoUtterances,
} from "./text";

describe("normalizeNarratorText", () => {
  it("collapses whitespace", () => {
    expect(normalizeNarratorText("  hello \n\t world  ")).toBe("hello world");
  });
});

describe("splitIntoUtterances", () => {
  it("returns empty for blank input", () => {
    expect(splitIntoUtterances("")).toEqual([]);
    expect(splitIntoUtterances("   ")).toEqual([]);
  });

  it("keeps short text as one utterance", () => {
    expect(splitIntoUtterances("Hello world.")).toEqual(["Hello world."]);
  });

  it("splits on sentence boundaries", () => {
    const longA = "A".repeat(100);
    const longB = "B".repeat(100);
    const text = `${longA}. ${longB}!`;
    const parts = splitIntoUtterances(text, 120);
    expect(parts.length).toBeGreaterThanOrEqual(2);
    expect(parts[0]).toContain("A");
    expect(parts[parts.length - 1]).toContain("B");
  });

  it("soft-wraps oversized sentences", () => {
    const words = Array.from({ length: 80 }, (_, i) => `w${i}`).join(" ");
    const parts = splitIntoUtterances(words, 40);
    expect(parts.length).toBeGreaterThan(1);
    for (const p of parts) {
      expect(p.length).toBeLessThanOrEqual(40);
    }
  });

  it("hard-splits single words longer than max", () => {
    const monster = "x".repeat(500);
    const parts = splitIntoUtterances(monster, 100);
    expect(parts.length).toBe(5);
    for (const p of parts) {
      expect(p.length).toBeLessThanOrEqual(100);
    }
  });

  it("never exceeds hard max even if requested max is huge", () => {
    const chapter = Array.from({ length: 200 }, (_, i) => `Sentence number ${i} goes here.`).join(
      " ",
    );
    expect(chapter.length).toBeGreaterThan(UTTERANCE_HARD_MAX);
    const parts = splitIntoUtterances(chapter, 50_000);
    expect(parts.length).toBeGreaterThan(1);
    for (const p of parts) {
      expect(p.length).toBeLessThanOrEqual(UTTERANCE_HARD_MAX);
    }
  });
});

describe("buildUtteranceQueue", () => {
  it("flattens paragraphs into utterances", () => {
    const q = buildUtteranceQueue(["One. Two.", "Three."], 280);
    expect(q).toEqual(["One. Two.", "Three."]);
  });

  it("chunks a huge chapter into model-sized pieces", () => {
    const paragraphs = Array.from({ length: 40 }, (_, i) => {
      return `Paragraph ${i}. ${"word ".repeat(120).trim()}.`;
    });
    const queue = buildUtteranceQueue(paragraphs, UTTERANCE_MAX_CHARS);
    expect(queue.length).toBeGreaterThan(40);
    for (const u of queue) {
      expect(u.length).toBeGreaterThan(0);
      expect(u.length).toBeLessThanOrEqual(UTTERANCE_HARD_MAX);
      expect(u.length).toBeLessThanOrEqual(UTTERANCE_MAX_CHARS);
    }
  });
});

describe("paragraphsFromDocument", () => {
  it("extracts block text", () => {
    const doc = document.implementation.createHTMLDocument("t");
    doc.body.innerHTML = "<p>Hello</p><p>  </p><li>World</li><div>ignored unless fallback</div>";
    expect(paragraphsFromDocument(doc)).toEqual(["Hello", "World"]);
  });

  it("falls back to body text", () => {
    const doc = document.implementation.createHTMLDocument("t");
    doc.body.innerHTML = "<div>Plain body</div>";
    expect(paragraphsFromDocument(doc)).toEqual(["Plain body"]);
  });

  it("handles null document", () => {
    expect(paragraphsFromDocument(null)).toEqual([]);
  });
});
