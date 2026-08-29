import { defaultDarkTheme, defaultLightTheme } from "./default";
import type { AppTheme } from "./tokens";

const themes = new Map<string, AppTheme>([
  [defaultLightTheme.id, defaultLightTheme],
  [defaultDarkTheme.id, defaultDarkTheme],
]);

/** Register an extra theme (e.g. sepia, high-contrast) or replace light/dark. */
export function registerAppTheme(theme: AppTheme): void {
  themes.set(theme.id, theme);
}

export function getAppTheme(id: string): AppTheme | undefined {
  return themes.get(id);
}

export function listAppThemes(): AppTheme[] {
  return [...themes.values()];
}

export function resolveSystemThemeId(): "light" | "dark" {
  return window.matchMedia("(prefers-color-scheme: light)").matches ? "light" : "dark";
}

export function resolveThemeId(preference: string): string {
  if (preference === "system") {
    return resolveSystemThemeId();
  }
  return themes.has(preference) ? preference : resolveSystemThemeId();
}

export { defaultDarkTheme, defaultLightTheme };
export type { AppTheme, ColorScheme, ThemeTokens } from "./tokens";
export { applyThemeTokens } from "./tokens";
