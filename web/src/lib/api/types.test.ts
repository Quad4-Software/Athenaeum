import { describe, expect, it } from "vitest";
import { isAudioFormat } from "./types";

describe("isAudioFormat", () => {
  it("detects audio extensions", () => {
    expect(isAudioFormat("mp3")).toBe(true);
    expect(isAudioFormat("m4b")).toBe(true);
  });

  it("rejects non-audio", () => {
    expect(isAudioFormat("epub")).toBe(false);
    expect(isAudioFormat("pdf")).toBe(false);
  });
});
