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

class Router {
  current = $state<Route>(match(window.location.pathname));

  constructor() {
    window.addEventListener("popstate", () => {
      this.current = match(window.location.pathname);
    });
  }

  navigate(path: string, replace = false) {
    const hash = path.indexOf("#");
    const target = hash === -1 ? path : path.slice(0, hash);
    const current = window.location.pathname + window.location.search;
    if (target === current) {
      this.current = match(window.location.pathname);
      return;
    }
    if (replace) {
      window.history.replaceState({}, "", target);
    } else {
      window.history.pushState({}, "", target);
    }
    this.current = match(pathnameOf(target));
  }
}

export const router = new Router();

export function link(node: HTMLElement, href: string) {
  function onClick(event: MouseEvent) {
    if (event.defaultPrevented || event.button !== 0) return;
    if (event.metaKey || event.ctrlKey || event.shiftKey || event.altKey) return;
    event.preventDefault();
    router.navigate(current);
  }
  let current = href;
  node.addEventListener("click", onClick);
  return {
    update(next: string) {
      current = next;
    },
    destroy() {
      node.removeEventListener("click", onClick);
    },
  };
}
