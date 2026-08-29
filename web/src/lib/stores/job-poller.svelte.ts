export type JobPollerOptions<T> = {
  fetchStatus: () => Promise<T>;
  isActive: (status: T) => boolean;
  intervalMs?: number;
};

/** Shared polling helper for long-running server jobs (scan, metadata match). */
export function createJobPoller<T>(opts: JobPollerOptions<T>) {
  const intervalMs = opts.intervalMs ?? 2000;

  class JobPoller {
    status = $state<T | null>(null);
    private timer: ReturnType<typeof setInterval> | null = null;
    private wasActive = false;

    startPolling(onComplete?: () => void) {
      this.stopPolling();
      void this.poll(onComplete);
      this.timer = setInterval(() => void this.poll(onComplete), intervalMs);
    }

    stopPolling() {
      if (this.timer) {
        clearInterval(this.timer);
        this.timer = null;
      }
    }

    private async poll(onComplete?: () => void) {
      try {
        const st = await opts.fetchStatus();
        this.status = st;
        const active = opts.isActive(st);
        if (this.wasActive && !active) {
          onComplete?.();
        }
        this.wasActive = active;
        if (!active && this.timer) {
          this.stopPolling();
        }
      } catch {
        // non-critical
      }
    }

    async checkActive(onComplete?: () => void) {
      try {
        const st = await opts.fetchStatus();
        this.status = st;
        this.wasActive = opts.isActive(st);
        if (this.wasActive) {
          this.startPolling(onComplete);
        }
      } catch {
        // non-critical
      }
    }
  }

  return new JobPoller();
}
