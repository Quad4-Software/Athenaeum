import type { AppTheme } from "./tokens";

/**
 * Default light and dark palettes. Forks edit these values to re-skin the app.
 * app.css contains matching fallbacks for first paint before JS runs.
 */
export const defaultLightTheme: AppTheme = {
  id: "light",
  label: "Light",
  tokens: {
    colorScheme: "light",
    bg: "oklch(0.98 0.003 270)",
    bgElevated: "oklch(1 0 0)",
    surface: "oklch(1 0 0)",
    surfaceHover: "oklch(0.96 0.004 270)",
    overlay: "oklch(0.2 0.01 270 / 0.4)",
    fg: "oklch(0.2 0.01 270)",
    fgMuted: "oklch(0.45 0.01 270)",
    fgSubtle: "oklch(0.6 0.01 270)",
    border: "oklch(0.9 0.005 270)",
    borderStrong: "oklch(0.82 0.008 270)",
    primary: "oklch(0.55 0.22 25)",
    primaryHover: "oklch(0.5 0.23 25)",
    primaryFg: "oklch(0.99 0 0)",
    accent: "oklch(0.55 0.2 25)",
    danger: "oklch(0.55 0.24 18)",
    success: "oklch(0.6 0.16 150)",
    ring: "oklch(0.55 0.22 25 / 0.5)",
    shadow: "0 10px 30px -16px rgb(0 0 0 / 0.25)",
  },
};

export const defaultDarkTheme: AppTheme = {
  id: "dark",
  label: "Dark",
  tokens: {
    colorScheme: "dark",
    bg: "oklch(0.16 0.006 270)",
    bgElevated: "oklch(0.2 0.008 270)",
    surface: "oklch(0.22 0.008 270)",
    surfaceHover: "oklch(0.27 0.01 270)",
    overlay: "oklch(0.12 0.006 270 / 0.7)",
    fg: "oklch(0.96 0.004 270)",
    fgMuted: "oklch(0.72 0.01 270)",
    fgSubtle: "oklch(0.55 0.01 270)",
    border: "oklch(0.32 0.01 270)",
    borderStrong: "oklch(0.42 0.012 270)",
    primary: "oklch(0.58 0.21 25)",
    primaryHover: "oklch(0.64 0.22 25)",
    primaryFg: "oklch(0.99 0 0)",
    accent: "oklch(0.7 0.16 25)",
    danger: "oklch(0.62 0.23 18)",
    success: "oklch(0.7 0.16 150)",
    ring: "oklch(0.58 0.21 25 / 0.6)",
    shadow: "0 10px 30px -12px rgb(0 0 0 / 0.6)",
  },
};
