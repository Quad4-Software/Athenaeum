/** TTS provider identifiers. */
export type NarratorProvider = "browser" | "kokoro";

export interface NarratorVoice {
  id: string;
  label: string;
  lang?: string;
  local?: boolean;
}

export interface NarratorStatus {
  enabled: boolean;
  provider: string;
  baseUrl: string;
  defaultVoice: string;
}

export interface TTSSettingsPublic {
  enabled: boolean;
  baseUrl: string;
  defaultVoice: string;
  apiKeySet: boolean;
  timeoutSec: number;
}

export type SpeakResult = "ended" | "cancelled" | "error";

export interface SpeakOptions {
  text: string;
  voiceId?: string;
  rate?: number;
  signal?: AbortSignal;
}

export interface NarratorEngine {
  readonly id: NarratorProvider;
  listVoices(): Promise<NarratorVoice[]>;
  speak(opts: SpeakOptions): Promise<SpeakResult>;
  pause(): void;
  resume(): void;
  cancel(): void;
}
