import { describe, expect, it } from "vitest";
import {
  KOKORO_LOCAL_MODEL_PATH,
  KOKORO_MODEL_ID,
  KOKORO_ORT_MJS,
  KOKORO_ORT_PATH,
  KOKORO_ORT_WASM,
  KOKORO_VOICE_HF_PREFIX,
  KOKORO_VOICE_LOCAL_PREFIX,
} from "./kokoro-paths";

describe("kokoro-paths", () => {
  it("keeps local paths under same-origin prefixes", () => {
    expect(KOKORO_LOCAL_MODEL_PATH).toBe("/models/");
    expect(KOKORO_ORT_PATH).toBe("/ort/");
    expect(KOKORO_ORT_MJS).toBe("/ort/ort-wasm-simd-threaded.jsep.mjs");
    expect(KOKORO_ORT_WASM).toBe("/ort/ort-wasm-simd-threaded.jsep.wasm");
    expect(KOKORO_VOICE_LOCAL_PREFIX).toBe(`/models/${KOKORO_MODEL_ID}/voices/`);
  });

  it("names the HF voice URL prefix that Vite rewrites", () => {
    expect(KOKORO_VOICE_HF_PREFIX).toContain("huggingface.co");
    expect(KOKORO_VOICE_HF_PREFIX).toContain("/voices/");
  });
});
