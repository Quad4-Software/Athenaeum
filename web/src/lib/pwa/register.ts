import { toast } from "$lib/stores/toast.svelte";
import { pwa } from "$lib/stores/pwa.svelte";
import { registerSW } from "virtual:pwa-register";

const UPDATE_CHECK_MS = 60 * 60 * 1000;

let updateInterval: number | null = null;
let visibilityCheck: (() => void) | null = null;

function onVisibilityChange() {
  if (document.visibilityState === "visible") visibilityCheck?.();
}

export function initPwa(offlineReadyMessage: string) {
  pwa.initInstall();

  if (!import.meta.env.PROD || !("serviceWorker" in navigator)) return;

  const updateSW = registerSW({
    immediate: true,
    onNeedRefresh() {
      pwa.setUpdateAvailable(() => {
        void updateSW(true);
      });
    },
    onOfflineReady() {
      toast.info(offlineReadyMessage);
    },
    onRegisteredSW(_url, registration) {
      if (!registration) return;
      const check = () => {
        void registration.update();
      };
      if (updateInterval) clearInterval(updateInterval);
      updateInterval = window.setInterval(check, UPDATE_CHECK_MS);
      visibilityCheck = check;
      document.removeEventListener("visibilitychange", onVisibilityChange);
      document.addEventListener("visibilitychange", onVisibilityChange);
    },
  });
}
