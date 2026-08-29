import { describe, expect, it } from "vitest";
import { isKokoroWasmAvailable, KNOWN_VOICES } from "./kokoro-wasm";

describe("kokoro-wasm", () => {
  it("reports WebAssembly availability", () => {
    expect(typeof isKokoroWasmAvailable()).toBe("boolean");
  });

  it("lists known voices with ids", () => {
    expect(KNOWN_VOICES.length).toBeGreaterThan(0);
    expect(KNOWN_VOICES.some((v) => v.id === "af_heart")).toBe(true);
  });
});
