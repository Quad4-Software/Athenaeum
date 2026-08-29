/**
 * Imperative confirm dialog. Call `confirmDialog.ask({...})` and await the result.
 * Mount ConfirmDialogHost once in App.svelte.
 */

export type ConfirmOptions = {
  title: string;
  message: string;
  confirmLabel?: string;
  cancelLabel?: string;
  danger?: boolean;
};

type Pending = ConfirmOptions & {
  resolve: (ok: boolean) => void;
};

class ConfirmDialogStore {
  open = $state(false);
  title = $state("");
  message = $state("");
  confirmLabel = $state("Confirm");
  cancelLabel = $state("Cancel");
  danger = $state(false);

  #pending: Pending | null = null;

  ask(opts: ConfirmOptions): Promise<boolean> {
    return new Promise((resolve) => {
      if (this.#pending) {
        this.#pending.resolve(false);
      }
      this.title = opts.title;
      this.message = opts.message;
      this.confirmLabel = opts.confirmLabel ?? "Confirm";
      this.cancelLabel = opts.cancelLabel ?? "Cancel";
      this.danger = opts.danger ?? false;
      this.#pending = { ...opts, resolve };
      this.open = true;
    });
  }

  accept() {
    this.#finish(true);
  }

  cancel() {
    this.#finish(false);
  }

  #finish(ok: boolean) {
    const pending = this.#pending;
    this.#pending = null;
    this.open = false;
    pending?.resolve(ok);
  }
}

export const confirmDialog = new ConfirmDialogStore();
