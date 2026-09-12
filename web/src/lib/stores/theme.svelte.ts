/**
 * Theme store: persists the user's theme preference and applies fork tokens.
 */

import { applyThemeTokens, brand, getAppTheme, resolveThemeId, storageKey } from "$lib/brand";
import { PersistedState, watch } from "runed";
import { MediaQuery } from "svelte/reactivity";

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

function isThemePreference(value: string): value is ThemePreference {
  return (
    value === "light" || value === "dark" || value === "system" || getAppTheme(value) !== undefined
  );
}

class ThemeStore {
  #persisted = new PersistedState<ThemePreference>(STORAGE_KEY, "system", {
    serializer: {
      // Stored as a bare string ("light" | "dark" | "system" | theme id), not JSON.
      serialize: (value) => value,
      deserialize: (value) => (isThemePreference(value) ? value : "system"),
    },
  });

  mode = $state<ThemeMode>("dark");
  activeThemeId = $state<string>("dark");

  constructor() {
    this.apply();
    if (typeof window !== "undefined") {
      const prefersLight = new MediaQuery("(prefers-color-scheme: light)");
      $effect.root(() => {
        watch(
          () => prefersLight.current,
          () => {
            if (this.preference === "system") this.apply();
          },
          { lazy: true },
        );
      });
    }
  }

  get preference(): ThemePreference {
    return this.#persisted.current;
  }

  set preference(preference: ThemePreference) {
    this.#persisted.current = preference;
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
