import type { NarratorEngine, NarratorVoice, SpeakOptions, SpeakResult } from "./types";

import { KOKORO_LOCAL_MODEL_PATH, KOKORO_MODEL_ID, KOKORO_ORT_PATH } from "./kokoro-paths";
import { UTTERANCE_HARD_MAX } from "./text";

const DEFAULT_VOICE = "af_heart";
const MAX_TEXT = UTTERANCE_HARD_MAX;

/** Voices aligned with docker/kokoro/app.py KNOWN_VOICES. */
export const KNOWN_VOICES: NarratorVoice[] = [
  { id: "af_heart", label: "Heart (US female)", lang: "en-us", local: true },
  { id: "af_bella", label: "Bella (US female)", lang: "en-us", local: true },
  { id: "af_sarah", label: "Sarah (US female)", lang: "en-us", local: true },
  { id: "am_adam", label: "Adam (US male)", lang: "en-us", local: true },
  { id: "am_michael", label: "Michael (US male)", lang: "en-us", local: true },
  { id: "bf_emma", label: "Emma (UK female)", lang: "en-gb", local: true },
  { id: "bm_george", label: "George (UK male)", lang: "en-gb", local: true },
];

type KokoroTTS = {
  generate(
    text: string,
    opts?: { voice?: string; speed?: number },
  ): Promise<{ toBlob(): Blob; toWav(): ArrayBuffer }>;
  list_voices(): void;
  readonly voices: Record<string, { name?: string; language?: string }>;
};

type KokoroModule = {
  KokoroTTS: {
    from_pretrained(
      modelId: string,
      opts: {
        dtype?: "fp32" | "fp16" | "q8" | "q4" | "q4f16";
        device?: "wasm" | "webgpu" | "cpu" | null;
      },
    ): Promise<KokoroTTS>;
  };
  env: { wasmPaths: string | Record<string, string> };
};

type TransformersEnv = {
  allowLocalModels: boolean;
  allowRemoteModels: boolean;
  localModelPath: string;
};

let pipelinePromise: Promise<KokoroTTS> | null = null;
let loading = false;
const loadingListeners = new Set<(value: boolean) => void>();

function setLoading(value: boolean) {
  if (loading === value) return;
  loading = value;
  for (const listener of loadingListeners) listener(value);
}

/** True when the browser exposes WebAssembly (required for ONNX Runtime). */
export function isKokoroWasmAvailable(): boolean {
  return typeof WebAssembly !== "undefined";
}

export function isKokoroWasmLoading(): boolean {
  return loading;
}

/** Subscribe to model init loading state. Returns unsubscribe. */
export function onKokoroWasmLoading(listener: (value: boolean) => void): () => void {
  loadingListeners.add(listener);
  listener(loading);
  return () => {
    loadingListeners.delete(listener);
  };
}

async function configureLocalRuntime(kokoroEnv: KokoroModule["env"]): Promise<void> {
  const { env: transformersEnv } = (await import("@huggingface/transformers")) as {
    env: TransformersEnv;
  };
  transformersEnv.allowLocalModels = true;
  transformersEnv.allowRemoteModels = false;
  transformersEnv.localModelPath = KOKORO_LOCAL_MODEL_PATH;
  kokoroEnv.wasmPaths = KOKORO_ORT_PATH;
}

async function createPipeline(): Promise<KokoroTTS> {
  const mod = (await import("kokoro-js")) as KokoroModule;
  const { KokoroTTS, env } = mod;
  await configureLocalRuntime(env);

  const canWebGPU =
    typeof navigator !== "undefined" &&
    "gpu" in navigator &&
    typeof (navigator as Navigator & { gpu?: unknown }).gpu !== "undefined";

  // Full builds embed q8 (model_quantized.onnx) only. No Hub fetch.
  if (canWebGPU) {
    try {
      return await KokoroTTS.from_pretrained(KOKORO_MODEL_ID, {
        dtype: "q8",
        device: "webgpu",
      });
    } catch {
      // Fall through to WASM (no SharedArrayBuffer / COOP-COEP required).
    }
  }

  return await KokoroTTS.from_pretrained(KOKORO_MODEL_ID, {
    dtype: "q8",
    device: "wasm",
  });
}

function ensurePipeline(): Promise<KokoroTTS> {
  if (!isKokoroWasmAvailable()) {
    return Promise.reject(new Error("WebAssembly is not available"));
  }
  if (!pipelinePromise) {
    setLoading(true);
    const pending = createPipeline()
      .catch((err) => {
        pipelinePromise = null;
        throw err;
      })
      .finally(() => {
        setLoading(false);
      });
    pipelinePromise = pending;
  }
  return pipelinePromise;
}

/** Warm the model (first call loads embedded weights). Safe to call repeatedly. */
export function preloadKokoroWasm(): Promise<void> {
  return ensurePipeline().then(() => undefined);
}

function clampSpeed(rate: number): number {
  if (!Number.isFinite(rate)) return 1;
  return Math.min(2, Math.max(0.5, rate));
}

function resolveVoice(voiceId: string | undefined): string {
  const id = (voiceId || DEFAULT_VOICE).trim() || DEFAULT_VOICE;
  if (KNOWN_VOICES.some((v) => v.id === id)) return id;
  return DEFAULT_VOICE;
}

export function createKokoroWasmEngine(): NarratorEngine {
  let audio: HTMLAudioElement | null = null;
  let objectUrl: string | null = null;

  const release = () => {
    if (audio) {
      audio.onended = null;
      audio.onerror = null;
      audio.pause();
      audio.removeAttribute("src");
      audio.load();
      audio = null;
    }
    if (objectUrl) {
      URL.revokeObjectURL(objectUrl);
      objectUrl = null;
    }
  };

  return {
    id: "kokoro",

    async listVoices(): Promise<NarratorVoice[]> {
      return KNOWN_VOICES.slice();
    },

    speak(opts: SpeakOptions): Promise<SpeakResult> {
      const text = opts.text.trim();
      if (!text) return Promise.resolve("ended");
      if (text.length > MAX_TEXT) return Promise.resolve("error");

      return new Promise((resolve) => {
        let finished = false;
        const finish = (result: SpeakResult) => {
          if (finished) return;
          finished = true;
          opts.signal?.removeEventListener("abort", onAbort);
          release();
          resolve(result);
        };

        const onAbort = () => finish("cancelled");
        if (opts.signal?.aborted) {
          finish("cancelled");
          return;
        }
        opts.signal?.addEventListener("abort", onAbort, { once: true });

        void (async () => {
          try {
            const tts = await ensurePipeline();
            if (opts.signal?.aborted) {
              finish("cancelled");
              return;
            }
            const raw = await tts.generate(text, {
              voice: resolveVoice(opts.voiceId),
              speed: clampSpeed(opts.rate ?? 1),
            });
            if (opts.signal?.aborted) {
              finish("cancelled");
              return;
            }
            const blob = raw.toBlob();
            if (!blob.size) {
              finish("error");
              return;
            }
            objectUrl = URL.createObjectURL(blob);
            audio = new Audio(objectUrl);
            audio.onended = () => finish("ended");
            audio.onerror = () => finish("error");
            await audio.play();
          } catch (err) {
            if (opts.signal?.aborted) {
              finish("cancelled");
              return;
            }
            if (err instanceof DOMException && err.name === "AbortError") {
              finish("cancelled");
              return;
            }
            finish("error");
          }
        })();
      });
    },

    pause() {
      audio?.pause();
    },

    resume() {
      void audio?.play().catch(() => undefined);
    },

    cancel() {
      release();
    },
  };
}
