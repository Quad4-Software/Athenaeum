<script lang="ts">
  import { Gauge, Headphones, Moon, Pause, Play, SkipBack, SkipForward, X } from "@lucide/svelte";
  import MenuList from "$lib/components/MenuList.svelte";
  import Popover from "$lib/components/Popover.svelte";
  import { audioPlayer } from "$lib/stores/audioPlayer.svelte";
  import { formatAudioTime, formatSleepRemaining } from "$lib/audio/format";
  import { router } from "$lib/router.svelte";
  import { i18n } from "$lib/stores/i18n.svelte";

  const MINI_SKIP = 10;
  const SLEEP_OPTIONS = [5, 15, 30, 45, 60, 90];
  const SPEEDS = [0.75, 1, 1.25, 1.5, 1.75, 2, 2.5, 3];

  let audioEl = $state<HTMLAudioElement | null>(null);
  let sleepOpen = $state(false);
  let speedOpen = $state(false);

  let showBar = $derived(audioPlayer.active && !audioPlayer.expanded);

  let sleepRemainingMs = $derived(
    audioPlayer.sleepEndsAt
      ? Math.max(0, audioPlayer.sleepEndsAt - (audioPlayer.sleepTick || Date.now()))
      : 0,
  );

  let playbackTime = $derived(audioPlayer.scrubbing ? audioPlayer.scrubValue : audioPlayer.current);

  $effect(() => {
    audioPlayer.bindAudio(audioEl);
    return () => audioPlayer.bindAudio(null);
  });

  $effect(() => {
    const bookId = audioPlayer.book?.id;
    if (!audioPlayer.active || bookId == null) return;
    const cleanup = audioPlayer.beginLifecycle();
    return cleanup;
  });

  $effect(() => {
    const bookId = audioPlayer.book?.id;
    const url = audioPlayer.playbackUrl;
    if (!audioPlayer.active || bookId == null || !url) return;
    audioPlayer.onStreamUrlChange();
  });

  function openFull() {
    const book = audioPlayer.book;
    if (!book) return;
    router.navigate(`/read/${book.id}`);
  }

  function stop() {
    audioPlayer.stop();
  }
</script>

<audio
  bind:this={audioEl}
  preload="auto"
  class="sr-only"
  onloadedmetadata={() => audioPlayer.onLoaded()}
  onplay={() => audioPlayer.onPlay()}
  onpause={() => audioPlayer.onPause()}
  ontimeupdate={() => audioPlayer.onTimeUpdate()}
  onended={() => audioPlayer.onEnded()}
  onstalled={() => audioPlayer.onStalled()}
  onerror={() => audioPlayer.onError()}
></audio>

{#if showBar && audioPlayer.book}
  {@const book = audioPlayer.book}
  <div
    class="audio-mini fixed inset-x-0 z-50 border-t border-border bg-bg-elevated/95 backdrop-blur-md"
    style:bottom="calc(var(--nav-bar-height, 0px) + env(safe-area-inset-bottom, 0px))"
    role="region"
    aria-label={i18n.t("audio.player")}
  >
    <div class="main-row">
      <button type="button" class="meta-btn" onclick={openFull}>
        <div class="cover">
          {#if audioPlayer.coverSrc}
            <img
              src={audioPlayer.coverSrc}
              alt=""
              class="cover-img"
              onerror={() => (audioPlayer.coverFailed = true)}
            />
          {:else}
            <div class="cover-fallback">
              <Headphones size={14} />
            </div>
          {/if}
        </div>
        <div class="meta-text">
          <span class="meta-title">{book.title}</span>
          {#if audioPlayer.currentChapter}
            <span class="meta-sub">{audioPlayer.currentChapter.title}</span>
          {:else if book.author}
            <span class="meta-sub">{book.author}</span>
          {/if}
        </div>
      </button>

      <div class="transport">
        <button
          type="button"
          class="skip-btn"
          aria-label={i18n.t("audio.rewind", { seconds: MINI_SKIP })}
          onclick={() => audioPlayer.seekBy(-MINI_SKIP)}
          disabled={!audioPlayer.duration}
        >
          <SkipBack size={16} />
        </button>

        <button
          type="button"
          class="play-btn"
          aria-label={audioPlayer.playing ? i18n.t("audio.pause") : i18n.t("audio.play")}
          onclick={() => audioPlayer.togglePlay()}
        >
          {#if audioPlayer.playing}
            <Pause size={20} />
          {:else}
            <Play size={20} class="play-offset" />
          {/if}
        </button>

        <button
          type="button"
          class="skip-btn"
          aria-label={i18n.t("audio.forward", { seconds: MINI_SKIP })}
          onclick={() => audioPlayer.seekBy(MINI_SKIP)}
          disabled={!audioPlayer.duration}
        >
          <SkipForward size={16} />
        </button>
      </div>

      <div class="tools">
        <Popover bind:open={speedOpen} placement="top" align="end" minWidth={120}>
          {#snippet trigger(toggle)}
            <button
              type="button"
              class="tool-btn"
              class:tool-btn--active={speedOpen}
              aria-label={i18n.t("audio.speed", { rate: audioPlayer.rate })}
              aria-expanded={speedOpen}
              onclick={toggle}
            >
              <Gauge size={15} />
              <span class="tool-rate">{audioPlayer.rate}x</span>
            </button>
          {/snippet}
          <MenuList
            title={i18n.t("audio.speedTitle")}
            items={SPEEDS.map((speed) => ({
              id: String(speed),
              label: `${speed}x`,
              active: audioPlayer.rate === speed,
              onclick: () => {
                audioPlayer.setRate(speed);
                speedOpen = false;
              },
            }))}
          />
        </Popover>

        <Popover bind:open={sleepOpen} placement="top" align="end" minWidth={220}>
          {#snippet trigger(toggle)}
            <button
              type="button"
              class="tool-btn"
              class:tool-btn--active={sleepOpen || !!audioPlayer.sleepEndsAt}
              aria-label={audioPlayer.sleepEndsAt
                ? i18n.t("audio.sleepIn", { time: formatSleepRemaining(sleepRemainingMs) })
                : i18n.t("audio.sleepTimer")}
              aria-expanded={sleepOpen}
              onclick={toggle}
            >
              <Moon size={15} />
              {#if audioPlayer.sleepEndsAt}
                <span class="tool-rate">{formatSleepRemaining(sleepRemainingMs)}</span>
              {/if}
            </button>
          {/snippet}
          <MenuList>
            <section class="sleep-menu">
              <p class="sleep-menu-label">{i18n.t("audio.sleepTimer")}</p>
              <div class="chip-row">
                {#each SLEEP_OPTIONS as min (min)}
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
                    class="chip chip--muted"
                    onclick={() => {
                      audioPlayer.clearSleepTimer();
                      sleepOpen = false;
                    }}
                  >
                    {i18n.t("audio.clearSleep")}
                  </button>
                {/if}
              </div>
            </section>
          </MenuList>
        </Popover>

        <button
          type="button"
          class="tool-btn tool-btn--close"
          aria-label={i18n.t("audio.stop")}
          onclick={stop}
        >
          <X size={15} />
        </button>
      </div>
    </div>

    <div class="seek-row">
      <span class="time">{formatAudioTime(playbackTime)}</span>
      <div class="seek-track" class:seek-track--disabled={!audioPlayer.duration}>
        <div class="seek-rail">
          <div class="seek-layer seek-buffer" style:width={`${audioPlayer.bufferedPercent}%`}></div>
          <div
            class="seek-layer seek-played"
            style:width={`${audioPlayer.duration > 0 ? (playbackTime / audioPlayer.duration) * 100 : 0}%`}
          ></div>
        </div>
        <input
          type="range"
          class="seek-input"
          min={0}
          max={audioPlayer.duration || 0}
          step={0.1}
          value={playbackTime}
          disabled={!audioPlayer.duration}
          aria-label={i18n.t("audio.seek")}
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
      <span class="time time-end">{formatAudioTime(audioPlayer.duration)}</span>
    </div>
  </div>
{/if}

<style>
  .sr-only {
    position: absolute;
    width: 1px;
    height: 1px;
    padding: 0;
    margin: -1px;
    overflow: hidden;
    clip: rect(0, 0, 0, 0);
    white-space: nowrap;
    border: 0;
  }

  .audio-mini {
    box-shadow: 0 -4px 20px -8px rgb(0 0 0 / 0.35);
  }

  .main-row {
    display: grid;
    grid-template-columns: minmax(0, 1fr) auto minmax(0, 1fr);
    align-items: center;
    gap: 0.5rem;
    padding: 0.4rem 0.625rem 0.15rem;
  }

  .meta-btn {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    min-width: 0;
    justify-self: start;
    padding: 0;
    border: 0;
    background: none;
    text-align: left;
    cursor: pointer;
    color: inherit;
  }

  .cover {
    width: 2.25rem;
    height: 2.25rem;
    flex-shrink: 0;
    overflow: hidden;
    border-radius: var(--radius-sm);
    background: var(--color-surface);
  }

  .cover-img {
    width: 100%;
    height: 100%;
    object-fit: cover;
  }

  .cover-fallback {
    display: grid;
    height: 100%;
    place-items: center;
    color: var(--color-muted);
  }

  .meta-text {
    min-width: 0;
    flex: 1;
    display: flex;
    flex-direction: column;
    gap: 0.0625rem;
  }

  .meta-title {
    font-size: 0.8125rem;
    font-weight: 600;
    color: var(--color-fg);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .meta-sub {
    font-size: 0.6875rem;
    color: var(--color-muted);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .transport {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 0.35rem;
    justify-self: center;
  }

  .tools {
    display: flex;
    align-items: center;
    justify-content: flex-end;
    gap: 0.125rem;
    justify-self: end;
  }

  .skip-btn {
    display: grid;
    place-items: center;
    width: 2.25rem;
    height: 2.25rem;
    padding: 0;
    border: 0;
    border-radius: 9999px;
    color: var(--color-fg);
    background: var(--color-surface);
    box-shadow: inset 0 0 0 1px var(--color-border);
    cursor: pointer;
  }

  .skip-btn:disabled {
    opacity: 0.35;
    cursor: default;
  }

  .play-btn {
    display: grid;
    place-items: center;
    width: 2.5rem;
    height: 2.5rem;
    border: 0;
    border-radius: 9999px;
    color: var(--color-primary-fg);
    background: var(--color-primary);
    box-shadow: 0 4px 14px -4px color-mix(in oklch, var(--color-primary) 55%, transparent);
    cursor: pointer;
  }

  .play-btn :global(.play-offset) {
    margin-left: 2px;
  }

  .tool-btn {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    gap: 0.125rem;
    min-width: 2rem;
    height: 2rem;
    padding: 0 0.3rem;
    border: 0;
    border-radius: 9999px;
    background: none;
    font-size: 0.625rem;
    font-weight: 600;
    color: var(--color-muted);
    cursor: pointer;
  }

  .tool-btn:hover {
    color: var(--color-fg);
    background: var(--color-surface);
  }

  .tool-btn--active {
    color: var(--color-primary);
    background: color-mix(in oklch, var(--color-primary) 10%, var(--color-surface));
  }

  .tool-btn--close {
    min-width: 2rem;
    padding: 0;
  }

  .tool-rate {
    display: none;
  }

  @media (min-width: 480px) {
    .tool-rate {
      display: inline;
    }
  }

  .seek-row {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    padding: 0.125rem 0.75rem 0.4375rem;
  }

  .time {
    flex-shrink: 0;
    width: 2.5rem;
    font-size: 0.625rem;
    font-variant-numeric: tabular-nums;
    color: var(--color-subtle);
    line-height: 1;
  }

  .time-end {
    text-align: right;
  }

  .seek-track {
    position: relative;
    flex: 1;
    min-width: 0;
    height: 1.25rem;
  }

  .seek-track--disabled {
    opacity: 0.45;
  }

  .seek-rail {
    position: absolute;
    left: 0;
    right: 0;
    top: 50%;
    height: 4px;
    transform: translateY(-50%);
    border-radius: 9999px;
    background: var(--color-surface);
    box-shadow: inset 0 0 0 1px var(--color-border);
    overflow: hidden;
  }

  .seek-layer {
    position: absolute;
    left: 0;
    top: 0;
    height: 100%;
    border-radius: 9999px;
    pointer-events: none;
  }

  .seek-buffer {
    background: color-mix(in oklch, var(--color-muted) 35%, transparent);
  }

  .seek-played {
    background: var(--color-primary);
  }

  .seek-input {
    position: absolute;
    inset: 0;
    z-index: 1;
    width: 100%;
    height: 100%;
    margin: 0;
    appearance: none;
    background: transparent;
    cursor: pointer;
  }

  .seek-input:disabled {
    cursor: default;
  }

  .seek-input::-webkit-slider-runnable-track {
    height: 4px;
    background: transparent;
  }

  .seek-input::-webkit-slider-thumb {
    appearance: none;
    width: 14px;
    height: 14px;
    margin-top: -5px;
    border-radius: 50%;
    border: 2px solid var(--color-primary-fg);
    background: var(--color-primary);
    box-shadow: 0 1px 4px rgb(0 0 0 / 0.3);
  }

  .seek-input::-moz-range-track {
    height: 4px;
    background: transparent;
    border: 0;
  }

  .seek-input::-moz-range-thumb {
    width: 14px;
    height: 14px;
    border-radius: 50%;
    border: 2px solid var(--color-primary-fg);
    background: var(--color-primary);
    box-shadow: 0 1px 4px rgb(0 0 0 / 0.3);
  }

  .sleep-menu {
    padding: 0.125rem 0;
  }

  .sleep-menu-label {
    margin: 0 0 0.5rem;
    font-size: 0.6875rem;
    font-weight: 600;
    letter-spacing: 0.04em;
    text-transform: uppercase;
    color: var(--color-subtle);
  }

  .chip-row {
    display: flex;
    flex-wrap: wrap;
    gap: 0.375rem;
  }

  .chip {
    border: 0;
    border-radius: 9999px;
    padding: 0.375rem 0.625rem;
    font-size: 0.75rem;
    font-weight: 500;
    color: var(--color-fg);
    background: var(--color-bg-elevated);
    box-shadow: inset 0 0 0 1px var(--color-border);
    cursor: pointer;
  }

  .chip--muted {
    color: var(--color-muted);
  }
</style>
