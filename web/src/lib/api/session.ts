import { apiOp } from "./generated/paths";

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

export const AUTH_SILENT_401 = new Set<string>([
  apiOp("GET__api_auth_me").path,
  apiOp("GET__api_auth_setup").path,
  apiOp("POST__api_auth_login").path,
  apiOp("POST__api_auth_refresh").path,
  apiOp("POST__api_auth_logout").path,
]);

export const AUTH_SILENT_403 = new Set<string>([]);
