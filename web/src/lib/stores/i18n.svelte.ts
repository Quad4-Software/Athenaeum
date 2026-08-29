import { api } from "$lib/api/client";
import type { LocaleInfo } from "$lib/i18n";
import {
  bundledLocaleCodes,
  detectBrowserLocale,
  loadBundledLocale,
  mergeMessages,
  translate,
  type Messages,
} from "$lib/i18n";
import { SvelteSet } from "svelte/reactivity";
import { brand } from "$lib/brand/config";
import { storageKey } from "$lib/brand/storage";

const STORAGE_KEY = storageKey("locale");
const DEFAULT_LOCALE = "en";

class I18nStore {
  locale = $state(DEFAULT_LOCALE);
  ready = $state(false);
  locales = $state<LocaleInfo[]>([]);

  #bundled = new Map<string, { messages: Messages; name: string }>();
  #custom = new Map<string, Messages>();
  #messages = $state<Messages>({});
  #fallback = $state<Messages>({});

  t(key: string, params?: Record<string, string | number>): string {
    return translate(this.#messages, key, params, this.#fallback);
  }

  /** Merge fork branding strings over bundled locales (e.g. app.title). */
  applyOverrides(overrides: Messages): void {
    this.#messages = mergeMessages(this.#messages, overrides);
    this.#fallback = mergeMessages(this.#fallback, overrides);
  }

  async #ensureBundled(code: string): Promise<{ messages: Messages; name: string } | null> {
    const cached = this.#bundled.get(code);
    if (cached) return cached;
    const loaded = await loadBundledLocale(code);
    if (loaded) this.#bundled.set(code, loaded);
    return loaded;
  }

  async init(): Promise<void> {
    if (this.ready) return;

    const fallback = await this.#ensureBundled(DEFAULT_LOCALE);
    this.#fallback = fallback?.messages ?? {};
    this.#messages = this.#fallback;

    const bundledList: LocaleInfo[] = bundledLocaleCodes().map((code) => ({
      code,
      name: this.#bundled.get(code)?.name || code,
      source: "bundled" as const,
    }));

    try {
      const remote = await api.listI18nLocales();
      const codes = new SvelteSet(bundledList.map((l) => l.code));
      const merged = [...bundledList];
      for (const loc of remote.locales) {
        if (loc.source === "custom" && !codes.has(loc.code)) {
          merged.push(loc);
          codes.add(loc.code);
          try {
            const msgs = await api.getI18nLocale(loc.code);
            this.#custom.set(loc.code, msgs);
          } catch {
            // skip unloadable custom locale
          }
        } else if (!codes.has(loc.code)) {
          merged.push(loc);
          codes.add(loc.code);
        }
      }
      this.locales = merged;
    } catch {
      this.locales = bundledList;
    }

    const available = this.locales.map((l) => l.code);
    const saved = localStorage.getItem(STORAGE_KEY);
    const initial = saved && available.includes(saved) ? saved : detectBrowserLocale(available);
    await this.setLocale(initial, false);
    this.applyOverrides({
      "app.title": brand.appName,
      "setup.welcome": `Welcome to ${brand.appName}`,
    });
    this.ready = true;
  }

  async setLocale(code: string, persist = true): Promise<void> {
    const bundled = await this.#ensureBundled(code);
    const custom = this.#custom.get(code);
    let remote: Messages | undefined;
    if (!bundled && !custom) {
      try {
        remote = await api.getI18nLocale(code);
        this.#custom.set(code, remote);
      } catch {
        if (code !== DEFAULT_LOCALE) {
          await this.setLocale(DEFAULT_LOCALE, persist);
          return;
        }
      }
    }
    if (bundled) {
      this.locales = this.locales.map((loc) =>
        loc.code === code && loc.source === "bundled"
          ? { ...loc, name: bundled.name || loc.name }
          : loc,
      );
    }
    const base = (await this.#ensureBundled(DEFAULT_LOCALE))?.messages ?? {};
    const next = mergeMessages(base, bundled?.messages ?? custom ?? remote ?? {});
    this.#messages = next;
    this.locale = code;
    if (persist) localStorage.setItem(STORAGE_KEY, code);
    document.documentElement.lang = code;
  }
}

export const i18n = new I18nStore();
