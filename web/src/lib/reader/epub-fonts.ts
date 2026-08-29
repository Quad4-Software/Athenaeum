import type { Rendition } from "epubjs";
import { storageKey } from "$lib/brand/storage";

type EpubThemes = Rendition["themes"] & {
  removeOverride(name: string): void;
};

export const EPUB_FONT_KEY = storageKey("epub-font-family");
const CUSTOM_FONT_THEME = "athenaeum-custom-font";
const CUSTOM_FONT_FAMILY = "Athenaeum Custom";
const DB_NAME = "athenaeum-fonts";
const DB_VERSION = 1;
const STORE = "epub";

export type BuiltInFontId = "book" | "serif" | "sans" | "mono" | "palatino" | "verdana";

export type EpubFontId = BuiltInFontId | "custom";

export interface BuiltInEpubFont {
  id: BuiltInFontId;
  labelKey: string;
  family?: string;
}

export const BUILTIN_EPUB_FONTS: BuiltInEpubFont[] = [
  { id: "book", labelKey: "reader.fontBook" },
  {
    id: "serif",
    labelKey: "reader.fontSerif",
    family: 'Georgia, "Times New Roman", Times, serif',
  },
  {
    id: "sans",
    labelKey: "reader.fontSans",
    family: 'system-ui, -apple-system, "Segoe UI", Roboto, Helvetica, Arial, sans-serif',
  },
  {
    id: "mono",
    labelKey: "reader.fontMono",
    family: 'ui-monospace, "Cascadia Code", "Segoe UI Mono", Menlo, Consolas, monospace',
  },
  {
    id: "palatino",
    labelKey: "reader.fontPalatino",
    family: '"Palatino Linotype", Palatino, "Book Antiqua", Georgia, serif',
  },
  {
    id: "verdana",
    labelKey: "reader.fontVerdana",
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
  if (builtIn?.family) {
    rendition.themes.font(builtIn.family);
  }
}
