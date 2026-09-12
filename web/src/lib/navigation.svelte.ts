import { type Route } from "$lib/router.svelte";
import { routes } from "$lib/routes";
import { library } from "$lib/stores/library.svelte";
import { collections } from "$lib/stores/collections.svelte";
import { i18n } from "$lib/stores/i18n.svelte";
import type { Crumb } from "$lib/components/Breadcrumbs.svelte";

export function breadcrumbsFor(route: Route, bookTitle?: string): Crumb[] {
  const items: Crumb[] = [{ label: i18n.t("nav.library"), href: routes.library() }];

  if (route.name === "settings") {
    items.push({ label: i18n.t("settings.title"), href: routes.settings("library") });
    return items;
  }

  if (route.name === "collections") {
    items.push({ label: i18n.t("collections.title") });
    return items;
  }

  if (route.name === "collection") {
    items.push({ label: i18n.t("collections.title"), href: routes.collections() });
    const id = Number(route.params.id);
    const col = collections.items.find((c) => c.id === id);
    items.push({ label: col?.name ?? i18n.t("nav.shelfFallback") });
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
        href: route.name === "reader" ? routes.book(route.params.id) : undefined,
      });
    }
    if (route.name === "reader") items.push({ label: i18n.t("nav.reading") });
  }

  return items;
}
