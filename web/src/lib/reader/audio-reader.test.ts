import { describe, expect, it, vi } from "vitest";
import type { Chapter } from "$lib/api/types";
import {
  AUDIO_SKIP_OPTIONS,
  AUDIO_SLEEP_OPTIONS,
  AUDIO_SPEEDS,
  audioCachePill,
  buildChapterMenuItems,
  buildSpeedMenuItems,
  buildTrackMenuItems,
  cacheBookTrackCount,
  chapterAtTime,
  clampChapterSeekSec,
  handleAudioKeys,
  isVolumeMutedIcon,
  seekDisplayTime,
  seekPlayedPercent,
  sleepRemainingMs,
  volumeSliderValue,
} from "./audio-reader";

describe("constants", () => {
  it("exports playback option lists", () => {
    expect(AUDIO_SPEEDS).toContain(1);
    expect(AUDIO_SPEEDS).toContain(3);
    expect(AUDIO_SKIP_OPTIONS).toEqual([10, 15, 30, 60]);
    expect(AUDIO_SLEEP_OPTIONS).toContain(15);
  });
});

describe("sleepRemainingMs", () => {
  it("returns 0 when unset", () => {
    expect(sleepRemainingMs(null, 1000)).toBe(0);
    expect(sleepRemainingMs(undefined, 1000)).toBe(0);
  });

  it("clamps expired timers to 0", () => {
    expect(sleepRemainingMs(500, 1000)).toBe(0);
  });

  it("returns remaining milliseconds", () => {
    expect(sleepRemainingMs(2500, 1000)).toBe(1500);
  });
});

describe("seek helpers", () => {
  it("picks scrub value while scrubbing", () => {
    expect(seekDisplayTime(true, 42, 10)).toBe(42);
    expect(seekDisplayTime(false, 42, 10)).toBe(10);
  });

  it("computes played percent", () => {
    expect(seekPlayedPercent(30, 100)).toBe(30);
    expect(seekPlayedPercent(10, 0)).toBe(0);
    expect(seekPlayedPercent(10, -1)).toBe(0);
  });
});

describe("chapter helpers", () => {
  const chapters: Chapter[] = [
    { index: 0, title: "Intro", startSec: 0 },
    { index: 1, title: "One", startSec: 60 },
    { index: 2, title: "Two", startSec: 120 },
  ];

  it("finds the chapter at a time", () => {
    expect(chapterAtTime([], 10)).toBeNull();
    expect(chapterAtTime(chapters, 0)?.title).toBe("Intro");
    expect(chapterAtTime(chapters, 59)?.title).toBe("Intro");
    expect(chapterAtTime(chapters, 60)?.title).toBe("One");
    expect(chapterAtTime(chapters, 200)?.title).toBe("Two");
  });

  it("clamps chapter seek targets", () => {
    expect(clampChapterSeekSec(-5, 100)).toBe(0);
    expect(clampChapterSeekSec(50, 100)).toBe(50);
    expect(clampChapterSeekSec(150, 100)).toBe(100);
    expect(clampChapterSeekSec(40, 0)).toBe(40);
  });
});

describe("playlist and menu builders", () => {
  it("builds track menu items with fallback labels", () => {
    const items = buildTrackMenuItems(
      [
        { index: 0, title: "A" },
        { index: 1, title: "" },
      ],
      1,
      (n) => `Track ${n}`,
    );
    expect(items).toEqual([
      { id: "0", label: "A", active: false, trackIndex: 0 },
      { id: "1", label: "Track 2", active: true, trackIndex: 1 },
    ]);
  });

  it("builds chapter menu items", () => {
    const items = buildChapterMenuItems(
      [{ index: 2, title: "Two", startSec: 120 }],
      2,
      (sec) => `t${sec}`,
    );
    expect(items).toEqual([
      {
        id: "2",
        label: "Two",
        hint: "t120",
        active: true,
        chapter: { index: 2, title: "Two", startSec: 120 },
      },
    ]);
  });

  it("builds speed menu items", () => {
    expect(buildSpeedMenuItems([1, 1.5], 1.5)).toEqual([
      { id: "1", label: "1x", active: false, speed: 1 },
      { id: "1.5", label: "1.5x", active: true, speed: 1.5 },
    ]);
  });

  it("uses at least one track for cache ops", () => {
    expect(cacheBookTrackCount(0)).toBe(1);
    expect(cacheBookTrackCount(4)).toBe(4);
  });
});

describe("status and volume helpers", () => {
  it("maps cache pill kinds", () => {
    expect(audioCachePill(true, false, true)).toBe("offlineReady");
    expect(audioCachePill(false, true, true)).toBe("offlineReady");
    expect(audioCachePill(false, false, false)).toBe("offline");
    expect(audioCachePill(false, false, true)).toBe("streaming");
  });

  it("maps mute/volume display", () => {
    expect(volumeSliderValue(true, 0.8)).toBe(0);
    expect(volumeSliderValue(false, 0.8)).toBe(0.8);
    expect(isVolumeMutedIcon(true, 1)).toBe(true);
    expect(isVolumeMutedIcon(false, 0)).toBe(true);
    expect(isVolumeMutedIcon(false, 0.5)).toBe(false);
  });
});

describe("handleAudioKeys", () => {
  it("toggles play on space", () => {
    const togglePlay = vi.fn();
    const seekBy = vi.fn();
    const event = new KeyboardEvent("keydown", { key: " ", code: "Space", cancelable: true });

    expect(handleAudioKeys(event, { togglePlay, seekBy, skipSeconds: 15 })).toBe(true);
    expect(event.defaultPrevented).toBe(true);
    expect(togglePlay).toHaveBeenCalledOnce();
    expect(seekBy).not.toHaveBeenCalled();
  });

  it("seeks with arrow keys", () => {
    const togglePlay = vi.fn();
    const seekBy = vi.fn();

    handleAudioKeys(new KeyboardEvent("keydown", { key: "ArrowLeft" }), {
      togglePlay,
      seekBy,
      skipSeconds: 30,
    });
    handleAudioKeys(new KeyboardEvent("keydown", { key: "ArrowRight" }), {
      togglePlay,
      seekBy,
      skipSeconds: 30,
    });

    expect(seekBy).toHaveBeenNthCalledWith(1, -30);
    expect(seekBy).toHaveBeenNthCalledWith(2, 30);
    expect(togglePlay).not.toHaveBeenCalled();
  });

  it("ignores shortcuts while typing", () => {
    const togglePlay = vi.fn();
    const input = document.createElement("input");
    document.body.appendChild(input);

    const event = {
      key: " ",
      code: "Space",
      target: input,
      preventDefault: vi.fn(),
    } as unknown as KeyboardEvent;

    expect(handleAudioKeys(event, { togglePlay, seekBy: vi.fn(), skipSeconds: 10 })).toBe(false);
    expect(togglePlay).not.toHaveBeenCalled();
    input.remove();
  });
});
