import { storageKey } from "$lib/brand/storage";
import { PersistedState } from "runed";
import { createBrowserEngine, isBrowserTTSAvailable } from "$lib/narrator/browser";
import { createServerEngine } from "$lib/narrator/server";
import {
  createKokoroEngine,
  fetchTTSStatus,
  isKokoroWasmAvailable,
  onKokoroWasmLoading,
  preloadKokoroWasm,
  resetKokoroWasm,
} from "$lib/narrator/kokoro";
import type { NarratorEngine, NarratorProvider, NarratorVoice } from "$lib/narrator/types";
import { toast } from "$lib/stores/toast.svelte";

const PROVIDER_KEY = storageKey("narrator-provider");
const VOICE_KEY = storageKey("narrator-voice");
const RATE_KEY = storageKey("narrator-rate");

const SPEEDS = [0.75, 1, 1.25, 1.5, 1.75, 2] as const;

export type NarratorErrorCode =
  | "unavailable"
  | "empty"
  | "speak_failed"
  | "kokoro_unavailable"
  | "server_unavailable"
  | "aborted";

class NarratorStore {
  #providerState = new PersistedState<NarratorProvider>(PROVIDER_KEY, "browser", {
    serializer: {
      // Stored as a bare provider string, not JSON.
      serialize: (value) => value,
      deserialize: (value) => (value === "kokoro" || value === "server" ? value : "browser"),
    },
  });
  #voiceState = new PersistedState<string>(VOICE_KEY, "", {
    serializer: {
      // Stored as a bare voice id string, not JSON.
      serialize: (value) => value,
      deserialize: (value) => value,
    },
  });
  #rateState = new PersistedState<number>(RATE_KEY, 1, {
    serializer: {
      // Stored as a bare number string, not JSON.
      serialize: (value) => String(value),
      deserialize: (value) => Number(value) || 1,
    },
  });

  active = $state(false);
  playing = $state(false);
  paused = $state(false);
  /** True when in-browser Kokoro WASM can run (WebAssembly present). */
  kokoroEnabled = $state(isKokoroWasmAvailable());
  /** True when the server TTS sidecar is configured by an admin. */
  serverEnabled = $state(false);
  kokoroLoading = $state(false);
  voices = $state<NarratorVoice[]>([]);
  voicesLoading = $state(false);
  index = $state(0);
  total = $state(0);
  currentText = $state("");
  error = $state<NarratorErrorCode | null>(null);
  bookTitle = $state("");

  readonly speeds = SPEEDS;

  private queue: string[] = [];
  private engine: NarratorEngine | null = null;
  private abort: AbortController | null = null;
  private generation = 0;
  private kokoroReady: Promise<boolean> | null = null;
  statusLoaded = false;

  constructor() {
    onKokoroWasmLoading((value) => {
      this.kokoroLoading = value;
    });
  }

  get provider(): NarratorProvider {
    return this.#providerState.current;
  }

  set provider(provider: NarratorProvider) {
    this.#providerState.current = provider;
  }

  get voiceId(): string {
    return this.#voiceState.current;
  }

  set voiceId(id: string) {
    this.#voiceState.current = id;
  }

  get rate(): number {
    return this.#rateState.current;
  }

  set rate(rate: number) {
    this.#rateState.current = rate;
  }

  get showBar(): boolean {
    return this.active;
  }

  get progressLabel(): string {
    if (!this.total) return "";
    return `${Math.min(this.index + 1, this.total)} / ${this.total}`;
  }

  /** 0..1 utterance progress for the current queue. */
  get progress(): number {
    if (!this.total) return 0;
    return Math.min(1, Math.max(0, this.index / this.total));
  }

  async refreshStatus(): Promise<void> {
    this.kokoroEnabled = isKokoroWasmAvailable();
    try {
      const status = await fetchTTSStatus();
      this.serverEnabled = status.enabled;
    } catch {
      this.serverEnabled = false;
    }
    if (!this.kokoroEnabled && this.provider === "kokoro") {
      this.provider = this.serverEnabled ? "server" : "browser";
    }
    if (!this.serverEnabled && this.provider === "server") {
      this.provider = "browser";
    }
    this.statusLoaded = true;
  }

  async loadVoices(): Promise<void> {
    this.voicesLoading = true;
    this.error = null;
    try {
      const engine = this.ensureEngine();
      this.voices = await engine.listVoices();
      if (this.voiceId && !this.voices.some((v) => v.id === this.voiceId)) {
        this.voiceId = this.voices[0]?.id ?? "";
      } else if (!this.voiceId && this.voices.length) {
        const preferred =
          this.voices.find((v) => v.lang?.startsWith("en") && v.local) ??
          this.voices.find((v) => v.lang?.startsWith("en")) ??
          this.voices[0];
        this.voiceId = preferred?.id ?? "";
      }
    } catch {
      this.voices = [];
      if (this.provider === "kokoro") this.error = "kokoro_unavailable";
    } finally {
      this.voicesLoading = false;
    }
  }

  setProvider(provider: NarratorProvider) {
    if (provider === "kokoro" && !isKokoroWasmAvailable()) {
      this.error = "kokoro_unavailable";
      return;
    }
    if (provider === "server" && !this.serverEnabled) {
      this.error = "server_unavailable";
      return;
    }
    if (provider === "browser" && !isBrowserTTSAvailable()) {
      this.error = "unavailable";
      return;
    }
    const wasActive = this.active;
    const queue = [...this.queue];
    const index = this.index;
    const title = this.bookTitle;
    this.stopInternal(false);
    this.provider = provider;
    this.engine = null;
    if (provider === "kokoro") {
      void this.ensureKokoroReady();
    }
    void this.loadVoices().then(() => {
      if (wasActive && queue.length) {
        void this.start(queue.slice(index), { title, resume: true });
      }
    });
  }

  setVoice(id: string) {
    this.voiceId = id;
  }

  setRate(rate: number) {
    const next = Number.isFinite(rate) ? Math.min(2, Math.max(0.5, rate)) : 1;
    this.rate = next;
  }

  async start(utterances: string[], opts?: { title?: string; resume?: boolean }): Promise<boolean> {
    const queue = utterances.map((u) => u.trim()).filter(Boolean);
    if (!queue.length) {
      this.error = "empty";
      return false;
    }

    if (this.provider === "browser" && !isBrowserTTSAvailable()) {
      this.error = "unavailable";
      return false;
    }
    if (this.provider === "kokoro" && !isKokoroWasmAvailable()) {
      this.error = "kokoro_unavailable";
      return false;
    }
    if (this.provider === "server" && !this.serverEnabled) {
      this.error = "server_unavailable";
      return false;
    }

    this.stopInternal(false);
    this.queue = queue;
    this.total = queue.length;
    this.index = 0;
    this.bookTitle = opts?.title ?? "";
    this.active = true;
    this.error = null;
    this.engine = this.ensureEngine();
    if (this.provider === "kokoro") {
      const ok = await this.ensureKokoroReady();
      if (!ok) {
        this.error = "kokoro_unavailable";
        this.active = false;
        return false;
      }
    }
    if (!this.voices.length) await this.loadVoices();
    void this.runQueue();
    return true;
  }

  private ensureKokoroReady(): Promise<boolean> {
    if (!isKokoroWasmAvailable()) return Promise.resolve(false);
    if (this.kokoroReady) return this.kokoroReady;

    let toastId: number | null = null;
    const timer = setTimeout(() => {
      toastId = toast.loading("Loading Kokoro model…");
    }, 300);

    this.kokoroReady = preloadKokoroWasm()
      .then(() => {
        clearTimeout(timer);
        if (toastId != null) toast.done(toastId, "Kokoro ready", "success");
        return true;
      })
      .catch(() => {
        clearTimeout(timer);
        resetKokoroWasm();
        this.kokoroReady = null;
        if (toastId != null) toast.done(toastId, "Kokoro failed to load", "error");
        else toast.error("Kokoro failed to load");
        this.error = "kokoro_unavailable";
        return false;
      });
    return this.kokoroReady;
  }

  togglePlay() {
    if (!this.active) return;
    if (this.playing && !this.paused) {
      this.pause();
    } else {
      this.resume();
    }
  }

  pause() {
    if (!this.active || this.paused) return;
    this.engine?.pause();
    this.paused = true;
    this.playing = false;
  }

  resume() {
    if (!this.active) return;
    if (this.paused) {
      this.engine?.resume();
      this.paused = false;
      this.playing = true;
      return;
    }
    if (!this.playing) void this.runQueue();
  }

  stop() {
    this.stopInternal(true);
  }

  skip() {
    if (!this.active) return;
    this.generation += 1;
    this.abort?.abort();
    this.engine?.cancel();
    this.index = Math.min(this.index + 1, this.total);
    if (this.index >= this.total) {
      this.stop();
      return;
    }
    this.paused = false;
    void this.runQueue();
  }

  private ensureEngine(): NarratorEngine {
    if (this.engine && this.engine.id === this.provider) return this.engine;
    this.engine =
      this.provider === "kokoro"
        ? createKokoroEngine()
        : this.provider === "server"
          ? createServerEngine()
          : createBrowserEngine();
    return this.engine;
  }

  private stopInternal(clearError: boolean) {
    this.generation += 1;
    this.abort?.abort();
    this.abort = null;
    this.engine?.cancel();
    this.active = false;
    this.playing = false;
    this.paused = false;
    this.queue = [];
    this.total = 0;
    this.index = 0;
    this.currentText = "";
    this.bookTitle = "";
    if (clearError) this.error = null;
  }

  private async runQueue() {
    const gen = ++this.generation;
    const engine = this.ensureEngine();
    while (this.active && gen === this.generation && this.index < this.queue.length) {
      if (this.paused) return;
      const text = this.queue[this.index];
      this.currentText = text;
      this.abort = new AbortController();
      this.playing = true;
      this.paused = false;

      const result = await engine.speak({
        text,
        voiceId: this.voiceId || undefined,
        rate: this.rate,
        signal: this.abort.signal,
      });

      if (gen !== this.generation || !this.active) return;
      if (result === "cancelled") {
        if (this.paused) return;
        return;
      }
      if (result === "error") {
        this.error =
          this.provider === "kokoro"
            ? "kokoro_unavailable"
            : this.provider === "server"
              ? "server_unavailable"
              : "speak_failed";
        this.playing = false;
        return;
      }
      this.index += 1;
    }
    if (this.active && gen === this.generation && this.index >= this.queue.length) {
      this.stop();
    }
  }
}

export const narrator = new NarratorStore();
