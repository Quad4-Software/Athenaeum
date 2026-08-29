/**
 * Theme store: persists the user's theme preference and applies fork tokens.
 */

import { applyThemeTokens, brand, getAppTheme, resolveThemeId, storageKey } from "$lib/brand";

export type ThemePreference = "light" | "dark" | "system" | (string & {});
export type ThemeMode = "light" | "dark";

const STORAGE_KEY = storageKey("theme");

function setThemeColorMeta(mode: ThemeMode): void {
  let el = document.querySelector('meta[name="theme-color"]');
  if (!el) {
    el = document.createElement("meta");
    el.setAttribute("name", "theme-color");
    document.head.appendChild(el);
  }
  el.setAttribute("content", brand.themeColor[mode]);
}

function initial(): ThemePreference {
  const saved = localStorage.getItem(STORAGE_KEY);
  if (saved === "light" || saved === "dark" || saved === "system") return saved;
  if (saved && getAppTheme(saved)) return saved;
  return "system";
}

class ThemeStore {
  preference = $state<ThemePreference>("system");
  mode = $state<ThemeMode>("dark");
  activeThemeId = $state<string>("dark");
  #media: MediaQueryList | null = null;
  #onSystemChange = () => {
    if (this.preference === "system") this.apply();
  };

  constructor() {
    this.preference = initial();
    this.apply();
    if (typeof window !== "undefined") {
      this.#media = window.matchMedia("(prefers-color-scheme: light)");
      this.#media.addEventListener("change", this.#onSystemChange);
    }
  }

  private apply() {
    const themeId = resolveThemeId(this.preference);
    const theme = getAppTheme(themeId);
    if (theme) {
      applyThemeTokens(theme);
      this.mode = theme.tokens.colorScheme;
      this.activeThemeId = themeId;
      setThemeColorMeta(theme.tokens.colorScheme);
      return;
    }
    this.mode = themeId === "light" ? "light" : "dark";
    this.activeThemeId = this.mode;
    document.documentElement.setAttribute("data-theme", this.mode);
    setThemeColorMeta(this.mode);
  }

  set(preference: ThemePreference) {
    this.preference = preference;
    localStorage.setItem(STORAGE_KEY, preference);
    this.apply();
  }

  toggle() {
    const next =
      this.mode === "dark"
        ? "light"
        : this.preference === "light"
          ? "dark"
          : this.mode === "light"
            ? "dark"
            : "light";
    this.set(next);
  }
}

export const theme = new ThemeStore();
