export type ComicFit = "contain" | "width" | "height";

export function comicFitClass(fit: ComicFit): string {
  switch (fit) {
    case "width":
      return "w-full h-auto";
    case "height":
      return "h-full w-auto max-w-none";
    default:
      return "max-h-full max-w-full object-contain";
  }
}

/** Zero-based index of the first page in the spread containing `page`. */
export function comicSpreadStart(page: number): number {
  if (page <= 0) return 0;
  const offset = page - 1;
  return 1 + 2 * Math.floor(offset / 2);
}

/** Zero-based page indices visible for the spread containing `page`. */
export function comicSpreadPages(
  page: number,
  total: number,
  spreadEnabled: boolean,
  wide: boolean,
): number[] {
  if (total <= 0) return [];
  const clamped = Math.min(Math.max(page, 0), total - 1);
  if (!spreadEnabled || !wide) return [clamped];
  const start = comicSpreadStart(clamped);
  if (start === 0) return [0];
  const pages = [start];
  if (start + 1 < total) pages.push(start + 1);
  return pages;
}

export function nextComicPage(
  page: number,
  total: number,
  spreadEnabled: boolean,
  wide: boolean,
): number {
  if (total <= 0) return page;
  const spread = comicSpreadPages(page, total, spreadEnabled, wide);
  const start = spread[0] ?? page;
  const size = spread.length || 1;
  const target = start + size;
  return target < total ? target : Math.max(total - 1, start);
}

export function prevComicPage(
  page: number,
  total: number,
  spreadEnabled: boolean,
  wide: boolean,
): number {
  if (total <= 0) return page;
  const spread = comicSpreadPages(page, total, spreadEnabled, wide);
  const start = spread[0] ?? page;
  if (start <= 0) return 0;
  return comicSpreadStart(start - 1);
}
