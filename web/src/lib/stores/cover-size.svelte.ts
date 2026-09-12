/**
 * Cover size preference for the library browse grid. Persisted as a bare
 * size letter; the grid derives its minimum column width from it.
 */

import { storageKey } from "$lib/brand/storage";
import { PersistedState } from "runed";

export type CoverSize = "s" | "m" | "l";

export const COVER_SIZES: readonly CoverSize[] = ["s", "m", "l"];

/** Minimum column width in px per size before the compact-density shrink. */
export const COVER_SIZE_MIN_PX: Record<CoverSize, number> = { s: 100, m: 140, l: 190 };

const COMPACT_COVER_SCALE = 0.8;
const STORAGE_KEY = storageKey("grid-cover-size");

export function isCoverSize(value: string): value is CoverSize {
  return value === "s" || value === "m" || value === "l";
}

/** Compact density keeps its tighter look by shrinking the chosen size. */
export function coverColumnMinPx(size: CoverSize, compact: boolean): number {
  const base = COVER_SIZE_MIN_PX[size];
  return compact ? Math.round(base * COMPACT_COVER_SCALE) : base;
}

class CoverSizeStore {
  #persisted = new PersistedState<CoverSize>(STORAGE_KEY, "m", {
    serializer: {
      // Stored as a bare string, not JSON.
      serialize: (value) => value,
      deserialize: (value) => (isCoverSize(value) ? value : "m"),
    },
  });

  get value(): CoverSize {
    return this.#persisted.current;
  }

  set value(next: CoverSize) {
    this.#persisted.current = next;
  }

  set(next: CoverSize) {
    this.value = next;
  }
}

export const coverSize = new CoverSizeStore();
