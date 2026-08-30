/**
 * Harden EPUB spine HTML for XSS without breaking epubjs.
 * Call from spine.hooks.content only (before blob serialize).
 * Do not run on the live iframe: that strips epubjs link onclick handlers.
 * Keeps <base> (resource resolution), styles, SVG, and text.
 * Strips scripts, embeds, event handlers, and dangerous URLs.
 * Does not inject CSP: script-src/base-uri policies break epubjs blob iframes.
 */

const DROP_TAGS = new Set(["script", "iframe", "object", "embed", "form", "input", "button"]);

/**
 * Harden an EPUB spine document before epubjs serializes it into a blob iframe.
 * Keeps text/layout for CFI, highlights, narration, styles, and SVG art.
 */
export function hardenEpubDocument(doc: Document): void {
  ensureHead(doc);
  stripHostileNodes(doc);
  hardenAnchors(doc);
}

function ensureHead(doc: Document): void {
  if (doc.getElementsByTagName("head")[0]) return;
  const head = doc.createElement("head");
  const root = doc.documentElement;
  if (!root) return;
  const body = doc.body;
  if (body?.parentNode === root) {
    root.insertBefore(head, body);
  } else {
    root.appendChild(head);
  }
}

function stripHostileNodes(doc: Document): void {
  const root = doc.documentElement;
  if (!root) return;
  const walker = doc.createTreeWalker(root, NodeFilter.SHOW_ELEMENT);
  const doomed: Element[] = [];
  let node = walker.nextNode();
  while (node) {
    const el = node as Element;
    const tag = el.tagName.toLowerCase();
    if (DROP_TAGS.has(tag) || shouldDropLink(el) || shouldDropMeta(el)) {
      doomed.push(el);
    } else {
      for (const attr of [...el.attributes]) {
        const name = attr.name.toLowerCase();
        if (name.startsWith("on") || name === "srcdoc") {
          el.removeAttribute(attr.name);
        }
        if (
          (name === "href" || name === "src" || name === "xlink:href") &&
          isUnsafeURL(attr.value)
        ) {
          el.removeAttribute(attr.name);
        }
      }
    }
    node = walker.nextNode();
  }
  for (const el of doomed) el.remove();
}

function shouldDropLink(el: Element): boolean {
  if (el.tagName.toLowerCase() !== "link") return false;
  const rel = (el.getAttribute("rel") ?? "").toLowerCase();
  if (rel.includes("stylesheet") || rel.includes("preload") || rel.includes("canonical")) {
    const href = el.getAttribute("href") ?? "";
    return isUnsafeURL(href);
  }
  return true;
}

function shouldDropMeta(el: Element): boolean {
  if (el.tagName.toLowerCase() !== "meta") return false;
  const httpEquiv = (el.getAttribute("http-equiv") ?? "").toLowerCase();
  if (!httpEquiv) return false;
  return httpEquiv === "refresh" || httpEquiv === "set-cookie";
}

function isUnsafeURL(raw: string): boolean {
  const v = raw
    .trim()
    .toLowerCase()
    .replace(/[\u0000-\u001f\u007f\s]+/g, "");
  return (
    v.startsWith("javascript:") ||
    v.startsWith("vbscript:") ||
    v.startsWith("data:text/html") ||
    v.startsWith("data:application/xhtml")
  );
}

function hardenAnchors(doc: Document): void {
  for (const a of doc.querySelectorAll("a[href]")) {
    const href = a.getAttribute("href") ?? "";
    if (/^https?:\/\//i.test(href)) {
      a.setAttribute("target", "_blank");
      a.setAttribute("rel", "noopener noreferrer");
    }
  }
}
