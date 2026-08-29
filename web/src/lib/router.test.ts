import { afterEach, describe, expect, it, vi } from "vitest";
import { loginGuardTarget, sanitizeLoginLocation, shouldRedirectToLogin } from "./auth-redirect";

function mockWindow(initialPath = "/", initialSearch = "") {
  let pathname = initialPath;
  let search = initialSearch;

  const applyUrl = (url: string) => {
    const q = url.indexOf("?");
    if (q === -1) {
      pathname = url;
      search = "";
      return;
    }
    pathname = url.slice(0, q);
    search = url.slice(q);
  };

  vi.stubGlobal("window", {
    location: {
      get pathname() {
        return pathname;
      },
      get search() {
        return search;
      },
      get href() {
        return `http://localhost${pathname}${search}`;
      },
    },
    addEventListener: vi.fn(),
    history: {
      pushState: vi.fn((_state: unknown, _title: string, url: string) => applyUrl(url)),
      replaceState: vi.fn((_state: unknown, _title: string, url: string) => applyUrl(url)),
    },
  });

  return {
    get pathname() {
      return pathname;
    },
    set pathname(value: string) {
      pathname = value;
    },
    get search() {
      return search;
    },
    set search(value: string) {
      search = value;
    },
  };
}

describe("router", () => {
  afterEach(() => {
    vi.unstubAllGlobals();
    vi.resetModules();
  });

  it("recognizes /login with query params as the login route", async () => {
    mockWindow("/", "");
    const { router } = await import("./router.svelte");
    router.navigate("/login?reason=required&next=%2F", true);

    expect(router.current.name).toBe("login");
    expect(router.current.path).toBe("/login");
  });

  it("recognizes other routes when query params are present", async () => {
    mockWindow("/", "");
    const { router } = await import("./router.svelte");
    router.navigate("/settings/library?tab=server", true);

    expect(router.current.name).toBe("settings");
    expect(router.current.path).toBe("/settings/library");
    expect(router.current.params.tab).toBe("library");
  });

  it("does not treat /login query urls as notfound", async () => {
    const loc = mockWindow("/", "");
    const { router } = await import("./router.svelte");

    router.navigate("/login?reason=required&next=%2Fbook%2F2", true);

    expect(router.current.name).not.toBe("notfound");
    expect(
      shouldRedirectToLogin({
        loading: false,
        needsLogin: true,
        routeName: router.current.name,
      }),
    ).toBe(false);
    expect(loc.pathname).toBe("/login");
    expect(loc.search).toBe("?reason=required&next=%2Fbook%2F2");
  });

  it("avoids compounding login redirects when the guard runs repeatedly", async () => {
    const loc = mockWindow("/", "");
    const { router } = await import("./router.svelte");

    for (let i = 0; i < 5; i++) {
      if (
        shouldRedirectToLogin({
          loading: false,
          needsLogin: true,
          routeName: router.current.name,
        })
      ) {
        router.navigate(loginGuardTarget(router.current.path), true);
      }
      const cleaned = sanitizeLoginLocation(loc.pathname, loc.search);
      if (cleaned) router.navigate(cleaned, true);
    }

    expect(router.current.name).toBe("login");
    expect(loc.search).toBe("?reason=required");
    expect(loc.search).not.toContain("next=");
  });

  it("syncs route state when navigating to the same url", async () => {
    mockWindow("/login", "?reason=required");
    const { router } = await import("./router.svelte");

    router.navigate("/login?reason=required", true);

    expect(router.current.name).toBe("login");
  });

  it("updates current route on browser popstate", async () => {
    const loc = mockWindow("/", "");
    const { router } = await import("./router.svelte");
    const popstate = (window.addEventListener as ReturnType<typeof vi.fn>).mock.calls.find(
      ([event]) => event === "popstate",
    )?.[1] as () => void;

    router.navigate("/book/9", true);
    loc.pathname = "/";
    loc.search = "";
    popstate();

    expect(router.current.name).toBe("library");
  });
});
