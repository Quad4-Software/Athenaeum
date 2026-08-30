import type { PageKeyHandlers } from "$lib/reader/reader-keys";
import { injectEpubContentBackground, type ReaderTheme } from "$lib/reader/epub-theme";
import { spinePreloadIndices } from "$lib/reader/reader-navigation";

export type EpubSpreadMode = "single" | "auto" | "always";
export type EpubjsSpread = "none" | "auto" | "always";

export const EPUB_MIN_FONT = 80;
export const EPUB_MAX_FONT = 160;
export const EPUB_FONT_STEP = 10;

export const EPUB_DEFAULT_FONT_PCT = 100;
export const EPUB_DEFAULT_LINE_HEIGHT = 1.6;
export const EPUB_DEFAULT_MARGIN_PX = 24;
export const EPUB_DEFAULT_SPREAD: EpubSpreadMode = "auto";

export const EPUB_HIGHLIGHT_STYLES: Record<string, Record<string, string>> = {
  yellow: { fill: "#fde047", "fill-opacity": "0.45", "mix-blend-mode": "multiply" },
  green: { fill: "#86efac", "fill-opacity": "0.45", "mix-blend-mode": "multiply" },
  blue: { fill: "#93c5fd", "fill-opacity": "0.45", "mix-blend-mode": "multiply" },
};

export const EPUB_SHORTCUT_ITEMS = [
  { keys: "<- / ->", action: "Previous / next page" },
  { keys: "Page Up / Down", action: "Previous / next page" },
  { keys: "+ / -", action: "Increase / decrease font size" },
  { keys: "?", action: "Show shortcuts" },
] as const;

export interface EpubDisplayPrefs {
  fontPct: number;
  theme: ReaderTheme;
  lineHeight: number;
  marginPx: number;
  spread: EpubSpreadMode;
}

export type EpubContentsFrame = { document?: Document; window?: Window };

export type EpubNarratorDecision =
  | { kind: "toggle-off" }
  | { kind: "error"; key: "narrator.errUnavailable" | "narrator.errKokoro" }
  | { kind: "start"; switchToBrowser: boolean };

/** Maps UI spread mode to epubjs rendition spread. */
export function epubjsSpread(mode: EpubSpreadMode): EpubjsSpread {
  if (mode === "single") return "none";
  if (mode === "always") return "always";
  return "auto";
}

/** Max-width utility for the reader surface under the given spread mode. */
export function epubSpreadWidthClass(mode: EpubSpreadMode): string {
  if (mode === "single") return "max-w-3xl";
  if (mode === "always") return "max-w-6xl";
  return "max-w-5xl";
}

export function isEpubSpreadMode(value: string | null | undefined): value is EpubSpreadMode {
  return value === "single" || value === "auto" || value === "always";
}

export function loadEpubSpreadMode(
  stored: string | null | undefined,
  fallback: EpubSpreadMode = EPUB_DEFAULT_SPREAD,
): EpubSpreadMode {
  return isEpubSpreadMode(stored) ? stored : fallback;
}

export function loadStoredNumber(stored: string | null | undefined, fallback: number): number {
  return Number(stored) || fallback;
}

const READER_THEMES: readonly ReaderTheme[] = ["light", "dark", "sepia", "night"];

export function isReaderTheme(value: string | null | undefined): value is ReaderTheme {
  return !!value && (READER_THEMES as readonly string[]).includes(value);
}

export function loadReaderTheme(
  stored: string | null | undefined,
  fallback: ReaderTheme = "light",
): ReaderTheme {
  return isReaderTheme(stored) ? stored : fallback;
}

export function clampEpubFontPct(pct: number): number {
  return Math.min(EPUB_MAX_FONT, Math.max(EPUB_MIN_FONT, pct));
}

export function nextEpubFontPct(pct: number): number {
  return clampEpubFontPct(pct + EPUB_FONT_STEP);
}

export function prevEpubFontPct(pct: number): number {
  return clampEpubFontPct(pct - EPUB_FONT_STEP);
}

export function epubHighlightStyles(color = "yellow"): Record<string, string> {
  return EPUB_HIGHLIGHT_STYLES[color] ?? EPUB_HIGHLIGHT_STYLES.yellow;
}

/** Safe progress fraction from CFI; missing locations yield 0. */
export function epubPercentFromCfi(
  percentageFromCfi: (cfi: string) => number,
  cfi: string,
): number {
  try {
    return percentageFromCfi(cfi) || 0;
  } catch {
    return 0;
  }
}

export function mergeEpubReaderPrefs(
  current: EpubDisplayPrefs,
  prefs: Record<string, unknown>,
): EpubDisplayPrefs {
  const next: EpubDisplayPrefs = { ...current };
  if (typeof prefs.fontPct === "number") next.fontPct = prefs.fontPct;
  if (typeof prefs.theme === "string" && isReaderTheme(prefs.theme)) next.theme = prefs.theme;
  if (typeof prefs.lineHeight === "number") next.lineHeight = prefs.lineHeight;
  if (typeof prefs.marginPx === "number") next.marginPx = prefs.marginPx;
  if (typeof prefs.spread === "string" && isEpubSpreadMode(prefs.spread)) {
    next.spread = prefs.spread;
  }
  return next;
}

export function normalizeEpubContents(
  raw: EpubContentsFrame | EpubContentsFrame[] | undefined | null,
): EpubContentsFrame[] {
  if (Array.isArray(raw)) return raw;
  if (raw) return [raw];
  return [];
}

export function epubFontUploadErrorKey(
  message: string,
): "reader.fontTooLarge" | "reader.fontUploadFailed" {
  return message.includes("large") ? "reader.fontTooLarge" : "reader.fontUploadFailed";
}

export function epubPageKeyHandlers(handlers: {
  prev: () => void;
  next: () => void;
  largerFont: () => void;
  smallerFont: () => void;
  openShortcuts: () => void;
}): PageKeyHandlers {
  return {
    prev: handlers.prev,
    next: handlers.next,
    zoomIn: handlers.largerFont,
    zoomOut: handlers.smallerFont,
    shortcuts: handlers.openShortcuts,
  };
}

export function decideEpubNarration(input: {
  narratorActive: boolean;
  provider: string;
  kokoroEnabled: boolean;
  browserAvailable: boolean;
}): EpubNarratorDecision {
  if (input.narratorActive) return { kind: "toggle-off" };
  if (input.provider === "browser" && !input.browserAvailable) {
    return { kind: "error", key: "narrator.errUnavailable" };
  }
  if (input.provider === "kokoro" && !input.kokoroEnabled) {
    if (!input.browserAvailable) return { kind: "error", key: "narrator.errKokoro" };
    return { kind: "start", switchToBrowser: true };
  }
  return { kind: "start", switchToBrowser: false };
}

/** Resolve font id when custom font presence and preference disagree. */
export function resolveEpubFontId(input: {
  preference: string;
  hasCustomFont: boolean;
  fallback?: string;
}): string {
  const fallback = input.fallback ?? "book";
  if (input.preference === "custom" && input.hasCustomFont) return "custom";
  if (input.preference === "custom" && !input.hasCustomFont) return fallback;
  return input.preference;
}

export function scheduleEpubLocationsGenerate(run: () => void): void {
  if (typeof requestIdleCallback === "function") {
    requestIdleCallback(run, { timeout: 5000 });
  } else {
    setTimeout(run, 2000);
  }
}

export function epubLoadErrorMessage(message: string | undefined, fallback: string): string {
  return message?.trim() || fallback;
}

export function persistEpubDisplayPrefs(
  keys: {
    font: string;
    theme: string;
    line: string;
    margin: string;
    spread: string;
  },
  prefs: EpubDisplayPrefs,
): void {
  localStorage.setItem(keys.font, String(prefs.fontPct));
  localStorage.setItem(keys.theme, prefs.theme);
  localStorage.setItem(keys.line, String(prefs.lineHeight));
  localStorage.setItem(keys.margin, String(prefs.marginPx));
  localStorage.setItem(keys.spread, prefs.spread);
}

export function canSelectEpubFont(fontId: string, hasCustomFont: boolean): boolean {
  if (fontId === "custom" && !hasCustomFont) return false;
  return true;
}

export interface EpubPrefKeys {
  font: string;
  theme: string;
  line: string;
  margin: string;
  spread: string;
}

/** Load display prefs from localStorage-like storage with defaults. */
export function loadInitialEpubDisplayPrefs(
  storage: { getItem(key: string): string | null } | null | undefined,
  keys: EpubPrefKeys,
): EpubDisplayPrefs {
  if (!storage) {
    return {
      fontPct: EPUB_DEFAULT_FONT_PCT,
      theme: "light",
      lineHeight: EPUB_DEFAULT_LINE_HEIGHT,
      marginPx: EPUB_DEFAULT_MARGIN_PX,
      spread: EPUB_DEFAULT_SPREAD,
    };
  }
  return {
    fontPct: loadStoredNumber(storage.getItem(keys.font), EPUB_DEFAULT_FONT_PCT),
    theme: loadReaderTheme(storage.getItem(keys.theme)),
    lineHeight: loadStoredNumber(storage.getItem(keys.line), EPUB_DEFAULT_LINE_HEIGHT),
    marginPx: loadStoredNumber(storage.getItem(keys.margin), EPUB_DEFAULT_MARGIN_PX),
    spread: loadEpubSpreadMode(storage.getItem(keys.spread)),
  };
}

export interface EpubThemeOverrideInput {
  fg: string;
  bg: string;
  lineHeight: number;
  marginPx: number;
}

/** Apply color, background, line-height, and margin theme overrides. */
export function applyEpubThemeOverrides(
  themes: { override(name: string, value: string): void },
  input: EpubThemeOverrideInput,
): void {
  themes.override("color", input.fg);
  themes.override("background", input.bg);
  themes.override("line-height", String(input.lineHeight));
  themes.override("margin-left", `${input.marginPx}px`);
  themes.override("margin-right", `${input.marginPx}px`);
}

/** Paint a highlight annotation onto an epubjs annotations API. */
export function paintEpubHighlight(
  annotations: {
    add(
      type: string,
      cfi: string,
      data: object,
      cb: () => void,
      className: string,
      styles: Record<string, string>,
    ): void;
  },
  cfi: string,
  id: number,
  color = "yellow",
): void {
  annotations.add("highlight", cfi, {}, () => undefined, `hl-${id}`, epubHighlightStyles(color));
}

type EpubSectionRequest = (url: string) => Promise<unknown>;

/** Preload adjacent spine sections around the current index. */
export function preloadEpubSpineSections(
  book: {
    load: EpubSectionRequest;
    spine: {
      length?: number;
      get(index: number): { load(request?: EpubSectionRequest): unknown } | undefined;
    };
  },
  currentIndex: number,
): void {
  const length = book.spine.length ?? 0;
  // section.load() without book.load hits the page origin for spine paths
  // like /titlepage.xhtml and 404s instead of reading the EPUB archive.
  const request = book.load.bind(book) as EpubSectionRequest;
  for (const index of spinePreloadIndices(currentIndex, length)) {
    const section = book.spine.get(index);
    if (!section) continue;
    void Promise.resolve(section.load(request)).catch(() => undefined);
  }
}

/** Set container background and inject into the active content document. */
export function applyEpubSurfaceBackground(
  container: HTMLElement | null | undefined,
  getContents: (() => { document?: Document } | undefined | null) | undefined,
  bg: string,
): void {
  if (container) container.style.backgroundColor = bg;
  try {
    const contents = getContents?.();
    if (contents?.document) injectEpubContentBackground(contents.document, bg);
  } catch {
    // iframe may be unavailable during transitions
  }
}

/** Persist theme, line-height, and margin keys used by applyReaderTheme. */
export function persistEpubThemePrefs(
  keys: Pick<EpubPrefKeys, "theme" | "line" | "margin">,
  prefs: Pick<EpubDisplayPrefs, "theme" | "lineHeight" | "marginPx">,
): void {
  localStorage.setItem(keys.theme, prefs.theme);
  localStorage.setItem(keys.line, String(prefs.lineHeight));
  localStorage.setItem(keys.margin, String(prefs.marginPx));
}

/** Gather iframe contents for narration, swallowing getContents failures. */
export function gatherEpubNarrateContents(
  getContents: () => EpubContentsFrame | EpubContentsFrame[] | undefined | null,
): EpubContentsFrame[] {
  try {
    return normalizeEpubContents(getContents());
  } catch {
    return [];
  }
}

/** Read trimmed selection text from a rendition contents window. */
export function readEpubSelectionText(getSelection: () => Selection | null | undefined): string {
  try {
    return getSelection()?.toString()?.trim() ?? "";
  } catch {
    return "";
  }
}

/** Resolve font id after custom font load, noting when preference must reset. */
export function resolveLoadedCustomFont(input: {
  preference: string;
  hasCustomFont: boolean;
  fallback?: string;
}): { fontId: string; shouldResetPreference: boolean } {
  return {
    fontId: resolveEpubFontId(input),
    shouldResetPreference: input.preference === "custom" && !input.hasCustomFont,
  };
}

export function createEpubPrefsSaver(opts: {
  delayMs?: number;
  isReady: () => boolean;
  getPrefs: () => EpubDisplayPrefs;
  save: (prefs: EpubDisplayPrefs) => Promise<unknown>;
}): { queue: () => void; clear: () => void } {
  let timer: ReturnType<typeof setTimeout> | null = null;
  return {
    queue() {
      if (!opts.isReady()) return;
      if (timer) clearTimeout(timer);
      timer = setTimeout(() => {
        void opts.save(opts.getPrefs()).catch(() => undefined);
      }, opts.delayMs ?? 600);
    },
    clear() {
      if (timer) clearTimeout(timer);
      timer = null;
    },
  };
}

/** Build the five localStorage keys used by the EPUB reader. */
export function buildEpubPrefKeys(keyFn: (name: string) => string): EpubPrefKeys {
  return {
    font: keyFn("epub-font"),
    theme: keyFn("epub-theme"),
    line: keyFn("epub-line"),
    margin: keyFn("epub-margin"),
    spread: keyFn("epub-spread"),
  };
}

/** Apply font-size override and persist the percentage. */
export function applyEpubFontPct(
  themes: { fontSize(value: string): void },
  pct: number,
  fontKey: string,
): void {
  themes.fontSize(`${pct}%`);
  localStorage.setItem(fontKey, String(pct));
}

/** Persist spread mode to localStorage. */
export function persistEpubSpreadMode(key: string, mode: EpubSpreadMode): void {
  localStorage.setItem(key, mode);
}

/** Paint a list of highlights onto the annotations API. */
export function paintEpubHighlightList(
  annotations: {
    add(
      type: string,
      cfi: string,
      data: object,
      cb: () => void,
      className: string,
      styles: Record<string, string>,
    ): void;
  },
  list: { location: string; id: number; color?: string }[],
): void {
  for (const hl of list) {
    paintEpubHighlight(annotations, hl.location, hl.id, hl.color || "yellow");
  }
}

/** Load highlights and paint them; failures are non-critical. */
export async function loadAndPaintEpubHighlights(
  annotations:
    | {
        add(
          type: string,
          cfi: string,
          data: object,
          cb: () => void,
          className: string,
          styles: Record<string, string>,
        ): void;
      }
    | undefined,
  load: () => Promise<{ location: string; id: number; color?: string }[]>,
): Promise<void> {
  if (!annotations) return;
  try {
    paintEpubHighlightList(annotations, await load());
  } catch {
    // non-critical
  }
}

/** Reset file input and return the first selected file. */
export function takeFileInput(event: Event): File | undefined {
  const input = event.currentTarget as HTMLInputElement;
  const file = input.files?.[0];
  input.value = "";
  return file;
}

export function errorMessage(e: unknown): string {
  return e instanceof Error ? e.message : "";
}

/** Resolve select value to a font id, or null when selection is blocked. */
export function epubFontIdFromSelect(event: Event, hasCustomFont: boolean): string | null {
  const value = (event.currentTarget as HTMLSelectElement).value;
  if (!canSelectEpubFont(value, hasCustomFont)) return null;
  return value;
}

/** Build utterance paragraphs from the active rendition contents. */
export function epubNarrationParagraphs(
  getContents: () => EpubContentsFrame | EpubContentsFrame[] | undefined | null,
  paragraphsFrom: (frames: EpubContentsFrame[]) => string[],
): string[] {
  return paragraphsFrom(gatherEpubNarrateContents(getContents));
}

export function mergeRemoteEpubPrefs(current: EpubDisplayPrefs, remote: unknown): EpubDisplayPrefs {
  return mergeEpubReaderPrefs(current, (remote ?? {}) as Record<string, unknown>);
}
