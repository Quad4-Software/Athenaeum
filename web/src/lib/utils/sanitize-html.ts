const ALLOWED_TAGS = new Set([
  "p",
  "br",
  "b",
  "i",
  "em",
  "strong",
  "h1",
  "h2",
  "h3",
  "h4",
  "ul",
  "ol",
  "li",
  "div",
  "span",
  "blockquote",
]);

const GLOBAL_ATTRS = new Set(["class", "title"]);
const TAG_ATTRS: Record<string, Set<string>> = {
  a: new Set(["href", "title"]),
};

function stripHrefNoise(href: string): string {
  let out = "";
  for (const ch of href.trim().toLowerCase()) {
    const code = ch.codePointAt(0) ?? 0;
    if (code < 32 || code === 127 || /\s/.test(ch)) continue;
    out += ch;
  }
  return out;
}

function isSafeHref(href: string): boolean {
  const v = stripHrefNoise(href);
  return v.startsWith("http://") || v.startsWith("https://") || v.startsWith("mailto:");
}

const DROP_TAGS = new Set(["script", "style", "iframe", "object", "embed", "link", "meta"]);

function sanitizeNode(node: Node): void {
  let i = 0;
  while (i < node.childNodes.length) {
    const child = node.childNodes[i];
    if (!child) break;
    if (child.nodeType === Node.TEXT_NODE) {
      i++;
      continue;
    }
    if (child.nodeType !== Node.ELEMENT_NODE) {
      child.remove();
      continue;
    }
    const el = child as HTMLElement;
    const tag = el.tagName.toLowerCase();
    if (DROP_TAGS.has(tag) || !/^[a-z][a-z0-9]*$/.test(tag) || !ALLOWED_TAGS.has(tag)) {
      if (DROP_TAGS.has(tag)) {
        el.remove();
        continue;
      }
      while (el.firstChild) {
        node.insertBefore(el.firstChild, el);
      }
      el.remove();
      continue;
    }
    const allowed = new Set(GLOBAL_ATTRS);
    const extra = TAG_ATTRS[tag];
    if (extra) for (const a of extra) allowed.add(a);
    for (const attr of [...el.attributes]) {
      if (!allowed.has(attr.name.toLowerCase())) {
        el.removeAttribute(attr.name);
        continue;
      }
      if (attr.name.toLowerCase() === "href" && !isSafeHref(attr.value)) {
        el.removeAttribute(attr.name);
      }
    }
    sanitizeNode(el);
    i++;
  }
}

/** Strip unsafe markup; allow basic formatting tags for book descriptions. */
export function sanitizeHtml(input: string): string {
  const raw = input.trim();
  if (!raw) return "";
  if (typeof DOMParser === "undefined") return raw.replace(/<[^>]+>/g, "");
  const doc = new DOMParser().parseFromString(raw, "text/html");
  sanitizeNode(doc.body);
  return doc.body.innerHTML.trim();
}

export function descriptionLooksLikeHtml(input: string): boolean {
  return /<[a-z][\s\S]*>/i.test(input);
}
