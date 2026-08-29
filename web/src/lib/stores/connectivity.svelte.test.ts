import { beforeEach, describe, expect, it, vi } from "vitest";

const navigate = vi.fn();
const routerState = {
  current: { name: "library", params: {} as Record<string, string>, path: "/" },
};

vi.mock("$lib/router.svelte", () => ({
  router: {
    get current() {
      return routerState.current;
    },
    navigate: (...args: unknown[]) => navigate(...args),
  },
}));

describe("ConnectivityStore offline navigation", () => {
  beforeEach(() => {
    navigate.mockClear();
    routerState.current = { name: "library", params: {}, path: "/" };
  });

  it("navigates to /error/offline when going offline on an app route", async () => {
    const { connectivity } = await import("./connectivity.svelte");
    connectivity.onOffline();
    expect(connectivity.online).toBe(false);
    expect(navigate).toHaveBeenCalledWith("/error/offline", true);
  });

  it("does not navigate when already on error, login, setup, or reader", async () => {
    const { connectivity } = await import("./connectivity.svelte");
    for (const name of ["error", "login", "setup", "reader"] as const) {
      navigate.mockClear();
      routerState.current = { name, params: {}, path: `/${name}` };
      connectivity.onOffline();
      expect(navigate).not.toHaveBeenCalled();
    }
  });

  it("returns from /error/offline to / when back online", async () => {
    const { connectivity } = await import("./connectivity.svelte");
    connectivity.markUnreachable();
    routerState.current = { name: "error", params: { code: "offline" }, path: "/error/offline" };
    connectivity.onOnline();
    expect(connectivity.online).toBe(true);
    expect(connectivity.serverReachable).toBe(true);
    expect(navigate).toHaveBeenCalledWith("/", true);
  });

  it("does not auto-navigate on markUnreachable", async () => {
    const { connectivity } = await import("./connectivity.svelte");
    routerState.current = { name: "library", params: {}, path: "/" };
    connectivity.onOnline();
    navigate.mockClear();
    connectivity.markUnreachable();
    expect(connectivity.online).toBe(true);
    expect(connectivity.disconnected).toBe(true);
    expect(navigate).not.toHaveBeenCalled();
  });
});
