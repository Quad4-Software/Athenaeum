/**
 * Demo mode lets the SPA run without a Go backend (GitHub Pages, local preview).
 * Enable with VITE_DEMO=1 at build time, ?demo=1, or localStorage athenaeum-demo=1.
 */
export function isDemoMode(): boolean {
  if (import.meta.env.VITE_DEMO === "true" || import.meta.env.VITE_DEMO === "1") {
    return true;
  }
  if (typeof window === "undefined") return false;
  try {
    const params = new URLSearchParams(window.location.search);
    if (params.get("demo") === "1" || params.get("demo") === "true") {
      window.localStorage.setItem("athenaeum-demo", "1");
      return true;
    }
    return window.localStorage.getItem("athenaeum-demo") === "1";
  } catch {
    return false;
  }
}
