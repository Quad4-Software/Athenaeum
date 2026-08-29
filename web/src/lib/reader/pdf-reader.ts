export type PagesPerView = 1 | 2 | 3;

export const PDF_LAYOUT_PAD = 32;
export const PDF_LAYOUT_GAP = 16;
export const PDF_MIN_SLOT_WIDTH = 120;
export const PDF_MIN_FIT_SCALE = 0.2;
export const PDF_ZOOM_STEP = 0.2;
export const PDF_ZOOM_MIN = 0.4;
export const PDF_ZOOM_MAX = 4;

export const PDF_SHORTCUT_ITEMS = [
  { keys: "<- / ->", action: "Previous / next page" },
  { keys: "+ / -", action: "Zoom in / out" },
  { keys: "?", action: "Show shortcuts" },
];

/** Parse stored pages-per-view preference. Invalid values become 1. */
export function parsePagesPerView(raw: string | null): PagesPerView {
  const n = Number(raw);
  return n === 2 || n === 3 ? n : 1;
}

/** Clamp a 1-based PDF page into [1, total]. */
export function clampPdfPage(page: number, total: number): number {
  if (total <= 0) return Math.max(page, 1);
  return Math.min(Math.max(page, 1), total);
}

export function nextPdfScale(scale: number): number {
  return Math.min(scale + PDF_ZOOM_STEP, PDF_ZOOM_MAX);
}

export function prevPdfScale(scale: number): number {
  return Math.max(scale - PDF_ZOOM_STEP, PDF_ZOOM_MIN);
}

export function pdfPrevDisabled(spreadPage: number): boolean {
  return spreadPage <= 1;
}

export function pdfNextDisabled(spreadPage: number, total: number): boolean {
  return total > 0 && spreadPage >= total;
}

/** Toolbar page indicator, e.g. "3–4 / 12" or "1 / ...". */
export function pdfPageLabel(spreadPage: number, pagesPerView: number, total: number): string {
  const range =
    pagesPerView > 1 ? `–${Math.min(spreadPage + pagesPerView - 1, total)}` : "";
  return `${spreadPage}${range} / ${total || "..."}`;
}

export function pdfZoomLabel(
  autoFit: boolean,
  scale: number,
  style: "short" | "long" = "short",
): string {
  if (autoFit) return style === "long" ? "Auto fit" : "Auto";
  return `${Math.round(scale * 100)}%`;
}

export function pdfProgressRatio(page: number, total: number): number {
  return total ? page / total : 0;
}

export function pdfSelectionLocation(pageNum: number, quote: string): string {
  return JSON.stringify({ page: pageNum, quote });
}

/** Resolve chapter/search/annotation locations to a 1-based page, or null. */
export function parsePdfJumpPage(location: string, total: number): number | null {
  try {
    const parsed = JSON.parse(location) as { page?: number };
    if (typeof parsed.page === "number" && parsed.page >= 1 && parsed.page <= total) {
      return parsed.page;
    }
  } catch {
    /* plain page number */
  }
  const n = Number(location);
  if (n >= 1 && n <= total) return n;
  return null;
}

export function highlightsForPdfPage<T extends { location: string }>(
  highlights: T[],
  pageNum: number,
): T[] {
  return highlights.filter((hl) => {
    try {
      const loc = JSON.parse(hl.location) as { page?: number };
      return loc.page === pageNum;
    } catch {
      return Number(hl.location) === pageNum;
    }
  });
}

export function pdfHighlightQuote(hl: { location: string; excerpt?: string }): string {
  let quote = hl.excerpt?.trim() ?? "";
  if (!quote) {
    try {
      quote = (JSON.parse(hl.location) as { quote?: string }).quote?.trim() ?? "";
    } catch {
      quote = "";
    }
  }
  return quote;
}

export function pdfSlotWidth(
  viewportWidth: number,
  slots: number,
  pad = PDF_LAYOUT_PAD,
  gap = PDF_LAYOUT_GAP,
  minWidth = PDF_MIN_SLOT_WIDTH,
): number {
  return Math.max((viewportWidth - pad - gap * Math.max(slots - 1, 0)) / Math.max(slots, 1), minWidth);
}

/** Unscaled page fit factor before applying the user zoom multiplier. */
export function pdfFitBaseScale(
  pageWidth: number,
  pageHeight: number,
  slotWidth: number,
  viewportHeight: number,
  pad: number,
  autoFit: boolean,
): number {
  const fitW = slotWidth / pageWidth;
  const fitH = (viewportHeight - pad) / pageHeight;
  return autoFit ? Math.min(fitW, fitH) : fitW;
}

export function pdfViewportScale(baseFit: number, targetScale: number): number {
  return Math.max(baseFit, PDF_MIN_FIT_SCALE) * targetScale;
}
