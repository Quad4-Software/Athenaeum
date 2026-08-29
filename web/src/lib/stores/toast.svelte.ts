export type ToastKind = "success" | "error" | "info" | "loading";

export interface Toast {
  id: number;
  kind: ToastKind;
  message: string;
  /** Optional 0..1 progress for loading toasts. */
  progress?: number;
  sticky?: boolean;
}

export type ToastUpdate = Partial<Pick<Toast, "kind" | "message" | "progress" | "sticky">>;

const DEFAULT_MS = 4000;
const ERROR_MS = 6500;
const SUCCESS_MS = 3500;

class ToastStore {
  items = $state<Toast[]>([]);
  private nextId = 1;
  private timers = new Map<number, ReturnType<typeof setTimeout>>();

  show(message: string, kind: ToastKind = "info", durationMs = DEFAULT_MS): number {
    const sticky = kind === "loading" || durationMs <= 0;
    const id = this.nextId++;
    this.items = [
      ...this.items,
      {
        id,
        kind,
        message,
        sticky,
      },
    ];
    if (!sticky) this.scheduleDismiss(id, durationMs);
    return id;
  }

  success(message: string): number {
    return this.show(message, "success", SUCCESS_MS);
  }

  error(message: string): number {
    return this.show(message, "error", ERROR_MS);
  }

  info(message: string): number {
    return this.show(message, "info", DEFAULT_MS);
  }

  /** Sticky in-progress toast. Returns id for update/done. */
  loading(message: string, progress?: number): number {
    const id = this.nextId++;
    this.items = [
      ...this.items,
      {
        id,
        kind: "loading",
        message,
        progress: clampProgress(progress),
        sticky: true,
      },
    ];
    return id;
  }

  update(id: number, patch: ToastUpdate): void {
    const idx = this.items.findIndex((t) => t.id === id);
    if (idx < 0) return;
    const cur = this.items[idx];
    const next: Toast = {
      ...cur,
      ...patch,
      progress: patch.progress !== undefined ? clampProgress(patch.progress) : cur.progress,
    };
    this.items = [...this.items.slice(0, idx), next, ...this.items.slice(idx + 1)];
  }

  /** Replace a loading toast with a final result and auto-dismiss. */
  done(id: number, message: string, kind: Exclude<ToastKind, "loading"> = "success"): void {
    this.clearTimer(id);
    const idx = this.items.findIndex((t) => t.id === id);
    if (idx < 0) {
      this.show(message, kind, kind === "error" ? ERROR_MS : SUCCESS_MS);
      return;
    }
    const next: Toast = {
      id,
      kind,
      message,
      sticky: false,
      progress: undefined,
    };
    this.items = [...this.items.slice(0, idx), next, ...this.items.slice(idx + 1)];
    this.scheduleDismiss(id, kind === "error" ? ERROR_MS : SUCCESS_MS);
  }

  /**
   * Show a loading toast while promise runs, then success/error.
   * Updates progress if onProgress is used by the caller via returned id.
   */
  async promise<T>(
    action: (ctx: { id: number; update: (patch: ToastUpdate) => void }) => Promise<T>,
    labels: {
      loading: string;
      success: string | ((value: T) => string);
      error: string | ((err: unknown) => string);
    },
  ): Promise<T> {
    const id = this.loading(labels.loading);
    try {
      const value = await action({
        id,
        update: (patch) => this.update(id, patch),
      });
      const msg = typeof labels.success === "function" ? labels.success(value) : labels.success;
      this.done(id, msg, "success");
      return value;
    } catch (err) {
      const msg = typeof labels.error === "function" ? labels.error(err) : labels.error;
      this.done(id, msg, "error");
      throw err;
    }
  }

  dismiss(id: number) {
    this.clearTimer(id);
    this.items = this.items.filter((t) => t.id !== id);
  }

  private scheduleDismiss(id: number, durationMs: number) {
    this.clearTimer(id);
    this.timers.set(
      id,
      setTimeout(() => {
        this.timers.delete(id);
        this.dismiss(id);
      }, durationMs),
    );
  }

  private clearTimer(id: number) {
    const t = this.timers.get(id);
    if (t) {
      clearTimeout(t);
      this.timers.delete(id);
    }
  }
}

function clampProgress(progress: number | undefined): number | undefined {
  if (progress === undefined || !Number.isFinite(progress)) return undefined;
  return Math.min(1, Math.max(0, progress));
}

export const toast = new ToastStore();
