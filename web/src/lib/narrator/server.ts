import { ensureCsrf, CSRF_HEADER } from "$lib/api/core";
import { opURL } from "$lib/api/op";
import type { NarratorEngine, NarratorVoice, SpeakOptions, SpeakResult } from "./types";
import { fetchTTSVoices } from "./kokoro";
import { UTTERANCE_HARD_MAX } from "./text";

const MAX_TEXT = UTTERANCE_HARD_MAX;

/**
 * Server-side narration engine. Synthesizes through the Athenaeum proxy
 * (POST /api/tts/synthesize) so weak devices only stream audio.
 */
export function createServerEngine(): NarratorEngine {
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
    id: "server",

    async listVoices(): Promise<NarratorVoice[]> {
      return fetchTTSVoices();
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
            const res = await fetch(opURL("POST__api_tts_synthesize"), {
              method: "POST",
              credentials: "same-origin",
              signal: opts.signal ?? null,
              headers: {
                "Content-Type": "application/json",
                Accept: "audio/*",
                [CSRF_HEADER]: await ensureCsrf(),
              },
              body: JSON.stringify({
                text,
                voice: opts.voiceId ?? "",
                speed: opts.rate ?? 1,
              }),
            });
            if (!res.ok) {
              finish("error");
              return;
            }
            const blob = await res.blob();
            if (opts.signal?.aborted) {
              finish("cancelled");
              return;
            }
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
            if (
              opts.signal?.aborted ||
              (err instanceof DOMException && err.name === "AbortError")
            ) {
              finish("cancelled");
              return;
            }
            console.error("[tts-server] speak failed", err);
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
