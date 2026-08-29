const REPLACEMENT = "\uFFFD";

const ENCODING_RE = /encoding\s*=\s*['"]([^'"]+)['"]|charset\s*=\s*['"]?([^'">\s]+)/i;

export function readDeclaredEncoding(bytes: Uint8Array): string | null {
  const head = new TextDecoder("ascii").decode(bytes.subarray(0, 512));
  const match = head.match(ENCODING_RE);
  const raw = match?.[1] ?? match?.[2];
  if (!raw) return null;
  return normalizeEncodingLabel(raw);
}

export function normalizeEncodingLabel(label: string): string {
  const key = label.trim().toLowerCase().replace(/[_-]/g, "");
  switch (key) {
    case "utf8":
      return "utf-8";
    case "latin1":
    case "iso88591":
      return "iso-8859-1";
    case "cp1252":
    case "windows1252":
      return "windows-1252";
    default:
      return label.trim();
  }
}

export function needsLegacyTextDecode(text: string): boolean {
  if (!text) return false;
  if (text.includes(REPLACEMENT)) return true;
  return false;
}

export function decodeEpubChapterBytes(bytes: Uint8Array): string {
  const declared = readDeclaredEncoding(bytes);
  if (declared && declared !== "utf-8") {
    try {
      return new TextDecoder(declared).decode(bytes);
    } catch {
      /* fall through */
    }
  }

  const utf8 = new TextDecoder("utf-8").decode(bytes);
  if (!needsLegacyTextDecode(utf8)) return utf8;

  for (const legacy of ["windows-1252", "iso-8859-1"]) {
    try {
      const decoded = new TextDecoder(legacy).decode(bytes);
      if (!decoded.includes(REPLACEMENT)) return decoded;
    } catch {
      /* try next */
    }
  }

  return utf8;
}

type ArchiveLike = {
  getText: (url: string, encoding?: string) => Promise<string | undefined>;
  zip?: { file: (path: string) => { async: (type: string) => Promise<Uint8Array> } | null };
};

export function patchEpubArchiveEncoding(book: { archive?: ArchiveLike | null }) {
  const archive = book.archive;
  if (!archive?.zip) return;

  const origGetText = archive.getText.bind(archive);
  archive.getText = (url: string, encoding?: string) => {
    const decodedUrl = decodeURIComponent(url.startsWith("/") ? url.slice(1) : url);
    const entry = archive.zip!.file(decodedUrl);
    if (!entry) return origGetText(url, encoding);

    return entry.async("uint8array").then((bytes) => decodeEpubChapterBytes(bytes));
  };
}
