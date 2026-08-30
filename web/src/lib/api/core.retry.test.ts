import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { _shouldRetry, request, ApiError, ensureCsrf, clearCsrfCache } from "./core";
import { brand } from "$lib/brand";

function clearDocumentCookies() {
  for (const part of document.cookie.split(";")) {
    const name = part.split("=")[0]?.trim();
    if (name) document.cookie = `${name}=; Max-Age=0; path=/`;
  }
}

describe("_shouldRetry", () => {
  it("retries network failures for GET", () => {
    expect(_shouldRetry("GET", null, 0)).toBe(true);
    expect(_shouldRetry("GET", null, 2)).toBe(false);
  });

  it("retries 502/503/504 for GET only", () => {
    expect(_shouldRetry("GET", 503, 0)).toBe(true);
    expect(_shouldRetry("GET", 500, 0)).toBe(false);
    expect(_shouldRetry("POST", 503, 0)).toBe(false);
  });
});

describe("request retry", () => {
  afterEach(() => {
    vi.unstubAllGlobals();
    vi.useRealTimers();
  });

  it("retries GET on 503 then succeeds", async () => {
    vi.useFakeTimers();
    let calls = 0;
    vi.stubGlobal(
      "fetch",
      vi.fn().mockImplementation(() => {
        calls++;
        if (calls < 3) {
          return Promise.resolve({
            ok: false,
            status: 503,
            statusText: "Service Unavailable",
            json: async () => ({ error: "unavailable" }),
          } as Response);
        }
        return Promise.resolve({
          ok: true,
          status: 200,
          json: async () => ({ ok: true }),
        } as Response);
      }),
    );

    const pending = request<{ ok: boolean }>("/api/health");
    await vi.runAllTimersAsync();
    await expect(pending).resolves.toEqual({ ok: true });
    expect(calls).toBe(3);
  });

  it("retries network errors for GET then fails", async () => {
    vi.useFakeTimers();
    vi.stubGlobal("fetch", vi.fn().mockRejectedValue(new TypeError("Failed to fetch")));

    const pending = request("/api/health");
    const expectation = expect(pending).rejects.toMatchObject({ status: 0, name: "ApiError" });
    await vi.runAllTimersAsync();
    await expectation;
    expect(fetch).toHaveBeenCalledTimes(3);
  });

  it("does not retry POST on 503", async () => {
    document.cookie = `${brand.csrfCookie}=test-csrf`;
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue({
        ok: false,
        status: 503,
        statusText: "Service Unavailable",
        json: async () => ({ error: "unavailable" }),
      } as Response),
    );
    await expect(request("/api/scan", { method: "POST" })).rejects.toBeInstanceOf(ApiError);
    expect(fetch).toHaveBeenCalledTimes(1);
  });
});

describe("ensureCsrf", () => {
  beforeEach(() => {
    clearDocumentCookies();
    clearCsrfCache();
  });

  afterEach(() => {
    vi.unstubAllGlobals();
    clearDocumentCookies();
    clearCsrfCache();
  });

  it("refetches when the cookie is gone even if memory still has a token", async () => {
    document.cookie = `${brand.csrfCookie}=stale-token`;
    await ensureCsrf();
    clearDocumentCookies();

    const fetchFn = vi.fn().mockImplementation(() => {
      document.cookie = `${brand.csrfCookie}=fresh-token`;
      return Promise.resolve({
        ok: true,
        status: 200,
        json: async () => ({ csrfToken: "fresh-token" }),
      } as Response);
    });
    vi.stubGlobal("fetch", fetchFn);

    await expect(ensureCsrf()).resolves.toBe("fresh-token");
    expect(fetchFn).toHaveBeenCalledWith("/api/auth/csrf", { credentials: "same-origin" });
  });
});
