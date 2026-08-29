import { router } from "$lib/router.svelte";

const SKIP_OFFLINE_NAV = new Set(["error", "login", "setup", "reader"]);

class ConnectivityStore {
  online = $state(typeof navigator !== "undefined" ? navigator.onLine : true);
  serverReachable = $state(true);

  constructor() {
    if (typeof window === "undefined") return;
    window.addEventListener("online", () => this.onOnline());
    window.addEventListener("offline", () => this.onOffline());
  }

  onOffline() {
    this.online = false;
    if (!SKIP_OFFLINE_NAV.has(router.current.name)) {
      router.navigate("/error/offline", true);
    }
  }

  onOnline() {
    this.online = true;
    if (router.current.name === "error" && router.current.params.code === "offline") {
      router.navigate("/", true);
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
