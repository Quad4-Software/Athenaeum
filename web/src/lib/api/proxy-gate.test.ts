import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import {
  _resetProxyGateForTests,
  clearProxyGateReloadGuard,
  isProxyGateRecovering,
  isProxyGateUnauthorized,
  maybeRecoverProxyGate,
  recoverFromProxyGate,
} from "./proxy-gate";

function response(status: number, contentType: string | null): Response {
  const headers = new Headers();
  if (contentType != null) headers.set("content-type", contentType);
  return {
    status,
    ok: status >= 200 && status < 300,
    headers,
  } as Response;
}

describe("isProxyGateUnauthorized", () => {
  it("detects HTML 401 from a reverse-proxy login page", () => {
    expect(isProxyGateUnauthorized(response(401, "text/html; charset=utf-8"))).toBe(true);
  });

  it("ignores Athenaeum JSON 401", () => {
    expect(isProxyGateUnauthorized(response(401, "application/json; charset=utf-8"))).toBe(false);
  });

  it("ignores 401 with no content-type", () => {
    expect(isProxyGateUnauthorized(response(401, null))).toBe(false);
  });

  it("ignores non-401 HTML", () => {
    expect(isProxyGateUnauthorized(response(403, "text/html"))).toBe(false);
    expect(isProxyGateUnauthorized(response(500, "text/html"))).toBe(false);
  });
});

describe("recoverFromProxyGate", () => {
  beforeEach(() => {
    _resetProxyGateForTests();
    clearProxyGateReloadGuard();
    sessionStorage.clear();
  });

  afterEach(() => {
    _resetProxyGateForTests();
    clearProxyGateReloadGuard();
    vi.unstubAllGlobals();
    sessionStorage.clear();
  });

  it("reloads once and marks recovering", () => {
    const reload = vi.fn();
    vi.stubGlobal("location", { reload });

    recoverFromProxyGate();
    expect(reload).toHaveBeenCalledTimes(1);
    expect(isProxyGateRecovering()).toBe(true);

    recoverFromProxyGate();
    expect(reload).toHaveBeenCalledTimes(1);
  });

  it("skips a second reload inside the cooldown window", () => {
    const reload = vi.fn();
    vi.stubGlobal("location", { reload });

    recoverFromProxyGate();
    expect(reload).toHaveBeenCalledTimes(1);

    _resetProxyGateForTests();
    recoverFromProxyGate();
    expect(reload).toHaveBeenCalledTimes(1);
    expect(isProxyGateRecovering()).toBe(true);
  });

  it("maybeRecoverProxyGate only acts on HTML 401", () => {
    const reload = vi.fn();
    vi.stubGlobal("location", { reload });

    expect(maybeRecoverProxyGate(response(401, "application/json"))).toBe(false);
    expect(reload).not.toHaveBeenCalled();

    expect(maybeRecoverProxyGate(response(401, "text/html"))).toBe(true);
    expect(reload).toHaveBeenCalledTimes(1);
  });

  it("clearProxyGateReloadGuard allows a later reload", () => {
    const reload = vi.fn();
    vi.stubGlobal("location", { reload });

    recoverFromProxyGate();
    expect(reload).toHaveBeenCalledTimes(1);

    clearProxyGateReloadGuard();
    _resetProxyGateForTests();
    recoverFromProxyGate();
    expect(reload).toHaveBeenCalledTimes(2);
  });
});
