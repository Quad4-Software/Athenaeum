import { afterEach, describe, expect, it, vi } from "vitest";
import {
  AUTH_SILENT_401,
  notifyForbidden,
  notifyUnauthorized,
  onForbidden,
  onUnauthorized,
} from "./session";

afterEach(() => {
  onUnauthorized(() => {});
  onForbidden(() => {});
});

describe("session handlers", () => {
  it("notifies the registered unauthorized handler", () => {
    const handler = vi.fn();
    onUnauthorized(handler);

    notifyUnauthorized("session_expired");

    expect(handler).toHaveBeenCalledWith("session_expired");
    expect(handler).toHaveBeenCalledTimes(1);
  });

  it("notifies the registered forbidden handler", () => {
    const handler = vi.fn();
    onForbidden(handler);

    notifyForbidden();

    expect(handler).toHaveBeenCalledTimes(1);
  });

  it("does nothing when no handler is registered", () => {
    onUnauthorized(() => {});
    expect(() => notifyUnauthorized("required")).not.toThrow();
  });
});

describe("AUTH_SILENT_401", () => {
  it("includes auth bootstrap endpoints", () => {
    expect(AUTH_SILENT_401.has("/api/auth/me")).toBe(true);
    expect(AUTH_SILENT_401.has("/api/auth/setup")).toBe(true);
    expect(AUTH_SILENT_401.has("/api/auth/login")).toBe(true);
    expect(AUTH_SILENT_401.has("/api/auth/refresh")).toBe(true);
    expect(AUTH_SILENT_401.has("/api/auth/logout")).toBe(true);
  });

  it("does not silence protected library endpoints", () => {
    expect(AUTH_SILENT_401.has("/api/books")).toBe(false);
    expect(AUTH_SILENT_401.has("/api/favorites")).toBe(false);
    expect(AUTH_SILENT_401.has("/api/library/stats")).toBe(false);
  });
});
