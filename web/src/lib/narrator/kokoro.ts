import { request, ApiError } from "$lib/api/core";
import { opURL } from "$lib/api/op";
import type { NarratorVoice, TTSJob, TTSSettingsPublic, TTSUserPrefs } from "./types";

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
  model: string;
  defaultVoice: string;
  responseFormat: string;
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

export async function fetchTTSPrefs(): Promise<TTSUserPrefs> {
  return request<TTSUserPrefs>(opURL("GET__api_tts_prefs"));
}

export async function saveTTSPrefs(prefs: TTSUserPrefs): Promise<TTSUserPrefs> {
  return request<TTSUserPrefs>(opURL("PUT__api_tts_prefs"), {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(prefs),
  });
}

export async function fetchTTSJobs(all = false): Promise<TTSJob[]> {
  const body = await request<{ jobs?: TTSJob[] }>(
    opURL("GET__api_tts_jobs") + (all ? "?all=1" : ""),
  );
  return body.jobs ?? [];
}

export async function queueTTSJob(bookId: number, voice?: string, speed?: number): Promise<TTSJob> {
  return request<TTSJob>(opURL("POST__api_tts_jobs"), {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ bookId, voice: voice ?? "", speed: speed ?? 0 }),
  });
}

export async function cancelTTSJob(id: number): Promise<void> {
  await request(opURL("POST__api_tts_jobs__id__cancel", { id }), { method: "POST" });
}

export async function retryTTSJob(id: number): Promise<void> {
  await request(opURL("POST__api_tts_jobs__id__retry", { id }), { method: "POST" });
}

export async function deleteTTSJob(id: number): Promise<void> {
  await request(opURL("DELETE__api_tts_jobs__id", { id }), { method: "DELETE" });
}
