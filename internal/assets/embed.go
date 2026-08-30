// Package assets embeds the production frontend build so the application
// can ship as a single self-contained binary. When --web-dir is set the
// server serves that directory instead. The Vite build writes its output
// into the dist directory here (see web/vite.config.ts).
package assets

import (
	"embed"
	"errors"
	"io/fs"
)

//go:embed all:dist
var distEmbed embed.FS

//go:embed all:fallback
var fallbackEmbed embed.FS

// DistFS contains the built single-page application under dist/.
// Vite-generated index.html is gitignored; fallback/index.html is used
// when dist/index.html is missing (fresh clone / go-only CI builds).
var DistFS fs.FS = overlayFS{primary: distEmbed, fallback: fallbackEmbed}

type overlayFS struct {
	primary  fs.FS
	fallback fs.FS
}

func (o overlayFS) Open(name string) (fs.File, error) {
	f, err := o.primary.Open(name)
	if err == nil {
		return f, nil
	}
	if !errors.Is(err, fs.ErrNotExist) {
		return nil, err
	}
	if name == "dist/index.html" {
		return o.fallback.Open("fallback/index.html")
	}
	return nil, err
}

func (o overlayFS) ReadDir(name string) ([]fs.DirEntry, error) {
	if rd, ok := o.primary.(fs.ReadDirFS); ok {
		entries, err := rd.ReadDir(name)
		if err == nil {
			return entries, nil
		}
		if !errors.Is(err, fs.ErrNotExist) {
			return nil, err
		}
	}
	if name == "dist" {
		if rd, ok := o.fallback.(fs.ReadDirFS); ok {
			fb, err := rd.ReadDir("fallback")
			if err != nil {
				return nil, err
			}
			out := make([]fs.DirEntry, 0, len(fb))
			for _, e := range fb {
				if e.Name() == "index.html" {
					out = append(out, e)
				}
			}
			return out, nil
		}
	}
	return nil, &fs.PathError{Op: "readdir", Path: name, Err: fs.ErrNotExist}
}
