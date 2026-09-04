import { brand } from "./config";

/**
 * IndexedDB / demo localStorage names. Keep these exact strings so existing
 * offline caches and demo flags stay valid across upgrades.
 */
export const IDB_BOOKS = "athenaeum-books";
export const IDB_AUDIO = "athenaeum-audio";
export const IDB_FONTS = "athenaeum-fonts";
export const DEMO_MODE_STORAGE_KEY = "athenaeum-demo";

/** Build a namespaced localStorage key for this fork. */
export function storageKey(suffix: string): string {
  return `${brand.storagePrefix}:${suffix}`;
}

/** Legacy keys that used underscores instead of colons. */
export function legacyStorageKey(suffix: string): string {
  return `${brand.storagePrefix}_${suffix}`;
}
