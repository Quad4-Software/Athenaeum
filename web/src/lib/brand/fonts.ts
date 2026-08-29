/**
 * UI and reading font catalogs. Families are self-hosted via @fontsource-variable.
 */

export type UiFontId =
  | "athenaeum"
  | "source-serif"
  | "literata"
  | "crimson"
  | "newsreader"
  | "ibm-plex"
  | "dm-sans"
  | "system";

export interface UiFontPreset {
  id: UiFontId;
  label: string;
  sample: string;
  /** CSS font-family used for the preview label and body text. */
  family: string;
  sans: string;
  display: string;
}

const SANS_FALLBACK =
  'ui-sans-serif, system-ui, -apple-system, "Segoe UI", Roboto, Helvetica, Arial, sans-serif';
const SERIF_FALLBACK = 'Georgia, "Times New Roman", Times, serif';
const SYSTEM_SANS = `system-ui, -apple-system, "Segoe UI", Roboto, Helvetica, Arial, sans-serif`;
const SYSTEM_SERIF = `ui-serif, Georgia, "Times New Roman", Times, serif`;

export const DEFAULT_UI_FONT: UiFontId = "athenaeum";

export const UI_FONT_PRESETS: UiFontPreset[] = [
  {
    id: "athenaeum",
    label: "Athenaeum",
    sample: "Aa The quick brown fox",
    family: '"Source Sans 3 Variable", Source Sans 3, ' + SANS_FALLBACK,
    sans: '"Source Sans 3 Variable", "Source Sans 3", ' + SANS_FALLBACK,
    display: '"Fraunces Variable", Fraunces, "Source Serif 4", ' + SERIF_FALLBACK,
  },
  {
    id: "source-serif",
    label: "Source Serif",
    sample: "Aa The quick brown fox",
    family: '"Source Serif 4 Variable", Source Serif 4, ' + SERIF_FALLBACK,
    sans: '"Source Serif 4 Variable", "Source Serif 4", ' + SERIF_FALLBACK,
    display: '"Source Serif 4 Variable", "Source Serif 4", ' + SERIF_FALLBACK,
  },
  {
    id: "literata",
    label: "Literata",
    sample: "Aa The quick brown fox",
    family: '"Literata Variable", Literata, ' + SERIF_FALLBACK,
    sans: '"Literata Variable", Literata, ' + SERIF_FALLBACK,
    display: '"Literata Variable", Literata, ' + SERIF_FALLBACK,
  },
  {
    id: "crimson",
    label: "Crimson Pro",
    sample: "Aa The quick brown fox",
    family: '"Crimson Pro Variable", "Crimson Pro", ' + SERIF_FALLBACK,
    sans: '"Crimson Pro Variable", "Crimson Pro", ' + SERIF_FALLBACK,
    display: '"Crimson Pro Variable", "Crimson Pro", ' + SERIF_FALLBACK,
  },
  {
    id: "newsreader",
    label: "Newsreader",
    sample: "Aa The quick brown fox",
    family: '"Newsreader Variable", Newsreader, ' + SERIF_FALLBACK,
    sans: '"Newsreader Variable", Newsreader, ' + SERIF_FALLBACK,
    display: '"Newsreader Variable", Newsreader, ' + SERIF_FALLBACK,
  },
  {
    id: "ibm-plex",
    label: "IBM Plex Sans",
    sample: "Aa The quick brown fox",
    family: '"IBM Plex Sans Variable", "IBM Plex Sans", ' + SANS_FALLBACK,
    sans: '"IBM Plex Sans Variable", "IBM Plex Sans", ' + SANS_FALLBACK,
    display: '"IBM Plex Sans Variable", "IBM Plex Sans", ' + SANS_FALLBACK,
  },
  {
    id: "dm-sans",
    label: "DM Sans",
    sample: "Aa The quick brown fox",
    family: '"DM Sans Variable", "DM Sans", ' + SANS_FALLBACK,
    sans: '"DM Sans Variable", "DM Sans", ' + SANS_FALLBACK,
    display: '"DM Sans Variable", "DM Sans", ' + SANS_FALLBACK,
  },
  {
    id: "system",
    label: "System",
    sample: "Aa The quick brown fox",
    family: SYSTEM_SANS,
    sans: SYSTEM_SANS,
    display: SYSTEM_SERIF,
  },
];

export function getUiFont(id: string | null | undefined): UiFontPreset {
  return UI_FONT_PRESETS.find((f) => f.id === id) ?? UI_FONT_PRESETS[0]!;
}

export function isUiFontId(value: string): value is UiFontId {
  return UI_FONT_PRESETS.some((f) => f.id === value);
}

/** Apply UI font CSS variables on :root. Safe to call before Svelte mounts. */
export function applyUiFont(id: UiFontId | string): void {
  if (typeof document === "undefined") return;
  const preset = getUiFont(id);
  const root = document.documentElement;
  root.style.setProperty("--font-sans", preset.sans);
  root.style.setProperty("--font-display", preset.display);
  root.setAttribute("data-ui-font", preset.id);
}
