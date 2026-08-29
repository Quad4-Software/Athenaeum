import { pdfPageCacheKey, preloadPageNumbers, visiblePageNumbers } from "./reader-navigation";

export interface SpreadLayout {
  targetScale: number;
  autoFit: boolean;
  slotWidth: number;
  viewportHeight: number;
}

export interface SpreadPagePlan {
  page: number;
  cacheKey: string;
  fromCache: boolean;
}

export function spreadPageCacheKey(page: number, layout: SpreadLayout): string {
  return pdfPageCacheKey(
    page,
    layout.targetScale,
    layout.autoFit,
    layout.slotWidth,
    layout.viewportHeight,
  );
}

export function buildVisibleSpreadPlan(
  spreadPage: number,
  pagesPerView: number,
  total: number,
  layout: SpreadLayout,
  hasCached: (key: string) => boolean,
): SpreadPagePlan[] {
  return visiblePageNumbers(spreadPage, pagesPerView, total).map((page) => {
    const cacheKey = spreadPageCacheKey(page, layout);
    return { page, cacheKey, fromCache: hasCached(cacheKey) };
  });
}

export function buildPreloadPlan(
  spreadPage: number,
  pagesPerView: number,
  total: number,
  layout: SpreadLayout,
  hasCached: (key: string) => boolean,
): Array<{ page: number; cacheKey: string }> {
  return preloadPageNumbers(spreadPage, pagesPerView, total)
    .filter((page) => !hasCached(spreadPageCacheKey(page, layout)))
    .map((page) => ({ page, cacheKey: spreadPageCacheKey(page, layout) }));
}
