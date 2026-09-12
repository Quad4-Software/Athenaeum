/**
 * UI store holds layout state shared across components: the collapsible
 * desktop sidebar and the mobile navigation drawer. Collapse preference is
 * persisted so the layout is stable across reloads.
 */

import { storageKey } from "$lib/brand/storage";
import { PersistedState } from "runed";

const COLLAPSE_KEY = storageKey("sidebar-collapsed");

class UiStore {
  #persistedCollapsed = new PersistedState<boolean>(COLLAPSE_KEY, false, {
    serializer: {
      // Stored as "1" / "0", matching the pre-existing format.
      serialize: (value) => (value ? "1" : "0"),
      deserialize: (value) => value === "1",
    },
  });

  mobileNavOpen = $state(false);
  mobileNavTrigger: HTMLElement | null = $state(null);
  pageTitle = $state("");

  get sidebarCollapsed(): boolean {
    return this.#persistedCollapsed.current;
  }

  set sidebarCollapsed(collapsed: boolean) {
    this.#persistedCollapsed.current = collapsed;
  }

  toggleSidebar() {
    this.sidebarCollapsed = !this.sidebarCollapsed;
  }

  openMobileNav(trigger?: HTMLElement | null) {
    this.mobileNavTrigger = trigger ?? null;
    this.mobileNavOpen = true;
  }

  closeMobileNav() {
    this.mobileNavOpen = false;
    const trigger = this.mobileNavTrigger;
    if (trigger && document.contains(trigger)) {
      trigger.focus();
    }
    this.mobileNavTrigger = null;
  }
}

export const ui = new UiStore();
