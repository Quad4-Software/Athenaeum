import type { I18nCatalog } from "./types";
import { request } from "./core";

export const i18nApi = {
  listI18nLocales: () => request<I18nCatalog>("/api/i18n/locales"),

  getI18nLocale: (code: string) =>
    request<Record<string, string>>(`/api/i18n/${encodeURIComponent(code)}`),

  getI18nTemplate: () => request<Record<string, string>>("/api/i18n/template"),
};
