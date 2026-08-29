import { describe, expect, it } from "vitest";
import { formatAudioTime, formatSleepRemaining } from "./format";

describe("formatAudioTime", () => {
  it("formats short durations", () => {
    expect(formatAudioTime(0)).toBe("0:00");
    expect(formatAudioTime(65)).toBe("1:05");
  });

  it("formats hour-long durations", () => {
    expect(formatAudioTime(3661)).toBe("1:01:01");
  });
});

describe("formatSleepRemaining", () => {
  it("formats remaining sleep time", () => {
    expect(formatSleepRemaining(125_000)).toBe("2:05");
  });
});
