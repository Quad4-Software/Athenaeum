import { router } from "$lib/router.svelte";

export type BookEditorPanel = "edit" | "identify";

class BookEditorIntentStore {
  target = $state<{ bookId: number; panel: BookEditorPanel } | null>(null);

  open(bookId: number, panel: BookEditorPanel) {
    this.target = { bookId, panel };
    router.navigate(`/book/${bookId}`);
  }

  consume(bookId: number): BookEditorPanel | null {
    if (this.target?.bookId !== bookId) return null;
    const panel = this.target.panel;
    this.target = null;
    return panel;
  }
}

export const bookEditorIntent = new BookEditorIntentStore();
