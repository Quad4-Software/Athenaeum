import * as pdfjs from "pdfjs-dist";
import type { PDFDocumentProxy, RenderTask, TextLayer } from "pdfjs-dist";
import type { Highlight } from "$lib/api/types";
import { PdfPageCache, readerViewportFillColor } from "$lib/reader/pdf-page-cache";
import {
  highlightsForPdfPage,
  pdfFitBaseScale,
  pdfHighlightQuote,
  pdfViewportScale,
} from "$lib/reader/pdf-reader";

export type PdfRenderLayout = {
  slotWidth: number;
  outerHeight: number;
  targetScale: number;
  autoFit: boolean;
  pad: number;
};

export type PdfSelectionCapture = (pageNum: number) => void;

/** Attach selection capture listeners to a text layer. */
export function bindPdfTextLayerEvents(
  layer: HTMLDivElement,
  pageNum: number,
  onCapture: PdfSelectionCapture,
) {
  layer.addEventListener("mouseup", () => onCapture(pageNum));
  layer.addEventListener("keyup", () => onCapture(pageNum));
}

/** Build a canvas + text-layer page wrapper for rendering. */
export function createPdfPageWrap(pageNum: number, onCapture: PdfSelectionCapture): HTMLDivElement {
  const wrap = document.createElement("div");
  wrap.className = "pdf-page relative h-fit shadow-[var(--shadow)]";
  wrap.dataset.page = String(pageNum);
  const canvas = document.createElement("canvas");
  canvas.className = "block h-full w-full";
  const layer = document.createElement("div");
  layer.className = "pdf-text-layer";
  bindPdfTextLayerEvents(layer, pageNum, onCapture);
  wrap.appendChild(canvas);
  wrap.appendChild(layer);
  return wrap;
}

/** Paint saved highlight quotes into a page text layer. */
export function applyPdfMarks(
  layer: HTMLDivElement,
  pageNum: number,
  savedHighlights: Highlight[],
) {
  for (const hl of highlightsForPdfPage(savedHighlights, pageNum)) {
    const quote = pdfHighlightQuote(hl);
    if (!quote) continue;
    const walker = document.createTreeWalker(layer, NodeFilter.SHOW_TEXT);
    let node: Node | null;
    while ((node = walker.nextNode())) {
      const text = node.textContent ?? "";
      const idx = text.indexOf(quote);
      if (idx < 0) continue;
      const range = document.createRange();
      range.setStart(node, idx);
      range.setEnd(node, idx + quote.length);
      const mark = document.createElement("mark");
      mark.className = "pdf-highlight-mark";
      try {
        range.surroundContents(mark);
      } catch {
        // overlapping ranges are skipped
      }
      break;
    }
  }
}

/** Render one PDF page into an existing wrap element. */
export async function renderPdfPageInto(
  wrap: HTMLDivElement,
  targetPage: number,
  cacheKey: string,
  pdfDoc: PDFDocumentProxy | undefined,
  layout: PdfRenderLayout,
  cancelled: () => boolean,
  activeRenderTasks: RenderTask[],
  shouldCache: boolean,
  pageCache: PdfPageCache,
  savedHighlights: Highlight[],
  setRenderTask: (task: RenderTask) => void,
  setTextLayerTask: (task: TextLayer) => void,
): Promise<void> {
  if (!pdfDoc) return;
  const canvas = wrap.querySelector("canvas");
  const layer = wrap.querySelector(".pdf-text-layer") as HTMLDivElement | null;
  if (!canvas || !layer) return;

  const pdfPage = await pdfDoc.getPage(targetPage);
  if (cancelled()) return;

  const unscaled = pdfPage.getViewport({ scale: 1 });
  const baseFit = pdfFitBaseScale(
    unscaled.width,
    unscaled.height,
    layout.slotWidth,
    layout.outerHeight,
    layout.pad,
    layout.autoFit,
  );
  const vp = pdfPage.getViewport({ scale: pdfViewportScale(baseFit, layout.targetScale) });
  const ctx = canvas.getContext("2d");
  if (!ctx) return;

  const ratio = window.devicePixelRatio || 1;
  canvas.width = Math.floor(vp.width * ratio);
  canvas.height = Math.floor(vp.height * ratio);
  canvas.style.width = `${Math.floor(vp.width)}px`;
  canvas.style.height = `${Math.floor(vp.height)}px`;
  wrap.style.width = `${Math.floor(vp.width)}px`;
  wrap.style.height = `${Math.floor(vp.height)}px`;
  ctx.setTransform(ratio, 0, 0, ratio, 0, 0);
  ctx.fillStyle = readerViewportFillColor();
  ctx.fillRect(0, 0, canvas.width / ratio, canvas.height / ratio);

  const task = pdfPage.render({ canvas, canvasContext: ctx, viewport: vp });
  activeRenderTasks.push(task);
  setRenderTask(task);
  await task.promise;
  if (cancelled()) return;

  const textContent = await pdfPage.getTextContent();
  if (cancelled()) return;

  const tl = new pdfjs.TextLayer({
    textContentSource: textContent,
    container: layer,
    viewport: vp,
  });
  setTextLayerTask(tl);
  await tl.render();
  if (cancelled()) return;

  applyPdfMarks(layer, targetPage, savedHighlights);
  if (cancelled()) return;
  if (shouldCache) pageCache.set(cacheKey, wrap);
}
