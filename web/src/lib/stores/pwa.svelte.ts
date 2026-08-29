class PwaStore {
  updateAvailable = $state(false);
  private reload?: () => void;

  setUpdateAvailable(reload: () => void) {
    this.updateAvailable = true;
    this.reload = reload;
  }

  applyUpdate() {
    this.reload?.();
  }

  dismissUpdate() {
    this.updateAvailable = false;
    this.reload = undefined;
  }
}

export const pwa = new PwaStore();
