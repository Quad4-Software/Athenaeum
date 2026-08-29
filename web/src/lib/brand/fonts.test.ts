import { describe, expect, it } from "vitest";
import { DEFAULT_UI_FONT, getUiFont, isUiFontId, UI_FONT_PRESETS } from "$lib/brand/fonts";

describe("ui fonts", () => {
  it("defaults to athenaeum", () => {
    expect(DEFAULT_UI_FONT).toBe("athenaeum");
    expect(getUiFont(null).id).toBe("athenaeum");
  });

  it("recognizes known ids", () => {
    expect(isUiFontId("literata")).toBe(true);
    expect(isUiFontId("nope")).toBe(false);
  });

  it("exposes preview families for every preset", () => {
    for (const preset of UI_FONT_PRESETS) {
      expect(preset.family.length).toBeGreaterThan(0);
      expect(preset.sample.length).toBeGreaterThan(0);
    }
  });
});
