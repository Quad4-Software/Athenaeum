import { ApiError } from "$lib/api/core";
import { toast } from "$lib/stores/toast.svelte";

export type ApiActionOptions = {
  success?: string;
  errorFallback: string;
  silentStatuses?: number[];
};

/** Run an API call with consistent toast success/error handling. */
export async function apiAction<T>(
  action: () => Promise<T>,
  opts: ApiActionOptions,
): Promise<T | undefined> {
  try {
    const result = await action();
    if (opts.success) toast.success(opts.success);
    return result;
  } catch (err) {
    if (err instanceof ApiError && opts.silentStatuses?.includes(err.status)) {
      return undefined;
    }
    toast.error(err instanceof Error ? err.message : opts.errorFallback);
    return undefined;
  }
}

/** Format an unknown error for inline form messages. */
export function apiErrorMessage(err: unknown, fallback: string): string {
  return err instanceof Error ? err.message : fallback;
}
