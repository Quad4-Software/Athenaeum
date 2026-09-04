#!/usr/bin/env bash
# Download Kokoro ONNX weights for embedding in the full web build.
# Slim builds skip these files. Voices come from the kokoro-js npm package at build time.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
DEST="${KOKORO_MODEL_DIR:-$ROOT/web/vendor/kokoro-model/Kokoro-82M-v1.0-ONNX}"
BASE="${KOKORO_MODEL_BASE:-https://huggingface.co/onnx-community/Kokoro-82M-v1.0-ONNX/resolve/main}"
# q8 weights (~92 MiB). fp32 is ~310 MiB and is not shipped.
ONNX_FILE="onnx/model_quantized.onnx"

mkdir -p "$DEST/onnx"

fetch() {
  local rel="$1"
  local out="$DEST/$rel"
  if [[ -f "$out" && -s "$out" ]]; then
    echo "skip $rel"
    return 0
  fi
  echo "fetch $rel"
  mkdir -p "$(dirname "$out")"
  curl -fL --retry 3 --retry-delay 2 -o "$out.partial" "$BASE/$rel"
  mv "$out.partial" "$out"
}

fetch "config.json"
fetch "tokenizer.json"
fetch "tokenizer_config.json"
fetch "$ONNX_FILE"

bytes="$(wc -c <"$DEST/$ONNX_FILE" | tr -d ' ')"
if [[ "$bytes" -lt 1000000 ]]; then
  echo "error: $ONNX_FILE looks too small ($bytes bytes); delete and re-run" >&2
  exit 1
fi

echo "kokoro model ready under $DEST ($(du -sh "$DEST" | awk '{print $1}'))"
