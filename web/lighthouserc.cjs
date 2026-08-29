/**
 * Lighthouse CI config for the Athenaeum SPA shell.
 * Audits the production binary (started by scripts/lighthouse-test.sh).
 *
 * Override thresholds with LIGHTHOUSE_MIN_SCORE (0-1, default 0.9).
 * Override base URL with LIGHTHOUSE_URL (default http://127.0.0.1:18080).
 */
const baseUrl = (process.env.LIGHTHOUSE_URL ?? "http://127.0.0.1:18080").replace(/\/$/, "");
const minScore = Number(process.env.LIGHTHOUSE_MIN_SCORE ?? "0.9");

module.exports = {
  ci: {
    collect: {
      numberOfRuns: 3,
      url: [`${baseUrl}/`, `${baseUrl}/login`],
      settings: {
        chromeFlags: "--no-sandbox --disable-dev-shm-usage --headless=new",
        // Production shell is what users install as a PWA.
        onlyCategories: ["performance", "accessibility", "best-practices", "seo"],
      },
    },
    assert: {
      assertions: {
        "categories:performance": ["error", { minScore }],
        "categories:accessibility": ["error", { minScore }],
        "categories:best-practices": ["error", { minScore }],
        "categories:seo": ["error", { minScore }],

        "first-contentful-paint": ["warn", { maxNumericValue: 3000 }],
        "largest-contentful-paint": ["warn", { maxNumericValue: 4000 }],
        "cumulative-layout-shift": ["warn", { maxNumericValue: 0.1 }],
        "total-blocking-time": ["warn", { maxNumericValue: 300 }],
        interactive: ["warn", { maxNumericValue: 4500 }],

        "total-byte-weight": ["warn", { maxNumericValue: 1_600_000 }],
        "unused-javascript": ["warn", { maxNumericValue: 350_000 }],
        // Server serves precompressed .br/.gz; Chromium often misreports
        // Content-Encoding on localhost, so keep this as warn.
        "uses-text-compression": "warn",
        "uses-responsive-images": "warn",
        "offscreen-images": "warn",
        "unminified-css": "error",
        "unminified-javascript": "error",
        "uses-long-cache-ttl": "warn",
        "errors-in-console": "warn",
      },
    },
    upload: {
      target: "filesystem",
      outputDir: ".lighthouseci",
    },
  },
};
