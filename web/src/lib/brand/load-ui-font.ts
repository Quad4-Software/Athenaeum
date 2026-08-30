/**
 * Lazy-load optional UI font CSS. Default Athenaeum fonts ship in app.css.
 */

import type { UiFontId } from "./fonts";

const loaded = new Set<UiFontId>(["athenaeum", "system"]);
const inflight = new Map<UiFontId, Promise<void>>();

const loaders: Partial<Record<UiFontId, () => Promise<unknown>>> = {
  "source-serif": () => import("@fontsource-variable/source-serif-4/wght.css"),
  literata: () => import("@fontsource-variable/literata/wght.css"),
  crimson: () => import("@fontsource-variable/crimson-pro/wght.css"),
  newsreader: () => import("@fontsource-variable/newsreader/wght.css"),
  "ibm-plex": () => import("@fontsource-variable/ibm-plex-sans/wght.css"),
  "dm-sans": () => import("@fontsource-variable/dm-sans/wght.css"),
};

/** Ensure CSS for the chosen UI font is loaded (no-op for defaults). */
export function loadUiFontCss(id: UiFontId): Promise<void> {
  if (loaded.has(id)) return Promise.resolve();
  const existing = inflight.get(id);
  if (existing) return existing;

  const loader = loaders[id];
  if (!loader) {
    loaded.add(id);
    return Promise.resolve();
  }

  const task = loader()
    .then(() => {
      loaded.add(id);
    })
    .finally(() => {
      inflight.delete(id);
    });
  inflight.set(id, task);
  return task;
}
