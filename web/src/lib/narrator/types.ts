/** TTS provider identifiers. */
export type NarratorProvider = "browser" | "kokoro" | "server";

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
  model: string;
  defaultVoice: string;
  responseFormat: string;
  apiKeySet: boolean;
  timeoutSec: number;
}

export interface TTSUserPrefs {
  voice: string;
  speed: number;
  schedEnabled: boolean;
  schedStart: string;
  schedEnd: string;
}

export interface TTSJob {
  id: number;
  userId: number;
  bookId: number;
  bookTitle?: string;
  voice: string;
  speed: number;
  status: "queued" | "running" | "done" | "failed" | "cancelled";
  totalChapters: number;
  doneChapters: number;
  outputDir?: string;
  error?: string;
  runAt?: number;
  createdAt: string;
  startedAt?: string;
  finishedAt?: string;
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
