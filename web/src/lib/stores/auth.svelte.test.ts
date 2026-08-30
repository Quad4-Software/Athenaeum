import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

const navigate = vi.fn();
const logout = vi.fn().mockResolvedValue(undefined);

vi.mock("$lib/router.svelte", () => ({
  router: {
    navigate,
    appPathname: (browserPath?: string) =>
      browserPath ??
      (typeof window !== "undefined" ? window.location.pathname : "/"),
  },
}));

vi.mock("$lib/api/client", () => ({
  api: { logout },
  clearCsrfCache: vi.fn(),
  ensureCsrf: vi.fn(),
  ApiError: class ApiError extends Error {
    constructor(
      public status: number,
      message: string,
    ) {
      super(message);
      this.name = "ApiError";
    }
  },
}));

vi.mock("$lib/stores/toast.svelte", () => ({
  toast: {
    success: vi.fn(),
    error: vi.fn(),
    info: vi.fn(),
  },
}));

function mockLocation(pathname: string, search = "") {
  vi.stubGlobal("window", {
    location: { pathname, search },
  });
}

describe("AuthStore redirects", () => {
  beforeEach(() => {
    navigate.mockClear();
    vi.resetModules();
  });

  afterEach(() => {
    vi.unstubAllGlobals();
    vi.resetModules();
  });

  it("redirects expired sessions from protected pages", async () => {
    mockLocation("/collections");
    const { auth } = await import("./auth.svelte");

    auth.handleUnauthorized("session_expired");

    expect(auth.user).toBeNull();
    expect(navigate).toHaveBeenCalledWith(
      "/login?reason=session_expired&next=%2Fcollections",
      true,
    );
  });

  it("does not redirect when already on login", async () => {
    mockLocation("/login", "?reason=required");
    const { auth } = await import("./auth.svelte");

    auth.handleUnauthorized("session_expired");

    expect(auth.user).toBeNull();
    expect(navigate).not.toHaveBeenCalled();
  });

  it("does not redirect when already on setup", async () => {
    mockLocation("/setup");
    const { auth } = await import("./auth.svelte");

    auth.handleUnauthorized("required");

    expect(navigate).not.toHaveBeenCalled();
  });

  it("sends forbidden users to the error page", async () => {
    mockLocation("/settings/library");
    const { auth } = await import("./auth.svelte");

    auth.handleForbidden();

    expect(navigate).toHaveBeenCalledWith("/error/forbidden", true);
  });

  it("does not send forbidden users away from login", async () => {
    mockLocation("/login");
    const { auth } = await import("./auth.svelte");

    auth.handleForbidden();

    expect(navigate).not.toHaveBeenCalled();
  });

  it("logs out to login with logged_out reason", async () => {
    mockLocation("/");
    const { auth } = await import("./auth.svelte");

    await auth.logout();

    expect(logout).toHaveBeenCalled();
    expect(navigate).toHaveBeenCalledWith("/login?reason=logged_out", true);
  });
});

describe("AuthStore.needsLogin", () => {
  beforeEach(() => {
    vi.resetModules();
  });

  afterEach(() => {
    vi.resetModules();
  });

  it("is true only when auth is enabled and no user is loaded", async () => {
    const { auth } = await import("./auth.svelte");

    auth.authEnabled = true;
    auth.setupNeeded = false;
    auth.user = null;
    expect(auth.needsLogin).toBe(true);

    auth.user = {
      id: 1,
      username: "alice",
      isAdmin: false,
      permissions: [],
      createdAt: "2026-01-01T00:00:00Z",
    };
    expect(auth.needsLogin).toBe(false);

    auth.user = null;
    auth.setupNeeded = true;
    expect(auth.needsLogin).toBe(false);

    auth.setupNeeded = false;
    auth.authEnabled = false;
    expect(auth.needsLogin).toBe(false);
  });
});
