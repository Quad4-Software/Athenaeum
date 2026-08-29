import { describe, expect, it } from "vitest";
import {
  clampPdfPage,
  highlightsForPdfPage,
  nextPdfScale,
  parsePagesPerView,
  parsePdfJumpPage,
  pdfFitBaseScale,
  pdfHighlightQuote,
  pdfNextDisabled,
  pdfPageLabel,
  pdfPrevDisabled,
  pdfProgressRatio,
  pdfSelectionLocation,
  pdfSlotWidth,
  pdfViewportScale,
  pdfZoomLabel,
  prevPdfScale,
} from "./pdf-reader";

describe("parsePagesPerView", () => {
  it("accepts 2 and 3", () => {
    expect(parsePagesPerView("2")).toBe(2);
    expect(parsePagesPerView("3")).toBe(3);
  });

  it("defaults invalid values to 1", () => {
    expect(parsePagesPerView(null)).toBe(1);
    expect(parsePagesPerView("1")).toBe(1);
    expect(parsePagesPerView("4")).toBe(1);
    expect(parsePagesPerView("nope")).toBe(1);
  });
});

describe("clampPdfPage", () => {
  it("clamps into the document range", () => {
    expect(clampPdfPage(0, 10)).toBe(1);
    expect(clampPdfPage(5, 10)).toBe(5);
    expect(clampPdfPage(99, 10)).toBe(10);
  });

  it("keeps at least page 1 when total is unknown", () => {
    expect(clampPdfPage(0, 0)).toBe(1);
    expect(clampPdfPage(3, 0)).toBe(3);
  });
});

describe("zoom scales", () => {
  it("steps in and out within bounds", () => {
    expect(nextPdfScale(1)).toBeCloseTo(1.2);
    expect(nextPdfScale(3.9)).toBe(4);
    expect(prevPdfScale(1)).toBeCloseTo(0.8);
    expect(prevPdfScale(0.5)).toBe(0.4);
  });
});

describe("nav disabled flags", () => {
  it("disables prev on the first page", () => {
    expect(pdfPrevDisabled(1)).toBe(true);
    expect(pdfPrevDisabled(2)).toBe(false);
  });

  it("disables next at or past the end once total is known", () => {
    expect(pdfNextDisabled(1, 0)).toBe(false);
    expect(pdfNextDisabled(9, 10)).toBe(false);
    expect(pdfNextDisabled(10, 10)).toBe(true);
  });
});

describe("pdfPageLabel", () => {
  it("shows a single page", () => {
    expect(pdfPageLabel(3, 1, 12)).toBe("3 / 12");
  });

  it("shows a spread range", () => {
    expect(pdfPageLabel(3, 2, 12)).toBe("3–4 / 12");
    expect(pdfPageLabel(11, 3, 12)).toBe("11–12 / 12");
  });

  it("shows ellipsis while total is unknown", () => {
    expect(pdfPageLabel(1, 1, 0)).toBe("1 / ...");
  });
});

describe("pdfZoomLabel", () => {
  it("labels auto-fit and percent zoom", () => {
    expect(pdfZoomLabel(true, 1, "short")).toBe("Auto");
    expect(pdfZoomLabel(true, 1, "long")).toBe("Auto fit");
    expect(pdfZoomLabel(false, 1.25)).toBe("125%");
  });
});

describe("pdfProgressRatio", () => {
  it("returns page over total", () => {
    expect(pdfProgressRatio(5, 10)).toBe(0.5);
    expect(pdfProgressRatio(1, 0)).toBe(0);
  });
});

describe("parsePdfJumpPage", () => {
  it("parses JSON page locations", () => {
    expect(parsePdfJumpPage(JSON.stringify({ page: 4, quote: "hi" }), 10)).toBe(4);
  });

  it("parses plain page numbers", () => {
    expect(parsePdfJumpPage("7", 10)).toBe(7);
  });

  it("rejects out-of-range pages", () => {
    expect(parsePdfJumpPage("0", 10)).toBeNull();
    expect(parsePdfJumpPage("11", 10)).toBeNull();
    expect(parsePdfJumpPage(JSON.stringify({ page: 99 }), 10)).toBeNull();
  });
});

describe("highlightsForPdfPage", () => {
  it("matches JSON page locations", () => {
    const highlights = [
      { location: JSON.stringify({ page: 2, quote: "a" }) },
      { location: JSON.stringify({ page: 2 }) },
    ];
    expect(highlightsForPdfPage(highlights, 2)).toHaveLength(2);
    expect(highlightsForPdfPage(highlights, 1)).toHaveLength(0);
  });

  it("falls back when location is not valid JSON", () => {
    expect(highlightsForPdfPage([{ location: "+3" }], 3)).toHaveLength(1);
  });
});

describe("pdfHighlightQuote", () => {
  it("prefers excerpt then embedded quote", () => {
    expect(pdfHighlightQuote({ location: "2", excerpt: " from excerpt " })).toBe("from excerpt");
    expect(pdfHighlightQuote({ location: JSON.stringify({ page: 2, quote: " embedded " }) })).toBe(
      "embedded",
    );
    expect(pdfHighlightQuote({ location: "2" })).toBe("");
  });
});

describe("pdfSelectionLocation", () => {
  it("serializes page and quote", () => {
    expect(JSON.parse(pdfSelectionLocation(5, "hello"))).toEqual({ page: 5, quote: "hello" });
  });
});

describe("layout helpers", () => {
  it("computes slot width with gaps", () => {
    expect(pdfSlotWidth(800, 2, 32, 16, 120)).toBe((800 - 32 - 16) / 2);
  });

  it("fits pages to width or contain", () => {
    expect(pdfFitBaseScale(200, 400, 400, 500, 32, false)).toBe(2);
    expect(pdfFitBaseScale(200, 400, 400, 500, 32, true)).toBeCloseTo((500 - 32) / 400);
  });

  it("applies zoom with a minimum fit floor", () => {
    expect(pdfViewportScale(0.1, 2)).toBeCloseTo(0.4);
    expect(pdfViewportScale(1, 1.5)).toBeCloseTo(1.5);
  });
});
