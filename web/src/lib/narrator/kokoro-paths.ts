/** Local folder name under /models/ (full build embeds these assets). */
export const KOKORO_MODEL_ID = "Kokoro-82M-v1.0-ONNX";

/** transformers.js localModelPath (trailing slash required). */
export const KOKORO_LOCAL_MODEL_PATH = "/models/";

/** Directory for ONNX Runtime Web wasm/mjs (self-hosted, not jsDelivr). */
export const KOKORO_ORT_PATH = "/ort/";

/** Exact ORT worker entry served from KOKORO_ORT_PATH (must match vite copyKokoroAssets). */
export const KOKORO_ORT_MJS = `${KOKORO_ORT_PATH}ort-wasm-simd-threaded.jsep.mjs`;

/** Exact ORT wasm binary served from KOKORO_ORT_PATH. */
export const KOKORO_ORT_WASM = `${KOKORO_ORT_PATH}ort-wasm-simd-threaded.jsep.wasm`;

/** HF voice URL prefix rewritten to this at build time (see vite kokoro plugin). */
export const KOKORO_VOICE_HF_PREFIX =
  "https://huggingface.co/onnx-community/Kokoro-82M-v1.0-ONNX/resolve/main/voices/";

export const KOKORO_VOICE_LOCAL_PREFIX = `/models/${KOKORO_MODEL_ID}/voices/`;
