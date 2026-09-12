/**
 * In-app route path builders, the single source for constructing URLs the
 * router understands. The matching side (path -> route) lives in the pattern
 * table in router.svelte.ts; keep both in sync.
 */
export const routes = {
  library: () => "/",
  login: () => "/login",
  setup: () => "/setup",
  invite: (token: string) => `/invite/${encodeURIComponent(token)}`,
  settings: (tab?: string) => (tab ? `/settings/${encodeURIComponent(tab)}` : "/settings"),
  collections: () => "/collections",
  collection: (id: string | number) => `/collections/${encodeURIComponent(String(id))}`,
  book: (id: string | number) => `/book/${encodeURIComponent(String(id))}`,
  reader: (id: string | number) => `/read/${encodeURIComponent(String(id))}`,
  error: (code: string | number) => `/error/${encodeURIComponent(String(code))}`,
} as const;
