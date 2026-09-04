import { ApiError, ensureCsrf, CSRF_HEADER } from "$lib/api/core";
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
  const res = await fetch("/api/tts/status", {
    credentials: "same-origin",
    headers: { Accept: "application/json" },
  });
  if (!res.ok) {
    if (res.status === 401 || res.status === 403) {
      return { enabled: false, defaultVoice: "" };
    }
    throw new ApiError(res.status, await readError(res));
  }
  return (await res.json()) as { enabled: boolean; defaultVoice: string };
}

export async function fetchTTSVoices(): Promise<NarratorVoice[]> {
  const res = await fetch("/api/tts/voices", {
    credentials: "same-origin",
    headers: { Accept: "application/json" },
  });
  if (!res.ok) {
    throw new ApiError(res.status, await readError(res));
  }
  const body = (await res.json()) as { voices?: { id: string; label?: string; lang?: string }[] };
  return (body.voices ?? []).map((v) => ({
    id: v.id,
    label: v.label || v.id,
    lang: v.lang,
    local: true,
  }));
}

export async function getTTSAdmin(): Promise<TTSSettingsPublic> {
  const res = await fetch("/api/admin/tts", {
    credentials: "same-origin",
    headers: { Accept: "application/json" },
  });
  if (!res.ok) throw new ApiError(res.status, await readError(res));
  return (await res.json()) as TTSSettingsPublic;
}

export async function saveTTSAdmin(config: {
  enabled: boolean;
  baseUrl: string;
  defaultVoice: string;
  apiKey?: string;
  timeoutSec: number;
}): Promise<TTSSettingsPublic> {
  const csrf = await ensureCsrf();
  const res = await fetch("/api/admin/tts", {
    method: "PUT",
    credentials: "same-origin",
    headers: {
      Accept: "application/json",
      "Content-Type": "application/json",
      [CSRF_HEADER]: csrf,
    },
    body: JSON.stringify(config),
  });
  if (!res.ok) throw new ApiError(res.status, await readError(res));
  return (await res.json()) as TTSSettingsPublic;
}

export async function testTTSAdmin(): Promise<{ ok: boolean; message: string }> {
  const csrf = await ensureCsrf();
  const res = await fetch("/api/admin/tts/test", {
    method: "POST",
    credentials: "same-origin",
    headers: {
      Accept: "application/json",
      [CSRF_HEADER]: csrf,
    },
  });
  if (!res.ok) throw new ApiError(res.status, await readError(res));
  return (await res.json()) as { ok: boolean; message: string };
}

async function readError(res: Response): Promise<string> {
  try {
    const body = (await res.json()) as { error?: string };
    if (body.error) return body.error;
  } catch {
    // ignore
  }
  return res.statusText || `HTTP ${res.status}`;
}
