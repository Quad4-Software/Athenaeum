import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { createBrowserEngine, isBrowserTTSAvailable } from "./browser";

class FakeUtterance {
  text: string;
  rate = 1;
  voice: SpeechSynthesisVoice | null = null;
  onend: ((ev: Event) => void) | null = null;
  onerror: ((ev: SpeechSynthesisErrorEvent) => void) | null = null;

  constructor(text: string) {
    this.text = text;
  }
}

function installFakeSpeech(opts?: { failSpeak?: boolean; voices?: SpeechSynthesisVoice[] }) {
  const voices = opts?.voices ?? [
    {
      name: "Test Voice",
      lang: "en-US",
      localService: true,
      default: true,
      voiceURI: "test-voice",
    } as SpeechSynthesisVoice,
  ];

  const synth = {
    paused: false,
    speaking: false,
    pending: false,
    getVoices: () => voices,
    speak: vi.fn((utter: FakeUtterance) => {
      if (opts?.failSpeak) {
        queueMicrotask(() =>
          utter.onerror?.({ error: "synthesis-failed" } as SpeechSynthesisErrorEvent),
        );
        return;
      }
      queueMicrotask(() => utter.onend?.(new Event("end")));
    }),
    cancel: vi.fn(),
    pause: vi.fn(function (this: { paused: boolean }) {
      this.paused = true;
    }),
    resume: vi.fn(function (this: { paused: boolean }) {
      this.paused = false;
    }),
    addEventListener: vi.fn(),
    removeEventListener: vi.fn(),
  };

  vi.stubGlobal("speechSynthesis", synth);
  vi.stubGlobal("SpeechSynthesisUtterance", FakeUtterance);
  return synth;
}

describe("browser TTS", () => {
  beforeEach(() => {
    vi.unstubAllGlobals();
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it("reports unavailable without speechSynthesis", () => {
    vi.stubGlobal("speechSynthesis", undefined);
    expect(isBrowserTTSAvailable()).toBe(false);
  });

  it("lists voices", async () => {
    installFakeSpeech();
    const engine = createBrowserEngine();
    const voices = await engine.listVoices();
    expect(voices).toHaveLength(1);
    expect(voices[0].id).toBe("test-voice");
  });

  it("speaks and resolves ended", async () => {
    const synth = installFakeSpeech();
    const engine = createBrowserEngine();
    const result = await engine.speak({ text: "Hello", voiceId: "test-voice", rate: 1.25 });
    expect(result).toBe("ended");
    expect(synth.speak).toHaveBeenCalledOnce();
  });

  it("returns error on synthesis failure", async () => {
    installFakeSpeech({ failSpeak: true });
    const engine = createBrowserEngine();
    await expect(engine.speak({ text: "Hello" })).resolves.toBe("error");
  });

  it("cancels when aborted", async () => {
    const synth = installFakeSpeech();
    synth.speak = vi.fn(() => {
      // never ends on its own
    });
    const engine = createBrowserEngine();
    const ac = new AbortController();
    const pending = engine.speak({ text: "Hello", signal: ac.signal });
    ac.abort();
    await expect(pending).resolves.toBe("cancelled");
    expect(synth.cancel).toHaveBeenCalled();
  });

  it("treats empty text as ended", async () => {
    installFakeSpeech();
    const engine = createBrowserEngine();
    await expect(engine.speak({ text: "  " })).resolves.toBe("ended");
  });
});
