export type Messages = Record<string, string>;

const META_PREFIX = "$";

export function flattenMessages(obj: Record<string, unknown>, prefix = ""): Messages | null {
  const out: Messages = {};
  for (const [key, val] of Object.entries(obj)) {
    if (key.startsWith(META_PREFIX)) continue;
    const fullKey = prefix ? `${prefix}.${key}` : key;
    if (typeof val === "string") {
      out[fullKey] = val;
      continue;
    }
    if (val && typeof val === "object" && !Array.isArray(val)) {
      const nested = flattenMessages(val as Record<string, unknown>, fullKey);
      if (!nested) return null;
      Object.assign(out, nested);
      continue;
    }
    return null;
  }
  return out;
}

export function parseLocaleJSON(data: unknown): { messages: Messages; name: string } | null {
  if (!data || typeof data !== "object" || Array.isArray(data)) return null;
  const root = data as Record<string, unknown>;
  const name = typeof root.$name === "string" ? root.$name : "";
  const messages = flattenMessages(root);
  if (!messages) return null;
  return { messages, name };
}

export function interpolate(template: string, params?: Record<string, string | number>): string {
  if (!params) return template;
  return template.replace(/\{(\w+)\}/g, (_, key: string) => {
    const val = params[key];
    return val === undefined ? `{${key}}` : String(val);
  });
}

export function translate(
  messages: Messages,
  key: string,
  params?: Record<string, string | number>,
  fallback?: Messages,
): string {
  const raw = messages[key] ?? fallback?.[key] ?? key;
  return interpolate(raw, params);
}

export function detectBrowserLocale(available: string[]): string {
  if (typeof navigator === "undefined") return "en";
  const saved = available;
  const langs = [navigator.language, ...(navigator.languages ?? [])];
  for (const lang of langs) {
    const norm = lang.replace("_", "-");
    if (saved.includes(norm)) return norm;
    const base = norm.split("-")[0];
    if (saved.includes(base)) return base;
  }
  return "en";
}

export function localeFileName(code: string): string {
  return `${code}.json`;
}

export function localeCodeFromPath(path: string): string | null {
  const match = path.match(/\/([^/]+)\.json$/);
  if (!match) return null;
  const code = match[1];
  if (!code || !/^[\w-]+$/.test(code)) return null;
  return code;
}
