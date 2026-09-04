/// <reference types="vitest/config" />
import { cpSync, createReadStream, existsSync, mkdirSync, statSync } from "node:fs";
import { createRequire } from "node:module";
import { dirname, extname, join, normalize, resolve } from "node:path";
import { defineConfig, type Plugin } from "vite";
import { svelte } from "@sveltejs/vite-plugin-svelte";
import tailwindcss from "@tailwindcss/vite";
import viteCompression from "vite-plugin-compression2";
import { VitePWA } from "vite-plugin-pwa";
import { brand } from "./src/lib/brand/config.js";
import {
  KOKORO_MODEL_ID,
  KOKORO_VOICE_HF_PREFIX,
  KOKORO_VOICE_LOCAL_PREFIX,
} from "./src/lib/narrator/kokoro-paths.js";

const require = createRequire(import.meta.url);
const rootDir = import.meta.dirname;
const isDemoBuild =
  process.env.VITE_DEMO === "1" ||
  process.env.VITE_DEMO === "true" ||
  process.env.VITE_DEMO === "yes";
const isSlimBuild =
  process.env.VITE_SLIM === "1" ||
  process.env.VITE_SLIM === "true" ||
  process.env.VITE_SLIM === "yes";
/** Full Kokoro embed is for production binaries only (not slim, not Pages demo). */
const omitKokoroAssets = isSlimBuild || isDemoBuild;
const pdfJsRoot = resolve(rootDir, "node_modules/pdfjs-dist");
const kokoroWasmEntry = resolve(rootDir, "src/lib/narrator/kokoro-wasm.ts");
const kokoroWasmSlim = resolve(rootDir, "src/lib/narrator/kokoro-wasm-slim.ts");
const kokoroModelRoot = resolve(rootDir, "vendor/kokoro-model", KOKORO_MODEL_ID);
const pdfJsDirs = ["wasm", "cmaps", "standard_fonts", "iccs"] as const;

function packageDir(name: string): string {
  let dir = dirname(require.resolve(name));
  while (dir !== rootDir && dir !== dirname(dir)) {
    if (existsSync(join(dir, "package.json"))) {
      return dir;
    }
    dir = dirname(dir);
  }
  throw new Error(`could not locate package root for ${name}`);
}

const pdfMimeTypes: Record<string, string> = {
  ".wasm": "application/wasm",
  ".js": "application/javascript",
  ".json": "application/json",
  ".bcmap": "application/octet-stream",
  ".icc": "application/vnd.iccprofile",
  ".pfb": "application/octet-stream",
};

function copyPdfJsAssets(destRoot: string) {
  const pdfDest = join(destRoot, "pdfjs");
  mkdirSync(pdfDest, { recursive: true });
  for (const dir of pdfJsDirs) {
    const src = join(pdfJsRoot, dir);
    if (!existsSync(src)) continue;
    cpSync(src, join(pdfDest, dir), { recursive: true });
  }
}

/** Swap kokoro-wasm for the slim stub so kokoro-js stays out of the graph. */
function slimKokoroPlugin(): Plugin {
  return {
    name: "slim-kokoro",
    enforce: "pre",
    resolveId(id, importer) {
      if (!omitKokoroAssets) return null;
      if (id === kokoroWasmEntry || id === kokoroWasmSlim) {
        return kokoroWasmSlim;
      }
      if (
        (id === "./kokoro-wasm" ||
          id === "./kokoro-wasm.ts" ||
          id.endsWith("/kokoro-wasm") ||
          id.endsWith("/kokoro-wasm.ts")) &&
        importer &&
        normalize(importer).includes(`${normalize(join("src", "lib", "narrator"))}`)
      ) {
        return kokoroWasmSlim;
      }
      return null;
    },
  };
}

function pdfJsAssetsPlugin(): Plugin {
  return {
    name: "pdfjs-assets",
    configureServer(server) {
      server.middlewares.use((req, res, next) => {
        if (!req.url?.startsWith("/pdfjs/")) {
          next();
          return;
        }
        const rel = decodeURIComponent(req.url.slice("/pdfjs/".length).split("?")[0] ?? "");
        const file = normalize(join(pdfJsRoot, rel));
        if (!file.startsWith(pdfJsRoot) || !existsSync(file) || !statSync(file).isFile()) {
          next();
          return;
        }
        const ext = extname(file).toLowerCase();
        res.setHeader("Content-Type", pdfMimeTypes[ext] ?? "application/octet-stream");
        createReadStream(file).pipe(res);
      });
    },
    closeBundle() {
      copyPdfJsAssets(
        isDemoBuild ? resolve(rootDir, "../site") : resolve(rootDir, "../internal/assets/dist"),
      );
    },
  };
}

const kokoroAssetMime: Record<string, string> = {
  ".wasm": "application/wasm",
  ".mjs": "text/javascript",
  ".js": "application/javascript",
  ".json": "application/json",
  ".onnx": "application/octet-stream",
  ".bin": "application/octet-stream",
};

function assertKokoroModelPresent() {
  const onnx = join(kokoroModelRoot, "onnx", "model_quantized.onnx");
  if (!existsSync(onnx) || !statSync(onnx).isFile()) {
    throw new Error(`Kokoro model missing at ${onnx}. Run: bash scripts/fetch-kokoro-models.sh`);
  }
}

function copyKokoroAssets(destRoot: string) {
  assertKokoroModelPresent();
  const transformersDist = join(packageDir("@huggingface/transformers"), "dist");
  const voicesSrc = join(packageDir("kokoro-js"), "voices");
  const modelDest = join(destRoot, "models", KOKORO_MODEL_ID);
  const ortDest = join(destRoot, "ort");

  mkdirSync(join(modelDest, "onnx"), { recursive: true });
  mkdirSync(ortDest, { recursive: true });
  cpSync(kokoroModelRoot, modelDest, { recursive: true });
  if (existsSync(voicesSrc)) {
    cpSync(voicesSrc, join(modelDest, "voices"), { recursive: true });
  }
  for (const name of ["ort-wasm-simd-threaded.jsep.mjs", "ort-wasm-simd-threaded.jsep.wasm"]) {
    const src = join(transformersDist, name);
    if (!existsSync(src)) {
      throw new Error(`ONNX Runtime asset missing: ${src}`);
    }
    cpSync(src, join(ortDest, name));
  }
}

/** Self-host Kokoro weights + ORT; rewrite HF voice URLs in kokoro-js. */
function kokoroAssetsPlugin(): Plugin {
  return {
    name: "kokoro-assets",
    enforce: "pre",
    transform(code, id) {
      if (omitKokoroAssets) return null;
      if (!id.includes("kokoro-js") && !id.includes("/kokoro/")) return null;
      if (!code.includes(KOKORO_VOICE_HF_PREFIX)) return null;
      return {
        code: code.replaceAll(KOKORO_VOICE_HF_PREFIX, KOKORO_VOICE_LOCAL_PREFIX),
        map: null,
      };
    },
    configureServer(server) {
      if (omitKokoroAssets) return;
      server.middlewares.use((req, res, next) => {
        const url = req.url ?? "";
        let file: string;
        if (url.startsWith("/ort/")) {
          const rel = decodeURIComponent(url.slice("/ort/".length).split("?")[0] ?? "");
          const root = join(packageDir("@huggingface/transformers"), "dist");
          file = normalize(join(root, rel));
          if (!file.startsWith(root)) {
            next();
            return;
          }
        } else if (url.startsWith(`/models/${KOKORO_MODEL_ID}/`)) {
          const rel = decodeURIComponent(
            url.slice(`/models/${KOKORO_MODEL_ID}/`.length).split("?")[0] ?? "",
          );
          if (rel.startsWith("voices/")) {
            const root = join(packageDir("kokoro-js"), "voices");
            file = normalize(join(root, rel.slice("voices/".length)));
            if (!file.startsWith(root)) {
              next();
              return;
            }
          } else {
            file = normalize(join(kokoroModelRoot, rel));
            if (!file.startsWith(kokoroModelRoot)) {
              next();
              return;
            }
          }
        } else {
          next();
          return;
        }
        if (!existsSync(file) || !statSync(file).isFile()) {
          next();
          return;
        }
        const ext = extname(file).toLowerCase();
        res.setHeader("Content-Type", kokoroAssetMime[ext] ?? "application/octet-stream");
        createReadStream(file).pipe(res);
      });
    },
    closeBundle() {
      if (omitKokoroAssets) return;
      copyKokoroAssets(
        isDemoBuild ? resolve(rootDir, "../site") : resolve(rootDir, "../internal/assets/dist"),
      );
    },
  };
}

// The production bundle is written directly into the Go embed package so
// the binary stays self-contained. Demo builds (VITE_DEMO=1) write a static
// SPA under ../site for GitHub Pages and similar hosts.
export default defineConfig({
  base: isDemoBuild ? "./" : "/",
  plugins: [
    svelte(),
    tailwindcss(),
    slimKokoroPlugin(),
    pdfJsAssetsPlugin(),
    kokoroAssetsPlugin(),
    VitePWA({
      registerType: "prompt",
      includeAssets: [
        "favicon.svg",
        "favicon.png",
        "apple-touch-icon.png",
        "pwa-192.png",
        "pwa-512.png",
        "robots.txt",
      ],
      manifest: {
        name: brand.appName,
        short_name: brand.appName,
        description: brand.appDescription,
        theme_color: brand.themeColor.dark,
        background_color: brand.themeColor.dark,
        display: "standalone",
        scope: isDemoBuild ? "./" : "/",
        start_url: isDemoBuild ? "./" : "/",
        icons: [
          {
            src: "favicon.svg",
            sizes: "any",
            type: "image/svg+xml",
            purpose: "any",
          },
          {
            src: "pwa-192.png",
            sizes: "192x192",
            type: "image/png",
            purpose: "any",
          },
          {
            src: "pwa-512.png",
            sizes: "512x512",
            type: "image/png",
            purpose: "maskable",
          },
        ],
      },
      workbox: {
        globPatterns: ["assets/**/*.{js,css,woff,woff2}", "index.html"],
        globIgnores: [
          "**/kokoro-*.js",
          "**/ort-*.js",
          "**/ort-*.wasm",
          "**/*ort-wasm*",
          "**/models/**",
          "**/ort/**",
          // Reader / captcha / TTS payloads are route-loaded, not shell.
          "**/pdf-*.js",
          "**/pdf.worker*",
          "**/epub-*.js",
          "**/altcha*.js",
          "**/SettingsView-*.js",
          "**/BookView-*.js",
          "**/ReaderView-*.js",
          "**/ComicReader-*.js",
          "**/AudioReader-*.js",
          "**/PdfReader-*.js",
          "**/EpubReader-*.js",
          "**/MobiReader-*.js",
          "**/CollectionsView-*.js",
          // Optional UI + reader fonts (latin files still fetched on demand).
          "**/literata-*.woff2",
          "**/crimson-*.woff2",
          "**/newsreader-*.woff2",
          "**/ibm-plex-*.woff2",
          "**/dm-sans-*.woff2",
          "**/source-serif-*.woff2",
          "**/source-sans-3-cyrillic*",
          "**/source-sans-3-greek*",
          "**/source-sans-3-vietnamese*",
          "**/fraunces-vietnamese*",
        ],
        // Kokoro/ONNX chunks exceed the default 2 MiB precache budget and are
        // loaded on demand via dynamic import (not part of the app shell).
        maximumFileSizeToCacheInBytes: 2 * 1024 * 1024,
        navigateFallback: "index.html",
        navigateFallbackDenylist: [/^\/api\//],
        cleanupOutdatedCaches: true,
        runtimeCaching: [
          {
            urlPattern: /^\/api\//,
            handler: "NetworkOnly",
          },
          {
            urlPattern: /^\/pdfjs\//,
            handler: "CacheFirst",
            options: {
              cacheName: "pdfjs-assets",
              expiration: {
                maxEntries: 80,
                maxAgeSeconds: 60 * 60 * 24 * 30,
              },
            },
          },
          {
            urlPattern: /\/assets\/(?:kokoro-|ort-).*/,
            handler: "CacheFirst",
            options: {
              cacheName: "kokoro-runtime",
              expiration: {
                maxEntries: 16,
                maxAgeSeconds: 60 * 60 * 24 * 30,
              },
            },
          },
          {
            urlPattern: /\/(?:models|ort)\//,
            handler: "CacheFirst",
            options: {
              cacheName: "kokoro-weights",
              expiration: {
                maxEntries: 64,
                maxAgeSeconds: 60 * 60 * 24 * 30,
              },
            },
          },
        ],
      },
      devOptions: {
        enabled: false,
      },
    }),
    viteCompression({
      algorithms: ["gzip", "brotliCompress"],
      exclude: [
        /\.(br|gz)$/,
        /sw\.js$/,
        /workbox-.*\.js$/,
        /manifest\.webmanifest$/,
        /\.onnx$/,
        /\.bin$/,
        /\/models\//,
        /\/ort\//,
      ],
    }),
  ],
  resolve: {
    alias: {
      $lib: resolve(rootDir, "src/lib"),
      ...(omitKokoroAssets ? { [kokoroWasmEntry]: kokoroWasmSlim } : {}),
    },
    // Svelte 5 exports map "browser" to the client runtime; without it Vite picks
    // the SSR stub and mount() throws lifecycle_function_unavailable in dev.
    conditions: ["browser", "module", "import", "default"],
  },
  build: {
    target: "es2020",
    cssMinify: true,
    sourcemap: false,
    // Skip gzip size accounting during build (saves Node heap on large assets).
    reportCompressedSize: false,
    chunkSizeWarningLimit: 1500,
    modulePreload: { polyfill: false },
    rollupOptions: {
      output: {
        manualChunks(id) {
          if (id.includes("pdfjs-dist")) return "pdf";
          if (id.includes("epubjs")) return "epub";
          if (
            !omitKokoroAssets &&
            (id.includes("kokoro-js") ||
              id.includes("@huggingface/transformers") ||
              id.includes("onnxruntime"))
          ) {
            return "kokoro";
          }
        },
      },
    },
    outDir: isDemoBuild ? resolve(rootDir, "../site") : resolve(rootDir, "../internal/assets/dist"),
    emptyOutDir: true,
  },
  optimizeDeps: {
    exclude: ["kokoro-js", "@huggingface/transformers"],
  },
  worker: {
    format: "es",
  },
  define: {
    "import.meta.env.VITE_DEMO": JSON.stringify(isDemoBuild ? "true" : ""),
    "import.meta.env.VITE_SLIM": JSON.stringify(isSlimBuild ? "true" : ""),
  },
  server: {
    proxy: isDemoBuild
      ? undefined
      : {
          "/api": "http://localhost:8080",
        },
  },
  test: {
    environment: "jsdom",
    globals: true,
    setupFiles: ["./vitest-setup.ts"],
    include: ["src/**/*.{test,spec}.{ts,svelte.ts}"],
    coverage: {
      provider: "v8",
      reporter: ["text", "html", "lcov"],
      reportsDirectory: "./coverage",
      include: ["src/lib/**/*.ts"],
      exclude: [
        "src/lib/api/generated/**",
        "src/**/*.{test,spec}.{ts,svelte.ts}",
        "src/lib/i18n/locales/**",
        "**/*.svelte",
        "**/*.svelte.ts",
      ],
      thresholds: {
        lines: 35,
        functions: 30,
        branches: 25,
        statements: 35,
      },
    },
  },
});
