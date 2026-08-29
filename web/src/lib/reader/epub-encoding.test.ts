import { describe, expect, it } from "vitest";
import {
  decodeEpubChapterBytes,
  needsLegacyTextDecode,
  normalizeEncodingLabel,
  readDeclaredEncoding,
} from "./epub-encoding";

describe("normalizeEncodingLabel", () => {
  it("maps common legacy labels", () => {
    expect(normalizeEncodingLabel("UTF-8")).toBe("utf-8");
    expect(normalizeEncodingLabel("windows-1252")).toBe("windows-1252");
    expect(normalizeEncodingLabel("ISO-8859-1")).toBe("iso-8859-1");
  });
});

describe("decodeEpubChapterBytes", () => {
  it("decodes valid utf-8 unchanged", () => {
    const bytes = new TextEncoder().encode("<p>Hello — world</p>");
    expect(decodeEpubChapterBytes(bytes)).toBe("<p>Hello — world</p>");
  });

  it("decodes windows-1252 smart punctuation", () => {
    const bytes = Uint8Array.from([
      0x3c, 0x70, 0x3e, 0x44, 0x6f, 0x63, 0x74, 0x6f, 0x72, 0x73, 0x97, 0x61, 0x6e, 0x64, 0x3c,
      0x2f, 0x70, 0x3e,
    ]);
    expect(decodeEpubChapterBytes(bytes)).toBe("<p>Doctors—and</p>");
  });

  it("honors declared legacy encoding", () => {
    const header = '<?xml version="1.0" encoding="windows-1252"?>';
    const body = Uint8Array.from([0x65, 0xe9, 0x6e]);
    const bytes = new Uint8Array([...new TextEncoder().encode(header), ...body]);
    expect(decodeEpubChapterBytes(bytes)).toContain("é");
  });

  it("falls back to iso-8859-1 when utf-8 produces replacement characters", () => {
    const bytes = Uint8Array.from([0xe9]); // é in latin-1/cp1252, invalid alone in utf-8
    expect(decodeEpubChapterBytes(bytes)).toBe("é");
  });
});

describe("needsLegacyTextDecode", () => {
  it("detects replacement characters", () => {
    expect(needsLegacyTextDecode("Charg\uFFFD Affaires")).toBe(true);
    expect(needsLegacyTextDecode("Chargé d'Affaires")).toBe(false);
  });

  it("returns false for empty strings", () => {
    expect(needsLegacyTextDecode("")).toBe(false);
  });
});

describe("readDeclaredEncoding", () => {
  it("reads xml encoding declarations", () => {
    const bytes = new TextEncoder().encode('<?xml version="1.0" encoding="ISO-8859-1"?>');
    expect(readDeclaredEncoding(bytes)).toBe("iso-8859-1");
  });

  it("reads html charset meta tags", () => {
    const bytes = new TextEncoder().encode('<meta charset="windows-1252">');
    expect(readDeclaredEncoding(bytes)).toBe("windows-1252");
  });

  it("returns null when no declaration exists", () => {
    const bytes = new TextEncoder().encode("<html><body>Hi</body></html>");
    expect(readDeclaredEncoding(bytes)).toBeNull();
  });
});
