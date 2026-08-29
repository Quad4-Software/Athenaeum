import { describe, expect, it } from "vitest";
import {
  isAuthPagePathname,
  loginGuardTarget,
  loginUrl,
  normalizeAuthRedirectReason,
  pathnameOf,
  safeReturnPath,
  sanitizeLoginLocation,
  shouldRedirectFromLogin,
  shouldRedirectToLogin,
  unauthorizedRedirect,
} from "./auth-redirect";

describe("auth-redirect paths", () => {
  it("strips query strings and hashes from pathnames", () => {
    expect(pathnameOf("/login?reason=required&next=%2F")).toBe("/login");
    expect(pathnameOf("/book/3?tab=info#section")).toBe("/book/3");
  });

  it("rejects unsafe and auth return targets", () => {
    expect(safeReturnPath(null)).toBeNull();
    expect(safeReturnPath("")).toBeNull();
    expect(safeReturnPath("//evil.example")).toBeNull();
    expect(safeReturnPath("https://evil.example")).toBeNull();
    expect(safeReturnPath("/login")).toBeNull();
    expect(safeReturnPath("/login?reason=required&next=%2F")).toBeNull();
    expect(safeReturnPath("/setup")).toBeNull();
    expect(safeReturnPath("/error/forbidden")).toBeNull();
    expect(safeReturnPath("/book/2")).toBe("/book/2");
    expect(safeReturnPath("/settings/library")).toBe("/settings/library");
  });

  it("builds login urls without nesting login in next", () => {
    expect(loginUrl("required", "/book/4")).toBe("/login?reason=required&next=%2Fbook%2F4");
    expect(loginUrl("session_expired", "/login?reason=required")).toBe(
      "/login?reason=session_expired",
    );
    expect(loginUrl("logged_out", "/")).toBe("/login?reason=logged_out");
    expect(loginUrl("required", "/")).toBe("/login?reason=required");
  });

  it("normalizes unknown redirect reasons to required", () => {
    expect(normalizeAuthRedirectReason("session_expired")).toBe("session_expired");
    expect(normalizeAuthRedirectReason("logged_out")).toBe("logged_out");
    expect(normalizeAuthRedirectReason("bogus")).toBe("required");
    expect(normalizeAuthRedirectReason(null)).toBe("required");
  });
});

describe("unauthorizedRedirect", () => {
  it("redirects protected pages to login with a safe return path", () => {
    expect(unauthorizedRedirect("/book/3", "session_expired")).toBe(
      "/login?reason=session_expired&next=%2Fbook%2F3",
    );
    expect(unauthorizedRedirect("/", "required")).toBe("/login?reason=required");
  });

  it("does not redirect when already on auth pages", () => {
    expect(unauthorizedRedirect("/login", "session_expired")).toBeNull();
    expect(unauthorizedRedirect("/setup", "required")).toBeNull();
  });
});

describe("login guard", () => {
  it("does not redirect to login when already on the login route", () => {
    expect(
      shouldRedirectToLogin({
        loading: false,
        needsLogin: true,
        routeName: "login",
      }),
    ).toBe(false);
  });

  it("does not redirect while auth is still loading", () => {
    expect(
      shouldRedirectToLogin({
        loading: true,
        needsLogin: true,
        routeName: "library",
      }),
    ).toBe(false);
  });

  it("redirects unauthenticated users away from protected routes", () => {
    expect(
      shouldRedirectToLogin({
        loading: false,
        needsLogin: true,
        routeName: "library",
      }),
    ).toBe(true);
    expect(
      shouldRedirectToLogin({
        loading: false,
        needsLogin: true,
        routeName: "settings",
      }),
    ).toBe(true);
  });

  it("skips setup and error routes", () => {
    expect(
      shouldRedirectToLogin({
        loading: false,
        needsLogin: true,
        routeName: "setup",
      }),
    ).toBe(false);
    expect(
      shouldRedirectToLogin({
        loading: false,
        needsLogin: true,
        routeName: "error",
      }),
    ).toBe(false);
  });

  it("documents the misrouted login loop case", () => {
    expect(
      shouldRedirectToLogin({
        loading: false,
        needsLogin: true,
        routeName: "notfound",
      }),
    ).toBe(true);
  });

  it("builds guard targets without nesting login paths", () => {
    expect(loginGuardTarget("/collections")).toBe("/login?reason=required&next=%2Fcollections");
    expect(loginGuardTarget("/login")).toBe("/login?reason=required");
  });
});

describe("post-login redirect", () => {
  it("sends authenticated users away from login", () => {
    expect(
      shouldRedirectFromLogin({
        loading: false,
        needsLogin: false,
        setupNeeded: false,
        routeName: "login",
      }),
    ).toBe(true);
  });

  it("keeps unauthenticated users on login", () => {
    expect(
      shouldRedirectFromLogin({
        loading: false,
        needsLogin: true,
        setupNeeded: false,
        routeName: "login",
      }),
    ).toBe(false);
  });
});

describe("sanitizeLoginLocation", () => {
  it("cleans nested login return paths", () => {
    const nested = "?reason=required&next=%2Flogin%3Freason%3Drequired%26next%3D%252F";
    expect(sanitizeLoginLocation("/login", nested)).toBe("/login?reason=required");
  });

  it("normalizes invalid reasons", () => {
    expect(sanitizeLoginLocation("/login", "?reason=weird&next=%2Fbook%2F1")).toBe(
      "/login?reason=required&next=%2Fbook%2F1",
    );
  });

  it("returns null when the login url is already clean", () => {
    expect(sanitizeLoginLocation("/login", "?reason=session_expired")).toBeNull();
    expect(sanitizeLoginLocation("/login", "?reason=required&next=%2Fbook%2F2")).toBeNull();
    expect(sanitizeLoginLocation("/book/1", "?reason=required")).toBeNull();
  });
});

describe("isAuthPagePathname", () => {
  it("recognizes auth and error pages", () => {
    expect(isAuthPagePathname("/login")).toBe(true);
    expect(isAuthPagePathname("/setup")).toBe(true);
    expect(isAuthPagePathname("/error/forbidden")).toBe(true);
    expect(isAuthPagePathname("/")).toBe(false);
    expect(isAuthPagePathname("/book/1")).toBe(false);
  });
});
