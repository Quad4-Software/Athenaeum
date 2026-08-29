import { test, expect, type APIRequestContext, type Page } from "@playwright/test";

const E2E_USER = "e2eadmin";
const E2E_PASS = "E2e-Admin-Pass1!";

async function ensureAdmin(request: APIRequestContext) {
  const setupRes = await request.get("/api/auth/setup");
  expect(setupRes.ok()).toBeTruthy();
  const setup = await setupRes.json();
  if (!setup.needed) return;

  const csrfRes = await request.get("/api/auth/csrf");
  expect(csrfRes.ok()).toBeTruthy();
  const csrf = (await csrfRes.json()).csrfToken as string;
  expect(csrf).toBeTruthy();

  const create = await request.post("/api/auth/setup", {
    headers: {
      "Content-Type": "application/json",
      "X-CSRF-Token": csrf,
    },
    data: { username: E2E_USER, password: E2E_PASS },
  });
  expect(create.status()).toBe(201);
}

async function gotoLogin(page: Page, request: APIRequestContext) {
  await ensureAdmin(request);
  await page.goto("/login");
  await expect(page.getByRole("heading", { name: /sign in/i }).first()).toBeVisible({
    timeout: 15_000,
  });
}

test.describe("auth flow", () => {
  test.beforeAll(async ({ request }) => {
    await ensureAdmin(request);
  });

  test("setup wizard or login page loads", async ({ page }) => {
    await page.goto("/");
    await expect(page.locator("body")).toBeVisible();
    const loginOrSetup = page.getByRole("heading", { name: /sign in|welcome/i });
    await expect(loginOrSetup.first()).toBeVisible({ timeout: 15_000 });
  });

  test("forbidden error route renders", async ({ page }) => {
    await page.goto("/error/forbidden");
    await expect(page.getByText(/access denied|zugriff verweigert/i)).toBeVisible({
      timeout: 15_000,
    });
  });

  test("offline error route renders", async ({ page }) => {
    await page.goto("/error/offline");
    await expect(page.getByText(/cannot reach|nicht erreichbar/i)).toBeVisible({
      timeout: 15_000,
    });
  });

  test("server error route renders", async ({ page }) => {
    await page.goto("/error/server");
    await expect(page.getByText(/something went wrong|etwas ist schiefgelaufen/i)).toBeVisible({
      timeout: 15_000,
    });
  });

  test("not-found error route renders", async ({ page }) => {
    await page.goto("/error/not-found");
    await expect(page.getByText(/not found|nicht gefunden/i)).toBeVisible({
      timeout: 15_000,
    });
  });
});

test.describe("login UI", () => {
  test("login form exposes username password and primary action", async ({ page, request }) => {
    await gotoLogin(page, request);
    await expect(page.getByLabel(/username|benutzername/i)).toBeVisible();
    await expect(page.getByLabel(/password|passwort/i)).toBeVisible();
    await expect(page.getByRole("button", { name: /sign in|anmelden/i })).toBeVisible();
  });

  test("theme toggle is reachable from login", async ({ page, request }) => {
    await gotoLogin(page, request);
    const toggle = page.getByRole("button", { name: /^theme$/i });
    await expect(toggle).toBeVisible();
    const before = await page.locator("html").getAttribute("data-theme");
    await toggle.click();
    const light = page.getByRole("button", { name: /light|hell/i });
    const dark = page.getByRole("button", { name: /dark|dunkel/i });
    if (await light.isVisible().catch(() => false)) {
      await (before === "light" ? dark : light).first().click();
    }
    await expect(page.locator("html")).toHaveAttribute("data-theme", /light|dark/);
  });

  test("invalid credentials show an error without crashing", async ({ page, request }) => {
    await gotoLogin(page, request);
    await page.getByLabel(/username|benutzername/i).fill("nope");
    await page.getByLabel(/password|passwort/i).fill("wrong-password");
    await page.getByRole("button", { name: /sign in|anmelden/i }).click();
    await expect(page.getByText(/failed|fehlgeschlagen|invalid|ungültig/i).first()).toBeVisible({
      timeout: 10_000,
    });
    await expect(page.getByRole("heading", { name: /sign in/i }).first()).toBeVisible();
  });
});

test.describe("public API contracts", () => {
  test("health endpoint returns ok JSON", async ({ request }) => {
    const res = await request.get("/api/health");
    expect(res.ok()).toBeTruthy();
    const body = await res.json();
    expect(body.status).toMatch(/ok|degraded/);
  });

  test("openapi.json exposes paths", async ({ request }) => {
    const res = await request.get("/api/openapi.json");
    expect(res.ok()).toBeTruthy();
    const body = await res.json();
    expect(body.openapi).toMatch(/^3\./);
    expect(Object.keys(body.paths ?? {}).length).toBeGreaterThan(10);
  });

  test("docs UI loads", async ({ page }) => {
    await page.goto("/docs");
    await expect(page.locator("body")).toBeVisible();
    await expect(page.getByText(/openapi|api|athenaeum/i).first()).toBeVisible({ timeout: 15_000 });
  });
});
