export type ColorScheme = "light" | "dark";

/** CSS custom properties applied to :root for app theming. */
export type ThemeTokens = {
  colorScheme: ColorScheme;
  bg: string;
  bgElevated: string;
  surface: string;
  surfaceHover: string;
  overlay: string;
  fg: string;
  fgMuted: string;
  fgSubtle: string;
  border: string;
  borderStrong: string;
  primary: string;
  primaryHover: string;
  primaryFg: string;
  accent: string;
  danger: string;
  success: string;
  ring: string;
  shadow: string;
};

export type AppTheme = {
  id: string;
  label: string;
  tokens: ThemeTokens;
};

const TOKEN_CSS_MAP: Record<keyof Omit<ThemeTokens, "colorScheme">, string> = {
  bg: "--bg",
  bgElevated: "--bg-elevated",
  surface: "--surface",
  surfaceHover: "--surface-hover",
  overlay: "--overlay",
  fg: "--fg",
  fgMuted: "--fg-muted",
  fgSubtle: "--fg-subtle",
  border: "--border",
  borderStrong: "--border-strong",
  primary: "--primary",
  primaryHover: "--primary-hover",
  primaryFg: "--primary-fg",
  accent: "--accent",
  danger: "--danger",
  success: "--success",
  ring: "--ring",
  shadow: "--shadow",
};

export function applyThemeTokens(theme: AppTheme): void {
  const root = document.documentElement;
  root.setAttribute("data-theme", theme.tokens.colorScheme);
  root.style.colorScheme = theme.tokens.colorScheme;
  for (const [key, cssVar] of Object.entries(TOKEN_CSS_MAP)) {
    const value = theme.tokens[key as keyof typeof TOKEN_CSS_MAP];
    root.style.setProperty(cssVar, value);
  }
}
