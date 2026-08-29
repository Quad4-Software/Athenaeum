#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
DEST="${1:-$ROOT/library/samples}"
mkdir -p "$DEST"

write_epub() {
  local out="$1"
  local title="$2"
  local author="$3"
  if [[ -f "$out" && -s "$out" ]]; then
    echo "skip $out"
    return 0
  fi
  echo "write $out"
  local tmp
  tmp="$(mktemp -d)"
  printf 'application/epub+zip' >"$tmp/mimetype"
  mkdir -p "$tmp/META-INF" "$tmp/OEBPS"
  cat >"$tmp/META-INF/container.xml" <<'EOF'
<?xml version="1.0"?>
<container version="1.0" xmlns="urn:oasis:names:tc:opendocument:xmlns:container">
  <rootfiles>
    <rootfile full-path="OEBPS/content.opf" media-type="application/oebps-package+xml"/>
  </rootfiles>
</container>
EOF
  cat >"$tmp/OEBPS/content.opf" <<EOF
<?xml version="1.0"?>
<package xmlns="http://www.idpf.org/2007/opf" version="3.0">
  <metadata xmlns:dc="http://purl.org/dc/elements/1.1/">
    <dc:title>${title}</dc:title>
    <dc:creator>${author}</dc:creator>
    <dc:language>en</dc:language>
  </metadata>
  <manifest>
    <item id="chapter" href="chapter.xhtml" media-type="application/xhtml+xml"/>
  </manifest>
  <spine>
    <itemref idref="chapter"/>
  </spine>
</package>
EOF
  cat >"$tmp/OEBPS/chapter.xhtml" <<EOF
<?xml version="1.0" encoding="UTF-8"?>
<html xmlns="http://www.w3.org/1999/xhtml">
  <head><title>${title}</title></head>
  <body><p>Sample chapter for ${title}.</p></body>
</html>
EOF
  (cd "$tmp" && zip -X0 "$out" mimetype >/dev/null && zip -Xr "$out" META-INF OEBPS >/dev/null)
  rm -rf "$tmp"
}

write_pdf() {
  local out="$1"
  local title="$2"
  local author="$3"
  if [[ -f "$out" && -s "$out" ]]; then
    echo "skip $out"
    return 0
  fi
  echo "write $out"
  cat >"$out" <<EOF
%PDF-1.4
1 0 obj<</Type/Catalog/Pages 2 0 R>>endobj
2 0 obj<</Type/Pages/Kids[3 0 R]/Count 1>>endobj
3 0 obj<</Type/Page/MediaBox[0 0 612 792]/Parent 2 0 R>>endobj
trailer<</Root 1 0 R/Info<</Title (${title})/Author (${author})>>>>
%%EOF
EOF
}

echo "Generating local sample books in $DEST ..."
write_epub "$DEST/Alice in Wonderland.epub" "Alice's Adventures in Wonderland" "Lewis Carroll"
write_epub "$DEST/Pride and Prejudice.epub" "Pride and Prejudice" "Jane Austen"
write_pdf "$DEST/Sherlock Holmes - Adventures.pdf" "The Adventures of Sherlock Holmes" "Arthur Conan Doyle"
write_pdf "$DEST/Frankenstein.pdf" "Frankenstein" "Mary Shelley"
write_pdf "$DEST/Popular Science 1931-01.pdf" "Popular Science 1931-01" "Popular Science"

if ! command -v ffmpeg >/dev/null 2>&1; then
  echo "ffmpeg not found; skipping sample MP3 (optional)"
else
  mp3="$DEST/Pride and Prejudice - Chapter 01.mp3"
  if [[ ! -f "$mp3" || ! -s "$mp3" ]]; then
    echo "write $mp3"
    ffmpeg -hide_banner -loglevel error \
      -f lavfi -i "sine=frequency=440:duration=3" \
      -metadata title="Pride and Prejudice - Chapter 01" \
      -metadata artist="Jane Austen" \
      -y "$mp3"
  else
    echo "skip $mp3"
  fi
fi

download() {
  local url="$1"
  local out="$2"
  echo "try remote -> $(basename "$out")"
  if curl -fsSL \
    --connect-timeout 10 \
    --max-time 120 \
    --retry 0 \
    --speed-time 30 \
    --speed-limit 1024 \
    "$url" -o "$out.tmp"; then
    mv "$out.tmp" "$out"
    echo "ok $(basename "$out")"
    return 0
  fi
  rm -f "$out.tmp"
  echo "skip remote $(basename "$out")"
  return 1
}

if [[ "${FETCH_REMOTE:-0}" == "1" ]]; then
  echo "Optional remote fetch enabled (FETCH_REMOTE=1) ..."
  download "https://www.gutenberg.org/ebooks/11.epub3.images" "$DEST/Alice in Wonderland.epub" || true
  download "https://www.gutenberg.org/ebooks/1342.epub3.images" "$DEST/Pride and Prejudice.epub" || true
fi

echo ""
echo "Samples ready in $DEST:"
ls -lh "$DEST"
echo ""
echo "Next: task run -- --library $DEST"
