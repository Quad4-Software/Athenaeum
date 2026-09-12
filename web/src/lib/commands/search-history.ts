import { PersistedState } from "runed";
import { storageKey } from "$lib/brand/storage";

const HISTORY_KEY = storageKey("recent-searches");
const MAX = 10;

const history = new PersistedState<string[]>(HISTORY_KEY, []);

/** Most-recent-first query list with dedupe and cap. Empty input is a no-op. */
export function nextSearchHistory(current: string[], query: string, max = MAX): string[] {
  const trimmed = query.trim();
  if (!trimmed) return current;
  return [trimmed, ...current.filter((item) => item !== trimmed)].slice(0, max);
}

export function listRecentSearches(): string[] {
  const list = history.current;
  return Array.isArray(list) ? list.filter((q) => typeof q === "string" && q.trim()) : [];
}

export function rememberSearch(query: string) {
  history.current = nextSearchHistory(listRecentSearches(), query);
}

export function clearRecentSearches() {
  history.current = [];
}
