import { PersistedState } from "runed";
import { storageKey } from "$lib/brand/storage";
import type { Book } from "$lib/api/types";

const RECENT_KEY = storageKey("recent-books");
const MAX = 8;

export type RecentBook = Pick<Book, "id" | "title" | "author" | "hasCover" | "modifiedAt">;

const recent = new PersistedState<RecentBook[]>(RECENT_KEY, []);

export function listRecentBooks(): RecentBook[] {
  const list = recent.current;
  return Array.isArray(list) ? list.slice(0, MAX) : [];
}

export function rememberBook(book: RecentBook) {
  const next = [book, ...listRecentBooks().filter((b) => b.id !== book.id)].slice(0, MAX);
  recent.current = next;
}
