import type { SidebarPrefs, SidebarSectionId } from "$lib/api/types";

import { storageKey } from "$lib/brand/storage";
import { PersistedState } from "runed";

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

function defaultPrefs(): SidebarPrefs {
  return { order: [...DEFAULT_ORDER], hidden: [] };
}

const prefsSerializer = {
  serialize: (prefs: SidebarPrefs) => JSON.stringify(prefs),
  deserialize: (raw: string): SidebarPrefs => {
    try {
      const parsed = JSON.parse(raw) as SidebarPrefs;
      const order = parsed.order?.length ? parsed.order : [...DEFAULT_ORDER];
      const hidden = parsed.hidden ?? [];
      for (const id of DEFAULT_ORDER) {
        if (!order.includes(id)) order.push(id);
      }
      return { order, hidden };
    } catch {
      return defaultPrefs();
    }
  },
};

const collapsedSerializer = {
  serialize: (collapsed: Record<string, boolean>) => JSON.stringify(collapsed),
  deserialize: (raw: string): Record<string, boolean> => {
    try {
      return JSON.parse(raw) as Record<string, boolean>;
    } catch {
      return {};
    }
  },
};

class SidebarPrefsStore {
  #prefs = new PersistedState<SidebarPrefs>(STORAGE_KEY, defaultPrefs(), {
    serializer: prefsSerializer,
  });
  #collapsed = new PersistedState<Record<string, boolean>>(
    COLLAPSED_SECTIONS_KEY,
    {},
    {
      serializer: collapsedSerializer,
    },
  );

  get order(): SidebarSectionId[] {
    return this.#prefs.current.order;
  }

  set order(order: SidebarSectionId[]) {
    this.#prefs.current = { order, hidden: this.hidden };
  }

  get hidden(): SidebarSectionId[] {
    return this.#prefs.current.hidden;
  }

  set hidden(hidden: SidebarSectionId[]) {
    this.#prefs.current = { order: this.order, hidden };
  }

  get sectionCollapsed(): Record<string, boolean> {
    return this.#collapsed.current;
  }

  set sectionCollapsed(collapsed: Record<string, boolean>) {
    this.#collapsed.current = collapsed;
  }

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
  }

  moveSection(id: SidebarSectionId, dir: -1 | 1) {
    const idx = this.order.indexOf(id);
    const next = idx + dir;
    if (idx < 0 || next < 0 || next >= this.order.length) return;
    const order = [...this.order];
    [order[idx], order[next]] = [order[next], order[idx]];
    this.order = order;
  }

  reset() {
    this.order = [...DEFAULT_ORDER];
    this.hidden = [];
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
  }

  setSectionExpanded(id: string, expanded: boolean) {
    this.sectionCollapsed = { ...this.sectionCollapsed, [id]: !expanded };
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
