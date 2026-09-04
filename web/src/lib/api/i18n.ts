import type { I18nCatalog } from "./types";
import { request } from "./core";
import { opURL } from "./op";

export const i18nApi = {
  listI18nLocales: () => request<I18nCatalog>(opURL("GET__api_i18n_locales")),

  getI18nLocale: (code: string) =>
    request<Record<string, string>>(opURL("GET__api_i18n__locale", { locale: code })),

  getI18nTemplate: () => request<Record<string, string>>(opURL("GET__api_i18n_template")),
};
