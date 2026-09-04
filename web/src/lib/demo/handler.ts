import type {
  Book,
  BookPage,
  BookQueryParams,
  Collection,
  HealthResponse,
  LibraryMount,
  LibraryStats,
  MetadataMatch,
  Progress,
  ReadingStats,
  ScanStatus,
  SeriesInfo,
  AuthorInfo,
  AudiobookTrack,
  Chapter,
  ComicManifest,
  MetadataProvider,
} from "$lib/api/types";
import { isAudioFormat } from "$lib/api/types";
import { opURL } from "$lib/api/op";
import { DEMO_CATALOG, allDemoBooks } from "./catalog";
import { demoCoverDataUrl } from "./covers";

type MutableBook = Book & { progressPercent?: number };

const books: MutableBook[] = allDemoBooks();
const favorites = new Set(DEMO_CATALOG.filter((e) => e.favorite).map((e) => e.id));
const progress = new Map<number, Progress>();
const shelves = new Map<string, number[]>();

for (const e of DEMO_CATALOG) {
  if (e.progress != null && e.progress > 0) {
    progress.set(e.id, {
      bookId: e.id,
      location: `demo:${e.progress.toFixed(2)}`,
      percent: e.progress,
      updatedAt: new Date().toISOString(),
    });
  }
  if (e.shelf) {
    const list = shelves.get(e.shelf) ?? [];
    list.push(e.id);
    shelves.set(e.shelf, list);
  }
}

function collections(): Collection[] {
  const now = new Date().toISOString();
  const out: Collection[] = [
    {
      id: 1,
      name: "Recently Added",
      description: "Books added in the last 30 days",
      kind: "auto",
      query: { addedDays: 30 },
      bookCount: books.length,
      createdAt: now,
    },
    {
      id: 2,
      name: "Audiobooks",
      description: "Public-domain audio titles",
      kind: "smart",
      query: { format: "audio" },
      bookCount: books.filter((b) => isAudioFormat(b.format)).length,
      createdAt: now,
    },
  ];
  let id = 3;
  for (const [name, ids] of shelves) {
    out.push({
      id: id++,
      name,
      description: "Public-domain demo shelf",
      kind: "manual",
      bookCount: ids.length,
      createdAt: now,
    });
  }
  return out;
}

function findBook(id: number): MutableBook | undefined {
  return books.find((b) => b.id === id);
}

function listFiltered(params: BookQueryParams): BookPage {
  let items = [...books];
  if (params.search) {
    const q = params.search.toLowerCase();
    items = items.filter(
      (b) =>
        b.title.toLowerCase().includes(q) ||
        b.author.toLowerCase().includes(q) ||
        (b.series?.toLowerCase().includes(q) ?? false) ||
        (b.description?.toLowerCase().includes(q) ?? false),
    );
  }
  if (params.format) {
    const f = params.format;
    items = items.filter((b) => {
      if (f === "audio") return isAudioFormat(b.format);
      if (f === "comic") return b.format === "cbz" || b.format === "cbr";
      if (f === "kindle") return b.format === "mobi" || b.format === "azw" || b.format === "azw3";
      if (f === "papers") return Boolean(b.doi || b.arxivId || b.pubmedId);
      return b.format === f;
    });
  }
  if (params.author) items = items.filter((b) => b.author === params.author);
  if (params.series) items = items.filter((b) => b.series === params.series);
  if (params.favorites) items = items.filter((b) => favorites.has(b.id));
  if (params.inProgress) {
    items = items.filter((b) => {
      const p = progress.get(b.id)?.percent ?? b.progressPercent ?? 0;
      return p > 0 && p < 1;
    });
  }
  if (params.collection != null) {
    const cols = collections();
    const col = cols.find((c) => c.id === params.collection);
    if (col?.kind === "manual" && col.name) {
      const ids = new Set(shelves.get(col.name) ?? []);
      items = items.filter((b) => ids.has(b.id));
    } else if (col?.query?.format === "audio") {
      items = items.filter((b) => isAudioFormat(b.format));
    }
  }

  switch (params.sort) {
    case "title":
      items.sort((a, b) => a.title.localeCompare(b.title));
      break;
    case "author":
      items.sort((a, b) => a.author.localeCompare(b.author) || a.title.localeCompare(b.title));
      break;
    case "oldest":
      items.sort((a, b) => a.addedAt.localeCompare(b.addedAt));
      break;
    case "progress":
      items.sort(
        (a, b) =>
          (b.progressPercent ?? 0) - (a.progressPercent ?? 0) || a.title.localeCompare(b.title),
      );
      break;
    default:
      items.sort((a, b) => b.addedAt.localeCompare(a.addedAt));
  }

  const limit = params.limit ?? 60;
  const offset = params.offset ?? 0;
  const total = items.length;
  const page = items.slice(offset, offset + limit).map((b) => ({
    ...b,
    progressPercent: progress.get(b.id)?.percent ?? b.progressPercent,
  }));
  return { items: page, total, limit, offset };
}

function parseQuery(path: string): BookQueryParams {
  const q = new URL(path, "http://demo.local").searchParams;
  return {
    search: q.get("search") ?? undefined,
    sort: (q.get("sort") as BookQueryParams["sort"]) ?? undefined,
    format: (q.get("format") as BookQueryParams["format"]) ?? undefined,
    series: q.get("series") ?? undefined,
    author: q.get("author") ?? undefined,
    library: q.get("library") ? Number(q.get("library")) : undefined,
    collection: q.get("collection") ? Number(q.get("collection")) : undefined,
    favorites: q.get("favorites") === "1",
    inProgress: q.get("inProgress") === "1",
    limit: q.get("limit") ? Number(q.get("limit")) : undefined,
    offset: q.get("offset") ? Number(q.get("offset")) : undefined,
  };
}

function stats(): LibraryStats {
  const audio = books.filter((b) => isAudioFormat(b.format)).length;
  const epub = books.filter((b) => b.format === "epub").length;
  const pdf = books.filter((b) => b.format === "pdf").length;
  const authors = new Set(books.map((b) => b.author));
  const series = new Set(books.map((b) => b.series).filter(Boolean));
  const inProg = [...progress.values()].filter((p) => p.percent > 0 && p.percent < 1).length;
  const completed = [...progress.values()].filter((p) => p.percent >= 1).length;
  return {
    totalBooks: books.length,
    epubCount: epub,
    pdfCount: pdf,
    audioCount: audio,
    totalSizeBytes: books.reduce((n, b) => n + b.fileSize, 0),
    authorCount: authors.size,
    seriesCount: series.size,
    libraryCount: 1,
    addedLast7Days: books.length,
    collectionCount: collections().length,
    readingInProgress: inProg,
    readingCompleted: completed,
    favoriteCount: favorites.size,
    scanning: false,
    authEnabled: false,
  };
}

function jsonResponse(data: unknown, status = 200): Response {
  if (data === undefined) return new Response(null, { status: status === 200 ? 204 : status });
  return new Response(JSON.stringify(data), {
    status,
    headers: { "Content-Type": "application/json" },
  });
}

function notFound(msg = "not found"): Response {
  return jsonResponse({ error: msg }, 404);
}

function fakeMetadata(book: Book): MetadataMatch[] {
  const entry = DEMO_CATALOG.find((e) => e.id === book.id);
  const cover = entry?.coverUrl ?? demoCoverDataUrl(book.title, book.author);
  return [
    {
      source: "openlibrary",
      sourceId: `demo-${book.id}`,
      title: book.title,
      author: book.author,
      description: book.description ?? `Public-domain edition of ${book.title}.`,
      language: book.language,
      series: book.series,
      seriesIndex: book.seriesIndex,
      coverUrl: cover,
      publishedYear: 1900,
    },
    {
      source: "demo",
      sourceId: `demo-alt-${book.id}`,
      title: `${book.title} (Annotated)`,
      author: book.author,
      description: `Alternate metadata match for ${book.title}.`,
      language: book.language ?? "en",
      coverUrl: cover,
      publishedYear: 1910,
    },
  ];
}

/** Handle an /api request in offline demo mode. Returns null if not handled. */
export async function handleDemoRequest(
  path: string,
  init?: RequestInit,
): Promise<Response | null> {
  const method = (init?.method ?? "GET").toUpperCase();
  const urlPath = path.split("?")[0] ?? path;

  if (urlPath === opURL("GET__api_health") && method === "GET") {
    const body: HealthResponse = { status: "ok", version: "demo", webVersion: "demo" };
    return jsonResponse(body);
  }

  if (urlPath === opURL("GET__api_auth_csrf") && method === "GET") {
    return jsonResponse({ ok: true });
  }

  if (urlPath === opURL("GET__api_auth_setup") && method === "GET") {
    return jsonResponse({
      needed: false,
      authEnabled: false,
      passwordPolicy: {
        minLength: 8,
        longLength: 12,
        minKinds: 3,
        requireLower: false,
        requireUpper: false,
        requireDigit: false,
        requireSymbol: false,
      },
    });
  }

  if (urlPath === opURL("GET__api_auth_methods") && method === "GET") {
    return jsonResponse({
      authEnabled: false,
      loginLocal: false,
      loginOidc: false,
      oidcAutoLaunch: false,
      passwordPolicy: {
        minLength: 8,
        longLength: 12,
        minKinds: 3,
        requireLower: false,
        requireUpper: false,
        requireDigit: false,
        requireSymbol: false,
      },
    });
  }

  if (urlPath === opURL("GET__api_i18n_locales") && method === "GET") {
    return jsonResponse({ locales: [] });
  }

  if (urlPath === opURL("GET__api_books") && method === "GET") {
    return jsonResponse(listFiltered(parseQuery(path)));
  }

  const bookMatch = urlPath.match(/^\/api\/books\/(\d+)(.*)$/);
  if (bookMatch) {
    const id = Number(bookMatch[1]);
    const rest = bookMatch[2] ?? "";
    const book = findBook(id);
    if (!book && rest !== "") return notFound();

    if (rest === "" && method === "GET") {
      if (!book) return notFound();
      return jsonResponse({
        ...book,
        progressPercent: progress.get(id)?.percent ?? book.progressPercent,
      });
    }

    if (rest === "" && method === "PUT" && book) {
      const body = init?.body ? (JSON.parse(String(init.body)) as Partial<Book>) : {};
      Object.assign(book, {
        title: body.title ?? book.title,
        author: body.author ?? book.author,
        series: body.series ?? book.series,
        seriesIndex: body.seriesIndex ?? book.seriesIndex,
        language: body.language ?? book.language,
        description: body.description ?? book.description,
        metaEdited: true,
        modifiedAt: new Date().toISOString(),
      });
      return jsonResponse(book);
    }

    if (rest === "/cover" && method === "GET" && book) {
      const entry = DEMO_CATALOG.find((e) => e.id === id);
      if (entry?.coverUrl) {
        return Response.redirect(entry.coverUrl, 302);
      }
      const svg = demoCoverDataUrl(book.title, book.author);
      const comma = svg.indexOf(",");
      const raw = decodeURIComponent(svg.slice(comma + 1));
      return new Response(raw, {
        status: 200,
        headers: { "Content-Type": "image/svg+xml", "Cache-Control": "public, max-age=3600" },
      });
    }

    if (rest === "/progress" && method === "GET") {
      return jsonResponse(
        progress.get(id) ?? {
          bookId: id,
          location: "",
          percent: 0,
          updatedAt: new Date().toISOString(),
        },
      );
    }

    if (rest === "/progress" && method === "PUT") {
      const body = init?.body
        ? (JSON.parse(String(init.body)) as Progress)
        : { location: "", percent: 0 };
      const p: Progress = {
        bookId: id,
        location: body.location ?? "",
        percent: body.percent ?? 0,
        updatedAt: new Date().toISOString(),
      };
      progress.set(id, p);
      const b = findBook(id);
      if (b) b.progressPercent = p.percent;
      return jsonResponse(p);
    }

    if (rest === "/favorite" && method === "PUT") {
      const body = init?.body ? (JSON.parse(String(init.body)) as { favorite?: boolean }) : {};
      if (body.favorite) favorites.add(id);
      else favorites.delete(id);
      return jsonResponse({ favorite: favorites.has(id) });
    }

    if (rest === "/chapters" && method === "GET") {
      const chapters: Chapter[] = [
        { index: 0, title: "Opening", startSec: 0 },
        { index: 1, title: "Middle", startSec: 600 },
        { index: 2, title: "Closing", startSec: 1200 },
      ];
      return jsonResponse(chapters);
    }

    if (rest === "/tracks" && method === "GET") {
      const tracks: AudiobookTrack[] = [
        { index: 0, title: "Opening", relPath: "01.mp3", format: "mp3", fileSize: 32000 },
        { index: 1, title: "Kiln", relPath: "02.mp3", format: "mp3", fileSize: 32000 },
        { index: 2, title: "Closing", relPath: "03.mp3", format: "mp3", fileSize: 32000 },
      ];
      return jsonResponse(tracks);
    }

    if (rest === "/pages" && method === "GET") {
      const manifest: ComicManifest = {
        total: 1,
        pages: [{ index: 0, name: "001.png", mimeType: "image/png" }],
      };
      return jsonResponse(manifest);
    }

    if (rest === "/bookmarks" && method === "GET") return jsonResponse([]);
    if (rest === "/highlights" && method === "GET") return jsonResponse([]);

    if (rest === "/metadata/search" && method === "POST" && book) {
      return jsonResponse({ matches: fakeMetadata(book) });
    }

    if (rest === "/metadata/apply" && method === "POST" && book) {
      const body = init?.body
        ? (JSON.parse(String(init.body)) as { match?: MetadataMatch; applyCover?: boolean })
        : {};
      const match = body.match;
      if (match) {
        book.title = match.title || book.title;
        book.author = match.author || book.author;
        book.description = match.description ?? book.description;
        book.language = match.language ?? book.language;
        book.series = match.series ?? book.series;
        book.seriesIndex = match.seriesIndex ?? book.seriesIndex;
        book.metaEdited = true;
        if (body.applyCover) book.hasCover = true;
        book.modifiedAt = new Date().toISOString();
      }
      return jsonResponse(book);
    }

    if (rest === "/reading-time" && method === "POST") {
      return jsonResponse({ ok: true });
    }

    if ((rest === "/file" || rest === "/download") && method === "GET") {
      return jsonResponse({ error: "demo mode: media streaming is disabled" }, 501);
    }
  }

  if (urlPath === opURL("GET__api_library_stats") && method === "GET") return jsonResponse(stats());

  if (urlPath === opURL("GET__api_library_scan_status") && method === "GET") {
    const status: ScanStatus = { scanning: false, indexed: books.length, skipped: 0 };
    return jsonResponse(status);
  }

  if (
    (urlPath === opURL("POST__api_library_scan") ||
      urlPath.match(/^\/api\/libraries\/\d+\/scan$/)) &&
    method === "POST"
  ) {
    return jsonResponse({ started: false });
  }

  if (urlPath === "/api/library/metadata/match/status" && method === "GET") {
    return jsonResponse({
      running: false,
      total: 0,
      done: 0,
      matched: 0,
      skipped: 0,
      failed: 0,
    });
  }

  if (urlPath === opURL("GET__api_series") && method === "GET") {
    const map = new Map<string, number>();
    for (const b of books) {
      if (!b.series) continue;
      map.set(b.series, (map.get(b.series) ?? 0) + 1);
    }
    const out: SeriesInfo[] = [...map.entries()].map(([name, count]) => ({ name, count }));
    out.sort((a, b) => a.name.localeCompare(b.name));
    return jsonResponse(out);
  }

  if (urlPath === opURL("GET__api_authors") && method === "GET") {
    const map = new Map<string, number>();
    for (const b of books) map.set(b.author, (map.get(b.author) ?? 0) + 1);
    const out: AuthorInfo[] = [...map.entries()].map(([name, count]) => ({ name, count }));
    out.sort((a, b) => a.name.localeCompare(b.name));
    return jsonResponse(out);
  }

  if (urlPath === opURL("GET__api_libraries") && method === "GET") {
    const libs: LibraryMount[] = [
      {
        id: 1,
        name: "Demo Library",
        mountPath: "./library",
        backend: "local",
        sortOrder: 0,
        bookCount: books.length,
        createdAt: new Date().toISOString(),
      },
    ];
    return jsonResponse(libs);
  }

  if (urlPath === opURL("GET__api_collections") && method === "GET")
    return jsonResponse(collections());

  if (urlPath === opURL("GET__api_favorites") && method === "GET") {
    return jsonResponse({ ids: [...favorites] });
  }

  if (urlPath === "/api/stats/reading" && method === "GET") {
    const body: ReadingStats = {
      totalReadSeconds: 12_600,
      booksInProgress: [...progress.values()].filter((p) => p.percent > 0 && p.percent < 1).length,
      booksCompleted: [...progress.values()].filter((p) => p.percent >= 1).length,
      currentStreakDays: 3,
    };
    return jsonResponse(body);
  }

  if (urlPath === "/api/metadata/providers" && method === "GET") {
    const providers: MetadataProvider[] = [
      { id: "demo", label: "Demo Provider", description: "Synthetic matches for offline demo" },
      { id: "openlibrary", label: "Open Library", description: "Unavailable offline" },
    ];
    return jsonResponse(providers);
  }

  if (urlPath === opURL("GET__api_docs") && method === "GET") {
    return jsonResponse({ title: "Athenaeum Demo API", baseUrl: "/api", groups: [] });
  }

  // Soft-success for mutating admin/settings calls in demo.
  if (method !== "GET" && urlPath.startsWith("/api/")) {
    return jsonResponse({ ok: true, demo: true });
  }

  if (urlPath.startsWith("/api/")) {
    return jsonResponse({ error: "demo mode: endpoint not available offline", path: urlPath }, 501);
  }

  return null;
}
