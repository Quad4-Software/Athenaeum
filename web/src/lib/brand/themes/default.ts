import type { AppTheme } from "./tokens";

/**
 * Default light and dark palettes. Forks edit these values to re-skin the app.
 * app.css contains matching fallbacks for first paint before JS runs.
 * Warm literary neutrals (hue ~55) with orange-red primary for CTAs.
 */
export const defaultLightTheme: AppTheme = {
  id: "light",
  label: "Light",
  tokens: {
    colorScheme: "light",
    bg: "oklch(0.97 0.01 75)",
    bgElevated: "oklch(0.99 0.006 75)",
    surface: "oklch(0.995 0.004 75)",
    surfaceHover: "oklch(0.94 0.012 70)",
    overlay: "oklch(0.25 0.02 55 / 0.45)",
    fg: "oklch(0.22 0.02 55)",
    fgMuted: "oklch(0.45 0.02 55)",
    fgSubtle: "oklch(0.58 0.015 55)",
    border: "oklch(0.88 0.015 70)",
    borderStrong: "oklch(0.78 0.02 65)",
    primary: "oklch(0.55 0.2 25)",
    primaryHover: "oklch(0.5 0.21 25)",
    primaryFg: "oklch(0.99 0 0)",
    accent: "oklch(0.55 0.18 25)",
    danger: "oklch(0.55 0.24 18)",
    success: "oklch(0.55 0.14 150)",
    ring: "oklch(0.55 0.2 25 / 0.5)",
    shadow: "0 12px 32px -18px rgb(40 28 18 / 0.35)",
  },
};

export const defaultDarkTheme: AppTheme = {
  id: "dark",
  label: "Dark",
  tokens: {
    colorScheme: "dark",
    bg: "oklch(0.15 0.014 55)",
    bgElevated: "oklch(0.18 0.016 55)",
    surface: "oklch(0.2 0.016 55)",
    surfaceHover: "oklch(0.25 0.018 55)",
    overlay: "oklch(0.1 0.012 55 / 0.72)",
    fg: "oklch(0.95 0.008 75)",
    fgMuted: "oklch(0.72 0.015 70)",
    fgSubtle: "oklch(0.55 0.015 65)",
    border: "oklch(0.3 0.018 55)",
    borderStrong: "oklch(0.4 0.02 55)",
    primary: "oklch(0.62 0.19 28)",
    primaryHover: "oklch(0.68 0.2 28)",
    primaryFg: "oklch(0.99 0 0)",
    accent: "oklch(0.72 0.14 35)",
    danger: "oklch(0.62 0.22 18)",
    success: "oklch(0.68 0.14 150)",
    ring: "oklch(0.62 0.19 28 / 0.55)",
    shadow: "0 14px 36px -14px rgb(0 0 0 / 0.65)",
  },
};
