import { storageKey } from "$lib/brand/storage";
import { createBrowserEngine, isBrowserTTSAvailable } from "$lib/narrator/browser";
import {
  createKokoroEngine,
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
  "unavailable" | "empty" | "speak_failed" | "kokoro_unavailable" | "aborted";

class NarratorStore {
  active = $state(false);
  playing = $state(false);
  paused = $state(false);
  provider = $state<NarratorProvider>(loadProvider());
  /** True when in-browser Kokoro WASM can run (WebAssembly present). */
  kokoroEnabled = $state(isKokoroWasmAvailable());
  kokoroLoading = $state(false);
  voiceId = $state(
    typeof localStorage !== "undefined" ? localStorage.getItem(VOICE_KEY) || "" : "",
  );
  rate = $state(
    typeof localStorage !== "undefined" ? Number(localStorage.getItem(RATE_KEY)) || 1 : 1,
  );
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
    if (!this.kokoroEnabled && this.provider === "kokoro") {
      this.provider = "browser";
      persistProvider("browser");
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
    persistProvider(provider);
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
    if (typeof localStorage !== "undefined") localStorage.setItem(VOICE_KEY, id);
  }

  setRate(rate: number) {
    const next = Number.isFinite(rate) ? Math.min(2, Math.max(0.5, rate)) : 1;
    this.rate = next;
    if (typeof localStorage !== "undefined") localStorage.setItem(RATE_KEY, String(next));
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
    this.engine = this.provider === "kokoro" ? createKokoroEngine() : createBrowserEngine();
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
        this.error = this.provider === "kokoro" ? "kokoro_unavailable" : "speak_failed";
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

function loadProvider(): NarratorProvider {
  if (typeof localStorage === "undefined") return "browser";
  const raw = localStorage.getItem(PROVIDER_KEY);
  return raw === "kokoro" ? "kokoro" : "browser";
}

function persistProvider(provider: NarratorProvider) {
  if (typeof localStorage !== "undefined") localStorage.setItem(PROVIDER_KEY, provider);
}

export const narrator = new NarratorStore();
