/** Playback speed choices offered in the audio player and mini player. */
export const AUDIO_SPEEDS = [0.75, 1, 1.25, 1.5, 1.75, 2, 2.5, 3] as const;

/** Fallback playback rate when a stored value is missing or unusable. */
export const AUDIO_RATE_DEFAULT = 1;

/**
 * Clamp a playback rate into the offered speed range. Keeps a stale
 * stored value like 4 from leaking onto the media element after the
 * option list shrinks.
 */
export function clampAudioRate(rate: number): number {
  if (!Number.isFinite(rate) || rate <= 0) return AUDIO_RATE_DEFAULT;
  return Math.min(Math.max(rate, AUDIO_SPEEDS[0]), AUDIO_SPEEDS[AUDIO_SPEEDS.length - 1]);
}

/** Skip-interval choices (seconds) offered in the audio reader. */
export const AUDIO_SKIP_OPTIONS = [10, 15, 30, 60] as const;

/** Sleep-timer choices (minutes) offered in the audio player and mini player. */
export const AUDIO_SLEEP_OPTIONS = [5, 15, 30, 45, 60, 90] as const;
