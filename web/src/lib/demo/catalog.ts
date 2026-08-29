import type { Book, BookFormat } from "$lib/api/types";

export interface DemoEntry {
  id: number;
  slug: string;
  title: string;
  author: string;
  series?: string;
  seriesIndex?: number;
  format: BookFormat;
  language: string;
  description: string;
  fileSize: number;
  progress?: number;
  favorite?: boolean;
  shelf?: string;
  /** Open Library cover for a public-domain work. */
  coverUrl?: string;
}

/**
 * Public-domain classics for the offline demo (Project Gutenberg / expired copyright).
 * Mirrored from internal/demo/catalog.go.
 */
export const DEMO_CATALOG: DemoEntry[] = [
  {
    id: 1,
    slug: "alice-in-wonderland",
    title: "Alice's Adventures in Wonderland",
    author: "Lewis Carroll",
    format: "epub",
    language: "en",
    description:
      "A girl falls down a rabbit hole into a world of peculiar creatures and impossible logic.",
    fileSize: 180000,
    progress: 0.42,
    favorite: true,
    shelf: "Classics",
    coverUrl: "https://covers.openlibrary.org/b/id/10527843-L.jpg",
  },
  {
    id: 2,
    slug: "pride-and-prejudice",
    title: "Pride and Prejudice",
    author: "Jane Austen",
    format: "epub",
    language: "en",
    description:
      "Elizabeth Bennet navigates manners, marriage, and misunderstanding in Georgian England.",
    fileSize: 720000,
    progress: 0.18,
    shelf: "Classics",
    coverUrl: "https://covers.openlibrary.org/b/id/14348537-L.jpg",
  },
  {
    id: 3,
    slug: "frankenstein",
    title: "Frankenstein",
    author: "Mary Shelley",
    format: "epub",
    language: "en",
    description:
      "A scientist's creation of life leads to tragedy, exile, and a chase across the ice.",
    fileSize: 450000,
    favorite: true,
    shelf: "Classics",
    coverUrl: "https://covers.openlibrary.org/b/id/12356249-L.jpg",
  },
  {
    id: 4,
    slug: "sherlock-holmes",
    title: "The Adventures of Sherlock Holmes",
    author: "Arthur Conan Doyle",
    format: "pdf",
    language: "en",
    description:
      "Twelve cases from Baker Street: scandals, red-headed leagues, and speckled bands.",
    fileSize: 1100000,
    progress: 0.55,
    coverUrl: "https://covers.openlibrary.org/b/id/6717853-L.jpg",
  },
  {
    id: 5,
    slug: "dracula",
    title: "Dracula",
    author: "Bram Stoker",
    format: "pdf",
    language: "en",
    description:
      "Letters and diaries chart Count Dracula's arrival in England and the hunt that follows.",
    fileSize: 980000,
    coverUrl: "https://covers.openlibrary.org/b/id/12216503-L.jpg",
  },
  {
    id: 6,
    slug: "moby-dick",
    title: "Moby-Dick; or, The Whale",
    author: "Herman Melville",
    format: "m4b",
    language: "en",
    description: "Ishmael ships aboard the Pequod as Captain Ahab hunts the white whale.",
    fileSize: 64000000,
    progress: 0.28,
    favorite: true,
    shelf: "Listen Next",
    coverUrl: "https://covers.openlibrary.org/b/id/12116552-L.jpg",
  },
  {
    id: 7,
    slug: "jungle-book",
    title: "The Jungle Book",
    author: "Rudyard Kipling",
    format: "mp3",
    language: "en",
    description: "Mowgli and other animal tales from the Indian jungle, first published in 1894.",
    fileSize: 42000000,
    shelf: "Listen Next",
    coverUrl: "https://covers.openlibrary.org/b/id/3344204-L.jpg",
  },
  {
    id: 8,
    slug: "christmas-carol",
    title: "A Christmas Carol",
    author: "Charles Dickens",
    format: "audiobook",
    language: "en",
    description: "Ebenezer Scrooge is visited by four spirits on Christmas Eve.",
    fileSize: 28000000,
    progress: 0.08,
    shelf: "Listen Next",
    coverUrl: "https://covers.openlibrary.org/b/id/12875748-L.jpg",
  },
  {
    id: 9,
    slug: "wizard-of-oz",
    title: "The Wonderful Wizard of Oz",
    author: "L. Frank Baum",
    format: "cbz",
    language: "en",
    description:
      "Dorothy is carried by a cyclone to Oz and seeks the Wizard with unusual companions.",
    fileSize: 2200000,
    favorite: true,
    shelf: "Classics",
    coverUrl: "https://covers.openlibrary.org/b/id/552443-L.jpg",
  },
  {
    id: 10,
    slug: "odyssey",
    title: "The Odyssey",
    author: "Homer",
    format: "epub",
    language: "en",
    description: "Odysseus wanders for ten years after Troy before returning to Ithaca.",
    fileSize: 640000,
    progress: 0.73,
    coverUrl: "https://covers.openlibrary.org/b/id/10876521-L.jpg",
  },
  {
    id: 11,
    slug: "metamorphosis",
    title: "The Metamorphosis",
    author: "Franz Kafka",
    format: "mobi",
    language: "en",
    description: "Gregor Samsa wakes to find himself transformed into a monstrous insect.",
    fileSize: 120000,
    coverUrl: "https://covers.openlibrary.org/b/id/12820198-L.jpg",
  },
  {
    id: 12,
    slug: "faust",
    title: "Faust",
    author: "Johann Wolfgang von Goethe",
    format: "epub",
    language: "de",
    description: "A scholar wagers his soul with Mephistopheles in Goethe's dramatic poem.",
    fileSize: 520000,
    shelf: "Classics",
    coverUrl: "https://covers.openlibrary.org/b/id/1002485-L.jpg",
  },
];

const ADDED_BASE = Date.UTC(2026, 7, 20, 12, 0, 0);

export function entryToBook(e: DemoEntry): Book {
  const added = new Date(ADDED_BASE - (e.id - 1) * 3600_000).toISOString();
  return {
    id: e.id,
    libraryId: 1,
    title: e.title,
    author: e.author,
    series: e.series,
    seriesIndex: e.seriesIndex,
    format: e.format,
    relPath: `demo/${e.slug}.${e.format === "audiobook" ? "" : e.format}`.replace(/\.$/, ""),
    fileSize: e.fileSize,
    hasCover: true,
    language: e.language,
    description: e.description,
    addedAt: added,
    modifiedAt: added,
    progressPercent: e.progress,
  };
}

export function allDemoBooks(): Book[] {
  return DEMO_CATALOG.map(entryToBook);
}
