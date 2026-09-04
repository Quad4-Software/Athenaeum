import type { NarratorEngine, NarratorVoice, SpeakOptions, SpeakResult } from "./types";

/**
 * Slim-build stand-in for kokoro-wasm.ts.
 * Vite aliases the real module here when VITE_SLIM=1 so kokoro-js and its
 * ONNX Runtime WASM assets never enter the production bundle.
 */

/** Voices aligned with docker/kokoro/app.py KNOWN_VOICES (UI labels only). */
export const KNOWN_VOICES: NarratorVoice[] = [
  { id: "af_heart", label: "Heart (US female)", lang: "en-us", local: true },
  { id: "af_bella", label: "Bella (US female)", lang: "en-us", local: true },
  { id: "af_sarah", label: "Sarah (US female)", lang: "en-us", local: true },
  { id: "am_adam", label: "Adam (US male)", lang: "en-us", local: true },
  { id: "am_michael", label: "Michael (US male)", lang: "en-us", local: true },
  { id: "bf_emma", label: "Emma (UK female)", lang: "en-gb", local: true },
  { id: "bm_george", label: "George (UK male)", lang: "en-gb", local: true },
];

/** Always false in slim binaries (no embedded Kokoro WASM). */
export function isKokoroWasmAvailable(): boolean {
  return false;
}

export function isKokoroWasmLoading(): boolean {
  return false;
}

/** No-op subscription. Returns unsubscribe. */
export function onKokoroWasmLoading(_listener: (value: boolean) => void): () => void {
  return () => undefined;
}

/** Resolves immediately. Slim builds never download a model. */
export function preloadKokoroWasm(): Promise<void> {
  return Promise.resolve();
}

/** No-op. Slim builds have nothing to reset. */
export function resetKokoroWasm(): void {}

/** Engine that always fails speak. Present so imports stay typed. */
export function createKokoroWasmEngine(): NarratorEngine {
  return {
    id: "kokoro",

    async listVoices(): Promise<NarratorVoice[]> {
      return KNOWN_VOICES.slice();
    },

    speak(_opts: SpeakOptions): Promise<SpeakResult> {
      return Promise.resolve("error");
    },

    pause() {},

    resume() {},

    cancel() {},
  };
}
