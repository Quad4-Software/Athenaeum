import type { AudiobookTrack, Chapter } from "$lib/api/types";
import { isTypingTarget } from "./reader-keys";

export const AUDIO_SPEEDS = [0.75, 1, 1.25, 1.5, 1.75, 2, 2.5, 3] as const;
export const AUDIO_SKIP_OPTIONS = [10, 15, 30, 60] as const;
export const AUDIO_SLEEP_OPTIONS = [5, 15, 30, 45, 60, 90] as const;

/** Milliseconds left on a sleep timer, or 0 when unset/expired. */
export function sleepRemainingMs(endsAt: number | null | undefined, nowMs: number): number {
  if (!endsAt) return 0;
  return Math.max(0, endsAt - nowMs);
}

/** Current or scrub position shown on the seek bar and time labels. */
export function seekDisplayTime(scrubbing: boolean, scrubValue: number, current: number): number {
  return scrubbing ? scrubValue : current;
}

/** Played fraction of the seek bar as a 0-100 percentage. */
export function seekPlayedPercent(time: number, duration: number): number {
  if (!(duration > 0)) return 0;
  return (time / duration) * 100;
}

/** Clamp a chapter start time into a playable seek position. */
export function clampChapterSeekSec(startSec: number, duration: number): number {
  return Math.min(Math.max(startSec, 0), duration || startSec);
}

/** Chapter whose start is at or before `timeSec`, or the first chapter. */
export function chapterAtTime(chapters: Chapter[], timeSec: number): Chapter | null {
  if (!chapters.length) return null;
  for (let i = chapters.length - 1; i >= 0; i--) {
    if (timeSec >= chapters[i].startSec) return chapters[i];
  }
  return chapters[0];
}

/** Track count used when clearing/releasing per-book cache entries. */
export function cacheBookTrackCount(playlistLength: number): number {
  return Math.max(playlistLength, 1);
}

export type AudioCachePill = "offlineReady" | "offline" | "streaming";

/** Which connectivity/cache status pill to show. */
export function audioCachePill(
  usingOffline: boolean,
  cacheComplete: boolean,
  online: boolean,
): AudioCachePill {
  if (usingOffline || cacheComplete) return "offlineReady";
  if (!online) return "offline";
  return "streaming";
}

export function volumeSliderValue(muted: boolean, volume: number): number {
  return muted ? 0 : volume;
}

export function isVolumeMutedIcon(muted: boolean, volume: number): boolean {
  return muted || volume === 0;
}

export interface AudioTrackMenuItem {
  id: string;
  label: string;
  active: boolean;
  trackIndex: number;
}

export function buildTrackMenuItems(
  playlist: Pick<AudiobookTrack, "index" | "title">[],
  activeIndex: number,
  fallbackLabel: (n: number) => string,
): AudioTrackMenuItem[] {
  return playlist.map((track, i) => ({
    id: String(track.index),
    label: track.title || fallbackLabel(i + 1),
    active: activeIndex === i,
    trackIndex: i,
  }));
}

export interface AudioChapterMenuItem {
  id: string;
  label: string;
  hint: string;
  active: boolean;
  chapter: Chapter;
}

export function buildChapterMenuItems(
  chapters: Chapter[],
  activeChapterIndex: number | undefined,
  formatTime: (sec: number) => string,
): AudioChapterMenuItem[] {
  return chapters.map((chapter) => ({
    id: String(chapter.index),
    label: chapter.title,
    hint: formatTime(chapter.startSec),
    active: activeChapterIndex === chapter.index,
    chapter,
  }));
}

export interface AudioSpeedMenuItem {
  id: string;
  label: string;
  active: boolean;
  speed: number;
}

export function buildSpeedMenuItems(
  speeds: readonly number[],
  activeRate: number,
): AudioSpeedMenuItem[] {
  return speeds.map((speed) => ({
    id: String(speed),
    label: `${speed}x`,
    active: activeRate === speed,
    speed,
  }));
}

export interface AudioKeyHandlers {
  togglePlay: () => void;
  seekBy: (deltaSeconds: number) => void;
  skipSeconds: number;
}

/** Space toggles play. Arrows seek by skip interval. Returns true when handled. */
export function handleAudioKeys(event: KeyboardEvent, handlers: AudioKeyHandlers): boolean {
  if (isTypingTarget(event.target)) return false;

  if (event.key === " " || event.code === "Space") {
    event.preventDefault();
    handlers.togglePlay();
    return true;
  }
  if (event.key === "ArrowLeft") {
    handlers.seekBy(-handlers.skipSeconds);
    return true;
  }
  if (event.key === "ArrowRight") {
    handlers.seekBy(handlers.skipSeconds);
    return true;
  }
  return false;
}
