import type { Messages } from "./core";
import { parseLocaleJSON } from "./core";

export type { Messages };
export {
  clearMissingKeys,
  detectBrowserLocale,
  flattenMessages,
  getMissingKeys,
  interpolate,
  parseLocaleJSON,
  setMissingKeyHandler,
  translate,
  localeCodeFromPath,
  localeFileName,
} from "./core";
export type { MissingKeyHandler, MissingKeyReason } from "./core";

export interface LocaleInfo {
  code: string;
  name: string;
  source: "bundled" | "custom";
}

const localeLoaders = import.meta.glob("./locales/*.json", {
  import: "default",
}) as Record<string, () => Promise<unknown>>;

function codeFromPath(path: string): string | null {
  return path.match(/\/([^/]+)\.json$/)?.[1] ?? null;
}

export function bundledLocaleCodes(): string[] {
  const codes: string[] = [];
  for (const path of Object.keys(localeLoaders)) {
    const code = codeFromPath(path);
    if (code) codes.push(code);
  }
  return codes.sort();
}

export async function loadBundledLocale(
  code: string,
): Promise<{ messages: Messages; name: string } | null> {
  const path = Object.keys(localeLoaders).find((p) => codeFromPath(p) === code);
  if (!path) return null;
  const data = await localeLoaders[path]();
  const parsed = parseLocaleJSON(data);
  if (!parsed || Object.keys(parsed.messages).length === 0) return null;
  return parsed;
}

export async function loadBundledLocales(): Promise<
  Map<string, { messages: Messages; name: string }>
> {
  const out = new Map<string, { messages: Messages; name: string }>();
  await Promise.all(
    Object.keys(localeLoaders).map(async (path) => {
      const code = codeFromPath(path);
      if (!code) return;
      const data = await localeLoaders[path]();
      const parsed = parseLocaleJSON(data);
      if (!parsed || Object.keys(parsed.messages).length === 0) return;
      out.set(code, parsed);
    }),
  );
  return out;
}

export function mergeMessages(base: Messages, overlay: Messages): Messages {
  return { ...base, ...overlay };
}
