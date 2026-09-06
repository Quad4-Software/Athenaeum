import { expect, type APIRequestContext } from "@playwright/test";

export const E2E_USER = "e2eadmin";
export const E2E_PASS = "E2e-Admin-Pass1!";

/** Create the first admin when setup is needed. Tolerates parallel workers. */
export async function ensureAdmin(request: APIRequestContext) {
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
  // 201 created; 409 when another parallel worker already finished setup
  expect([201, 409]).toContain(create.status());
}
