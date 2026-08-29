#!/usr/bin/env node
/**
 * Sync locale JSON keys from en.json into sibling locales.
 * Missing keys are filled with the English string (marked for translation).
 * Extra keys are reported. Use --check to fail without writing.
 */
import { readFileSync, writeFileSync, readdirSync } from "node:fs";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";

const here = dirname(fileURLToPath(import.meta.url));
const localesDir = join(here, "../web/src/lib/i18n/locales");
const checkOnly = process.argv.includes("--check");
const fill = !process.argv.includes("--report-only");

function isObject(v) {
  return v !== null && typeof v === "object" && !Array.isArray(v);
}

function syncObject(en, other, path = "") {
  const missing = [];
  const extra = [];
  const out = { ...other };

  for (const key of Object.keys(en)) {
    if (key.startsWith("$")) continue;
    const childPath = path ? `${path}.${key}` : key;
    if (!(key in other)) {
      missing.push(childPath);
      if (fill) out[key] = en[key];
      continue;
    }
    if (isObject(en[key]) && isObject(other[key])) {
      const nested = syncObject(en[key], other[key], childPath);
      missing.push(...nested.missing);
      extra.push(...nested.extra);
      out[key] = nested.value;
    } else if (isObject(en[key]) !== isObject(other[key])) {
      missing.push(`${childPath} (type mismatch)`);
      if (fill) out[key] = en[key];
    }
  }

  for (const key of Object.keys(other)) {
    if (key.startsWith("$")) continue;
    if (!(key in en)) {
      extra.push(path ? `${path}.${key}` : key);
      if (fill) delete out[key];
    }
  }

  return { value: out, missing, extra };
}

const en = JSON.parse(readFileSync(join(localesDir, "en.json"), "utf8"));
const files = readdirSync(localesDir).filter((f) => f.endsWith(".json") && f !== "en.json");

let failed = false;
for (const file of files) {
  const code = file.replace(/\.json$/, "");
  const raw = JSON.parse(readFileSync(join(localesDir, file), "utf8"));
  const { value, missing, extra } = syncObject(en, raw);
  if (missing.length || extra.length) {
    console.log(`\n[${code}]`);
    if (missing.length) console.log(`  missing (${missing.length}): ${missing.slice(0, 20).join(", ")}${missing.length > 20 ? "…" : ""}`);
    if (extra.length) console.log(`  extra (${extra.length}): ${extra.slice(0, 20).join(", ")}${extra.length > 20 ? "…" : ""}`);
    failed = true;
  } else {
    console.log(`[${code}] ok`);
  }
  if (!checkOnly && fill && (missing.length || extra.length)) {
    writeFileSync(join(localesDir, file), `${JSON.stringify(value, null, 2)}\n`);
    console.log(`  wrote ${file}`);
  }
}

if (checkOnly && failed) {
  console.error("\ni18n key drift detected (run: task i18n:sync)");
  process.exit(1);
}

if (!failed) {
  console.log("\nAll locales match en key structure.");
}
