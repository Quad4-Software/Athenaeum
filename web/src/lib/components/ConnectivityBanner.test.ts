import { beforeEach, describe, expect, it, vi } from "vitest";
import { render, screen } from "@testing-library/svelte";
import ConnectivityBanner from "./ConnectivityBanner.svelte";

const { connectivityState } = vi.hoisted(() => {
  const state = {
    online: true,
    serverReachable: true,
    get disconnected() {
      return !this.online || !this.serverReachable;
    },
  };
  return { connectivityState: state };
});

vi.mock("$lib/stores/connectivity.svelte", () => ({
  connectivity: connectivityState,
}));

vi.mock("$lib/brand", () => ({
  brand: { appName: "Athenaeum" },
}));

vi.mock("$lib/stores/i18n.svelte", () => ({
  i18n: {
    t: (key: string, params?: Record<string, string | number>) => {
      const messages: Record<string, string> = {
        "connectivity.offline":
          "You are offline. Some features may not work until your connection returns.",
        "connectivity.unreachable": `Cannot reach the server. Check that ${params?.app} is running and try again.`,
      };
      return messages[key] ?? key;
    },
  },
}));

describe("ConnectivityBanner", () => {
  beforeEach(() => {
    connectivityState.online = true;
    connectivityState.serverReachable = true;
  });

  it("shows offline message when browser is offline", () => {
    connectivityState.online = false;
    render(ConnectivityBanner);
    expect(
      screen.getByText(
        "You are offline. Some features may not work until your connection returns.",
      ),
    ).toBeInTheDocument();
  });

  it("shows unreachable message with brand name when server is down", () => {
    connectivityState.serverReachable = false;
    render(ConnectivityBanner);
    expect(
      screen.getByText("Cannot reach the server. Check that Athenaeum is running and try again."),
    ).toBeInTheDocument();
  });

  it("hides when connected", () => {
    render(ConnectivityBanner);
    expect(screen.queryByRole("status")).toBeNull();
  });
});
