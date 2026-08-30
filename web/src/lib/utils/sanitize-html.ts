const DESCRIPTION_TAGS = new Set([
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
  "a",
]);

const READER_TAGS = new Set([
  ...DESCRIPTION_TAGS,
  "h5",
  "h6",
  "table",
  "thead",
  "tbody",
  "tr",
  "th",
  "td",
  "img",
  "hr",
  "pre",
  "code",
  "sub",
  "sup",
]);

const GLOBAL_ATTRS = new Set(["class", "title"]);
const TAG_ATTRS: Record<string, Set<string>> = {
  a: new Set(["href", "title"]),
  img: new Set(["src", "alt", "title", "width", "height"]),
};

const DROP_TAGS = new Set([
  "script",
  "style",
  "iframe",
  "object",
  "embed",
  "link",
  "meta",
  "form",
  "input",
  "button",
  "svg",
  "math",
]);

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

function isSafeImgSrc(src: string): boolean {
  const v = stripHrefNoise(src);
  return (
    v.startsWith("http://") ||
    v.startsWith("https://") ||
    v.startsWith("data:image/") ||
    v.startsWith("/")
  );
}

function sanitizeNode(node: Node, allowedTags: Set<string>): void {
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
    if (DROP_TAGS.has(tag) || !/^[a-z][a-z0-9]*$/.test(tag) || !allowedTags.has(tag)) {
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
      const name = attr.name.toLowerCase();
      if (!allowed.has(name)) {
        el.removeAttribute(attr.name);
        continue;
      }
      if (name === "href" && !isSafeHref(attr.value)) {
        el.removeAttribute(attr.name);
      }
      if (name === "src" && !isSafeImgSrc(attr.value)) {
        el.removeAttribute(attr.name);
      }
    }
    sanitizeNode(el, allowedTags);
    i++;
  }
}

function runSanitize(input: string, allowedTags: Set<string>): string {
  const raw = input.trim();
  if (!raw) return "";
  if (typeof DOMParser === "undefined") return raw.replace(/<[^>]+>/g, "");
  const doc = new DOMParser().parseFromString(raw, "text/html");
  sanitizeNode(doc.body, allowedTags);
  return doc.body.innerHTML.trim();
}

/** Strip unsafe markup; allow basic formatting tags for book descriptions. */
export function sanitizeHtml(input: string): string {
  return runSanitize(input, DESCRIPTION_TAGS);
}

/** Sanitize HTML from book body content (MOBI sections) before {@html}. */
export function sanitizeReaderHtml(input: string): string {
  return runSanitize(input, READER_TAGS);
}

export function descriptionLooksLikeHtml(input: string): boolean {
  return /<[a-z][\s\S]*>/i.test(input);
}
