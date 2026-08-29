import type { AuthRedirectReason } from "$lib/api/session";

const AUTH_PATHS = new Set(["/login", "/setup"]);
const AUTH_REDIRECT_REASONS = new Set<AuthRedirectReason>([
  "required",
  "session_expired",
  "logged_out",
]);

export function pathnameOf(url: string): string {
  let path = url;
  const hash = path.indexOf("#");
  if (hash !== -1) path = path.slice(0, hash);
  const q = path.indexOf("?");
  return q === -1 ? path : path.slice(0, q);
}

export function safeReturnPath(path: string | null | undefined): string | null {
  if (!path || !path.startsWith("/") || path.startsWith("//")) return null;
  const base = pathnameOf(path);
  if (AUTH_PATHS.has(base) || base.startsWith("/error/")) return null;
  return base;
}

export function loginUrl(reason: AuthRedirectReason, returnTo?: string | null): string {
  const params = new URLSearchParams({ reason });
  const next = safeReturnPath(returnTo);
  if (next && next !== "/") params.set("next", next);
  return `/login?${params}`;
}

export function normalizeAuthRedirectReason(reason: string | null | undefined): AuthRedirectReason {
  if (reason && AUTH_REDIRECT_REASONS.has(reason as AuthRedirectReason)) {
    return reason as AuthRedirectReason;
  }
  return "required";
}

export function unauthorizedRedirect(
  pathname: string,
  reason: AuthRedirectReason = "required",
): string | null {
  if (pathname === "/login" || pathname === "/setup" || pathname.startsWith("/invite/"))
    return null;
  return loginUrl(reason, safeReturnPath(pathname));
}

export function shouldRedirectToLogin(state: {
  loading: boolean;
  needsLogin: boolean;
  routeName: string;
}): boolean {
  if (state.loading || !state.needsLogin) return false;
  return (
    state.routeName !== "login" &&
    state.routeName !== "setup" &&
    state.routeName !== "error" &&
    state.routeName !== "invite"
  );
}

export function loginGuardTarget(routePath: string): string {
  return loginUrl("required", safeReturnPath(routePath));
}

export function shouldRedirectFromLogin(state: {
  loading: boolean;
  needsLogin: boolean;
  setupNeeded: boolean;
  routeName: string;
}): boolean {
  return !state.loading && !state.needsLogin && !state.setupNeeded && state.routeName === "login";
}

export function sanitizeLoginLocation(pathname: string, search: string): string | null {
  if (pathname !== "/login") return null;
  const params = new URLSearchParams(search);
  const cleaned = loginUrl(normalizeAuthRedirectReason(params.get("reason")), params.get("next"));
  const current = pathname + search;
  return cleaned !== current ? cleaned : null;
}

export function isAuthPagePathname(pathname: string): boolean {
  return (
    pathname === "/login" ||
    pathname === "/setup" ||
    pathname.startsWith("/invite/") ||
    pathname.startsWith("/error/")
  );
}
