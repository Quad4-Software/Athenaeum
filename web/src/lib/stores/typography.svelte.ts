/**
 * UI typography preference: persists and applies --font-sans / --font-display.
 */

import { applyUiFont, DEFAULT_UI_FONT, isUiFontId, type UiFontId } from "$lib/brand/fonts";
import { loadUiFontCss } from "$lib/brand/load-ui-font";
import { storageKey } from "$lib/brand/storage";

const STORAGE_KEY = storageKey("ui-font");

function load(): UiFontId {
  if (typeof localStorage === "undefined") return DEFAULT_UI_FONT;
  const saved = localStorage.getItem(STORAGE_KEY);
  if (saved && isUiFontId(saved)) return saved;
  return DEFAULT_UI_FONT;
}

class TypographyStore {
  id = $state<UiFontId>(DEFAULT_UI_FONT);

  constructor() {
    this.id = load();
    void loadUiFontCss(this.id).then(() => applyUiFont(this.id));
  }

  set(id: UiFontId) {
    this.id = id;
    localStorage.setItem(STORAGE_KEY, id);
    void loadUiFontCss(id).then(() => applyUiFont(id));
  }
}

export const typography = new TypographyStore();
