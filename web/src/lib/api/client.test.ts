import { afterEach, describe, expect, it, vi } from "vitest";
import { api, ApiError } from "./client";
import { onUnauthorized } from "./session";
import { _resetProxyGateForTests } from "./proxy-gate";

function mockFetch(body: unknown, status = 200, contentType = "application/json") {
  const headers = new Headers({ "content-type": contentType });
  const fn = vi.fn().mockResolvedValue({
    ok: status >= 200 && status < 300,
    status,
    statusText: "",
    headers,
    json: async () => body,
  } as Response);
  vi.stubGlobal("fetch", fn);
  return fn;
}

afterEach(() => {
  vi.unstubAllGlobals();
  _resetProxyGateForTests();
  sessionStorage.clear();
});

describe("api client", () => {
  it("builds a query string from list params", async () => {
    const fetchFn = mockFetch({ items: [], total: 0, limit: 60, offset: 0 });
    await api.listBooks({ search: "dune", sort: "title", limit: 20, offset: 40 });
    const url = fetchFn.mock.calls[0][0] as string;
    expect(url).toContain("/api/books?");
    expect(url).toContain("search=dune");
    expect(url).toContain("sort=title");
    expect(url).toContain("limit=20");
    expect(url).toContain("offset=40");
  });

  it("throws ApiError carrying the server message on failure", async () => {
    mockFetch({ error: "book not found" }, 404);
    await expect(api.getBook(99)).rejects.toMatchObject({
      name: "ApiError",
      status: 404,
      message: "book not found",
    });
  });

  it("exposes file and cover urls", () => {
    expect(api.coverUrl(3)).toBe("/api/books/3/cover");
    expect(api.fileUrl(3)).toBe("/api/books/3/file");
  });

  it("ApiError is an Error", () => {
    expect(new ApiError(500, "boom")).toBeInstanceOf(Error);
  });

  it("sends CSRF header on mutating requests", async () => {
    document.cookie = "athenaeum_csrf=test-csrf-token";
    const fetchFn = mockFetch({ ok: true });
    await api.logout();
    const init = fetchFn.mock.calls[0][1] as RequestInit;
    expect(init.headers).toMatchObject({ "X-CSRF-Token": "test-csrf-token" });
    expect(init.credentials).toBe("same-origin");
  });

  it("retries once after refreshing session on 401", async () => {
    let calls = 0;
    vi.stubGlobal(
      "fetch",
      vi.fn().mockImplementation((path: string, _init?: RequestInit) => {
        calls++;
        const headers = new Headers({ "content-type": "application/json" });
        if (path === "/api/auth/refresh") {
          return Promise.resolve({
            ok: true,
            status: 200,
            statusText: "",
            headers,
            json: async () => ({}),
          } as Response);
        }
        if (calls === 1) {
          return Promise.resolve({
            ok: false,
            status: 401,
            statusText: "Unauthorized",
            headers,
            json: async () => ({ error: "authentication required" }),
          } as Response);
        }
        return Promise.resolve({
          ok: true,
          status: 200,
          statusText: "",
          headers,
          json: async () => ({ items: [], total: 0, limit: 60, offset: 0 }),
        } as Response);
      }),
    );
    document.cookie = "athenaeum_csrf=csrf";
    const page = await api.listBooks();
    expect(page.total).toBe(0);
    expect(calls).toBeGreaterThanOrEqual(3);
  });

  it("notifies session handler when refresh fails on 401", async () => {
    const handler = vi.fn();
    onUnauthorized(handler);
    let calls = 0;
    vi.stubGlobal(
      "fetch",
      vi.fn().mockImplementation((path: string) => {
        calls++;
        const headers = new Headers({ "content-type": "application/json" });
        if (path === "/api/auth/refresh") {
          return Promise.resolve({
            ok: false,
            status: 401,
            statusText: "Unauthorized",
            headers,
            json: async () => ({ error: "authentication required" }),
          } as Response);
        }
        return Promise.resolve({
          ok: false,
          status: 401,
          statusText: "Unauthorized",
          headers,
          json: async () => ({ error: "authentication required" }),
        } as Response);
      }),
    );
    document.cookie = "athenaeum_csrf=csrf";
    await expect(api.listBooks()).rejects.toBeInstanceOf(ApiError);
    expect(handler).toHaveBeenCalledWith("session_expired");
    expect(calls).toBeGreaterThanOrEqual(2);
  });

  it("notifies session handler when retry still returns 401 after refresh", async () => {
    const handler = vi.fn();
    onUnauthorized(handler);
    vi.stubGlobal(
      "fetch",
      vi.fn().mockImplementation((path: string) => {
        const headers = new Headers({ "content-type": "application/json" });
        if (path === "/api/auth/refresh") {
          return Promise.resolve({
            ok: true,
            status: 200,
            statusText: "",
            headers,
            json: async () => ({}),
          } as Response);
        }
        return Promise.resolve({
          ok: false,
          status: 401,
          statusText: "Unauthorized",
          headers,
          json: async () => ({ error: "authentication required" }),
        } as Response);
      }),
    );
    document.cookie = "athenaeum_csrf=csrf";
    await expect(api.listBooks()).rejects.toBeInstanceOf(ApiError);
    expect(handler).toHaveBeenCalledWith("session_expired");
  });

  it("does not notify session handler for silent auth bootstrap 401s", async () => {
    const handler = vi.fn();
    onUnauthorized(handler);
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue({
        ok: false,
        status: 401,
        statusText: "Unauthorized",
        headers: new Headers({ "content-type": "application/json" }),
        json: async () => ({ error: "authentication required" }),
      } as Response),
    );

    await expect(api.me()).rejects.toBeInstanceOf(ApiError);
    expect(handler).not.toHaveBeenCalled();
  });

  it("notifies session handler for protected endpoints with query strings", async () => {
    const handler = vi.fn();
    onUnauthorized(handler);
    vi.stubGlobal(
      "fetch",
      vi.fn().mockImplementation((path: string) => {
        const headers = new Headers({ "content-type": "application/json" });
        if (path === "/api/auth/refresh") {
          return Promise.resolve({
            ok: false,
            status: 401,
            statusText: "Unauthorized",
            headers,
            json: async () => ({ error: "authentication required" }),
          } as Response);
        }
        return Promise.resolve({
          ok: false,
          status: 401,
          statusText: "Unauthorized",
          headers,
          json: async () => ({ error: "authentication required" }),
        } as Response);
      }),
    );
    document.cookie = "athenaeum_csrf=csrf";

    await expect(api.listBooks({ sort: "recent", favorites: true })).rejects.toBeInstanceOf(
      ApiError,
    );
    expect(handler).toHaveBeenCalledWith("session_expired");
  });

  it("reloads on HTML 401 from a reverse proxy instead of soft login redirect", async () => {
    const handler = vi.fn();
    onUnauthorized(handler);
    const reload = vi.fn();
    vi.stubGlobal("location", { reload });
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue({
        ok: false,
        status: 401,
        statusText: "Unauthorized",
        headers: new Headers({ "content-type": "text/html; charset=utf-8" }),
        json: async () => {
          throw new Error("not json");
        },
      } as Response),
    );

    await expect(api.listBooks()).rejects.toMatchObject({
      name: "ApiError",
      status: 401,
      message: "gateway authentication required",
    });
    expect(handler).not.toHaveBeenCalled();
    expect(reload).toHaveBeenCalledTimes(1);
  });
});
