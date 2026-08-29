import { api } from "$lib/api/client";
import type { AudiobookTrack, Book, Chapter } from "$lib/api/types";
import { audioCache, type AudioCacheStatus } from "$lib/audio/cache";
import {
  bindMediaSession,
  setMediaPlaybackState,
  setMediaPositionState,
} from "$lib/audio/media-session";
import {
  audioPercentFromSizes,
  encodeAudioLocation,
  parseAudioLocation,
} from "$lib/audio/progress";
import { storageKey } from "$lib/brand/storage";

export type AudioProgressFn = (
  location: string,
  percent: number,
  seconds: number,
  trackIndex: number,
) => void;

const RATE_KEY = storageKey("audio-rate");
const SKIP_KEY = storageKey("audio-skip");

class AudioPlayerStore {
  book = $state<Book | null>(null);
  streamUrl = $state("");
  active = $state(false);
  expanded = $state(false);

  trackIndex = $state(0);
  playlist = $state<AudiobookTrack[]>([]);

  playing = $state(false);
  current = $state(0);
  duration = $state(0);
  rate = $state(
    typeof localStorage !== "undefined" ? Number(localStorage.getItem(RATE_KEY)) || 1 : 1,
  );
  volume = $state(1);
  muted = $state(false);
  skipSeconds = $state(
    typeof localStorage !== "undefined" ? Number(localStorage.getItem(SKIP_KEY)) || 10 : 10,
  );
  scrubbing = $state(false);
  scrubValue = $state(0);

  chapters = $state<Chapter[]>([]);
  sleepEndsAt = $state<number | null>(null);
  sleepTick = $state(0);
  online = $state(typeof navigator !== "undefined" ? navigator.onLine : true);
  usingOffline = $state(false);
  coverFailed = $state(false);
  cacheStatus = $state<AudioCacheStatus>({
    cachedBytes: 0,
    totalBytes: 0,
    complete: false,
    downloading: false,
    error: null,
  });

  initialSeconds = 0;
  resumeApplied = false;

  private audio: HTMLAudioElement | null = null;
  private onProgress: AudioProgressFn | null = null;
  private progressTimer: ReturnType<typeof setInterval> | null = null;
  private sleepTimer: ReturnType<typeof setInterval> | null = null;
  private unsubCache: (() => void) | null = null;
  private releaseMedia: (() => void) | null = null;
  private releaseConnectivity: (() => void) | null = null;

  get showBar(): boolean {
    return this.active && !this.expanded;
  }

  get playbackUrl(): string {
    const book = this.book;
    if (!book) return "";
    if (book.format === "audiobook" && this.playlist.length) {
      return api.fileUrl(book.id, this.playlist[this.trackIndex]?.index ?? 0);
    }
    return this.streamUrl;
  }

  get coverSrc(): string {
    const book = this.book;
    if (!book || !book.hasCover || this.coverFailed) return "";
    return api.coverUrl(book.id, book.modifiedAt);
  }

  get bufferedEnd(): number {
    const audio = this.audio;
    const duration = this.duration;
    if (!audio || !duration) return 0;
    try {
      const ranges = audio.buffered;
      for (let i = 0; i < ranges.length; i++) {
        if (ranges.start(i) <= audio.currentTime && audio.currentTime <= ranges.end(i)) {
          return ranges.end(i);
        }
      }
      if (ranges.length > 0) return ranges.end(ranges.length - 1);
    } catch {
      // ignore
    }
    return 0;
  }

  get cachePercent(): number {
    return this.cacheStatus.totalBytes > 0
      ? Math.min(100, (this.cacheStatus.cachedBytes / this.cacheStatus.totalBytes) * 100)
      : 0;
  }

  get bufferedPercent(): number {
    return this.duration > 0 ? (this.bufferedEnd / this.duration) * 100 : 0;
  }

  get playbackTime(): number {
    return this.scrubbing ? this.scrubValue : this.current;
  }

  get currentChapter(): Chapter | null {
    if (!this.chapters.length) return null;
    const t = this.playbackTime;
    for (let i = this.chapters.length - 1; i >= 0; i--) {
      if (t >= this.chapters[i].startSec) return this.chapters[i];
    }
    return this.chapters[0];
  }

  bindAudio(el: HTMLAudioElement | null) {
    this.audio = el;
  }

  startSession(book: Book, url: string, initialLocation = "", onProgress?: AudioProgressFn) {
    const sameBook = this.book?.id === book.id;
    this.book = book;
    this.streamUrl = url;
    this.active = true;
    this.onProgress = onProgress ?? null;
    if (!sameBook) {
      const loc = parseAudioLocation(initialLocation);
      this.initialSeconds = loc.seconds;
      this.resumeApplied = false;
      this.trackIndex = loc.trackIndex;
      this.playlist = [];
      this.coverFailed = false;
      this.current = 0;
      this.duration = 0;
    }
  }

  setExpanded(expanded: boolean) {
    this.expanded = expanded;
  }

  stop() {
    this.lifecycleBookId = null;
    this.teardown();
    this.book = null;
    this.streamUrl = "";
    this.active = false;
    this.expanded = false;
    this.playing = false;
    this.current = 0;
    this.duration = 0;
    this.playlist = [];
    this.trackIndex = 0;
    this.chapters = [];
    this.onProgress = null;
    if (this.audio) {
      this.audio.pause();
      this.audio.removeAttribute("src");
      this.audio.load();
    }
  }

  private lifecycleBookId: number | null = null;

  beginLifecycle(): () => void {
    const book = this.book;
    if (!book) return () => undefined;

    const bookId = book.id;
    if (this.lifecycleBookId === bookId) {
      return () => undefined;
    }

    if (this.lifecycleBookId != null) {
      this.teardownPartial();
    }
    this.lifecycleBookId = bookId;

    if (book.format === "audiobook") {
      void api.getAudiobookTracks(bookId).then((items) => {
        if (this.book?.id !== bookId) return;
        this.playlist = items;
        if (this.trackIndex >= items.length) this.trackIndex = Math.max(0, items.length - 1);
        this.bindCacheStatus(bookId);
        this.startTrackPrefetch();
        void this.setupSource();
      });
    } else {
      this.bindCacheStatus(bookId);
      this.startTrackPrefetch();
    }

    const onOnline = () => {
      this.online = true;
      if (this.book?.id === bookId) this.startTrackPrefetch();
    };
    const onOffline = () => {
      this.online = false;
      void this.setupSource();
    };
    window.addEventListener("online", onOnline);
    window.addEventListener("offline", onOffline);
    this.releaseConnectivity = () => {
      window.removeEventListener("online", onOnline);
      window.removeEventListener("offline", onOffline);
    };

    void this.setupSource();
    void api.getChapters(bookId).then((items) => {
      if (this.book?.id === bookId) this.chapters = items;
    });

    this.progressTimer = setInterval(() => this.reportProgress(true), 5000);
    this.sleepTimer = setInterval(() => {
      this.sleepTick = Date.now();
    }, 1000);

    this.releaseMedia = bindMediaSession(
      {
        title: book.title,
        artist: book.author,
        album: book.series,
        artworkUrl: this.coverSrc || undefined,
      },
      {
        onPlay: () => void this.audio?.play(),
        onPause: () => this.audio?.pause(),
        onSeekBackward: () => this.seekBy(-this.skipSeconds),
        onSeekForward: () => this.seekBy(this.skipSeconds),
        onSeekTo: (seconds) => {
          if (!this.audio || !this.duration) return;
          this.audio.currentTime = Math.min(Math.max(seconds, 0), this.duration);
          this.reportProgress(true);
        },
      },
    );

    return () => {
      if (this.lifecycleBookId !== bookId) return;
      this.lifecycleBookId = null;
      this.teardownPartial();
    };
  }

  private teardownPartial() {
    if (this.progressTimer) clearInterval(this.progressTimer);
    if (this.sleepTimer) clearInterval(this.sleepTimer);
    this.progressTimer = null;
    this.sleepTimer = null;
    this.releaseConnectivity?.();
    this.releaseConnectivity = null;
    this.unsubCache?.();
    this.unsubCache = null;
    this.releaseMedia?.();
    this.releaseMedia = null;
    if (this.book) {
      audioCache.releaseBook(this.book.id, Math.max(this.playlist.length, 1));
    }
  }

  private teardown() {
    this.teardownPartial();
  }

  private bindCacheStatus(bookId: number) {
    this.unsubCache?.();
    const key = this.playlist.length > 1 ? this.trackCacheKey() : String(bookId);
    this.unsubCache = audioCache.subscribe(key, (status) => {
      if (this.book?.id !== bookId) return;
      if (this.playlist.length > 1) {
        void audioCache.getBookStatus(bookId, this.playlist.length).then((agg) => {
          if (this.book?.id === bookId) this.cacheStatus = agg;
        });
      } else {
        this.cacheStatus = status;
      }
      if (status.complete) void this.setupSource();
    });
  }

  private trackCacheKey(trackIndex = this.trackIndex): string {
    const book = this.book;
    if (!book) return "";
    if (book.format === "audiobook" && this.playlist.length > 1) {
      return `${book.id}:${trackIndex}`;
    }
    return String(book.id);
  }

  private trackFileSize(trackIndex = this.trackIndex): number {
    const book = this.book;
    if (!book) return 0;
    if (book.format === "audiobook" && this.playlist.length) {
      return this.playlist[trackIndex]?.fileSize || 0;
    }
    return book.fileSize;
  }

  private startTrackPrefetch() {
    const book = this.book;
    if (!book) return;
    if (book.format === "audiobook" && this.playlist.length > 1) {
      for (let i = 0; i < this.playlist.length; i++) {
        const track = this.playlist[i];
        audioCache.startPrefetch(
          `${book.id}:${i}`,
          api.fileUrl(book.id, track.index),
          track.fileSize,
          book.modifiedAt,
        );
      }
      return;
    }
    audioCache.startPrefetch(book.id, this.playbackUrl, book.fileSize, book.modifiedAt);
  }

  /** Public wrapper for UI download button. */
  startTrackPrefetchPublic() {
    this.startTrackPrefetch();
  }

  async setupSource() {
    const audio = this.audio;
    const book = this.book;
    if (!audio || !book) return;
    const url = this.playbackUrl;
    const key = this.trackCacheKey();
    const fileSize = this.trackFileSize();
    const playbackUrl = await audioCache.resolvePlaybackUrl(key, url, fileSize, book.modifiedAt);
    this.usingOffline = playbackUrl.startsWith("blob:");
    if (audio.src !== playbackUrl) {
      const wasPlaying = !audio.paused;
      const at = !this.resumeApplied ? this.initialSeconds : audio.currentTime;
      audio.src = playbackUrl;
      audio.load();
      if (at > 0) audio.currentTime = at;
      if (wasPlaying) void audio.play().catch(() => undefined);
    }
  }

  selectTrack(index: number) {
    if (!this.playlist.length || index < 0 || index >= this.playlist.length) return;
    if (index === this.trackIndex) return;
    this.reportProgress(true);
    this.trackIndex = index;
    this.initialSeconds = 0;
    this.resumeApplied = true;
    this.current = 0;
    void this.setupSource().then(() => void this.audio?.play());
  }

  togglePlay() {
    const audio = this.audio;
    if (!audio) return;
    if (audio.paused) void audio.play();
    else audio.pause();
  }

  seekBy(delta: number) {
    const audio = this.audio;
    if (!audio || !this.duration) return;
    audio.currentTime = Math.min(Math.max(audio.currentTime + delta, 0), this.duration);
    this.reportProgress(true);
  }

  seekToChapter(chapter: Chapter) {
    const audio = this.audio;
    if (!audio) return;
    const next = Math.min(Math.max(chapter.startSec, 0), this.duration || chapter.startSec);
    audio.currentTime = next;
    this.current = next;
    this.scrubValue = next;
    this.reportProgress(true);
  }

  setRate(next: number) {
    this.rate = next;
    localStorage.setItem(RATE_KEY, String(next));
    if (this.audio) this.audio.playbackRate = next;
    this.reportProgress(true);
  }

  setVolume(next: number) {
    this.volume = next;
    this.muted = next === 0;
    if (this.audio) {
      this.audio.volume = next;
      this.audio.muted = this.muted;
    }
  }

  toggleMute() {
    this.muted = !this.muted;
    if (this.audio) this.audio.muted = this.muted;
  }

  applyScrub() {
    const audio = this.audio;
    if (!audio) return;
    audio.currentTime = this.scrubValue;
    this.scrubbing = false;
    this.reportProgress(true);
  }

  setSleepTimer(minutes: number) {
    this.sleepEndsAt = Date.now() + minutes * 60 * 1000;
  }

  clearSleepTimer() {
    this.sleepEndsAt = null;
  }

  reportProgress(force = false) {
    const audio = this.audio;
    if (!audio || !this.duration) return;
    const trackPct = this.duration > 0 ? audio.currentTime / this.duration : 0;
    const pct =
      this.playlist.length > 1
        ? audioPercentFromSizes(this.playlist, this.trackIndex, trackPct)
        : trackPct;
    const location = encodeAudioLocation(this.trackIndex, audio.currentTime);
    if (this.onProgress) {
      this.onProgress(location, pct, audio.currentTime, this.trackIndex);
    } else if (this.book) {
      void api.saveProgress(this.book.id, { location, percent: pct }).catch(() => undefined);
    }
    if (force) {
      setMediaPositionState(this.duration, audio.currentTime, this.rate);
    }
  }

  onLoaded() {
    const audio = this.audio;
    if (!audio) return;
    this.duration = audio.duration || 0;
    audio.playbackRate = this.rate;
    audio.volume = this.volume;
    audio.muted = this.muted;
    if (!this.resumeApplied && this.initialSeconds > 0 && this.initialSeconds < this.duration) {
      audio.currentTime = this.initialSeconds;
      this.current = this.initialSeconds;
      this.resumeApplied = true;
    }
    this.reportProgress(true);
  }

  onPlay() {
    this.playing = true;
    setMediaPlaybackState(true);
  }

  onPause() {
    this.playing = false;
    setMediaPlaybackState(false);
    this.reportProgress(true);
  }

  onTimeUpdate() {
    const audio = this.audio;
    if (!audio || this.scrubbing) return;
    this.current = audio.currentTime;
    if (this.sleepEndsAt && Date.now() >= this.sleepEndsAt) {
      this.sleepEndsAt = null;
      audio.pause();
    }
    setMediaPositionState(this.duration, this.current, this.rate);
  }

  onEnded() {
    const audio = this.audio;
    const book = this.book;
    if (
      book?.format === "audiobook" &&
      this.playlist.length &&
      this.trackIndex < this.playlist.length - 1
    ) {
      this.trackIndex += 1;
      this.initialSeconds = 0;
      this.resumeApplied = true;
      this.current = 0;
      this.reportProgress(true);
      void this.setupSource().then(() => void audio?.play());
      return;
    }
    this.playing = false;
    this.reportProgress(true);
  }

  onStalled() {
    if (this.cacheStatus.complete) void this.setupSource();
  }

  onError() {
    if (!this.usingOffline && this.cacheStatus.cachedBytes > 0) {
      void this.setupSource();
    }
  }

  onStreamUrlChange() {
    if (!this.book) return;
    void this.setupSource();
    this.startTrackPrefetch();
  }
}

export const audioPlayer = new AudioPlayerStore();
