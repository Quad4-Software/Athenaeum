/** Playback speed choices offered in the audio player and mini player. */
export const AUDIO_SPEEDS = [0.75, 1, 1.25, 1.5, 1.75, 2, 2.5, 3] as const;

/** Skip-interval choices (seconds) offered in the audio reader. */
export const AUDIO_SKIP_OPTIONS = [10, 15, 30, 60] as const;

/** Sleep-timer choices (minutes) offered in the audio player and mini player. */
export const AUDIO_SLEEP_OPTIONS = [5, 15, 30, 45, 60, 90] as const;
