/**
 * Capture showcase screenshots of the offline demo SPA into ./showcase.
 *
 * Usage:
 *   task showcase
 *   # or after task build:demo:
 *   node scripts/showcase-screenshots.mjs
 */
import { createRequire } from "node:module";
import { mkdir, writeFile } from "node:fs/promises";
import { dirname, resolve } from "node:path";
import { fileURLToPath } from "node:url";
import { createServer } from "node:http";
import { createReadStream, existsSync, statSync } from "node:fs";
import { extname, join, normalize } from "node:path";

const require = createRequire(import.meta.url);
const { chromium } = require("../web/node_modules/@playwright/test");

const __dirname = dirname(fileURLToPath(import.meta.url));
const root = resolve(__dirname, "..");
const siteDir = resolve(root, "site");
const outDir = resolve(root, "showcase");

const mime = {
  ".html": "text/html; charset=utf-8",
  ".js": "text/javascript; charset=utf-8",
  ".css": "text/css; charset=utf-8",
  ".svg": "image/svg+xml",
  ".png": "image/png",
  ".woff2": "font/woff2",
  ".json": "application/json",
  ".webmanifest": "application/manifest+json",
};

function startStaticServer(dir, port) {
  return new Promise((resolvePromise, reject) => {
    const server = createServer((req, res) => {
      const urlPath = decodeURIComponent((req.url ?? "/").split("?")[0] || "/");
      let filePath = normalize(join(dir, urlPath === "/" ? "index.html" : urlPath));
      if (!filePath.startsWith(dir)) {
        res.writeHead(403).end();
        return;
      }
      if (!existsSync(filePath) || !statSync(filePath).isFile()) {
        filePath = join(dir, "index.html");
      }
      if (!existsSync(filePath)) {
        res.writeHead(404).end("missing demo build (run pnpm build:demo)");
        return;
      }
      res.writeHead(200, { "Content-Type": mime[extname(filePath)] ?? "application/octet-stream" });
      createReadStream(filePath).pipe(res);
    });
    server.once("error", reject);
    server.listen(port, "127.0.0.1", () => {
      resolvePromise({
        url: `http://127.0.0.1:${port}`,
        close: () =>
          new Promise((r, j) => {
            server.close((err) => (err ? j(err) : r()));
          }),
      });
    });
  });
}

async function settle(page) {
  await page.waitForLoadState("networkidle").catch(() => undefined);
  await page.waitForTimeout(800);
  const toastClose = page
    .locator('[role="status"] button, .toast button')
    .filter({ hasText: /×|x|close|dismiss/i })
    .first();
  if (await toastClose.isVisible().catch(() => false)) {
    await toastClose.click().catch(() => undefined);
    await page.waitForTimeout(200);
  }
  // Hide offline-ready toast if still visible.
  await page
    .evaluate(() => {
      document.querySelectorAll('[role="status"], .toast').forEach((el) => {
        if (/offline/i.test(el.textContent || "")) el.remove();
      });
    })
    .catch(() => undefined);
}

async function shot(page, name) {
  const path = join(outDir, `${name}.png`);
  await page.screenshot({ path, fullPage: false });
  console.log("wrote", path);
}


async function writeThemeSplit() {
  const dark = join(outDir, "library-dark.png");
  const light = join(outDir, "library-light.png");
  const split = join(outDir, "library-theme-split.png");
  if (!existsSync(dark) || !existsSync(light)) return;

  const { spawnSync } = await import("node:child_process");
  const script = join(root, "scripts/showcase-theme-split.py");
  const result = spawnSync("python3", [script, dark, light, split], { encoding: "utf8" });
  if (result.status !== 0) {
    console.warn("theme split failed:", result.stderr || result.stdout);
    return;
  }
  if (result.stdout?.trim()) console.log(result.stdout.trim());
}

async function syncDerivedAssets() {
  const { copyFile, mkdir } = await import("node:fs/promises");
  const { spawnSync } = await import("node:child_process");
  const magick = existsSync("/usr/bin/magick") ? "magick" : "convert";

  const publicDir = join(root, "web/public/showcase");
  const docsDir = join(root, "docs/static/img/showcase");
  const distDir = join(root, "internal/assets/dist/showcase");
  await mkdir(publicDir, { recursive: true });
  await mkdir(docsDir, { recursive: true });
  await mkdir(distDir, { recursive: true });

  for (const theme of ["light", "dark"]) {
    const src = join(outDir, `library-${theme}.png`);
    if (!existsSync(src)) continue;
    const jpg = join(publicDir, `library-${theme}.jpg`);
    const jpgResult = spawnSync(magick, [src, "-quality", "82", jpg], { encoding: "utf8" });
    if (jpgResult.status === 0) {
      console.log("wrote", jpg);
      await copyFile(jpg, join(distDir, `library-${theme}.jpg`));
    }
  }

  const docsShots = [
    "library-light",
    "library-dark",
    "library-mobile-light",
    "library-mobile-dark",
    "book-detail-light",
    "book-detail-dark",
    "settings-light",
    "settings-dark",
  ];
  for (const name of docsShots) {
    const src = join(outDir, `${name}.png`);
    if (!existsSync(src)) continue;
    await copyFile(src, join(docsDir, `${name}.png`));
  }
}

async function openDemo(browser, url, theme, viewport = { width: 1440, height: 900 }) {
  const page = await browser.newPage({ viewport, colorScheme: theme });
  await page.addInitScript((t) => {
    localStorage.setItem("athenaeum:theme", t);
    localStorage.setItem("athenaeum-demo", "1");
  }, theme);
  await page.goto(`${url}/?demo=1`, { waitUntil: "domcontentloaded" });
  await page.waitForFunction(
    (t) => document.documentElement.getAttribute("data-theme") === t,
    theme,
    { timeout: 15_000 },
  );
  await settle(page);
  return page;
}

async function captureTheme(browser, url, theme) {
  const page = await openDemo(browser, url, theme);
  await shot(page, `library-${theme}`);

  const firstCard = page.locator('a[href^="/book/"]').first();
  if (await firstCard.count()) {
    await firstCard.click();
    await page.waitForFunction(
      (t) => document.documentElement.getAttribute("data-theme") === t,
      theme,
      { timeout: 10_000 },
    );
    await settle(page);
    await shot(page, `book-detail-${theme}`);
  }

  await page.goto(`${url}/settings?demo=1`, { waitUntil: "domcontentloaded" });
  await page.waitForFunction(
    (t) => document.documentElement.getAttribute("data-theme") === t,
    theme,
    { timeout: 10_000 },
  );
  await settle(page);
  await shot(page, `settings-${theme}`);
  await page.close();

  const mobile = await openDemo(browser, url, theme, { width: 390, height: 844 });
  await shot(mobile, `library-mobile-${theme}`);
  await mobile.close();
}

async function main() {
  if (!existsSync(join(siteDir, "index.html"))) {
    console.error("missing demo build: run task build:demo first");
    process.exit(1);
  }

  await mkdir(outDir, { recursive: true });
  const server = await startStaticServer(siteDir, 4177);
  const browser = await chromium.launch();
  try {
    await captureTheme(browser, server.url, "light");
    await captureTheme(browser, server.url, "dark");

    // Convenience aliases used by older README links.
    const { copyFile } = await import("node:fs/promises");
    await copyFile(join(outDir, "library-light.png"), join(outDir, "library.png"));
    await copyFile(join(outDir, "book-detail-light.png"), join(outDir, "book-detail.png"));
    await copyFile(join(outDir, "settings-light.png"), join(outDir, "settings.png"));
    await copyFile(join(outDir, "library-mobile-light.png"), join(outDir, "library-mobile.png"));

    await writeThemeSplit();
    await syncDerivedAssets();

    await writeFile(
      join(outDir, "README.txt"),
      [
        "Showcase screenshots generated by scripts/showcase-screenshots.mjs",
        "Regenerate with: task showcase",
        "",
        "Light: library-light, book-detail-light, settings-light, library-mobile-light",
        "Dark:  library-dark, book-detail-dark, settings-dark, library-mobile-dark",
        "Aliases: library, book-detail, settings, library-mobile (= light)",
        "Hero: library-theme-split.png (dark/light diagonal or half composite)",
        "Also syncs web/public/showcase/*.jpg and docs/static/img/showcase/*.png",
        "",
      ].join("\n"),
    );
  } finally {
    await browser.close();
    await server.close();
  }
}

main().catch((err) => {
  console.error(err);
  process.exit(1);
});
