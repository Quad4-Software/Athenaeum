import type { Book, MetadataProvider } from "$lib/api/types";

/** Editable metadata fields as strings, matching the edit form inputs. */
export interface BookEditFields {
  title: string;
  author: string;
  series: string;
  seriesIndex: string;
  language: string;
  description: string;
  doi: string;
  arxivId: string;
  pubmedId: string;
  journal: string;
  volume: string;
  issue: string;
  pages: string;
  publishedYear: string;
}

/** External metadata search criteria. */
export interface MetadataSearchFields {
  title: string;
  author: string;
  isbn: string;
  asin: string;
  doi: string;
  arxivId: string;
  pubmedId: string;
}

export function emptyEditFields(): BookEditFields {
  return {
    title: "",
    author: "",
    series: "",
    seriesIndex: "",
    language: "",
    description: "",
    doi: "",
    arxivId: "",
    pubmedId: "",
    journal: "",
    volume: "",
    issue: "",
    pages: "",
    publishedYear: "",
  };
}

export function emptySearchFields(): MetadataSearchFields {
  return { title: "", author: "", isbn: "", asin: "", doi: "", arxivId: "", pubmedId: "" };
}

export function bookToEditFields(b: Book): BookEditFields {
  return {
    title: b.title,
    author: b.author,
    series: b.series ?? "",
    seriesIndex: b.seriesIndex != null && b.seriesIndex > 0 ? String(b.seriesIndex) : "",
    language: b.language ?? "",
    description: b.description ?? "",
    doi: b.doi ?? "",
    arxivId: b.arxivId ?? "",
    pubmedId: b.pubmedId ?? "",
    journal: b.journal ?? "",
    volume: b.volume ?? "",
    issue: b.issue ?? "",
    pages: b.pages ?? "",
    publishedYear: b.publishedYear != null && b.publishedYear > 0 ? String(b.publishedYear) : "",
  };
}

export function bookToSearchFields(b: Book): MetadataSearchFields {
  return {
    title: b.title,
    author: b.author,
    isbn: "",
    asin: "",
    doi: b.doi ?? "",
    arxivId: b.arxivId ?? "",
    pubmedId: b.pubmedId ?? "",
  };
}

/** True when any search field has input. */
export function hasSearchInput(search: MetadataSearchFields): boolean {
  return Object.values(search).some((v) => v.trim().length > 0);
}

/** Display label for a provider id, falling back to the id. */
export function providerLabel(providers: MetadataProvider[], id: string): string {
  return providers.find((p) => p.id === id)?.label ?? id;
}
