import { storageKey } from "$lib/brand/storage";
import type { Book } from "$lib/api/types";

const RECENT_KEY = storageKey("recent-books");
const MAX = 8;

export type RecentBook = Pick<Book, "id" | "title" | "author" | "hasCover" | "modifiedAt">;

function load(): RecentBook[] {
  if (typeof localStorage === "undefined") return [];
  try {
    const raw = localStorage.getItem(RECENT_KEY);
    if (!raw) return [];
    const parsed = JSON.parse(raw) as RecentBook[];
    return Array.isArray(parsed) ? parsed.slice(0, MAX) : [];
  } catch {
    return [];
  }
}

export function listRecentBooks(): RecentBook[] {
  return load();
}

export function rememberBook(book: RecentBook) {
  const next = [book, ...load().filter((b) => b.id !== book.id)].slice(0, MAX);
  localStorage.setItem(RECENT_KEY, JSON.stringify(next));
}
