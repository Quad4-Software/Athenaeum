<script lang="ts">
  import { Gauge, Pause, Play, SkipForward, Volume2, X } from "@lucide/svelte";
  import MenuList from "$lib/components/MenuList.svelte";
  import Popover from "$lib/components/Popover.svelte";
  import { narrator } from "$lib/stores/narrator.svelte";
  import { i18n } from "$lib/stores/i18n.svelte";
  import { toast } from "$lib/stores/toast.svelte";
  import type { NarratorProvider } from "$lib/narrator/types";

  let speedOpen = $state(false);
  let voiceOpen = $state(false);
  let providerOpen = $state(false);

  let showBar = $derived(narrator.showBar);

  $effect(() => {
    const code = narrator.error;
    if (!code) return;
    const key =
      code === "unavailable"
        ? "narrator.errUnavailable"
        : code === "empty"
          ? "narrator.errEmpty"
          : code === "kokoro_unavailable"
            ? "narrator.errKokoro"
            : "narrator.errSpeak";
    toast.error(i18n.t(key));
    narrator.error = null;
  });

  $effect(() => {
    if (!showBar) return;
    void narrator.refreshStatus().then(() => narrator.loadVoices());
  });
</script>

{#if showBar}
  <div
    class="narrator-mini fixed inset-x-0 z-50 border-t border-border bg-bg-elevated/95 backdrop-blur-md"
    style:bottom="calc(var(--nav-bar-height, 0px) + env(safe-area-inset-bottom, 0px))"
    role="region"
    aria-label={i18n.t("narrator.player")}
  >
    <div class="main-row">
      <div class="meta">
        <Volume2 size={16} class="shrink-0 text-muted" aria-hidden="true" />
        <div class="min-w-0">
          <p class="truncate text-sm font-medium text-fg">
            {narrator.bookTitle || i18n.t("narrator.player")}
          </p>
          <p class="truncate text-xs text-muted">
            {#if narrator.kokoroLoading}
              {i18n.t("narrator.kokoroLoading")}
            {:else}
              {narrator.progressLabel}
              {#if narrator.currentText}
                <span class="text-fg-subtle"> · {narrator.currentText}</span>
              {/if}
            {/if}
          </p>
          {#if narrator.total > 0}
            <div
              class="mt-1 h-1 w-full max-w-xs overflow-hidden rounded-full bg-border/50"
              role="progressbar"
              aria-valuemin={0}
              aria-valuemax={narrator.total}
              aria-valuenow={Math.min(narrator.index + 1, narrator.total)}
              aria-label={narrator.progressLabel}
            >
              <div
                class="h-full rounded-full bg-accent transition-[width] duration-200"
                style:width="{Math.round(narrator.progress * 100)}%"
              ></div>
            </div>
          {/if}
        </div>
      </div>

      <div class="controls">
        <button
          type="button"
          class="btn btn-ghost"
          aria-label={narrator.playing && !narrator.paused
            ? i18n.t("narrator.pause")
            : i18n.t("narrator.play")}
          onclick={() => narrator.togglePlay()}
        >
          {#if narrator.playing && !narrator.paused}
            <Pause size={18} />
          {:else}
            <Play size={18} />
          {/if}
        </button>
        <button
          type="button"
          class="btn btn-ghost"
          aria-label={i18n.t("narrator.skip")}
          onclick={() => narrator.skip()}
        >
          <SkipForward size={18} />
        </button>

        <Popover bind:open={speedOpen} placement="top" align="end" minWidth={140}>
          {#snippet trigger(toggle)}
            <button
              type="button"
              class="btn btn-ghost text-xs tabular-nums"
              aria-label={i18n.t("narrator.speed", { rate: narrator.rate })}
              aria-expanded={speedOpen}
              onclick={toggle}
            >
              <Gauge size={16} />
              {narrator.rate}x
            </button>
          {/snippet}
          <MenuList
            items={narrator.speeds.map((r) => ({
              id: String(r),
              label: `${r}x`,
              active: r === narrator.rate,
              onclick: () => {
                narrator.setRate(r);
                speedOpen = false;
              },
            }))}
          />
        </Popover>

        <Popover bind:open={voiceOpen} placement="top" align="end" minWidth={220}>
          {#snippet trigger(toggle)}
            <button
              type="button"
              class="btn btn-ghost text-xs"
              aria-label={i18n.t("narrator.voice")}
              aria-expanded={voiceOpen}
              onclick={() => {
                void narrator.loadVoices();
                toggle();
              }}
            >
              {i18n.t("narrator.voice")}
            </button>
          {/snippet}
          {#if narrator.voicesLoading}
            <p class="px-3 py-2 text-xs text-muted">{i18n.t("common.loading")}</p>
          {:else if !narrator.voices.length}
            <p class="px-3 py-2 text-xs text-muted">{i18n.t("narrator.noVoices")}</p>
          {:else}
            <MenuList
              items={narrator.voices.map((v) => ({
                id: v.id,
                label: v.label,
                active: v.id === narrator.voiceId,
                onclick: () => {
                  narrator.setVoice(v.id);
                  voiceOpen = false;
                },
              }))}
            />
          {/if}
        </Popover>

        <Popover bind:open={providerOpen} placement="top" align="end" minWidth={180}>
          {#snippet trigger(toggle)}
            <button
              type="button"
              class="btn btn-ghost text-xs"
              aria-label={i18n.t("narrator.provider")}
              aria-expanded={providerOpen}
              onclick={toggle}
            >
              {narrator.provider === "kokoro"
                ? i18n.t("narrator.providerKokoro")
                : i18n.t("narrator.providerBrowser")}
            </button>
          {/snippet}
          <MenuList
            items={[
              {
                id: "browser",
                label: i18n.t("narrator.providerBrowser"),
                active: narrator.provider === "browser",
                onclick: () => {
                  narrator.setProvider("browser" as NarratorProvider);
                  providerOpen = false;
                },
              },
              ...(narrator.kokoroEnabled
                ? [
                    {
                      id: "kokoro",
                      label: i18n.t("narrator.providerKokoro"),
                      active: narrator.provider === "kokoro",
                      onclick: () => {
                        narrator.setProvider("kokoro" as NarratorProvider);
                        providerOpen = false;
                      },
                    },
                  ]
                : []),
            ]}
          />
        </Popover>

        <button
          type="button"
          class="btn btn-ghost"
          aria-label={i18n.t("narrator.stop")}
          onclick={() => narrator.stop()}
        >
          <X size={18} />
        </button>
      </div>
    </div>
  </div>
{/if}

<style>
  .narrator-mini {
    --mini-h: 4.5rem;
  }

  .main-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 0.5rem;
    min-height: var(--mini-h);
    padding: 0.5rem 0.75rem;
    max-width: 72rem;
    margin-inline: auto;
  }

  .meta {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    min-width: 0;
    flex: 1;
  }

  .controls {
    display: flex;
    align-items: center;
    gap: 0.15rem;
    flex-shrink: 0;
  }

  @media (max-width: 640px) {
    .main-row {
      flex-wrap: wrap;
    }

    .controls {
      width: 100%;
      justify-content: flex-end;
    }
  }
</style>
