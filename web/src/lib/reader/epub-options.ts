import type { BookOptions } from "epubjs/types/book";
import type { RenditionOptions } from "epubjs/types/rendition";

/** epubjs treats extensionless URLs as unpacked EPUB directories; API file URLs need binary mode. */
export function epubOpenOptions(): BookOptions {
  return {
    openAs: "epub",
    requestCredentials: true,
  } as unknown as BookOptions;
}

/**
 * Blob URLs for spine sections. allow-scripts is required so epubjs can attach
 * link handlers inside the iframe. hardenEpubDocument strips book scripts first.
 */
export function epubRenderOptions(spread: "none" | "auto" | "always"): RenditionOptions {
  return {
    width: "100%",
    height: "100%",
    flow: "paginated",
    spread,
    method: "blobUrl",
    allowScriptedContent: true,
  } as RenditionOptions;
}

/** epubjs injectIdentifier expects a head element in every spine document. */
export function ensureEpubDocumentHead(doc: Document) {
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

export { hardenEpubDocument } from "./epub-harden";
