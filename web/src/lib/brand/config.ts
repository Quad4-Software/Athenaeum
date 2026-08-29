/**
 * Fork branding: edit this file to rename the app and set defaults.
 * Also update web/public/favicon.svg and matching i18n app.title if you use translations.
 */
export const brand = {
  /** Display name shown in sidebar, settings, and document title */
  appName: "Athenaeum",

  /** HTML meta description */
  appDescription:
    "Self-hosted EPUB, PDF, and audiobook library with search, collections, and reading progress.",

  /**
   * Prefix for localStorage keys (e.g. "mylib" -> "mylib:theme").
   * Change when forking so prefs do not clash with upstream.
   */
  storagePrefix: "athenaeum",

  /** Exported server config download filename */
  configExportName: "athenaeum-config.json",

  /** API key prefix shown in settings documentation */
  apiKeyPrefix: "ath_",

  /** CSRF cookie name (must match server auth.CSRFCookie) */
  csrfCookie: "athenaeum_csrf",

  /** Initial html data-theme before the theme store loads */
  defaultThemeId: "dark" as const,

  /** theme-color meta tag for mobile browser chrome (hex) */
  themeColor: {
    light: "#f7f3eb",
    dark: "#1c1712",
  },
} as const;

export type BrandConfig = typeof brand;
