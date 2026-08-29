export function errorCodeFromSlug(slug: string): number | "offline" {
  switch (slug) {
    case "unauthorized":
      return 401;
    case "forbidden":
      return 403;
    case "not-found":
      return 404;
    case "server":
      return 500;
    case "offline":
      return "offline";
    default: {
      const n = Number(slug);
      if (Number.isFinite(n) && n >= 400) return n;
      return 404;
    }
  }
}
