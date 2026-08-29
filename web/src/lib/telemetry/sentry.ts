import type { TelemetryConfig } from "$lib/api/types";

type SentryModule = typeof import("@sentry/svelte");

let sentry: SentryModule | null = null;
let active = false;

export async function initSentry(cfg: TelemetryConfig) {
  if (active || !cfg.sentryDsn || import.meta.env.DEV) return;

  sentry = await import("@sentry/svelte");
  sentry.init({
    dsn: cfg.sentryDsn,
    environment: cfg.environment || undefined,
    release: cfg.release || undefined,
    tracesSampleRate: cfg.tracesSampleRate ?? 0,
  });
  active = true;
}

export function captureException(err: unknown) {
  if (!active || !sentry) return;
  sentry.captureException(err);
}

export function captureApiError(path: string, status: number, message: string) {
  if (!active || !sentry || status < 500) return;
  sentry.captureMessage(`API ${status} ${path}: ${message}`, "error");
}
