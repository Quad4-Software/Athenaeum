import { brand } from "./config";
import { applyUiFont, DEFAULT_UI_FONT, isUiFontId } from "./fonts";
import { loadUiFontCss } from "./load-ui-font";
import { storageKey } from "./storage";
import { applyThemeTokens, getAppTheme, resolveThemeId } from "./themes";

function setMeta(name: string, content: string): void {
  let el = document.querySelector(`meta[name="${name}"]`);
  if (!el) {
    el = document.createElement("meta");
    el.setAttribute("name", name);
    document.head.appendChild(el);
  }
  el.setAttribute("content", content);
}

/** Apply fork branding before the Svelte app mounts. */
export function initBrand(): void {
  document.title = brand.appName;
  setMeta("description", brand.appDescription);

  // Match ThemeStore.initial() so this does not undo the store after module init.
  const saved = localStorage.getItem(storageKey("theme"));
  const preference =
    saved === "light" || saved === "dark" || saved === "system" || (saved && getAppTheme(saved))
      ? saved
      : "system";
  const themeId = resolveThemeId(preference);
  const theme = getAppTheme(themeId);
  if (theme) {
    applyThemeTokens(theme);
    setMeta("theme-color", brand.themeColor[theme.tokens.colorScheme]);
  }

  const savedFont = localStorage.getItem(storageKey("ui-font"));
  const fontId = savedFont && isUiFontId(savedFont) ? savedFont : DEFAULT_UI_FONT;
  void loadUiFontCss(fontId).then(() => applyUiFont(fontId));
}

export { brand } from "./config";
export { storageKey, legacyStorageKey, IDB_BOOKS, IDB_AUDIO, IDB_FONTS, DEMO_MODE_STORAGE_KEY } from "./storage";
export { applyUiFont, DEFAULT_UI_FONT, getUiFont, isUiFontId, UI_FONT_PRESETS } from "./fonts";
export type { UiFontId, UiFontPreset } from "./fonts";
export {
  applyThemeTokens,
  getAppTheme,
  listAppThemes,
  registerAppTheme,
  resolveSystemThemeId,
  resolveThemeId,
} from "./themes";
export type { AppTheme, ThemeTokens } from "./themes";
