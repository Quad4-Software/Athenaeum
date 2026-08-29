import { brand } from "./config";

/** Build a namespaced localStorage key for this fork. */
export function storageKey(suffix: string): string {
  return `${brand.storagePrefix}:${suffix}`;
}

/** Legacy keys that used underscores instead of colons. */
export function legacyStorageKey(suffix: string): string {
  return `${brand.storagePrefix}_${suffix}`;
}
