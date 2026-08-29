import type { Rendition } from "epubjs";
import { storageKey } from "$lib/brand/storage";

import sourceSerifUrl from "@fontsource-variable/source-serif-4/files/source-serif-4-latin-wght-normal.woff2?url";
import literataUrl from "@fontsource-variable/literata/files/literata-latin-wght-normal.woff2?url";
import crimsonUrl from "@fontsource-variable/crimson-pro/files/crimson-pro-latin-wght-normal.woff2?url";
import newsreaderUrl from "@fontsource-variable/newsreader/files/newsreader-latin-wght-normal.woff2?url";
import ibmPlexUrl from "@fontsource-variable/ibm-plex-sans/files/ibm-plex-sans-latin-wght-normal.woff2?url";
import dmSansUrl from "@fontsource-variable/dm-sans/files/dm-sans-latin-wght-normal.woff2?url";
import sourceSansUrl from "@fontsource-variable/source-sans-3/files/source-sans-3-latin-wght-normal.woff2?url";
import frauncesUrl from "@fontsource-variable/fraunces/files/fraunces-latin-wght-normal.woff2?url";

type EpubThemes = Rendition["themes"] & {
  removeOverride(name: string): void;
};

export const EPUB_FONT_KEY = storageKey("epub-font-family");
const CUSTOM_FONT_THEME = "athenaeum-custom-font";
const VENDORED_FONT_THEME = "athenaeum-vendored-font";
const CUSTOM_FONT_FAMILY = "Athenaeum Custom";
const DB_NAME = "athenaeum-fonts";
const DB_VERSION = 1;
const STORE = "epub";
const FONT_SAMPLE = "Aa The quick brown fox";

export type BuiltInFontId =
  | "book"
  | "serif"
  | "sans"
  | "mono"
  | "palatino"
  | "verdana"
  | "source-serif"
  | "literata"
  | "crimson"
  | "newsreader"
  | "ibm-plex"
  | "dm-sans"
  | "source-sans"
  | "fraunces";

export type EpubFontId = BuiltInFontId | "custom";

export interface VendoredFace {
  family: string;
  url: string;
  weight: string;
}

export interface BuiltInEpubFont {
  id: BuiltInFontId;
  labelKey: string;
  /** Display name used when i18n is unavailable (also for previews). */
  label: string;
  sample?: string;
  family?: string;
  face?: VendoredFace;
}

export const BUILTIN_EPUB_FONTS: BuiltInEpubFont[] = [
  { id: "book", labelKey: "reader.fontBook", label: "Book default", sample: FONT_SAMPLE },
  {
    id: "source-serif",
    labelKey: "reader.fontSourceSerif",
    label: "Source Serif",
    sample: FONT_SAMPLE,
    family: '"Source Serif 4 Variable", Georgia, serif',
    face: {
      family: "Source Serif 4 Variable",
      url: sourceSerifUrl,
      weight: "200 900",
    },
  },
  {
    id: "literata",
    labelKey: "reader.fontLiterata",
    label: "Literata",
    sample: FONT_SAMPLE,
    family: '"Literata Variable", Georgia, serif',
    face: {
      family: "Literata Variable",
      url: literataUrl,
      weight: "200 900",
    },
  },
  {
    id: "crimson",
    labelKey: "reader.fontCrimson",
    label: "Crimson Pro",
    sample: FONT_SAMPLE,
    family: '"Crimson Pro Variable", Georgia, serif',
    face: {
      family: "Crimson Pro Variable",
      url: crimsonUrl,
      weight: "200 900",
    },
  },
  {
    id: "newsreader",
    labelKey: "reader.fontNewsreader",
    label: "Newsreader",
    sample: FONT_SAMPLE,
    family: '"Newsreader Variable", Georgia, serif',
    face: {
      family: "Newsreader Variable",
      url: newsreaderUrl,
      weight: "200 800",
    },
  },
  {
    id: "source-sans",
    labelKey: "reader.fontSourceSans",
    label: "Source Sans",
    sample: FONT_SAMPLE,
    family: '"Source Sans 3 Variable", system-ui, sans-serif',
    face: {
      family: "Source Sans 3 Variable",
      url: sourceSansUrl,
      weight: "200 900",
    },
  },
  {
    id: "ibm-plex",
    labelKey: "reader.fontIbmPlex",
    label: "IBM Plex Sans",
    sample: FONT_SAMPLE,
    family: '"IBM Plex Sans Variable", system-ui, sans-serif',
    face: {
      family: "IBM Plex Sans Variable",
      url: ibmPlexUrl,
      weight: "100 700",
    },
  },
  {
    id: "dm-sans",
    labelKey: "reader.fontDmSans",
    label: "DM Sans",
    sample: FONT_SAMPLE,
    family: '"DM Sans Variable", system-ui, sans-serif',
    face: {
      family: "DM Sans Variable",
      url: dmSansUrl,
      weight: "100 1000",
    },
  },
  {
    id: "fraunces",
    labelKey: "reader.fontFraunces",
    label: "Fraunces",
    sample: FONT_SAMPLE,
    family: '"Fraunces Variable", Georgia, serif',
    face: {
      family: "Fraunces Variable",
      url: frauncesUrl,
      weight: "100 900",
    },
  },
  {
    id: "serif",
    labelKey: "reader.fontSerif",
    label: "Serif",
    sample: FONT_SAMPLE,
    family: 'Georgia, "Times New Roman", Times, serif',
  },
  {
    id: "sans",
    labelKey: "reader.fontSans",
    label: "Sans-serif",
    sample: FONT_SAMPLE,
    family: 'system-ui, -apple-system, "Segoe UI", Roboto, Helvetica, Arial, sans-serif',
  },
  {
    id: "mono",
    labelKey: "reader.fontMono",
    label: "Monospace",
    sample: FONT_SAMPLE,
    family: 'ui-monospace, "Cascadia Code", "Segoe UI Mono", Menlo, Consolas, monospace',
  },
  {
    id: "palatino",
    labelKey: "reader.fontPalatino",
    label: "Palatino",
    sample: FONT_SAMPLE,
    family: '"Palatino Linotype", Palatino, "Book Antiqua", Georgia, serif',
  },
  {
    id: "verdana",
    labelKey: "reader.fontVerdana",
    label: "Verdana",
    sample: FONT_SAMPLE,
    family: "Verdana, Geneva, Tahoma, sans-serif",
  },
];

export interface StoredCustomFont {
  fileName: string;
  mimeType: string;
  data: ArrayBuffer;
}

export function isBuiltInFontId(value: string): value is BuiltInFontId {
  return BUILTIN_EPUB_FONTS.some((f) => f.id === value);
}

export function loadFontPreference(): EpubFontId {
  if (typeof localStorage === "undefined") return "book";
  const saved = localStorage.getItem(EPUB_FONT_KEY);
  if (saved === "custom") return "custom";
  if (saved && isBuiltInFontId(saved)) return saved;
  return "book";
}

export function saveFontPreference(id: EpubFontId): void {
  localStorage.setItem(EPUB_FONT_KEY, id);
}

function openFontDb(): Promise<IDBDatabase> {
  return new Promise((resolve, reject) => {
    const req = indexedDB.open(DB_NAME, DB_VERSION);
    req.onerror = () => reject(req.error ?? new Error("Failed to open font database"));
    req.onupgradeneeded = () => {
      const db = req.result;
      if (!db.objectStoreNames.contains(STORE)) {
        db.createObjectStore(STORE);
      }
    };
    req.onsuccess = () => resolve(req.result);
  });
}

export async function loadCustomFont(): Promise<StoredCustomFont | null> {
  if (typeof indexedDB === "undefined") return null;
  const db = await openFontDb();
  return new Promise((resolve, reject) => {
    const tx = db.transaction(STORE, "readonly");
    const req = tx.objectStore(STORE).get("custom");
    req.onerror = () => reject(req.error ?? new Error("Failed to read custom font"));
    req.onsuccess = () => {
      db.close();
      resolve((req.result as StoredCustomFont | undefined) ?? null);
    };
  });
}

export async function saveCustomFont(file: File): Promise<StoredCustomFont> {
  if (file.size > 4 * 1024 * 1024) {
    throw new Error("Font file is too large (max 4 MB)");
  }
  const stored: StoredCustomFont = {
    fileName: file.name,
    mimeType: file.type || guessFontMime(file.name),
    data: await file.arrayBuffer(),
  };
  const db = await openFontDb();
  await new Promise<void>((resolve, reject) => {
    const tx = db.transaction(STORE, "readwrite");
    tx.oncomplete = () => {
      db.close();
      resolve();
    };
    tx.onerror = () => reject(tx.error ?? new Error("Failed to save custom font"));
    tx.objectStore(STORE).put(stored, "custom");
  });
  return stored;
}

export async function clearCustomFont(): Promise<void> {
  if (typeof indexedDB === "undefined") return;
  const db = await openFontDb();
  await new Promise<void>((resolve, reject) => {
    const tx = db.transaction(STORE, "readwrite");
    tx.oncomplete = () => {
      db.close();
      resolve();
    };
    tx.onerror = () => reject(tx.error ?? new Error("Failed to clear custom font"));
    tx.objectStore(STORE).delete("custom");
  });
}

function guessFontMime(fileName: string): string {
  const lower = fileName.toLowerCase();
  if (lower.endsWith(".woff2")) return "font/woff2";
  if (lower.endsWith(".woff")) return "font/woff";
  if (lower.endsWith(".otf")) return "font/otf";
  return "font/ttf";
}

function formatFromMime(mimeType: string): string {
  if (mimeType.includes("woff2")) return "woff2";
  if (mimeType.includes("woff")) return "woff";
  if (mimeType.includes("otf")) return "opentype";
  return "truetype";
}

export function customFontCss(stored: StoredCustomFont): string {
  revokeCustomFontBlobUrl();
  const blob = new Blob([stored.data], { type: stored.mimeType });
  const url = URL.createObjectURL(blob);
  customFontBlobUrl = url;
  const format = formatFromMime(stored.mimeType);
  return `@font-face{font-family:"${CUSTOM_FONT_FAMILY}";src:url("${url}") format("${format}");font-display:swap;}body,p,div,span,li{font-family:"${CUSTOM_FONT_FAMILY}",serif!important;}`;
}

export function vendoredFontCss(face: VendoredFace): string {
  const href = absoluteFontUrl(face.url);
  return `@font-face{font-family:"${face.family}";src:url("${href}") format("woff2-variations");font-weight:${face.weight};font-display:swap;font-style:normal;}`;
}

function absoluteFontUrl(url: string): string {
  if (typeof window === "undefined") return url;
  try {
    return new URL(url, window.location.href).href;
  } catch {
    return url;
  }
}

let customFontBlobUrl: string | null = null;

export function revokeCustomFontBlobUrl(): void {
  if (customFontBlobUrl) {
    URL.revokeObjectURL(customFontBlobUrl);
    customFontBlobUrl = null;
  }
}

export function applyEpubFont(
  rendition: Rendition,
  fontId: EpubFontId,
  customFont: StoredCustomFont | null,
): void {
  if (fontId === "custom" && customFont) {
    rendition.themes.registerCss(CUSTOM_FONT_THEME, customFontCss(customFont));
    rendition.themes.font(`"${CUSTOM_FONT_FAMILY}", serif`);
    return;
  }

  revokeCustomFontBlobUrl();

  if (fontId === "book") {
    (rendition.themes as EpubThemes).removeOverride("font-family");
    return;
  }

  const builtIn = BUILTIN_EPUB_FONTS.find((f) => f.id === fontId);
  if (!builtIn?.family) return;

  if (builtIn.face) {
    rendition.themes.registerCss(VENDORED_FONT_THEME, vendoredFontCss(builtIn.face));
  }
  rendition.themes.font(builtIn.family);
}
