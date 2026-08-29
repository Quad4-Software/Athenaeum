package demo

import (
	"bytes"
	"hash/fnv"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"
)

var palette = []color.RGBA{
	{0x8B, 0x1E, 0x1E, 0xFF},
	{0x1E, 0x3A, 0x5F, 0xFF},
	{0x2F, 0x4F, 0x3E, 0xFF},
	{0x5C, 0x3A, 0x21, 0xFF},
	{0x3D, 0x2B, 0x4F, 0xFF},
	{0x1F, 0x4E, 0x5F, 0xFF},
	{0x6B, 0x2D, 0x3C, 0xFF},
	{0x2C, 0x3E, 0x50, 0xFF},
}

func bumpChannel(c uint8, delta int) uint8 {
	v := int(c) + delta
	if v > 255 {
		return 255
	}
	if v < 0 {
		return 0
	}
	return uint8(v) // #nosec G115 -- clamped to 0..255
}

func colorFor(s string) (bg, accent color.RGBA) {
	h := fnv.New32a()
	_, _ = h.Write([]byte(s))
	n := h.Sum32()
	bg = palette[int(n)%len(palette)]
	accent = color.RGBA{
		R: bumpChannel(bg.R, 40),
		G: bumpChannel(bg.G, 35),
		B: bumpChannel(bg.B, 30),
		A: 255,
	}
	return bg, accent
}

// WriteCoverPNG writes a generated cover image for the given title/author.
func WriteCoverPNG(path, title, author string) error {
	data, err := EncodeCoverBuffer(title, author)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}

func truncate(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n-1]) + "..."
}

// Tiny 5x7 glyph map for ASCII cover labels (no external font dependency).
var glyphs = map[rune][]string{
	'A':  {"01110", "10001", "10001", "11111", "10001", "10001", "10001"},
	'B':  {"11110", "10001", "10001", "11110", "10001", "10001", "11110"},
	'C':  {"01110", "10001", "10000", "10000", "10000", "10001", "01110"},
	'D':  {"11110", "10001", "10001", "10001", "10001", "10001", "11110"},
	'E':  {"11111", "10000", "10000", "11110", "10000", "10000", "11111"},
	'F':  {"11111", "10000", "10000", "11110", "10000", "10000", "10000"},
	'G':  {"01110", "10001", "10000", "10111", "10001", "10001", "01110"},
	'H':  {"10001", "10001", "10001", "11111", "10001", "10001", "10001"},
	'I':  {"11111", "00100", "00100", "00100", "00100", "00100", "11111"},
	'J':  {"00111", "00010", "00010", "00010", "00010", "10010", "01100"},
	'K':  {"10001", "10010", "10100", "11000", "10100", "10010", "10001"},
	'L':  {"10000", "10000", "10000", "10000", "10000", "10000", "11111"},
	'M':  {"10001", "11011", "10101", "10001", "10001", "10001", "10001"},
	'N':  {"10001", "11001", "10101", "10011", "10001", "10001", "10001"},
	'O':  {"01110", "10001", "10001", "10001", "10001", "10001", "01110"},
	'P':  {"11110", "10001", "10001", "11110", "10000", "10000", "10000"},
	'Q':  {"01110", "10001", "10001", "10001", "10101", "10010", "01101"},
	'R':  {"11110", "10001", "10001", "11110", "10100", "10010", "10001"},
	'S':  {"01111", "10000", "10000", "01110", "00001", "00001", "11110"},
	'T':  {"11111", "00100", "00100", "00100", "00100", "00100", "00100"},
	'U':  {"10001", "10001", "10001", "10001", "10001", "10001", "01110"},
	'V':  {"10001", "10001", "10001", "10001", "10001", "01010", "00100"},
	'W':  {"10001", "10001", "10001", "10001", "10101", "11011", "10001"},
	'X':  {"10001", "10001", "01010", "00100", "01010", "10001", "10001"},
	'Y':  {"10001", "10001", "01010", "00100", "00100", "00100", "00100"},
	'Z':  {"11111", "00001", "00010", "00100", "01000", "10000", "11111"},
	'0':  {"01110", "10001", "10011", "10101", "11001", "10001", "01110"},
	'1':  {"00100", "01100", "00100", "00100", "00100", "00100", "01110"},
	'2':  {"01110", "10001", "00001", "00010", "00100", "01000", "11111"},
	'3':  {"11110", "00001", "00001", "01110", "00001", "00001", "11110"},
	'4':  {"00010", "00110", "01010", "10010", "11111", "00010", "00010"},
	'5':  {"11111", "10000", "11110", "00001", "00001", "10001", "01110"},
	'6':  {"01110", "10000", "10000", "11110", "10001", "10001", "01110"},
	'7':  {"11111", "00001", "00010", "00100", "01000", "01000", "01000"},
	'8':  {"01110", "10001", "10001", "01110", "10001", "10001", "01110"},
	'9':  {"01110", "10001", "10001", "01111", "00001", "00001", "01110"},
	' ':  {"00000", "00000", "00000", "00000", "00000", "00000", "00000"},
	'-':  {"00000", "00000", "00000", "11111", "00000", "00000", "00000"},
	'.':  {"00000", "00000", "00000", "00000", "00000", "01100", "01100"},
	',':  {"00000", "00000", "00000", "00000", "01100", "00100", "01000"},
	':':  {"00000", "01100", "01100", "00000", "01100", "01100", "00000"},
	'\'': {"00100", "00100", "01000", "00000", "00000", "00000", "00000"},
}

func drawLabel(img *image.RGBA, x, y int, text string, c color.RGBA) {
	cx := x
	for _, r := range text {
		g, ok := glyphs[r]
		if !ok {
			if r >= 'a' && r <= 'z' {
				g, ok = glyphs[r-32]
			}
		}
		if !ok {
			g = glyphs[' ']
		}
		for row, line := range g {
			for col, bit := range line {
				if bit == '1' {
					img.Set(cx+col*2, y+row*2, c)
					img.Set(cx+col*2+1, y+row*2, c)
					img.Set(cx+col*2, y+row*2+1, c)
					img.Set(cx+col*2+1, y+row*2+1, c)
				}
			}
		}
		cx += 12
		if cx > img.Bounds().Max.X-20 {
			break
		}
	}
}

// EncodeCoverBuffer encodes a cover into a buffer without touching disk.
func EncodeCoverBuffer(title, author string) ([]byte, error) {
	const w, h = 400, 600
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	bg, accent := colorFor(title + "|" + author)
	draw.Draw(img, img.Bounds(), &image.Uniform{C: bg}, image.Point{}, draw.Src)
	draw.Draw(img, image.Rect(0, 80, w, 160), &image.Uniform{C: accent}, image.Point{}, draw.Src)
	draw.Draw(img, image.Rect(36, 200, 56, h-80), &image.Uniform{C: color.RGBA{R: 240, G: 240, B: 235, A: 255}}, image.Point{}, draw.Src)
	draw.Draw(img, image.Rect(70, 220, w-40, h-100), &image.Uniform{C: color.RGBA{R: 20, G: 18, B: 16, A: 180}}, image.Point{}, draw.Src)
	drawLabel(img, 84, 260, truncate(title, 22), color.RGBA{R: 250, G: 246, B: 240, A: 255})
	drawLabel(img, 84, 300, truncate(author, 24), color.RGBA{R: 200, G: 190, B: 175, A: 255})
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
