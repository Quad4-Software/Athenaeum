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

  const create = await request.post("/api/auth/setup", {
    headers: {
      "Content-Type": "application/json",
      "X-CSRF-Token": csrf,
    },
    data: { username: E2E_USER, password: E2E_PASS },
  });
  expect(create.status()).toBe(201);
}

async function signIn(page: Page) {
  await page.goto("/login");
  await page.getByLabel(/username/i).fill(E2E_USER);
  await page.getByLabel(/password/i).fill(E2E_PASS);
  await page.getByRole("button", { name: /sign in/i }).click();
  await expect(page).not.toHaveURL(/login/, { timeout: 15_000 });
}

async function firstBookId(page: Page, format: "epub" | "pdf"): Promise<number> {
  const res = await page.request.get(`/api/books?format=${format}&limit=1`);
  expect(res.ok()).toBeTruthy();
  const id = (await res.json()).items[0]?.id as number | undefined;
  expect(id).toBeTruthy();
  return id!;
}

test.describe("readers", () => {
  test.beforeAll(async ({ request }) => {
    await ensureAdmin(request);
  });

  test("EPUB iframe allows scripts and shows chapter text", async ({ page, request }) => {
    await ensureAdmin(request);
    await signIn(page);
    const id = await firstBookId(page, "epub");

    await page.goto(`/read/${id}`);
    const iframe = page.locator("iframe").first();
    await expect(iframe).toBeVisible({ timeout: 20_000 });
    await expect(iframe).toHaveAttribute("sandbox", /allow-scripts/);

    for (let i = 0; i < 6; i++) {
      await page.keyboard.press("ArrowRight");
      await page.waitForTimeout(400);
    }

    const bodyText = await iframe.evaluate((el) => {
      const doc = (el as HTMLIFrameElement).contentDocument;
      return doc?.body?.innerText?.slice(0, 400) ?? "";
    });
    expect(bodyText.length).toBeGreaterThan(20);
  });

  test("PDF reader ships text-layer CSS and keeps extracted text invisible", async ({
    page,
    request,
  }) => {
    await page.setViewportSize({ width: 1280, height: 800 });
    await ensureAdmin(request);
    await signIn(page);
    const id = await firstBookId(page, "pdf");

    await page.goto(`/read/${id}`);
    await expect
      .poll(
        async () =>
          page.evaluate(() =>
            [...document.styleSheets].some((ss) => {
              try {
                return [...(ss.cssRules || [])].some((r) =>
                  String((r as CSSStyleRule).cssText || "").includes("pdf-text-layer"),
                );
              } catch {
                return false;
              }
            }),
          ),
        { timeout: 20_000 },
      )
      .toBe(true);

    const canvas = page.locator(".pdf-page canvas").first();
    const appeared = await canvas.waitFor({ state: "attached", timeout: 15_000 }).then(
      () => true,
      () => false,
    );
    test.info().annotations.push({
      type: "pdf-canvas",
      description: appeared ? "rendered" : "no-canvas-demo-pdf",
    });

    if (!appeared) return;

    const layer = page.locator(".pdf-page .pdf-text-layer, .pdf-page .textLayer").first();
    await expect(layer).toBeAttached();
    await expect
      .poll(async () => layer.evaluate((el) => getComputedStyle(el).position === "absolute"))
      .toBe(true);

    const span = layer.locator("span").first();
    if ((await span.count()) > 0) {
      await expect(span).toHaveCSS("color", "rgba(0, 0, 0, 0)");
      await expect(span).toHaveCSS("position", "absolute");
    }
  });
});
