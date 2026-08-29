/**
 * @type {import('@stryker-mutator/api/core').PartialStrykerOptions}
 */
export default {
  packageManager: "pnpm",
  testRunner: "vitest",
  checkers: [],
  reporters: ["clear-text", "progress", "html"],
  htmlReporter: { fileName: "reports/mutation/mutation.html" },
  coverageAnalysis: "perTest",
  mutate: [
    "src/lib/utils/sanitize-html.ts",
    "src/lib/utils/password-strength.ts",
    "src/lib/utils/format.ts",
  ],
  vitest: {
    configFile: "vite.config.ts",
  },
  thresholds: {
    high: 80,
    low: 60,
    break: 50,
  },
  timeoutMS: 60000,
  concurrency: 4,
};
