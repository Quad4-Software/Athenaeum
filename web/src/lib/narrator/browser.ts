import type { NarratorEngine, NarratorVoice, SpeakOptions, SpeakResult } from "./types";

const VOICES_WAIT_MS = 1500;

function synthesis(): SpeechSynthesis | null {
  if (typeof window === "undefined") return null;
  return window.speechSynthesis ?? null;
}

/** True when the browser exposes SpeechSynthesis. */
export function isBrowserTTSAvailable(): boolean {
  const synth = synthesis();
  return !!synth && typeof SpeechSynthesisUtterance !== "undefined";
}

function waitForVoices(synth: SpeechSynthesis): Promise<SpeechSynthesisVoice[]> {
  const existing = synth.getVoices();
  if (existing.length > 0) return Promise.resolve(existing);

  return new Promise((resolve) => {
    let settled = false;
    const done = (voices: SpeechSynthesisVoice[]) => {
      if (settled) return;
      settled = true;
      synth.removeEventListener("voiceschanged", onChange);
      clearTimeout(timer);
      resolve(voices);
    };
    const onChange = () => done(synth.getVoices());
    synth.addEventListener("voiceschanged", onChange);
    const timer = setTimeout(() => done(synth.getVoices()), VOICES_WAIT_MS);
  });
}

function voiceLabel(v: SpeechSynthesisVoice): string {
  const local = v.localService ? "local" : "remote";
  return `${v.name} (${v.lang}, ${local})`;
}

export function createBrowserEngine(): NarratorEngine {
  let current: SpeechSynthesisUtterance | null = null;

  return {
    id: "browser",

    async listVoices(): Promise<NarratorVoice[]> {
      const synth = synthesis();
      if (!synth) return [];
      const voices = await waitForVoices(synth);
      return voices
        .map((v) => ({
          id: v.voiceURI || v.name,
          label: voiceLabel(v),
          lang: v.lang,
          local: v.localService,
        }))
        .sort((a, b) => a.label.localeCompare(b.label));
    },

    speak(opts: SpeakOptions): Promise<SpeakResult> {
      const synth = synthesis();
      if (!synth) return Promise.resolve("error");
      const text = opts.text.trim();
      if (!text) return Promise.resolve("ended");

      return new Promise((resolve) => {
        let finished = false;
        const finish = (result: SpeakResult) => {
          if (finished) return;
          finished = true;
          opts.signal?.removeEventListener("abort", onAbort);
          if (current === utter) current = null;
          resolve(result);
        };

        const utter = new SpeechSynthesisUtterance(text);
        current = utter;
        utter.rate = clampRate(opts.rate ?? 1);

        if (opts.voiceId) {
          const voices = synth.getVoices();
          const match = voices.find((v) => v.voiceURI === opts.voiceId || v.name === opts.voiceId);
          if (match) utter.voice = match;
        }

        const onAbort = () => {
          try {
            synth.cancel();
          } catch {
            // ignore
          }
          finish("cancelled");
        };

        if (opts.signal?.aborted) {
          finish("cancelled");
          return;
        }
        opts.signal?.addEventListener("abort", onAbort, { once: true });

        utter.onend = () => finish("ended");
        utter.onerror = (ev) => {
          if (ev.error === "canceled" || ev.error === "interrupted") {
            finish("cancelled");
            return;
          }
          finish("error");
        };

        try {
          // Chrome can stall if a previous cancel left synthesis paused.
          if (synth.paused) synth.resume();
          synth.speak(utter);
        } catch {
          finish("error");
        }
      });
    },

    pause() {
      const synth = synthesis();
      if (synth && !synth.paused) synth.pause();
    },

    resume() {
      const synth = synthesis();
      if (synth?.paused) synth.resume();
    },

    cancel() {
      const synth = synthesis();
      current = null;
      try {
        synth?.cancel();
      } catch {
        // ignore
      }
    },
  };
}

function clampRate(rate: number): number {
  if (!Number.isFinite(rate)) return 1;
  return Math.min(2, Math.max(0.5, rate));
}
