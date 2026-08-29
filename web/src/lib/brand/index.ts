import { brand } from "./config";
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
}

export { brand } from "./config";
export { storageKey, legacyStorageKey } from "./storage";
export {
  applyThemeTokens,
  getAppTheme,
  listAppThemes,
  registerAppTheme,
  resolveSystemThemeId,
  resolveThemeId,
} from "./themes";
export type { AppTheme, ThemeTokens } from "./themes";
