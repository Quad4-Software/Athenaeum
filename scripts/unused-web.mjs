#!/usr/bin/env node
/**
 * Lightweight unused frontend dependency check (knip alternative).
 * Scans web/ source and config for package name references.
 */
import { readFileSync, readdirSync, statSync } from "node:fs";
import { join, extname } from "node:path";
import { fileURLToPath } from "node:url";
import { dirname } from "node:path";

const webRoot = join(dirname(fileURLToPath(import.meta.url)), "../web");
const pkg = JSON.parse(readFileSync(join(webRoot, "package.json"), "utf8"));

const ignore = new Set([
  "@lhci/cli",
  "@stryker-mutator/core",
  "@stryker-mutator/vitest-runner",
  "@vitest/coverage-v8",
  "@playwright/test",
  "typescript",
  "svelte-check",
  "prettier",
  "prettier-plugin-svelte",
  "eslint",
  "@eslint/js",
  "eslint-plugin-svelte",
  "typescript-eslint",
  "globals",
  "@types/node",
  "@tsconfig/svelte",
  "jsdom",
  "@testing-library/jest-dom",
  "@testing-library/svelte",
  "@testing-library/user-event",
  "fast-check",
  "vitest",
  "workbox-window",
]);

const exts = new Set([".ts", ".js", ".svelte", ".cjs", ".mjs", ".json", ".css", ".html"]);

function walk(dir, out = []) {
  for (const name of readdirSync(dir)) {
    if (name === "node_modules" || name === "coverage" || name === "dist" || name === ".svelte-kit") {
      continue;
    }
    const p = join(dir, name);
    const st = statSync(p);
    if (st.isDirectory()) walk(p, out);
    else if (exts.has(extname(name))) out.push(p);
  }
  return out;
}

const files = [
  ...walk(join(webRoot, "src")),
  join(webRoot, "vite.config.ts"),
  join(webRoot, "vitest-setup.ts"),
  join(webRoot, "playwright.config.ts"),
  join(webRoot, "eslint.config.js"),
  join(webRoot, "svelte.config.js"),
  join(webRoot, "lighthouserc.cjs"),
  join(webRoot, "stryker.config.js"),
].filter((p) => {
  try {
    statSync(p);
    return true;
  } catch {
    return false;
  }
});

const blob = files.map((f) => readFileSync(f, "utf8")).join("\n");

const deps = {
  ...pkg.dependencies,
  ...pkg.devDependencies,
};

const unused = [];
for (const name of Object.keys(deps).sort()) {
  if (ignore.has(name)) continue;
  const patterns = [name, name.replace(/^@/, "").split("/").pop()].filter(Boolean);
  const hit = patterns.some((p) => blob.includes(p) || blob.includes(`from "${name}"`) || blob.includes(`from '${name}'`));
  if (!hit) unused.push(name);
}

if (unused.length) {
  console.log("Possibly unused dependencies:");
  for (const u of unused) console.log(`  - ${u}`);
  process.exit(1);
}

console.log(`Checked ${Object.keys(deps).length} packages against ${files.length} files: none unused.`);
