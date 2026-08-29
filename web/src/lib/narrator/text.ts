const BLOCK_SELECTOR = "p, li, h1, h2, h3, h4, h5, h6, blockquote, dd, figcaption, td, pre, dt";

/** Default utterance length for browser and Kokoro (chars). */
export const UTTERANCE_MAX_CHARS = 400;

/** Hard ceiling matching Kokoro synthesize limits. Never exceed this per speak. */
export const UTTERANCE_HARD_MAX = 1800;

/** Normalize whitespace for TTS. */
export function normalizeNarratorText(raw: string): string {
  return raw.replace(/\s+/g, " ").trim();
}

/**
 * Collect block-level text from an EPUB iframe document.
 * Falls back to body text when no block elements have content.
 */
export function paragraphsFromDocument(doc: Document | null | undefined): string[] {
  if (!doc?.body) return [];
  const blocks = doc.body.querySelectorAll(BLOCK_SELECTOR);
  const out: string[] = [];
  for (const el of blocks) {
    const t = normalizeNarratorText(el.textContent ?? "");
    if (t.length > 0) out.push(t);
  }
  if (out.length === 0) {
    const t = normalizeNarratorText(doc.body.textContent ?? "");
    if (t) out.push(t);
  }
  return out;
}

export interface EpubContentsLike {
  document?: Document;
  window?: Window;
}

/**
 * Paragraphs from the current EPUB view, starting at the first block that
 * intersects the viewport (so "read from here" skips scrolled-past text).
 * Never walks the whole book — only the loaded content documents.
 */
export function paragraphsFromContents(contents: EpubContentsLike[]): string[] {
  const out: string[] = [];
  for (const c of contents) {
    const doc = c.document;
    if (!doc?.body) continue;
    const win = c.window;
    const blocks = Array.from(doc.body.querySelectorAll(BLOCK_SELECTOR));
    let started = !win;
    for (const el of blocks) {
      if (!started && win) {
        try {
          const rect = (el as HTMLElement).getBoundingClientRect();
          if (rect.bottom <= 0) continue;
          started = true;
        } catch {
          started = true;
        }
      }
      const t = normalizeNarratorText(el.textContent ?? "");
      if (t.length > 0) out.push(t);
    }
    if (out.length === 0) {
      const t = normalizeNarratorText(doc.body.textContent ?? "");
      if (t) out.push(t);
    }
  }
  return out;
}

/**
 * Split text into speakable utterances. Prefer sentence boundaries, then
 * soft-wrap long runs so browser and Kokoro stay within practical limits.
 */
export function splitIntoUtterances(text: string, maxLen = UTTERANCE_MAX_CHARS): string[] {
  const limit = Math.min(Math.max(32, maxLen), UTTERANCE_HARD_MAX);
  const normalized = normalizeNarratorText(text);
  if (!normalized) return [];
  if (normalized.length <= limit) return [normalized];

  const sentences = normalized.match(/[^.!?…]+[.!?…]+|[^.!?…]+$/g) ?? [normalized];
  const chunks: string[] = [];
  let buf = "";

  const flush = () => {
    const t = buf.trim();
    if (t) chunks.push(t);
    buf = "";
  };

  for (const sentence of sentences) {
    const s = sentence.trim();
    if (!s) continue;
    if (s.length > limit) {
      flush();
      for (const piece of softWrap(s, limit)) chunks.push(piece);
      continue;
    }
    if (!buf) {
      buf = s;
      continue;
    }
    if (buf.length + 1 + s.length <= limit) {
      buf = `${buf} ${s}`;
    } else {
      flush();
      buf = s;
    }
  }
  flush();
  return chunks;
}

function softWrap(text: string, maxLen: number): string[] {
  const words = text.split(" ");
  const out: string[] = [];
  let buf = "";

  const pushBuf = () => {
    if (buf) {
      out.push(buf);
      buf = "";
    }
  };

  for (const w of words) {
    if (w.length > maxLen) {
      pushBuf();
      for (let i = 0; i < w.length; i += maxLen) {
        out.push(w.slice(i, i + maxLen));
      }
      continue;
    }
    if (!buf) {
      buf = w;
      continue;
    }
    if (buf.length + 1 + w.length <= maxLen) {
      buf = `${buf} ${w}`;
    } else {
      out.push(buf);
      buf = w;
    }
  }
  pushBuf();
  return out;
}

/**
 * Expand paragraphs into an utterance queue.
 * Caps each piece under UTTERANCE_HARD_MAX so models never see a whole chapter.
 */
export function buildUtteranceQueue(paragraphs: string[], maxLen = UTTERANCE_MAX_CHARS): string[] {
  const limit = Math.min(Math.max(32, maxLen), UTTERANCE_HARD_MAX);
  const queue: string[] = [];
  for (const p of paragraphs) {
    for (const u of splitIntoUtterances(p, limit)) {
      if (u.length <= UTTERANCE_HARD_MAX) {
        queue.push(u);
      } else {
        queue.push(...softWrap(u, UTTERANCE_HARD_MAX));
      }
    }
  }
  return queue;
}
