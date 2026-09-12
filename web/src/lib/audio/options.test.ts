import { describe, expect, it } from "vitest";
import {
  AUDIO_RATE_DEFAULT,
  AUDIO_SKIP_OPTIONS,
  AUDIO_SLEEP_OPTIONS,
  AUDIO_SPEEDS,
  clampAudioRate,
} from "./options";

describe("audio options", () => {
  it("offers 1x playback speed", () => {
    expect(AUDIO_SPEEDS).toContain(1);
  });

  it("keeps speeds sorted ascending", () => {
    const sorted = [...AUDIO_SPEEDS].sort((a, b) => a - b);
    expect([...AUDIO_SPEEDS]).toEqual(sorted);
  });

  it("keeps sleep options sorted ascending minutes", () => {
    const sorted = [...AUDIO_SLEEP_OPTIONS].sort((a, b) => a - b);
    expect([...AUDIO_SLEEP_OPTIONS]).toEqual(sorted);
  });

  it("keeps skip options sorted ascending seconds", () => {
    const sorted = [...AUDIO_SKIP_OPTIONS].sort((a, b) => a - b);
    expect([...AUDIO_SKIP_OPTIONS]).toEqual(sorted);
  });
});

describe("clampAudioRate", () => {
  it("passes through rates inside the offered range", () => {
    expect(clampAudioRate(1)).toBe(1);
    expect(clampAudioRate(1.3)).toBe(1.3);
    expect(clampAudioRate(AUDIO_SPEEDS[0])).toBe(AUDIO_SPEEDS[0]);
    expect(clampAudioRate(AUDIO_SPEEDS[AUDIO_SPEEDS.length - 1])).toBe(
      AUDIO_SPEEDS[AUDIO_SPEEDS.length - 1],
    );
  });

  it("clamps stale stored values into the offered range", () => {
    expect(clampAudioRate(4)).toBe(AUDIO_SPEEDS[AUDIO_SPEEDS.length - 1]);
    expect(clampAudioRate(0.5)).toBe(AUDIO_SPEEDS[0]);
  });

  it("falls back to the default for unusable values", () => {
    expect(clampAudioRate(Number.NaN)).toBe(AUDIO_RATE_DEFAULT);
    expect(clampAudioRate(Number.POSITIVE_INFINITY)).toBe(AUDIO_RATE_DEFAULT);
    expect(clampAudioRate(0)).toBe(AUDIO_RATE_DEFAULT);
    expect(clampAudioRate(-2)).toBe(AUDIO_RATE_DEFAULT);
  });
});
