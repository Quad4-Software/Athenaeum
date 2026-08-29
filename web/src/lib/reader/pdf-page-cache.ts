function defaultMaxEntries(): number {
  const mem = (navigator as Navigator & { deviceMemory?: number }).deviceMemory;
  if (mem != null && mem <= 4) return 4;
  return 6;
}

function releaseCanvasMemory(wrap: HTMLDivElement): void {
  for (const canvas of wrap.querySelectorAll("canvas")) {
    canvas.width = 0;
    canvas.height = 0;
  }
}

export class PdfPageCache {
  private entries = new Map<string, HTMLDivElement>();
  readonly maxEntries: number;

  constructor(maxEntries = defaultMaxEntries()) {
    this.maxEntries = maxEntries;
  }

  has(key: string): boolean {
    return this.entries.has(key);
  }

  take(key: string): HTMLDivElement | null {
    const wrap = this.entries.get(key);
    if (!wrap) return null;
    this.entries.delete(key);
    return wrap;
  }

  set(key: string, wrap: HTMLDivElement): void {
    if (this.entries.has(key)) this.evict(key);
    wrap.remove();
    this.entries.set(key, wrap);
    while (this.entries.size > this.maxEntries) {
      const oldest = this.entries.keys().next().value;
      if (oldest === undefined || oldest === key) break;
      this.evict(oldest);
    }
  }

  private evict(key: string): void {
    const wrap = this.entries.get(key);
    if (wrap) releaseCanvasMemory(wrap);
    this.entries.delete(key);
  }

  clear(): void {
    for (const key of this.entries.keys()) this.evict(key);
  }
}

export function readerViewportFillColor(): string {
  if (typeof document === "undefined") return "#16161a";
  const value = getComputedStyle(document.documentElement)
    .getPropertyValue("--color-bg-elevated")
    .trim();
  return value || "#16161a";
}
