<script lang="ts">
  import {
    ChevronDown,
    CloudOff,
    Download,
    Gauge,
    Headphones,
    ListMusic,
    Moon,
    Pause,
    Play,
    RotateCcw,
    SkipBack,
    SkipForward,
    Trash2,
    Volume2,
    VolumeX,
    Wifi,
    WifiOff,
  } from "@lucide/svelte";
  import MenuList from "$lib/components/MenuList.svelte";
  import Popover from "$lib/components/Popover.svelte";
  import type { Book, Chapter } from "$lib/api/types";
  import { audioCache } from "$lib/audio/cache";
  import { formatAudioTime, formatSleepRemaining } from "$lib/audio/format";
  import {
    AUDIO_SKIP_OPTIONS,
    AUDIO_SLEEP_OPTIONS,
    AUDIO_SPEEDS,
    audioCachePill,
    buildChapterMenuItems,
    buildSpeedMenuItems,
    buildTrackMenuItems,
    cacheBookTrackCount,
    handleAudioKeys,
    isVolumeMutedIcon,
    seekDisplayTime,
    seekPlayedPercent,
    sleepRemainingMs,
    volumeSliderValue,
  } from "$lib/reader/audio-reader";
  import { audioPlayer } from "$lib/stores/audioPlayer.svelte";
  import { narrator } from "$lib/stores/narrator.svelte";
  import { i18n } from "$lib/stores/i18n.svelte";
  import { storageKey } from "$lib/brand/storage";

  interface Props {
    book: Book;
    url: string;
    initialLocation?: string;
    onProgress?: (location: string, percent: number, seconds: number, trackIndex: number) => void;
  }

  let { book, url, initialLocation = "", onProgress }: Props = $props();

  const SKIP_KEY = storageKey("audio-skip");

  let showMore = $state(false);
  let chaptersOpen = $state(false);
  let tracksOpen = $state(false);
  let speedOpen = $state(false);
  let sleepOpen = $state(false);

  let sleepLeftMs = $derived(
    sleepRemainingMs(audioPlayer.sleepEndsAt, audioPlayer.sleepTick || Date.now()),
  );
  let seekTime = $derived(
    seekDisplayTime(audioPlayer.scrubbing, audioPlayer.scrubValue, audioPlayer.current),
  );
  let playedPct = $derived(seekPlayedPercent(seekTime, audioPlayer.duration));
  let cachePill = $derived(
    audioCachePill(
      audioPlayer.usingOffline,
      audioPlayer.cacheStatus.complete,
      audioPlayer.online,
    ),
  );
  let trackItems = $derived(
    buildTrackMenuItems(audioPlayer.playlist, audioPlayer.trackIndex, (n) =>
      i18n.t("audio.trackN", { n: String(n) }),
    ),
  );
  let chapterItems = $derived(
    buildChapterMenuItems(
      audioPlayer.chapters,
      audioPlayer.currentChapter?.index,
      formatAudioTime,
    ),
  );
  let speedItems = $derived(buildSpeedMenuItems(AUDIO_SPEEDS, audioPlayer.rate));

  function seekToChapter(chapter: Chapter) {
    audioPlayer.seekToChapter(chapter);
    chaptersOpen = false;
  }

  function onKey(event: KeyboardEvent) {
    handleAudioKeys(event, {
      togglePlay: () => audioPlayer.togglePlay(),
      seekBy: (delta) => audioPlayer.seekBy(delta),
      skipSeconds: audioPlayer.skipSeconds,
    });
  }

  $effect(() => {
    narrator.stop();
    audioPlayer.startSession(book, url, initialLocation, onProgress);
    audioPlayer.setExpanded(true);
    return () => audioPlayer.setExpanded(false);
  });
</script>

<svelte:window onkeydown={onKey} />

<div class="player">
  <div class="player-body">
    <div class="status-row">
      {#if cachePill === "offlineReady"}
        <span class="status-pill status-pill--ok">
          <CloudOff size={12} />
          {i18n.t("audio.offlineReady")}
        </span>
      {:else if cachePill === "offline"}
        <span class="status-pill status-pill--warn">
          <WifiOff size={12} />
          {i18n.t("audio.offline")}
        </span>
      {:else}
        <span class="status-pill">
          <Wifi size={12} />
          {i18n.t("audio.streaming")}
        </span>
      {/if}
      {#if audioPlayer.cacheStatus.downloading}
        <span class="status-pill"
          >{i18n.t("audio.caching", { pct: String(Math.round(audioPlayer.cachePercent)) })}</span
        >
      {:else if audioPlayer.cachePercent > 0 && !audioPlayer.cacheStatus.complete}
        <span class="status-pill"
          >{i18n.t("audio.cached", { pct: String(Math.round(audioPlayer.cachePercent)) })}</span
        >
      {/if}
      {#if audioPlayer.cacheStatus.error && audioPlayer.online}
        <span class="status-pill status-pill--warn">{i18n.t("audio.cachePaused")}</span>
      {/if}
    </div>

    <div class="artwork-block">
      <div class="artwork">
        {#if audioPlayer.coverSrc}
          <img
            src={audioPlayer.coverSrc}
            alt={`Cover of ${book.title}`}
            class="artwork-img"
            onerror={() => (audioPlayer.coverFailed = true)}
          />
        {:else}
          <div class="artwork-fallback">
            <Headphones size={44} strokeWidth={1.25} />
          </div>
        {/if}
      </div>
      <div class="meta">
        <p class="meta-title">{book.title}</p>
        {#if book.author}
          <p class="meta-author">{book.author}</p>
        {/if}
        {#if audioPlayer.currentChapter}
          <p class="meta-chapter">{audioPlayer.currentChapter.title}</p>
        {/if}
      </div>
    </div>

    {#if audioPlayer.sleepEndsAt}
      <div class="sleep-banner">
        <Moon size={14} />
        <span>Sleep in {formatSleepRemaining(sleepLeftMs)}</span>
        <button type="button" class="sleep-cancel" onclick={() => audioPlayer.clearSleepTimer()}
          >Cancel</button
        >
      </div>
    {/if}
  </div>

  <div class="player-dock">
    <div class="seek-block">
      <div class="seek-track" class:seek-track--disabled={!audioPlayer.duration}>
        <div class="seek-layer seek-cache" style:width={`${audioPlayer.cachePercent}%`}></div>
        <div class="seek-layer seek-buffer" style:width={`${audioPlayer.bufferedPercent}%`}></div>
        <div class="seek-layer seek-played" style:width={`${playedPct}%`}></div>
        <input
          type="range"
          class="seek-input"
          min={0}
          max={audioPlayer.duration || 0}
          step={0.1}
          value={seekTime}
          disabled={!audioPlayer.duration}
          aria-label="Seek"
          onpointerdown={() => {
            audioPlayer.scrubbing = true;
            audioPlayer.scrubValue = audioPlayer.current;
          }}
          oninput={(e) => (audioPlayer.scrubValue = Number(e.currentTarget.value))}
          onchange={() => audioPlayer.applyScrub()}
          onkeyup={(e) => {
            if (e.key === "Enter") audioPlayer.applyScrub();
          }}
        />
      </div>
      <div class="time-row">
        <span>{formatAudioTime(seekTime)}</span>
        <span>{formatAudioTime(audioPlayer.duration)}</span>
      </div>
    </div>

    <div class="transport">
      <button
        type="button"
        class="transport-skip"
        aria-label={i18n.t("audio.rewind", { seconds: String(audioPlayer.skipSeconds) })}
        onclick={() => audioPlayer.seekBy(-audioPlayer.skipSeconds)}
        disabled={!audioPlayer.duration}
      >
        <SkipBack size={20} />
        <span class="transport-skip-label">{audioPlayer.skipSeconds}</span>
      </button>

      <button
        type="button"
        class="transport-play"
        aria-label={audioPlayer.playing ? i18n.t("audio.pause") : i18n.t("audio.play")}
        onclick={() => audioPlayer.togglePlay()}
        disabled={!audioPlayer.duration}
      >
        {#if audioPlayer.playing}
          <Pause size={26} />
        {:else}
          <Play size={26} class="play-offset" />
        {/if}
      </button>

      <button
        type="button"
        class="transport-skip"
        aria-label={i18n.t("audio.forward", { seconds: String(audioPlayer.skipSeconds) })}
        onclick={() => audioPlayer.seekBy(audioPlayer.skipSeconds)}
        disabled={!audioPlayer.duration}
      >
        <span class="transport-skip-label">{audioPlayer.skipSeconds}</span>
        <SkipForward size={20} />
      </button>
    </div>

    <div class="toolbar">
      {#if audioPlayer.playlist.length > 1}
        <Popover bind:open={tracksOpen} placement="top" align="start" minWidth={260}>
          {#snippet trigger(toggle)}
            <button
              type="button"
              class="toolbar-item"
              class:toolbar-item--active={tracksOpen}
              aria-expanded={tracksOpen}
              onclick={toggle}
            >
              <ListMusic size={15} class="toolbar-icon" />
              <span>{i18n.t("audio.tracks")}</span>
              <ChevronDown size={14} class="toolbar-chevron" />
            </button>
          {/snippet}
          <MenuList
            title={i18n.t("audio.tracks")}
            items={trackItems.map((item) => ({
              id: item.id,
              label: item.label,
              active: item.active,
              onclick: () => {
                audioPlayer.selectTrack(item.trackIndex);
                tracksOpen = false;
              },
            }))}
          />
        </Popover>
      {/if}

      {#if audioPlayer.chapters.length > 0}
        <Popover bind:open={chaptersOpen} placement="top" align="start" minWidth={240}>
          {#snippet trigger(toggle)}
            <button
              type="button"
              class="toolbar-item"
              class:toolbar-item--active={chaptersOpen}
              aria-expanded={chaptersOpen}
              onclick={toggle}
            >
              <ListMusic size={15} class="toolbar-icon" />
              <span>{i18n.t("audio.chapters")}</span>
              <ChevronDown size={14} class="toolbar-chevron" />
            </button>
          {/snippet}
          <MenuList
            title={i18n.t("audio.chapters")}
            items={chapterItems.map((item) => ({
              id: item.id,
              label: item.label,
              hint: item.hint,
              active: item.active,
              onclick: () => seekToChapter(item.chapter),
            }))}
          />
        </Popover>
      {/if}

      <Popover bind:open={speedOpen} placement="top" align="start" minWidth={120}>
        {#snippet trigger(toggle)}
          <button
            type="button"
            class="toolbar-item"
            class:toolbar-item--active={speedOpen}
            aria-expanded={speedOpen}
            onclick={toggle}
          >
            <Gauge size={15} class="toolbar-icon" />
            <span>{audioPlayer.rate}x</span>
            <ChevronDown size={14} class="toolbar-chevron" />
          </button>
        {/snippet}
        <MenuList
          title={i18n.t("audio.speed")}
          items={speedItems.map((item) => ({
            id: item.id,
            label: item.label,
            active: item.active,
            onclick: () => {
              audioPlayer.setRate(item.speed);
              speedOpen = false;
            },
          }))}
        />
      </Popover>

      <Popover bind:open={sleepOpen} placement="top" align="center" minWidth={220}>
        {#snippet trigger(toggle)}
          <button
            type="button"
            class="toolbar-item"
            class:toolbar-item--active={sleepOpen || !!audioPlayer.sleepEndsAt}
            aria-expanded={sleepOpen}
            onclick={toggle}
          >
            <Moon size={15} class="toolbar-icon" />
            <span>Sleep</span>
            <ChevronDown size={14} class="toolbar-chevron" />
          </button>
        {/snippet}
        <MenuList>
          <section class="more-section">
            <p class="more-label">Sleep timer</p>
            <div class="chip-row">
              {#each AUDIO_SLEEP_OPTIONS as min (min)}
                <button
                  type="button"
                  class="chip"
                  onclick={() => {
                    audioPlayer.setSleepTimer(min);
                    sleepOpen = false;
                  }}
                >
                  {min}m
                </button>
              {/each}
              {#if audioPlayer.sleepEndsAt}
                <button
                  type="button"
                  class="chip"
                  onclick={() => {
                    audioPlayer.clearSleepTimer();
                    sleepOpen = false;
                  }}
                >
                  <RotateCcw size={12} /> Clear
                </button>
              {/if}
            </div>
          </section>
        </MenuList>
      </Popover>

      <div class="toolbar-item toolbar-volume">
        <button
          type="button"
          class="toolbar-icon-btn"
          aria-label={audioPlayer.muted ? "Unmute" : "Mute"}
          onclick={() => audioPlayer.toggleMute()}
        >
          {#if isVolumeMutedIcon(audioPlayer.muted, audioPlayer.volume)}
            <VolumeX size={15} />
          {:else}
            <Volume2 size={15} />
          {/if}
        </button>
        <input
          type="range"
          class="volume-input"
          min={0}
          max={1}
          step={0.05}
          value={volumeSliderValue(audioPlayer.muted, audioPlayer.volume)}
          aria-label="Volume"
          oninput={(e) => audioPlayer.setVolume(Number(e.currentTarget.value))}
        />
      </div>

      <Popover bind:open={showMore} placement="top" align="end" minWidth={260}>
        {#snippet trigger(toggle)}
          <button
            type="button"
            class="toolbar-item toolbar-more"
            class:toolbar-more--open={showMore}
            aria-expanded={showMore}
            onclick={toggle}
          >
            <ChevronDown size={15} class="toolbar-chevron" />
            <span>More</span>
          </button>
        {/snippet}
        <MenuList>
          <section class="more-section">
            <p class="more-label">Skip interval</p>
            <div class="chip-row">
              {#each AUDIO_SKIP_OPTIONS as sec (sec)}
                <button
                  type="button"
                  class="chip"
                  class:chip--active={audioPlayer.skipSeconds === sec}
                  onclick={() => {
                    audioPlayer.skipSeconds = sec;
                    localStorage.setItem(SKIP_KEY, String(sec));
                  }}
                >
                  {sec}s
                </button>
              {/each}
            </div>
          </section>

          <section class="more-section">
            <p class="more-label">{i18n.t("audio.offlineCache")}</p>
            <p class="more-hint">
              {i18n.t("audio.offlineCacheHint", {
                pct: String(Math.round(audioPlayer.cachePercent)),
              })}
            </p>
            <div class="chip-row">
              <button
                type="button"
                class="chip"
                onclick={() => audioPlayer.startTrackPrefetchPublic()}
              >
                <Download size={14} />
                {audioPlayer.cacheStatus.downloading
                  ? i18n.t("audio.downloading")
                  : i18n.t("audio.download")}
              </button>
              <button
                type="button"
                class="chip chip--danger"
                onclick={() =>
                  void audioCache.clearBook(
                    book.id,
                    cacheBookTrackCount(audioPlayer.playlist.length),
                  )}
              >
                <Trash2 size={14} />
                {i18n.t("audio.clearCache")}
              </button>
            </div>
          </section>
        </MenuList>
      </Popover>
    </div>
  </div>
</div>

<style src="./AudioReader.css"></style>
