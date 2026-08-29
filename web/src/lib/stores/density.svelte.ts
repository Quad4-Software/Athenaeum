/**
 * Cover grid density preference for the library browse views.
 */

import { storageKey } from "$lib/brand/storage";

export type GridDensity = "comfortable" | "compact";

const STORAGE_KEY = storageKey("grid-density");

function load(): GridDensity {
  if (typeof localStorage === "undefined") return "comfortable";
  const v = localStorage.getItem(STORAGE_KEY);
  return v === "compact" ? "compact" : "comfortable";
}

class DensityStore {
  value = $state<GridDensity>(load());

  set(next: GridDensity) {
    this.value = next;
    localStorage.setItem(STORAGE_KEY, next);
  }

  toggle() {
    this.set(this.value === "compact" ? "comfortable" : "compact");
  }
}

export const density = new DensityStore();
