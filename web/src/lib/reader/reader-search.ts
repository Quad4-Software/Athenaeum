import type { Book } from "epubjs";
import type Section from "epubjs/types/section";
import type { PDFDocumentProxy } from "pdfjs-dist";

export interface ReaderSearchHit {
  id: string;
  label: string;
  excerpt: string;
  location: string;
}

export interface ReaderChapter {
  id: string;
  label: string;
  depth?: number;
  hint?: string;
  location: string;
}

export interface EpubNavItem {
  label: string;
  href: string;
  subitems?: EpubNavItem[];
}

export function flattenEpubToc(items: EpubNavItem[], depth = 0): ReaderChapter[] {
  const out: ReaderChapter[] = [];
  for (const item of items) {
    if (item.href && item.label) {
      out.push({
        id: item.href,
        label: item.label,
        depth,
        location: item.href,
      });
    }
    if (item.subitems?.length) {
      out.push(...flattenEpubToc(item.subitems, depth + 1));
    }
  }
  return out;
}

export async function loadEpubChapters(book: Book): Promise<ReaderChapter[]> {
  await book.ready;
  const nav = await book.loaded.navigation;
  const toc = (nav.toc ?? []) as EpubNavItem[];
  return flattenEpubToc(toc);
}

interface EpubSearchMatch {
  cfi: string;
  excerpt: string;
}

type SearchableSection = Section & {
  search(query: string): EpubSearchMatch[];
};

export async function searchEpub(book: Book, query: string): Promise<ReaderSearchHit[]> {
  const q = query.trim();
  if (!q) return [];
  await book.ready;
  const spineItems = await book.loaded.spine;
  const hits: ReaderSearchHit[] = [];
  let index = 0;
  for (const item of spineItems) {
    const section = book.spine.get(item.index) as SearchableSection;
    await section.load(book.load.bind(book));
    const matches = section.search(q);
    for (const match of matches) {
      const href = item.href || String(item.index);
      hits.push({
        id: `epub-${index++}`,
        label: href.split("/").pop() || href,
        excerpt: match.excerpt?.trim() || q,
        location: match.cfi,
      });
    }
    section.unload();
  }
  return hits;
}

interface PdfOutlineNode {
  title: string;
  dest?: string | unknown[] | null;
  items?: PdfOutlineNode[];
}

async function resolvePdfDest(
  doc: PDFDocumentProxy,
  dest: string | unknown[] | null | undefined,
): Promise<number | null> {
  if (!dest) return null;
  try {
    let explicitDest: unknown = dest;
    if (typeof dest === "string") {
      explicitDest = await doc.getDestination(dest);
    }
    if (!Array.isArray(explicitDest) || explicitDest.length === 0) return null;
    const ref = explicitDest[0];
    const pageIndex = await doc.getPageIndex(
      ref as Parameters<PDFDocumentProxy["getPageIndex"]>[0],
    );
    return pageIndex + 1;
  } catch {
    return null;
  }
}

async function flattenPdfOutline(
  doc: PDFDocumentProxy,
  nodes: PdfOutlineNode[],
  depth = 0,
): Promise<ReaderChapter[]> {
  const out: ReaderChapter[] = [];
  for (const node of nodes) {
    const page = await resolvePdfDest(doc, node.dest);
    if (node.title && page) {
      out.push({
        id: `pdf-outline-${page}-${node.title}`,
        label: node.title,
        hint: depth > 0 ? undefined : `Page ${page}`,
        location: String(page),
      });
    }
    if (node.items?.length) {
      out.push(...(await flattenPdfOutline(doc, node.items, depth + 1)));
    }
  }
  return out;
}

export async function loadPdfChapters(doc: PDFDocumentProxy): Promise<ReaderChapter[]> {
  try {
    const outline = (await doc.getOutline()) as PdfOutlineNode[] | null;
    if (!outline?.length) return [];
    return flattenPdfOutline(doc, outline);
  } catch {
    return [];
  }
}

export async function searchPdf(doc: PDFDocumentProxy, query: string): Promise<ReaderSearchHit[]> {
  const q = query.trim().toLowerCase();
  if (!q) return [];
  const hits: ReaderSearchHit[] = [];
  for (let pageNum = 1; pageNum <= doc.numPages; pageNum++) {
    const page = await doc.getPage(pageNum);
    const textContent = await page.getTextContent();
    const text = textContent.items
      .map((item) => ("str" in item ? item.str : ""))
      .join(" ")
      .replace(/\s+/g, " ")
      .trim();
    if (!text) continue;
    const lower = text.toLowerCase();
    let from = 0;
    let matchIndex = 0;
    while (from < lower.length) {
      const pos = lower.indexOf(q, from);
      if (pos < 0) break;
      const start = Math.max(0, pos - 60);
      const end = Math.min(text.length, pos + q.length + 60);
      let excerpt = text.slice(start, end);
      if (start > 0) excerpt = `...${excerpt}`;
      if (end < text.length) excerpt = `${excerpt}...`;
      hits.push({
        id: `pdf-${pageNum}-${matchIndex++}`,
        label: `Page ${pageNum}`,
        excerpt,
        location: JSON.stringify({ page: pageNum, quote: text.slice(pos, pos + q.length) }),
      });
      from = pos + q.length;
      if (matchIndex > 40) break;
    }
    if (hits.length > 120) break;
  }
  return hits;
}
