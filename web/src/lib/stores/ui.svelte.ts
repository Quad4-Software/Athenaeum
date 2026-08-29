/**
 * UI store holds layout state shared across components: the collapsible
 * desktop sidebar and the mobile navigation drawer. Collapse preference is
 * persisted so the layout is stable across reloads.
 */

import { storageKey } from "$lib/brand/storage";

const COLLAPSE_KEY = storageKey("sidebar-collapsed");

class UiStore {
  sidebarCollapsed = $state(false);
  mobileNavOpen = $state(false);
  mobileNavTrigger: HTMLElement | null = $state(null);
  pageTitle = $state("");

  constructor() {
    this.sidebarCollapsed = localStorage.getItem(COLLAPSE_KEY) === "1";
  }

  toggleSidebar() {
    this.sidebarCollapsed = !this.sidebarCollapsed;
    localStorage.setItem(COLLAPSE_KEY, this.sidebarCollapsed ? "1" : "0");
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
