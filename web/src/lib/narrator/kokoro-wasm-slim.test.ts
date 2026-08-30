import { describe, expect, it } from "vitest";
import {
  createKokoroWasmEngine,
  isKokoroWasmAvailable,
  isKokoroWasmLoading,
  KNOWN_VOICES,
  preloadKokoroWasm,
} from "./kokoro-wasm-slim";

describe("kokoro-wasm-slim", () => {
  it("reports Kokoro WASM as unavailable", () => {
    expect(isKokoroWasmAvailable()).toBe(false);
    expect(isKokoroWasmLoading()).toBe(false);
  });

  it("keeps the known voice list for UI parity", () => {
    expect(KNOWN_VOICES.length).toBeGreaterThan(0);
    expect(KNOWN_VOICES.some((v) => v.id === "af_heart")).toBe(true);
  });

  it("preload resolves without loading a model", async () => {
    await expect(preloadKokoroWasm()).resolves.toBeUndefined();
  });

  it("engine speak fails closed", async () => {
    const engine = createKokoroWasmEngine();
    expect(engine.id).toBe("kokoro");
    await expect(engine.speak({ text: "hello" })).resolves.toBe("error");
  });
});
