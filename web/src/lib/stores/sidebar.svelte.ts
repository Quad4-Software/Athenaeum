import type { SidebarPrefs, SidebarSectionId } from "$lib/api/types";

import { storageKey } from "$lib/brand/storage";

const STORAGE_KEY = storageKey("sidebar-prefs");
const COLLAPSED_SECTIONS_KEY = storageKey("sidebar-section-collapsed");

const DEFAULT_ORDER: SidebarSectionId[] = [
  "continue",
  "favorites",
  "formats",
  "libraries",
  "reading",
  "shelves",
  "series",
];

function load(): SidebarPrefs {
  if (typeof localStorage === "undefined") {
    return { order: [...DEFAULT_ORDER], hidden: [] };
  }
  try {
    const raw = localStorage.getItem(STORAGE_KEY);
    if (!raw) return { order: [...DEFAULT_ORDER], hidden: [] };
    const parsed = JSON.parse(raw) as SidebarPrefs;
    const order = parsed.order?.length ? parsed.order : [...DEFAULT_ORDER];
    const hidden = parsed.hidden ?? [];
    for (const id of DEFAULT_ORDER) {
      if (!order.includes(id)) order.push(id);
    }
    return { order, hidden };
  } catch {
    return { order: [...DEFAULT_ORDER], hidden: [] };
  }
}

function save(prefs: SidebarPrefs) {
  localStorage.setItem(STORAGE_KEY, JSON.stringify(prefs));
}

function loadCollapsedSections(): Record<string, boolean> {
  if (typeof localStorage === "undefined") return {};
  try {
    const raw = localStorage.getItem(COLLAPSED_SECTIONS_KEY);
    if (!raw) return {};
    return JSON.parse(raw) as Record<string, boolean>;
  } catch {
    return {};
  }
}

function saveCollapsedSections(collapsed: Record<string, boolean>) {
  localStorage.setItem(COLLAPSED_SECTIONS_KEY, JSON.stringify(collapsed));
}

class SidebarPrefsStore {
  order = $state<SidebarSectionId[]>(load().order);
  hidden = $state<SidebarSectionId[]>(load().hidden);
  sectionCollapsed = $state<Record<string, boolean>>(loadCollapsedSections());

  visibleSections(): SidebarSectionId[] {
    return this.order.filter((id) => !this.hidden.includes(id));
  }

  isHidden(id: SidebarSectionId): boolean {
    return this.hidden.includes(id);
  }

  toggleSection(id: SidebarSectionId) {
    if (this.hidden.includes(id)) {
      this.hidden = this.hidden.filter((x) => x !== id);
    } else {
      this.hidden = [...this.hidden, id];
    }
    save({ order: this.order, hidden: this.hidden });
  }

  moveSection(id: SidebarSectionId, dir: -1 | 1) {
    const idx = this.order.indexOf(id);
    const next = idx + dir;
    if (idx < 0 || next < 0 || next >= this.order.length) return;
    const order = [...this.order];
    [order[idx], order[next]] = [order[next], order[idx]];
    this.order = order;
    save({ order: this.order, hidden: this.hidden });
  }

  reset() {
    this.order = [...DEFAULT_ORDER];
    this.hidden = [];
    save({ order: this.order, hidden: this.hidden });
  }

  isSectionExpanded(id: string, defaultExpanded = true): boolean {
    if (id in this.sectionCollapsed) {
      return !this.sectionCollapsed[id];
    }
    return defaultExpanded;
  }

  toggleSectionExpanded(id: string) {
    const expanded = this.isSectionExpanded(id);
    this.sectionCollapsed = { ...this.sectionCollapsed, [id]: expanded };
    saveCollapsedSections(this.sectionCollapsed);
  }

  setSectionExpanded(id: string, expanded: boolean) {
    this.sectionCollapsed = { ...this.sectionCollapsed, [id]: !expanded };
    saveCollapsedSections(this.sectionCollapsed);
  }
}

export const sidebarPrefs = new SidebarPrefsStore();

export const SIDEBAR_SECTION_LABELS: Record<SidebarSectionId, string> = {
  libraries: "Libraries",
  formats: "Browse by format",
  favorites: "Favorites",
  continue: "Continue reading",
  reading: "Reading lists",
  series: "Series",
  shelves: "Shelves",
};
