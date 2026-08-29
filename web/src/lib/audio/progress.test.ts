import { describe, expect, it } from "vitest";
import { audioPercentFromSizes, encodeAudioLocation, parseAudioLocation } from "./progress";

describe("audio progress location", () => {
  it("encodes and parses track + seconds", () => {
    const loc = encodeAudioLocation(2, 125.5);
    expect(loc).toBe("audio:2:125.500");
    expect(parseAudioLocation(loc)).toEqual({ trackIndex: 2, seconds: 125.5 });
  });

  it("parses legacy bare seconds", () => {
    expect(parseAudioLocation("42.5")).toEqual({ trackIndex: 0, seconds: 42.5 });
  });

  it("handles empty location", () => {
    expect(parseAudioLocation("")).toEqual({ trackIndex: 0, seconds: 0 });
    expect(parseAudioLocation(null)).toEqual({ trackIndex: 0, seconds: 0 });
  });
});

describe("audioPercentFromSizes", () => {
  it("weights by file size", () => {
    const tracks = [{ fileSize: 100 }, { fileSize: 100 }];
    expect(audioPercentFromSizes(tracks, 0, 0.5)).toBeCloseTo(0.25);
    expect(audioPercentFromSizes(tracks, 1, 0.5)).toBeCloseTo(0.75);
  });

  it("falls back to track count when sizes are zero", () => {
    const tracks = [{ fileSize: 0 }, { fileSize: 0 }];
    expect(audioPercentFromSizes(tracks, 0, 0.5)).toBeCloseTo(0.25);
  });
});
