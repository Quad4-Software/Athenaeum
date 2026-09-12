/**
 * UI typography preference: persists and applies --font-sans / --font-display.
 */

import { applyUiFont, DEFAULT_UI_FONT, isUiFontId, type UiFontId } from "$lib/brand/fonts";
import { loadUiFontCss } from "$lib/brand/load-ui-font";
import { storageKey } from "$lib/brand/storage";
import { PersistedState } from "runed";

const STORAGE_KEY = storageKey("ui-font");

class TypographyStore {
  #persisted = new PersistedState<UiFontId>(STORAGE_KEY, DEFAULT_UI_FONT, {
    serializer: {
      // Stored as a bare font id string, not JSON.
      serialize: (value) => value,
      deserialize: (value) => (isUiFontId(value) ? value : DEFAULT_UI_FONT),
    },
  });

  get id(): UiFontId {
    return this.#persisted.current;
  }

  set id(id: UiFontId) {
    this.#persisted.current = id;
  }

  constructor() {
    void loadUiFontCss(this.id).then(() => applyUiFont(this.id));
  }

  set(id: UiFontId) {
    this.id = id;
    void loadUiFontCss(id).then(() => applyUiFont(id));
  }
}

export const typography = new TypographyStore();
