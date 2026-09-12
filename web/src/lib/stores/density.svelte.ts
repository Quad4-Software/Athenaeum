/**
 * Cover grid density preference for the library browse views.
 */

import { storageKey } from "$lib/brand/storage";
import { PersistedState } from "runed";

export type GridDensity = "comfortable" | "compact";

const STORAGE_KEY = storageKey("grid-density");

class DensityStore {
  #persisted = new PersistedState<GridDensity>(STORAGE_KEY, "comfortable", {
    serializer: {
      // Stored as a bare string, not JSON.
      serialize: (value) => value,
      deserialize: (value) => (value === "compact" ? "compact" : "comfortable"),
    },
  });

  get value(): GridDensity {
    return this.#persisted.current;
  }

  set value(next: GridDensity) {
    this.#persisted.current = next;
  }

  set(next: GridDensity) {
    this.value = next;
  }

  toggle() {
    this.set(this.value === "compact" ? "comfortable" : "compact");
  }
}

export const density = new DensityStore();
