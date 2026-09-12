import { router } from "$lib/router.svelte";
import { routes } from "$lib/routes";
import { watch } from "runed";
import { online as onlineWindow } from "svelte/reactivity/window";

const SKIP_OFFLINE_NAV = new Set(["error", "login", "setup", "reader"]);

class ConnectivityStore {
  online = $state(typeof navigator !== "undefined" ? navigator.onLine : true);
  serverReachable = $state(true);

  constructor() {
    if (typeof window === "undefined") return;
    $effect.root(() => {
      watch(
        () => onlineWindow.current,
        (now, prev) => {
          if (now === prev || now === undefined) return;
          if (now) this.onOnline();
          else this.onOffline();
        },
        { lazy: true },
      );
    });
  }

  onOffline() {
    this.online = false;
    if (!SKIP_OFFLINE_NAV.has(router.current.name)) {
      router.navigate(routes.error("offline"), true);
    }
  }

  onOnline() {
    this.online = true;
    if (router.current.name === "error" && router.current.params.code === "offline") {
      router.navigate(routes.library(), true);
      this.markReachable();
    }
  }

  markUnreachable() {
    this.serverReachable = false;
  }

  markReachable() {
    this.serverReachable = true;
  }

  get disconnected(): boolean {
    return !this.online || !this.serverReachable;
  }
}

export const connectivity = new ConnectivityStore();
