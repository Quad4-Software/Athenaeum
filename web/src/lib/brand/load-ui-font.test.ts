import { beforeEach, describe, expect, it, vi } from "vitest";

vi.mock("@fontsource-variable/literata/wght.css", () => ({}));

describe("loadUiFontCss", () => {
  beforeEach(() => {
    vi.resetModules();
  });

  it("is a no-op for default and system fonts", async () => {
    const { loadUiFontCss } = await import("./load-ui-font");
    await expect(loadUiFontCss("athenaeum")).resolves.toBeUndefined();
    await expect(loadUiFontCss("system")).resolves.toBeUndefined();
  });

  it("loads optional font CSS once", async () => {
    const { loadUiFontCss } = await import("./load-ui-font");
    await loadUiFontCss("literata");
    await loadUiFontCss("literata");
  });
});
