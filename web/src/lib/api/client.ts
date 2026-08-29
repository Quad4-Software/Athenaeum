import type { HealthResponse } from "./types";
import { request, ApiError, restoreSession, ensureCsrf, clearCsrfCache } from "./core";
import { booksApi } from "./books";
import { librariesApi } from "./libraries";
import { collectionsApi } from "./collections";
import { authApi } from "./auth";
import { adminApi } from "./admin";
import { i18nApi } from "./i18n";
import { apiOp, apiPath, apiOperations } from "./generated/paths";

export { ApiError, restoreSession, ensureCsrf, clearCsrfCache, apiOp, apiPath, apiOperations };

export const api = {
  health: () => request<HealthResponse>(apiPath("/api/health")),
  ...booksApi,
  ...librariesApi,
  ...collectionsApi,
  ...authApi,
  ...adminApi,
  ...i18nApi,
};
