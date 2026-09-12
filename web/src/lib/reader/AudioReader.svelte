<script lang="ts">
  import {
    ChevronsLeft,
    ChevronsRight,
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
  import Dropdown from "$lib/components/Dropdown.svelte";
  import Slider from "$lib/components/Slider.svelte";
  import type { MenuItem } from "$lib/components/menu";
  import { PersistedState } from "runed";
  import "$lib/reader/AudioReader.css";
  import type { Book } from "$lib/api/types";
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
  const skipPref = new PersistedState<number>(SKIP_KEY, audioPlayer.skipSeconds);

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
    audioCachePill(audioPlayer.usingOffline, audioPlayer.cacheStatus.complete, audioPlayer.online),
  );
  let multiTrack = $derived(audioPlayer.playlist.length > 1);
  let trackMenuItems = $derived<MenuItem[]>(
    buildTrackMenuItems(audioPlayer.playlist, audioPlayer.trackIndex, (n) =>
      i18n.t("audio.trackN", { n: String(n) }),
    ).map((item) => ({
      id: item.id,
      label: item.label,
      active: item.active,
      onclick: () => audioPlayer.selectTrack(item.trackIndex),
    })),
  );
  let chapterMenuItems = $derived<MenuItem[]>(
    buildChapterMenuItems(
      audioPlayer.chapters,
      audioPlayer.currentChapter?.index,
      formatAudioTime,
    ).map((item) => ({
      id: item.id,
      label: item.label,
      hint: item.hint,
      active: item.active,
      onclick: () => audioPlayer.seekToChapter(item.chapter),
    })),
  );
  let speedMenuItems = $derived<MenuItem[]>(
    buildSpeedMenuItems(AUDIO_SPEEDS, audioPlayer.rate).map((item) => ({
      id: item.id,
      label: item.label,
      active: item.active,
      onclick: () => audioPlayer.setRate(item.speed),
    })),
  );

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
            <Headphones size={48} strokeWidth={1.25} />
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
        {:else if multiTrack}
          <p class="meta-chapter">
            {audioPlayer.playlist[audioPlayer.trackIndex]?.title ||
              i18n.t("audio.trackN", { n: String(audioPlayer.trackIndex + 1) })}
          </p>
        {/if}
      </div>
    </div>

    {#if audioPlayer.sleepEndsAt}
      <div class="sleep-banner">
        <Moon size={14} />
        <span>{i18n.t("audio.sleepIn", { time: formatSleepRemaining(sleepLeftMs) })}</span>
        <button type="button" class="sleep-cancel" onclick={() => audioPlayer.clearSleepTimer()}
          >{i18n.t("audio.clearSleep")}</button
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
        <Slider
          class="seek-slider"
          value={seekTime}
          min={0}
          max={audioPlayer.duration || 1}
          step={0.1}
          disabled={!audioPlayer.duration}
          ariaLabel={i18n.t("audio.seek")}
          onchange={(v) => {
            if (!audioPlayer.scrubbing) audioPlayer.scrubbing = true;
            audioPlayer.scrubValue = v;
          }}
          oncommit={(v) => {
            audioPlayer.scrubValue = v;
            audioPlayer.applyScrub();
          }}
        />
      </div>
      <div class="time-row">
        <span>{formatAudioTime(seekTime)}</span>
        <span>{formatAudioTime(audioPlayer.duration)}</span>
      </div>
    </div>

    <div class="transport">
      {#if multiTrack}
        <button
          type="button"
          class="transport-edge"
          aria-label={i18n.t("audio.prevTrack")}
          onclick={() => audioPlayer.selectTrack(audioPlayer.trackIndex - 1)}
          disabled={audioPlayer.trackIndex <= 0}
        >
          <ChevronsLeft size={22} />
        </button>
      {/if}

      <button
        type="button"
        class="transport-skip"
        aria-label={i18n.t("audio.rewind", { seconds: String(audioPlayer.skipSeconds) })}
        onclick={() => audioPlayer.seekBy(-audioPlayer.skipSeconds)}
        disabled={!audioPlayer.duration}
      >
        <SkipBack size={22} />
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
          <Pause size={28} />
        {:else}
          <Play size={28} class="play-offset" />
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
        <SkipForward size={22} />
      </button>

      {#if multiTrack}
        <button
          type="button"
          class="transport-edge"
          aria-label={i18n.t("audio.nextTrack")}
          onclick={() => audioPlayer.selectTrack(audioPlayer.trackIndex + 1)}
          disabled={audioPlayer.trackIndex >= audioPlayer.playlist.length - 1}
        >
          <ChevronsRight size={22} />
        </button>
      {/if}
    </div>

    <div class="toolbar">
      {#if multiTrack}
        <Dropdown
          bind:open={tracksOpen}
          side="top"
          align="center"
          minWidth={260}
          title={i18n.t("audio.tracks")}
          items={trackMenuItems}
        >
          {#snippet trigger(props)}
            <button
              type="button"
              class="toolbar-item"
              class:toolbar-item--active={tracksOpen}
              {...props}
            >
              <ListMusic size={15} class="toolbar-icon" />
              <span>{i18n.t("audio.tracks")}</span>
            </button>
          {/snippet}
        </Dropdown>
      {/if}

      {#if audioPlayer.chapters.length > 0}
        <Dropdown
          bind:open={chaptersOpen}
          side="top"
          align="center"
          minWidth={240}
          title={i18n.t("audio.chapters")}
          items={chapterMenuItems}
        >
          {#snippet trigger(props)}
            <button
              type="button"
              class="toolbar-item"
              class:toolbar-item--active={chaptersOpen}
              {...props}
            >
              <ListMusic size={15} class="toolbar-icon" />
              <span>{i18n.t("audio.chapters")}</span>
            </button>
          {/snippet}
        </Dropdown>
      {/if}

      <Dropdown
        bind:open={speedOpen}
        side="top"
        align="center"
        minWidth={120}
        title={i18n.t("audio.speedTitle")}
        items={speedMenuItems}
      >
        {#snippet trigger(props)}
          <button
            type="button"
            class="toolbar-item"
            class:toolbar-item--active={speedOpen}
            {...props}
          >
            <Gauge size={15} class="toolbar-icon" />
            <span>{audioPlayer.rate}x</span>
          </button>
        {/snippet}
      </Dropdown>

      <Popover bind:open={sleepOpen} placement="top" align="center" minWidth={220}>
        {#snippet trigger(props)}
          <button
            type="button"
            class="toolbar-item"
            class:toolbar-item--active={sleepOpen || !!audioPlayer.sleepEndsAt}
            {...props}
          >
            <Moon size={15} class="toolbar-icon" />
            <span>{i18n.t("audio.sleepTimer")}</span>
          </button>
        {/snippet}
        <MenuList>
          <section class="more-section">
            <p class="more-label">{i18n.t("audio.sleepTimer")}</p>
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
                  <RotateCcw size={12} />
                  {i18n.t("audio.clearSleep")}
                </button>
              {/if}
            </div>
          </section>
        </MenuList>
      </Popover>

      <div class="toolbar-item toolbar-volume" title={i18n.t("audio.volume")}>
        <button
          type="button"
          class="toolbar-icon-btn"
          aria-label={audioPlayer.muted ? i18n.t("audio.unmute") : i18n.t("audio.mute")}
          onclick={() => audioPlayer.toggleMute()}
        >
          {#if isVolumeMutedIcon(audioPlayer.muted, audioPlayer.volume)}
            <VolumeX size={15} />
          {:else}
            <Volume2 size={15} />
          {/if}
        </button>
        <Slider
          class="volume-slider"
          value={volumeSliderValue(audioPlayer.muted, audioPlayer.volume)}
          min={0}
          max={1}
          step={0.05}
          ariaLabel={i18n.t("audio.volume")}
          onchange={(v) => audioPlayer.setVolume(v)}
        />
      </div>

      <Popover bind:open={showMore} placement="top" align="center" minWidth={260}>
        {#snippet trigger(props)}
          <button
            type="button"
            class="toolbar-item"
            class:toolbar-item--active={showMore}
            {...props}
          >
            <span>{i18n.t("audio.more")}</span>
          </button>
        {/snippet}
        <MenuList>
          <section class="more-section">
            <p class="more-label">{i18n.t("audio.skipInterval")}</p>
            <div class="chip-row">
              {#each AUDIO_SKIP_OPTIONS as sec (sec)}
                <button
                  type="button"
                  class="chip"
                  class:chip--active={audioPlayer.skipSeconds === sec}
                  onclick={() => {
                    audioPlayer.skipSeconds = sec;
                    skipPref.current = sec;
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

<style>
  /* Seek bar keeps its layered cache/buffer/played track; the slider only
     supplies the hit area and thumb. */
  :global(.slider.seek-slider) {
    z-index: 1;
    height: 1.35rem;
  }

  :global(.slider.seek-slider::before) {
    height: 5px;
    background: transparent;
  }

  :global(.slider.seek-slider .slider-range) {
    height: 5px;
    background: transparent;
  }

  :global(.slider.seek-slider .slider-thumb) {
    width: 15px;
    height: 15px;
    border-color: var(--color-primary-fg);
    background: var(--color-primary);
    box-shadow: 0 1px 4px rgb(0 0 0 / 0.35);
  }

  :global(.slider.volume-slider) {
    flex: 1;
    min-width: 3.5rem;
    height: 1rem;
  }

  :global(.slider.volume-slider::before) {
    height: 3px;
  }

  :global(.slider.volume-slider .slider-range) {
    height: 3px;
    background: transparent;
  }

  :global(.slider.volume-slider .slider-thumb) {
    width: 10px;
    height: 10px;
    border: 0;
    background: var(--color-fg);
    box-shadow: none;
  }
</style>
