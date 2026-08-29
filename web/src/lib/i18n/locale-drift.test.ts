import { readFileSync, readdirSync } from "node:fs";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";
import { describe, expect, it } from "vitest";

const here = dirname(fileURLToPath(import.meta.url));
const localesDir = join(here, "locales");

type Json = string | { [k: string]: Json };

function flatten(obj: Json, prefix = ""): Set<string> {
  const out = new Set<string>();
  if (typeof obj !== "object" || obj === null) {
    if (prefix) out.add(prefix);
    return out;
  }
  for (const [k, v] of Object.entries(obj)) {
    if (k.startsWith("$")) continue;
    const path = prefix ? `${prefix}.${k}` : k;
    if (typeof v === "object" && v !== null) {
      for (const child of flatten(v, path)) out.add(child);
    } else {
      out.add(path);
    }
  }
  return out;
}

function loadLocale(code: string): Set<string> {
  const raw = JSON.parse(readFileSync(join(localesDir, `${code}.json`), "utf8")) as Json;
  return flatten(raw);
}

describe("i18n locale drift", () => {
  const codes = readdirSync(localesDir)
    .filter((f) => f.endsWith(".json"))
    .map((f) => f.replace(/\.json$/, ""));

  it("ships en as the canonical key set", () => {
    expect(codes).toContain("en");
    expect(loadLocale("en").size).toBeGreaterThan(100);
  });

  it("keeps non-en locales in key parity with en", () => {
    const en = loadLocale("en");
    for (const code of codes) {
      if (code === "en") continue;
      const keys = loadLocale(code);
      const missing = [...en].filter((k) => !keys.has(k)).sort();
      const extra = [...keys].filter((k) => !en.has(k)).sort();
      expect(missing, `${code} missing keys`).toEqual([]);
      expect(extra, `${code} extra keys`).toEqual([]);
    }
  });
});

describe("backend i18n subset drift", () => {
  it("keeps Go default locale keys as a subset of web en", () => {
    const webEn = loadLocale("en");
    const goPath = join(here, "../../../../internal/i18n/locales/en.json");
    const goRaw = JSON.parse(readFileSync(goPath, "utf8")) as Json;
    const goKeys = flatten(goRaw);
    const missing = [...goKeys].filter((k) => !webEn.has(k)).sort();
    expect(missing, "Go locale keys missing from web en").toEqual([]);
  });
});
