/** Encoded audiobook resume location: audio:{track}:{seconds} or legacy bare seconds. */

export interface AudioLocation {
  trackIndex: number;
  seconds: number;
}

const PREFIX = "audio:";

export function encodeAudioLocation(trackIndex: number, seconds: number): string {
  const track = Math.max(0, Math.floor(trackIndex));
  const sec = Math.max(0, seconds);
  return `${PREFIX}${track}:${sec.toFixed(3)}`;
}

export function parseAudioLocation(location: string | undefined | null): AudioLocation {
  if (!location) return { trackIndex: 0, seconds: 0 };
  if (location.startsWith(PREFIX)) {
    const rest = location.slice(PREFIX.length);
    const colon = rest.indexOf(":");
    if (colon >= 0) {
      const trackIndex = Number(rest.slice(0, colon));
      const seconds = Number(rest.slice(colon + 1));
      return {
        trackIndex: Number.isFinite(trackIndex) && trackIndex >= 0 ? Math.floor(trackIndex) : 0,
        seconds: Number.isFinite(seconds) && seconds >= 0 ? seconds : 0,
      };
    }
  }
  const seconds = Number(location);
  return {
    trackIndex: 0,
    seconds: Number.isFinite(seconds) && seconds >= 0 ? seconds : 0,
  };
}

/** Overall book percent from track sizes when durations are unknown. */
export function audioPercentFromSizes(
  tracks: { fileSize: number }[],
  trackIndex: number,
  trackPercent: number,
): number {
  if (!tracks.length) return Math.min(1, Math.max(0, trackPercent));
  const total = tracks.reduce((sum, t) => sum + Math.max(0, t.fileSize), 0);
  if (total <= 0) {
    return Math.min(1, Math.max(0, (trackIndex + trackPercent) / tracks.length));
  }
  let before = 0;
  for (let i = 0; i < trackIndex && i < tracks.length; i++) {
    before += Math.max(0, tracks[i].fileSize);
  }
  const current = tracks[Math.min(trackIndex, tracks.length - 1)]?.fileSize ?? 0;
  const within = Math.max(0, current) * Math.min(1, Math.max(0, trackPercent));
  return Math.min(1, (before + within) / total);
}
