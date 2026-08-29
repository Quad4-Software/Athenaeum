export function visiblePageNumbers(
  spreadPage: number,
  pagesPerView: number,
  total: number,
): number[] {
  const pages: number[] = [];
  for (let i = 0; i < pagesPerView; i++) {
    const n = spreadPage + i;
    if (n >= 1 && n <= total) pages.push(n);
  }
  return pages;
}

export function adjacentSpreadStartPages(
  spreadPage: number,
  pagesPerView: number,
  total: number,
): number[] {
  const starts: number[] = [];
  const prevStart = spreadPage - pagesPerView;
  if (prevStart >= 1) starts.push(prevStart);
  const nextStart = spreadPage + pagesPerView;
  if (nextStart <= total) starts.push(nextStart);
  return starts;
}

export function preloadPageNumbers(
  spreadPage: number,
  pagesPerView: number,
  total: number,
): number[] {
  if (total <= 0) return [];
  const visible = new Set(visiblePageNumbers(spreadPage, pagesPerView, total));
  const pages: number[] = [];
  for (const start of adjacentSpreadStartPages(spreadPage, pagesPerView, total)) {
    for (let i = 0; i < pagesPerView; i++) {
      const n = start + i;
      if (n >= 1 && n <= total && !visible.has(n)) pages.push(n);
    }
  }
  return pages;
}

export function spinePreloadIndices(currentIndex: number, spineLength: number): number[] {
  if (spineLength <= 0 || currentIndex < 0) return [];
  const indices: number[] = [];
  if (currentIndex > 0) indices.push(currentIndex - 1);
  if (currentIndex + 1 < spineLength) indices.push(currentIndex + 1);
  return indices;
}

export function pdfPageCacheKey(
  pageNum: number,
  scale: number,
  autoFit: boolean,
  slotWidth: number,
  viewportHeight: number,
): string {
  return `${pageNum}|${scale}|${autoFit ? 1 : 0}|${Math.round(slotWidth)}|${Math.round(viewportHeight)}`;
}

export function prevSpreadStart(spreadPage: number, pagesPerView: number): number {
  return Math.max(1, spreadPage - pagesPerView);
}

export function nextSpreadStart(spreadPage: number, pagesPerView: number, total: number): number {
  const nextStart = spreadPage + pagesPerView;
  if (nextStart <= total) return nextStart;
  if (spreadPage < total) return total;
  return spreadPage;
}

export function clampSpreadStart(spreadPage: number, pagesPerView: number, total: number): number {
  if (total <= 0) return spreadPage;
  return Math.min(spreadPage, Math.max(1, total - pagesPerView + 1));
}

export function shouldCommitSpreadMount(
  mountId: number,
  activeMountId: number,
  cancelled: boolean,
): boolean {
  return !cancelled && mountId === activeMountId;
}
