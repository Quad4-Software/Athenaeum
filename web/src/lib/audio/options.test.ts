import { describe, expect, it } from "vitest";
import { AUDIO_SKIP_OPTIONS, AUDIO_SLEEP_OPTIONS, AUDIO_SPEEDS } from "./options";

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
