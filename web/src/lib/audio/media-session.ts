export interface MediaSessionMeta {
  title: string;
  artist?: string;
  album?: string;
  artworkUrl?: string;
}

export interface MediaSessionHandlers {
  onPlay?: () => void;
  onPause?: () => void;
  onSeekBackward?: () => void;
  onSeekForward?: () => void;
  onSeekTo?: (seconds: number) => void;
}

export function bindMediaSession(
  meta: MediaSessionMeta,
  handlers: MediaSessionHandlers,
): () => void {
  if (typeof navigator === "undefined" || !("mediaSession" in navigator)) {
    return () => undefined;
  }

  const artwork = meta.artworkUrl
    ? [{ src: meta.artworkUrl, sizes: "512x512", type: "image/jpeg" }]
    : [];

  navigator.mediaSession.metadata = new MediaMetadata({
    title: meta.title,
    artist: meta.artist ?? "",
    album: meta.album ?? "",
    artwork,
  });

  const set = (
    action: MediaSessionAction,
    fn?: (() => void) | ((details: MediaSessionActionDetails) => void),
  ) => {
    if (!fn) {
      try {
        navigator.mediaSession.setActionHandler(action, null);
      } catch {
        // ignore unsupported actions
      }
      return;
    }
    try {
      navigator.mediaSession.setActionHandler(action, fn as () => void);
    } catch {
      // ignore unsupported actions
    }
  };

  set("play", handlers.onPlay);
  set("pause", handlers.onPause);
  set("seekbackward", handlers.onSeekBackward);
  set("seekforward", handlers.onSeekForward);
  set(
    "seekto",
    handlers.onSeekTo
      ? (details) => {
          const t = details.seekTime;
          if (t != null && Number.isFinite(t)) handlers.onSeekTo?.(t);
        }
      : undefined,
  );

  return () => {
    set("play");
    set("pause");
    set("seekbackward");
    set("seekforward");
    set("seekto");
    navigator.mediaSession.metadata = null;
    navigator.mediaSession.playbackState = "none";
  };
}

export function setMediaPlaybackState(playing: boolean) {
  if (typeof navigator === "undefined" || !("mediaSession" in navigator)) return;
  navigator.mediaSession.playbackState = playing ? "playing" : "paused";
}

export function setMediaPositionState(duration: number, position: number, rate: number) {
  if (typeof navigator === "undefined" || !("mediaSession" in navigator)) return;
  if (!Number.isFinite(duration) || duration <= 0) return;
  try {
    navigator.mediaSession.setPositionState({
      duration,
      playbackRate: rate,
      position: Math.min(Math.max(position, 0), duration),
    });
  } catch {
    // ignore
  }
}
