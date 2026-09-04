/** Session key used to avoid reload loops when APIs stay gated after one refresh. */
const PROXY_GATE_RELOAD_KEY = "athenaeum_proxy_gate_reload";

/** Ignore a second automatic reload within this window after the first. */
const PROXY_GATE_RELOAD_COOLDOWN_MS = 15_000;

let recovering = false;

/**
 * True when an authenticating reverse proxy (NetBird password/SSO, etc.)
 * answered instead of Athenaeum. Athenaeum API errors are JSON; gate login
 * pages are HTML.
 */
export function isProxyGateUnauthorized(res: Response): boolean {
  if (res.status !== 401) return false;
  const ct = (res.headers.get("content-type") ?? "").toLowerCase();
  if (!ct) return false;
  if (ct.includes("application/json")) return false;
  return ct.includes("text/html");
}

export function isProxyGateRecovering(): boolean {
  return recovering;
}

/** Clear the cooldown after a healthy API response so a later expiry can reload again. */
export function clearProxyGateReloadGuard(): void {
  if (typeof sessionStorage === "undefined") return;
  try {
    sessionStorage.removeItem(PROXY_GATE_RELOAD_KEY);
  } catch {
    // private mode or blocked storage
  }
}

/**
 * Full document reload so the proxy can serve its login page.
 * Returns true when the response was a proxy gate (caller must stop app auth handling).
 */
export function maybeRecoverProxyGate(res: Response): boolean {
  if (!isProxyGateUnauthorized(res)) return false;
  recoverFromProxyGate();
  return true;
}

export function recoverFromProxyGate(): void {
  if (typeof window === "undefined") return;
  if (recovering) return;
  recovering = true;

  const now = Date.now();
  try {
    const last = Number(sessionStorage.getItem(PROXY_GATE_RELOAD_KEY) ?? "");
    if (Number.isFinite(last) && last > 0 && now - last < PROXY_GATE_RELOAD_COOLDOWN_MS) {
      return;
    }
    sessionStorage.setItem(PROXY_GATE_RELOAD_KEY, String(now));
  } catch {
    // still reload once even if storage is unavailable
  }

  window.location.reload();
}

/** @internal Reset in-memory recovery latch between unit tests. */
export function _resetProxyGateForTests(): void {
  recovering = false;
}
