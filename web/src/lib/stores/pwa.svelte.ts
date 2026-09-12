import { useEventListener, watch } from "runed";
import { MediaQuery } from "svelte/reactivity";

type BeforeInstallPromptEvent = Event & {
  prompt: () => Promise<void>;
  userChoice: Promise<{ outcome: "accepted" | "dismissed" }>;
};

class PwaStore {
  updateAvailable = $state(false);
  /** True when running as an installed PWA (standalone display). */
  installed = $state(false);
  /** Browser fired beforeinstallprompt and install has not completed. */
  canInstall = $state(false);
  /** Human-readable reason when install is unavailable. */
  installUnavailableReason = $state("");

  private reload?: () => void;
  private deferredPrompt: BeforeInstallPromptEvent | null = null;
  private listening = false;

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

  /** Start listening for installability. Safe to call more than once. */
  initInstall() {
    if (typeof window === "undefined" || this.listening) return;
    this.listening = true;

    this.refreshInstalled();
    this.refreshUnavailableReason();

    const standalone = new MediaQuery("(display-mode: standalone)");
    $effect.root(() => {
      useEventListener(window, "beforeinstallprompt", (event) => {
        event.preventDefault();
        this.deferredPrompt = event as BeforeInstallPromptEvent;
        this.canInstall = true;
        this.installUnavailableReason = "";
      });

      useEventListener(window, "appinstalled", () => {
        this.deferredPrompt = null;
        this.canInstall = false;
        this.installed = true;
        this.installUnavailableReason = "";
      });

      watch(
        () => standalone.current,
        () => {
          this.refreshInstalled();
          this.refreshUnavailableReason();
        },
        { lazy: true },
      );
    });
  }

  async promptInstall(): Promise<"accepted" | "dismissed" | "unavailable"> {
    if (!this.deferredPrompt) {
      this.refreshUnavailableReason();
      return "unavailable";
    }
    const prompt = this.deferredPrompt;
    this.deferredPrompt = null;
    this.canInstall = false;
    await prompt.prompt();
    const { outcome } = await prompt.userChoice;
    if (outcome === "accepted") {
      this.installed = true;
    } else {
      this.refreshUnavailableReason();
    }
    return outcome;
  }

  private refreshInstalled() {
    const standalone =
      window.matchMedia("(display-mode: standalone)").matches ||
      ("standalone" in navigator &&
        Boolean((navigator as Navigator & { standalone?: boolean }).standalone));
    this.installed = standalone;
    if (standalone) {
      this.canInstall = false;
      this.installUnavailableReason = "";
    }
  }

  private refreshUnavailableReason() {
    if (this.installed || this.canInstall) {
      this.installUnavailableReason = "";
      return;
    }
    if (!window.isSecureContext) {
      this.installUnavailableReason = "Requires HTTPS (or localhost).";
      return;
    }
    if (!("serviceWorker" in navigator)) {
      this.installUnavailableReason = "This browser does not support service workers.";
      return;
    }
    if (!import.meta.env.PROD) {
      this.installUnavailableReason = "Install is available in production builds only.";
      return;
    }
    this.installUnavailableReason =
      "Install is not offered yet. Use the browser menu (Install app / Add to Home Screen), or wait until the app is eligible.";
  }
}

export const pwa = new PwaStore();
