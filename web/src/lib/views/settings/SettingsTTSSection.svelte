<script lang="ts">
  import { AudioLines, RefreshCw } from "@lucide/svelte";
  import Button from "$lib/components/Button.svelte";
  import Skeleton from "$lib/components/Skeleton.svelte";
  import EmptyState from "$lib/components/EmptyState.svelte";
  import { auth } from "$lib/stores/auth.svelte";
  import { narrator } from "$lib/stores/narrator.svelte";
  import { toast } from "$lib/stores/toast.svelte";
  import { confirmDialog } from "$lib/stores/confirm.svelte";
  import { i18n } from "$lib/stores/i18n.svelte";
  import { ApiError } from "$lib/api/client";
  import { router } from "$lib/router.svelte";
  import { routes } from "$lib/routes";
  import {
    cancelTTSJob,
    deleteTTSJob,
    fetchTTSJobs,
    fetchTTSPrefs,
    fetchTTSVoices,
    retryTTSJob,
    saveTTSPrefs,
  } from "$lib/narrator/kokoro";
  import type { NarratorVoice, TTSJob } from "$lib/narrator/types";
  import { untrack } from "svelte";

  const SPEEDS = [0.75, 1, 1.25, 1.5, 1.75, 2];
  const POLL_MS = 15000;

  let loading = $state(true);
  let saving = $state(false);
  let voices = $state<NarratorVoice[]>([]);
  let voice = $state("");
  let speed = $state(1);
  let schedEnabled = $state(false);
  let schedStart = $state("");
  let schedEnd = $state("");

  let jobs = $state<TTSJob[]>([]);
  let jobsLoading = $state(false);
  let pollTimer: ReturnType<typeof setTimeout> | null = null;

  $effect(() => {
    if (!auth.user) return;
    untrack(() => {
      void load();
    });
    return () => {
      if (pollTimer) clearTimeout(pollTimer);
    };
  });

  async function load() {
    loading = true;
    await narrator.refreshStatus();
    if (!narrator.serverEnabled) {
      loading = false;
      return;
    }
    try {
      const [prefs, v] = await Promise.all([fetchTTSPrefs(), fetchTTSVoices()]);
      voices = v;
      voice = prefs.voice || v[0]?.id || "";
      speed = prefs.speed || 1;
      schedEnabled = prefs.schedEnabled;
      schedStart = prefs.schedStart;
      schedEnd = prefs.schedEnd;
      await refreshJobs();
    } catch (e) {
      toast.error(e instanceof ApiError ? e.message : i18n.t("settings.ttsLoadFailed"));
    } finally {
      loading = false;
    }
  }

  async function refreshJobs() {
    jobsLoading = jobs.length === 0;
    try {
      jobs = await fetchTTSJobs();
    } catch {
      // Keep the last list on transient failures.
    } finally {
      jobsLoading = false;
    }
    if (pollTimer) clearTimeout(pollTimer);
    if (jobs.some((j) => j.status === "queued" || j.status === "running")) {
      pollTimer = setTimeout(() => void refreshJobs(), POLL_MS);
    }
  }

  async function save(event: Event) {
    event.preventDefault();
    saving = true;
    try {
      await saveTTSPrefs({
        voice,
        speed,
        schedEnabled,
        schedStart: schedEnabled ? schedStart : "",
        schedEnd: schedEnabled ? schedEnd : "",
      });
      toast.success(i18n.t("settings.ttsSaved"));
    } catch (e) {
      toast.error(e instanceof ApiError ? e.message : i18n.t("settings.ttsSaveFailed"));
    } finally {
      saving = false;
    }
  }

  async function cancelJob(job: TTSJob) {
    const ok = await confirmDialog.ask({
      title: i18n.t("settings.ttsJobCancelTitle"),
      message: i18n.t("settings.ttsJobCancelBody", { title: job.bookTitle || `#${job.bookId}` }),
      confirmLabel: i18n.t("settings.ttsJobCancel"),
      cancelLabel: i18n.t("confirm.cancel"),
      danger: true,
    });
    if (!ok) return;
    try {
      await cancelTTSJob(job.id);
      await refreshJobs();
    } catch (e) {
      toast.error(e instanceof ApiError ? e.message : i18n.t("settings.ttsJobActionFailed"));
    }
  }

  async function retryJob(job: TTSJob) {
    try {
      await retryTTSJob(job.id);
      await refreshJobs();
    } catch (e) {
      toast.error(e instanceof ApiError ? e.message : i18n.t("settings.ttsJobActionFailed"));
    }
  }

  async function removeJob(job: TTSJob) {
    try {
      await deleteTTSJob(job.id);
      await refreshJobs();
    } catch (e) {
      toast.error(e instanceof ApiError ? e.message : i18n.t("settings.ttsJobActionFailed"));
    }
  }

  function jobStatusLabel(job: TTSJob): string {
    return i18n.t(`settings.ttsJobStatus.${job.status}`);
  }

  function jobProgressLabel(job: TTSJob): string {
    if (job.status !== "running" || job.totalChapters === 0) return "";
    return i18n.t("settings.ttsJobProgress", {
      done: job.doneChapters,
      total: job.totalChapters,
    });
  }
</script>

{#if narrator.serverEnabled}
  <div class="rounded-[var(--radius-card)] border border-border bg-surface p-5">
    <p class="text-sm font-medium text-fg">{i18n.t("settings.ttsPrefsTitle")}</p>
    <p class="mt-1 text-xs text-muted">{i18n.t("settings.ttsPrefsHint")}</p>
    {#if loading}
      <div class="mt-3">
        <Skeleton height="6rem" rounded="lg" />
      </div>
    {:else}
      <form class="mt-3 space-y-3" onsubmit={save}>
        <div class="grid gap-3 sm:grid-cols-2">
          <label class="block text-sm text-fg">
            <span class="mb-1 block text-xs text-muted">{i18n.t("narrator.voice")}</span>
            <select class="field-input" bind:value={voice}>
              {#each voices as v (v.id)}
                <option value={v.id}>{v.label || v.id}</option>
              {/each}
            </select>
          </label>
          <label class="block text-sm text-fg">
            <span class="mb-1 block text-xs text-muted">{i18n.t("settings.ttsSpeed")}</span>
            <select class="field-input" bind:value={speed}>
              {#each SPEEDS as r (r)}
                <option value={r}>{r}x</option>
              {/each}
            </select>
          </label>
        </div>
        <label class="flex items-center gap-2 text-sm text-fg">
          <input type="checkbox" bind:checked={schedEnabled} />
          {i18n.t("settings.ttsSchedule")}
        </label>
        {#if schedEnabled}
          <div class="grid gap-3 sm:grid-cols-2">
            <label class="block text-sm text-fg">
              <span class="mb-1 block text-xs text-muted"
                >{i18n.t("settings.ttsScheduleStart")}</span
              >
              <input type="time" class="field-input" bind:value={schedStart} required />
            </label>
            <label class="block text-sm text-fg">
              <span class="mb-1 block text-xs text-muted">{i18n.t("settings.ttsScheduleEnd")}</span>
              <input type="time" class="field-input" bind:value={schedEnd} required />
            </label>
          </div>
        {/if}
        <div class="pt-1">
          <Button type="submit" size="sm" loading={saving}>{i18n.t("settings.ttsSave")}</Button>
        </div>
      </form>
    {/if}
  </div>

  <div class="rounded-[var(--radius-card)] border border-border bg-surface p-5">
    <div class="flex flex-wrap items-center justify-between gap-2">
      <p class="text-sm font-medium text-fg">{i18n.t("settings.ttsJobsTitle")}</p>
      <button
        type="button"
        class="btn btn-ghost text-xs ring-1 ring-border"
        onclick={() => void refreshJobs()}
        aria-label={i18n.t("common.refresh")}
      >
        <RefreshCw size={14} />
      </button>
    </div>
    {#if jobsLoading}
      <div class="mt-3">
        <Skeleton height="4rem" rounded="lg" />
      </div>
    {:else if jobs.length === 0}
      <div class="mt-3">
        <EmptyState
          size="sm"
          title={i18n.t("settings.ttsJobsEmptyTitle")}
          body={i18n.t("settings.ttsJobsEmptyBody")}
        >
          {#snippet icon(size)}
            <AudioLines {size} />
          {/snippet}
        </EmptyState>
      </div>
    {:else}
      <ul class="mt-3 divide-y divide-border rounded-lg border border-border">
        {#each jobs as job (job.id)}
          <li class="px-3 py-2">
            <div class="flex flex-wrap items-start justify-between gap-2">
              <div class="min-w-0">
                <button
                  type="button"
                  class="text-sm font-medium text-fg hover:underline"
                  onclick={() => router.navigate(routes.book(job.bookId))}
                >
                  {job.bookTitle || `#${job.bookId}`}
                </button>
                <p class="text-xs text-muted">
                  {jobStatusLabel(job)}
                  {#if jobProgressLabel(job)}
                    · {jobProgressLabel(job)}{/if}
                  {#if job.voice}
                    · {job.voice}{/if}
                </p>
                {#if job.error}
                  <p class="text-xs text-danger">{job.error}</p>
                {/if}
              </div>
              <div class="flex shrink-0 gap-1">
                {#if job.status === "queued" || job.status === "running"}
                  <button
                    type="button"
                    class="btn btn-ghost text-xs text-danger"
                    onclick={() => cancelJob(job)}
                  >
                    {i18n.t("settings.ttsJobCancel")}
                  </button>
                {:else}
                  {#if job.status !== "done"}
                    <button
                      type="button"
                      class="btn btn-ghost text-xs"
                      onclick={() => retryJob(job)}
                    >
                      {i18n.t("settings.ttsJobRetry")}
                    </button>
                  {/if}
                  <button
                    type="button"
                    class="btn btn-ghost text-xs text-danger"
                    onclick={() => removeJob(job)}
                  >
                    {i18n.t("confirm.delete")}
                  </button>
                {/if}
              </div>
            </div>
            {#if job.status === "running" && job.totalChapters > 0}
              <div class="mt-2 h-1 w-full overflow-hidden rounded-full bg-border/50">
                <div
                  class="h-full rounded-full bg-accent transition-[width] duration-300"
                  style:width="{Math.round((job.doneChapters / job.totalChapters) * 100)}%"
                ></div>
              </div>
            {/if}
          </li>
        {/each}
      </ul>
    {/if}
  </div>
{/if}
