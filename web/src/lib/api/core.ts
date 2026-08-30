import type { BookQueryParams } from "./types";
import { connectivity } from "$lib/stores/connectivity.svelte";
import { captureApiError } from "$lib/telemetry/sentry";
import { brand } from "$lib/brand";
import { AUTH_SILENT_401, AUTH_SILENT_403, notifyUnauthorized, notifyForbidden } from "./session";
import { isDemoMode } from "$lib/demo/mode";

/**
 * ApiError carries the HTTP status alongside a human-readable message so
 * callers can branch on, for example, 404 responses.
 */
export class ApiError extends Error {
  constructor(
    public status: number,
    message: string,
  ) {
    super(message);
    this.name = "ApiError";
  }
}

const CSRF_COOKIE = brand.csrfCookie;
export const CSRF_HEADER = "X-CSRF-Token";

const MUTATING = new Set(["POST", "PUT", "PATCH", "DELETE"]);

function readCSRFCookie(): string {
  if (typeof document === "undefined") return "";
  const match = document.cookie.match(new RegExp(`(?:^|; )${CSRF_COOKIE}=([^;]*)`));
  return match ? decodeURIComponent(match[1]) : "";
}

let csrfReady: Promise<string> | null = null;

export async function ensureCsrf(): Promise<string> {
  if (isDemoMode()) return "demo-csrf";

  // Cookie is the source of truth. Never keep a parallel in-memory token
  // (failed refresh used to clear the cookie and leave login broken).
  const cookie = readCSRFCookie();
  if (cookie) return cookie;
  if (!csrfReady) {
    csrfReady = fetch("/api/auth/csrf", { credentials: "same-origin" })
      .then(async (res) => {
        if (!res.ok) throw new ApiError(res.status, "failed to fetch csrf token");
        let bodyToken = "";
        try {
          const body = (await res.json()) as { csrfToken?: string };
          bodyToken = body.csrfToken?.trim() ?? "";
        } catch {
          // Cookie is the source of truth when the body is empty.
        }
        const cookieToken = readCSRFCookie();
        const token = cookieToken || bodyToken;
        if (!token) throw new ApiError(500, "csrf token missing from response");
        return token;
      })
      .finally(() => {
        csrfReady = null;
      });
  }
  return csrfReady;
}

/** Kept for call sites after login/logout. Cookie is the CSRF source of truth. */
export function clearCsrfCache(): void {
  csrfReady = null;
}

let refreshPromise: Promise<boolean> | null = null;

function apiPath(path: string): string {
  const q = path.indexOf("?");
  return q === -1 ? path : path.slice(0, q);
}

async function tryRefresh(): Promise<boolean> {
  if (refreshPromise) return refreshPromise;
  refreshPromise = (async () => {
    try {
      const csrf = await ensureCsrf();
      const res = await fetch("/api/auth/refresh", {
        method: "POST",
        credentials: "same-origin",
        headers: { Accept: "application/json", [CSRF_HEADER]: csrf },
      });
      if (!res.ok) {
        clearCsrfCache();
      }
      return res.ok;
    } catch {
      clearCsrfCache();
      return false;
    } finally {
      refreshPromise = null;
    }
  })();
  return refreshPromise;
}

/** Attempt to rotate cookies from a refresh token before protected API calls. */
export async function restoreSession(): Promise<boolean> {
  return tryRefresh();
}

const RETRYABLE_STATUS = new Set([502, 503, 504]);
const MAX_ATTEMPTS = 3;

function isIdempotent(method: string): boolean {
  return method === "GET" || method === "HEAD";
}

function shouldRetry(method: string, status: number | null, attempt: number): boolean {
  if (attempt >= MAX_ATTEMPTS - 1) return false;
  if (!isIdempotent(method)) return false;
  if (status === null) return true;
  return RETRYABLE_STATUS.has(status);
}

function retryDelayMs(attempt: number): number {
  return Math.min(1000 * 2 ** attempt, 4000);
}

function sleep(ms: number): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

/** @internal Exported for unit tests. */
export function _shouldRetry(method: string, status: number | null, attempt: number): boolean {
  return shouldRetry(method, status, attempt);
}

export async function request<T>(
  path: string,
  init?: RequestInit,
  authRetried = false,
): Promise<T> {
  return requestWithRetry(path, init, authRetried, 0);
}

async function requestWithRetry<T>(
  path: string,
  init: RequestInit | undefined,
  authRetried: boolean,
  attempt: number,
): Promise<T> {
  if (isDemoMode()) {
    const { handleDemoRequest } = await import("$lib/demo/handler");
    const demoRes = await handleDemoRequest(path, init);
    if (demoRes) {
      if (!demoRes.ok) {
        let message = demoRes.statusText;
        try {
          const body = (await demoRes.clone().json()) as { error?: string };
          if (body.error) message = body.error;
        } catch {
          // ignore
        }
        throw new ApiError(demoRes.status, message);
      }
      connectivity.markReachable();
      if (demoRes.status === 204) return undefined as T;
      return (await demoRes.json()) as T;
    }
  }

  const method = init?.method ?? "GET";
  const headers: Record<string, string> = {
    Accept: "application/json",
    ...(init?.headers as Record<string, string> | undefined),
  };
  if (MUTATING.has(method)) {
    headers[CSRF_HEADER] = await ensureCsrf();
  }

  let res: Response;
  try {
    res = await fetch(path, {
      credentials: "same-origin",
      ...init,
      headers,
    });
  } catch (err) {
    connectivity.markUnreachable();
    if (shouldRetry(method, null, attempt)) {
      await sleep(retryDelayMs(attempt));
      return requestWithRetry(path, init, authRetried, attempt + 1);
    }
    const message = err instanceof Error ? err.message : "network error";
    captureApiError(apiPath(path), 0, message);
    throw new ApiError(0, message);
  }

  if (res.status === 401) {
    const base = apiPath(path);
    if (!authRetried && base !== "/api/auth/refresh" && base !== "/api/auth/login") {
      const refreshed = await tryRefresh();
      if (refreshed) return requestWithRetry(path, init, true, attempt);
    }
    if (!AUTH_SILENT_401.has(base)) {
      notifyUnauthorized("session_expired");
    }
  }

  if (res.status === 403 && !AUTH_SILENT_403.has(apiPath(path))) {
    notifyForbidden();
  }

  if (!res.ok) {
    if (shouldRetry(method, res.status, attempt)) {
      await sleep(retryDelayMs(attempt));
      return requestWithRetry(path, init, authRetried, attempt + 1);
    }
    let message = res.statusText;
    try {
      const body = (await res.json()) as { error?: string };
      if (body.error) message = body.error;
    } catch {
      // ignore non-JSON error bodies
    }
    if (res.status >= 500 || res.status === 0) {
      connectivity.markUnreachable();
    }
    captureApiError(apiPath(path), res.status, message);
    throw new ApiError(res.status, message);
  }
  connectivity.markReachable();
  if (res.status === 204) return undefined as T;
  return (await res.json()) as T;
}

export function buildQuery(params: BookQueryParams): string {
  const q = new URLSearchParams();
  if (params.search) q.set("search", params.search);
  if (params.sort) q.set("sort", params.sort);
  if (params.format) q.set("format", params.format);
  if (params.series) q.set("series", params.series);
  if (params.author) q.set("author", params.author);
  if (params.library != null) q.set("library", String(params.library));
  if (params.collection != null) q.set("collection", String(params.collection));
  if (params.favorites) q.set("favorites", "1");
  if (params.inProgress) q.set("inProgress", "1");
  if (params.limit != null) q.set("limit", String(params.limit));
  if (params.offset != null) q.set("offset", String(params.offset));
  const s = q.toString();
  return s ? `?${s}` : "";
}
