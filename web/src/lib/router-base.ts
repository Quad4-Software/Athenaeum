/**
 * Resolve the SPA mount path for root installs and subdirectory demos
 * (for example /demo or /reader/demo on GitHub Pages).
 */

const ROUTE_SUFFIX_PATTERNS: RegExp[] = [
  /\/book\/[^/]+\/?$/,
  /\/read\/[^/]+\/?$/,
  /\/collections\/[^/]+\/?$/,
  /\/invite\/[^/]+\/?$/,
  /\/error\/[^/]+\/?$/,
  /\/settings(?:\/[^/]+)?\/?$/,
  /\/collections\/?$/,
  /\/login\/?$/,
  /\/setup\/?$/,
];

/** Absolute Vite base without trailing slash, or null when base is relative. */
export function normalizeViteBase(viteBaseUrl: string): string | null {
  const raw = (viteBaseUrl || "/").trim();
  if (raw === "./" || raw === "." || raw === "") return null;
  if (!raw.startsWith("/")) return null;
  return raw.replace(/\/+$/, "") || "";
}

/**
 * When Vite base is relative (demo builds), treat the directory that served
 * index.html as the mount path. Known app routes are stripped first.
 */
export function inferBaseFromPathname(pathname: string): string {
  let path = pathname;
  const hash = path.indexOf("#");
  if (hash !== -1) path = path.slice(0, hash);
  const q = path.indexOf("?");
  if (q !== -1) path = path.slice(0, q);
  if (!path.startsWith("/")) path = `/${path}`;

  for (const re of ROUTE_SUFFIX_PATTERNS) {
    if (!re.test(path)) continue;
    const stripped = path.replace(re, "");
    return stripped.replace(/\/+$/, "") || "";
  }

  if (path === "/" || path === "") return "";
  return path.replace(/\/+$/, "");
}

/** App mount path with no trailing slash. Empty string means site root. */
export function resolveAppBase(pathname: string, viteBaseUrl: string): string {
  const fromVite = normalizeViteBase(viteBaseUrl);
  if (fromVite !== null) return fromVite;
  return inferBaseFromPathname(pathname);
}

/** Map a browser pathname to an app-relative path starting with /. */
export function stripBase(pathname: string, base: string): string {
  const path = pathname || "/";
  if (!base) return path.startsWith("/") ? path : `/${path}`;
  if (path === base || path === `${base}/`) return "/";
  if (path.startsWith(`${base}/`)) {
    const rest = path.slice(base.length);
    return rest.startsWith("/") ? rest : `/${rest}`;
  }
  return path.startsWith("/") ? path : `/${path}`;
}

/** Prefix an app-relative path (optional query) with the mount base. */
export function withBase(appPath: string, base: string): string {
  const hash = appPath.indexOf("#");
  const withoutHash = hash === -1 ? appPath : appPath.slice(0, hash);
  const hashPart = hash === -1 ? "" : appPath.slice(hash);

  const q = withoutHash.indexOf("?");
  const pathPart = q === -1 ? withoutHash : withoutHash.slice(0, q);
  const search = q === -1 ? "" : withoutHash.slice(q);

  let path = pathPart || "/";
  if (!path.startsWith("/")) path = `/${path}`;

  if (!base) return `${path}${search}${hashPart}`;
  if (path === "/") return `${base}/${search}${hashPart}`;
  return `${base}${path}${search}${hashPart}`;
}
