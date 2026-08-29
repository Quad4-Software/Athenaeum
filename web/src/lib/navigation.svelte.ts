import { type Route } from "$lib/router.svelte";
import { library } from "$lib/stores/library.svelte";
import { collections } from "$lib/stores/collections.svelte";
import type { Crumb } from "$lib/components/Breadcrumbs.svelte";

export function breadcrumbsFor(route: Route, bookTitle?: string): Crumb[] {
  const items: Crumb[] = [{ label: "Library", href: "/" }];

  if (route.name === "settings") {
    items.push({ label: "Settings", href: "/settings/library" });
    return items;
  }

  if (route.name === "collections") {
    items.push({ label: "Collections" });
    return items;
  }

  if (route.name === "collection") {
    items.push({ label: "Collections", href: "/collections" });
    const id = Number(route.params.id);
    const col = collections.items.find((c) => c.id === id);
    items.push({ label: col?.name ?? "Shelf" });
    return items;
  }

  if (library.collectionFilter != null) {
    const col = collections.items.find((c) => c.id === library.collectionFilter);
    if (col) items.push({ label: col.name });
  } else if (library.seriesFilter) {
    items.push({ label: library.seriesFilter });
  } else if (library.formatFilter) {
    items.push({ label: library.formatFilter.toUpperCase() });
  }

  if (route.name === "book" || route.name === "reader") {
    if (bookTitle) {
      items.push({
        label: bookTitle,
        href: route.name === "reader" ? `/book/${route.params.id}` : undefined,
      });
    }
    if (route.name === "reader") items.push({ label: "Reading" });
  }

  return items;
}

import { legacyStorageKey } from "$lib/brand/storage";

export function pdfReaderMode(): "canvas" | "native" {
  if (typeof localStorage === "undefined") return "canvas";
  return localStorage.getItem(legacyStorageKey("pdf_mode")) === "native" ? "native" : "canvas";
}

export function setPdfReaderMode(mode: "canvas" | "native") {
  localStorage.setItem(legacyStorageKey("pdf_mode"), mode);
}
