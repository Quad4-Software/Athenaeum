import { DEMO_MODE_STORAGE_KEY } from "$lib/brand/storage";

/**
 * Demo mode lets the SPA run without a Go backend (GitHub Pages, local preview).
 * Enable with VITE_DEMO=1 at build time, ?demo=1, or localStorage demo flag.
 */
export function isDemoMode(): boolean {
  if (import.meta.env.VITE_DEMO === "true" || import.meta.env.VITE_DEMO === "1") {
    return true;
  }
  if (typeof window === "undefined") return false;
  try {
    const params = new URLSearchParams(window.location.search);
    if (params.get("demo") === "1" || params.get("demo") === "true") {
      window.localStorage.setItem(DEMO_MODE_STORAGE_KEY, "1");
      return true;
    }
    return window.localStorage.getItem(DEMO_MODE_STORAGE_KEY) === "1";
  } catch {
    return false;
  }
}
