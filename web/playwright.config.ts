import { defineConfig } from "@playwright/test";

export default defineConfig({
  testDir: "./e2e",
  timeout: 60_000,
  use: {
    baseURL: process.env.ATHENAEUM_E2E_URL ?? "http://127.0.0.1:8080",
    trace: "on-first-retry",
  },
  webServer: process.env.ATHENAEUM_E2E_URL
    ? undefined
    : {
        command:
          "../bin/athenaeum --addr 127.0.0.1:8080 --data /tmp/athenaeum-e2e-data --library /tmp/athenaeum-e2e-lib --demo",
        port: 8080,
        reuseExistingServer: !process.env.CI,
      },
});
