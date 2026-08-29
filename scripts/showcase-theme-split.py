#!/usr/bin/env python3
"""Diagonal composite of library-dark (top-left) and library-light (bottom-right)."""
from pathlib import Path
import sys

from PIL import Image
import numpy as np

def main() -> int:
    if len(sys.argv) != 4:
        print("usage: showcase-theme-split.py dark.png light.png out.png", file=sys.stderr)
        return 2
    dark_path, light_path, out_path = map(Path, sys.argv[1:])
    size = (1440, 900)
    dark = Image.open(dark_path).convert("RGB").resize(size, Image.Resampling.LANCZOS)
    light = Image.open(light_path).convert("RGB").resize(size, Image.Resampling.LANCZOS)
    w, h = size
    ys, xs = np.mgrid[0:h, 0:w]
    # Soft diagonal band (~24px) from top-right to bottom-left style: dark on left.
    t = (xs / (w - 1) + ys / (h - 1)) / 2.0
    soft = 12 / max(w, h)
    alpha = np.clip((t - 0.5) / soft + 0.5, 0.0, 1.0)[..., None]
    d = np.asarray(dark, dtype=np.float32)
    l = np.asarray(light, dtype=np.float32)
    out = d * (1.0 - alpha) + l * alpha
    Image.fromarray(np.clip(out, 0, 255).astype(np.uint8), "RGB").save(out_path)
    print(f"wrote {out_path}")
    return 0

if __name__ == "__main__":
    raise SystemExit(main())
