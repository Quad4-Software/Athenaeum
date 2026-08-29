import { resolveAppBase, stripBase, withBase } from "./router-base";

export type RouteName =
  | "library"
  | "book"
  | "reader"
  | "settings"
  | "login"
  | "setup"
  | "invite"
  | "collections"
  | "collection"
  | "error"
  | "notfound";

export interface Route {
  name: RouteName;
  params: Record<string, string>;
  path: string;
}

const patterns: { name: RouteName; regex: RegExp; keys: string[] }[] = [
  { name: "library", regex: /^\/$/, keys: [] },
  { name: "login", regex: /^\/login$/, keys: [] },
  { name: "setup", regex: /^\/setup$/, keys: [] },
  { name: "invite", regex: /^\/invite\/([^/]+)$/, keys: ["token"] },
  { name: "settings", regex: /^\/settings(?:\/([^/]+))?$/, keys: ["tab"] },
  { name: "collections", regex: /^\/collections$/, keys: [] },
  { name: "collection", regex: /^\/collections\/([^/]+)$/, keys: ["id"] },
  { name: "book", regex: /^\/book\/([^/]+)$/, keys: ["id"] },
  { name: "reader", regex: /^\/read\/([^/]+)$/, keys: ["id"] },
  { name: "error", regex: /^\/error\/([^/]+)$/, keys: ["code"] },
];

function pathnameOf(url: string): string {
  let path = url;
  const hash = path.indexOf("#");
  if (hash !== -1) path = path.slice(0, hash);
  const q = path.indexOf("?");
  return q === -1 ? path : path.slice(0, q);
}

function match(pathname: string): Route {
  for (const p of patterns) {
    const m = p.regex.exec(pathname);
    if (m) {
      const params: Record<string, string> = {};
      p.keys.forEach((k, i) => (params[k] = decodeURIComponent(m[i + 1])));
      return { name: p.name, params, path: pathname };
    }
  }
  return { name: "notfound", params: {}, path: pathname };
}

const appBase = resolveAppBase(window.location.pathname, import.meta.env.BASE_URL);

function appPathFromLocation(browserPath: string): string {
  return stripBase(pathnameOf(browserPath), appBase);
}

class Router {
  /** Mount path with no trailing slash. Empty when the app is at site root. */
  readonly base = appBase;

  current = $state<Route>(match(appPathFromLocation(window.location.pathname)));

  constructor() {
    window.addEventListener("popstate", () => {
      this.current = match(appPathFromLocation(window.location.pathname));
    });
  }

  /** Browser URL for an app-relative path (keeps query/hash). */
  href(path: string): string {
    return withBase(path, this.base);
  }

  /** App-relative pathname for the current browser location. */
  appPathname(browserPath = window.location.pathname): string {
    return appPathFromLocation(browserPath);
  }

  navigate(path: string, replace = false) {
    const target = this.href(path);
    const current = window.location.pathname + window.location.search;
    if (target === current) {
      this.current = match(appPathFromLocation(window.location.pathname));
      return;
    }
    if (replace) {
      window.history.replaceState({}, "", target);
    } else {
      window.history.pushState({}, "", target);
    }
    this.current = match(appPathFromLocation(pathnameOf(target)));
  }
}

export const router = new Router();

export function link(node: HTMLElement, href: string) {
  function applyHref(next: string) {
    current = next;
    if (node instanceof HTMLAnchorElement) {
      node.setAttribute("href", router.href(next));
    }
  }

  function onClick(event: MouseEvent) {
    if (event.defaultPrevented || event.button !== 0) return;
    if (event.metaKey || event.ctrlKey || event.shiftKey || event.altKey) return;
    event.preventDefault();
    router.navigate(current);
  }

  let current = href;
  applyHref(href);
  node.addEventListener("click", onClick);
  return {
    update(next: string) {
      applyHref(next);
    },
    destroy() {
      node.removeEventListener("click", onClick);
    },
  };
}
