export type AuthRedirectReason = "required" | "session_expired" | "logged_out";

let unauthorizedHandler: ((reason: AuthRedirectReason) => void) | null = null;
let forbiddenHandler: (() => void) | null = null;

export function onUnauthorized(handler: (reason: AuthRedirectReason) => void) {
  unauthorizedHandler = handler;
}

export function onForbidden(handler: () => void) {
  forbiddenHandler = handler;
}

export function notifyUnauthorized(reason: AuthRedirectReason = "required") {
  unauthorizedHandler?.(reason);
}

export function notifyForbidden() {
  forbiddenHandler?.();
}

export const AUTH_SILENT_401 = new Set([
  "/api/auth/me",
  "/api/auth/setup",
  "/api/auth/login",
  "/api/auth/refresh",
  "/api/auth/logout",
]);

export const AUTH_SILENT_403 = new Set<string>([]);
