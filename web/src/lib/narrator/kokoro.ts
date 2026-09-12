import { request, ApiError } from "$lib/api/core";
import { opURL } from "$lib/api/op";
import type { NarratorVoice, TTSSettingsPublic } from "./types";

export { createKokoroWasmEngine as createKokoroEngine } from "./kokoro-wasm";
export {
  isKokoroWasmAvailable,
  isKokoroWasmLoading,
  onKokoroWasmLoading,
  preloadKokoroWasm,
  resetKokoroWasm,
  KNOWN_VOICES,
} from "./kokoro-wasm";

export async function fetchTTSStatus(): Promise<{
  enabled: boolean;
  defaultVoice: string;
}> {
  try {
    return await request<{ enabled: boolean; defaultVoice: string }>(opURL("GET__api_tts_status"));
  } catch (e) {
    // TTS status is probed opportunistically; auth failures mean "disabled".
    if (e instanceof ApiError && (e.status === 401 || e.status === 403)) {
      return { enabled: false, defaultVoice: "" };
    }
    throw e;
  }
}

export async function fetchTTSVoices(): Promise<NarratorVoice[]> {
  const body = await request<{
    voices?: { id: string; label?: string; lang?: string }[];
  }>(opURL("GET__api_tts_voices"));
  return (body.voices ?? []).map((v) => ({
    id: v.id,
    label: v.label || v.id,
    lang: v.lang,
    local: true,
  }));
}

export async function getTTSAdmin(): Promise<TTSSettingsPublic> {
  return request<TTSSettingsPublic>(opURL("GET__api_admin_tts"));
}

export async function saveTTSAdmin(config: {
  enabled: boolean;
  baseUrl: string;
  defaultVoice: string;
  apiKey?: string;
  timeoutSec: number;
}): Promise<TTSSettingsPublic> {
  return request<TTSSettingsPublic>(opURL("PUT__api_admin_tts"), {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(config),
  });
}

export async function testTTSAdmin(): Promise<{ ok: boolean; message: string }> {
  return request<{ ok: boolean; message: string }>(opURL("POST__api_admin_tts_test"), {
    method: "POST",
  });
}
