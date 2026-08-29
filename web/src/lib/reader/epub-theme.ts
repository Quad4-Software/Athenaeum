export type ReaderTheme = "light" | "dark" | "sepia" | "night";

export const EPUB_PALETTE: Record<ReaderTheme, { fg: string; bg: string }> = {
  light: { fg: "#1a1a1a", bg: "#ffffff" },
  dark: { fg: "#e6e6e6", bg: "#16161a" },
  sepia: { fg: "#3b2f2f", bg: "#f4ecd8" },
  night: { fg: "#b8c0cc", bg: "#0d1117" },
};

export function resolvedEpubTheme(readerTheme: ReaderTheme, appDark: boolean): ReaderTheme {
  if (readerTheme === "light" || readerTheme === "sepia") return readerTheme;
  if (readerTheme === "night") return "night";
  if (appDark) return "dark";
  return readerTheme === "dark" ? "dark" : "light";
}

export function injectEpubContentBackground(doc: Document, bg: string): void {
  doc.documentElement.style.backgroundColor = bg;
  if (doc.body) doc.body.style.backgroundColor = bg;
}
