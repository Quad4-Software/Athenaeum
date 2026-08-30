package server

import (
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"athenaeum/internal/assets"
)

// spaHandler serves the single-page application from webDir when set,
// otherwise from the embedded production build. Unknown paths that are
// not static assets fall back to index.html so client-side routing works
// on deep links and refreshes.
func spaHandler(webDir string) (http.Handler, error) {
	root, err := spaRoot(webDir)
	if err != nil {
		return nil, err
	}
	fileServer := http.FileServer(http.FS(root))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clean := strings.TrimPrefix(path.Clean(r.URL.Path), "/")
		if clean == "" {
			clean = "index.html"
		}

		if _, err := fs.Stat(root, clean); err != nil {
			if looksLikeStaticAsset(clean) {
				writeHTMLError(w, http.StatusNotFound, "Not found",
					"The requested file was not found.")
				return
			}
			r2 := r.Clone(r.Context())
			r2.URL.Path = "/"
			setCache(w, "/")
			fileServer.ServeHTTP(w, r2)
			return
		}

		if serveCompressed(w, r, root, clean) {
			return
		}

		setCache(w, clean)
		fileServer.ServeHTTP(w, r)
	}), nil
}

func spaRoot(webDir string) (fs.FS, error) {
	if webDir == "" {
		sub, err := fs.Sub(assets.DistFS, "dist")
		if err != nil {
			return nil, err
		}
		return sub, nil
	}
	info, err := os.Stat(webDir)
	if err != nil {
		return nil, fmt.Errorf("web-dir %q: %w", webDir, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("web-dir %q is not a directory", webDir)
	}
	if _, err := os.Stat(filepath.Join(webDir, "index.html")); err != nil {
		return nil, fmt.Errorf("web-dir %q: missing index.html", webDir)
	}
	abs, err := filepath.Abs(webDir)
	if err != nil {
		return nil, err
	}
	return jailedDirFS{root: abs}, nil
}

// jailedDirFS is like os.DirFS but rejects paths whose resolved target
// escapes the configured root (symlink jail).
type jailedDirFS struct {
	root string
}

func (d jailedDirFS) Open(name string) (fs.File, error) {
	if !fs.ValidPath(name) {
		return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrInvalid}
	}
	full := filepath.Join(d.root, filepath.FromSlash(name))
	root, err := filepath.EvalSymlinks(d.root)
	if err != nil {
		root = d.root
	}
	root = filepath.Clean(root)
	resolved, err := filepath.EvalSymlinks(full)
	if err != nil {
		return nil, &fs.PathError{Op: "open", Path: name, Err: err}
	}
	resolved = filepath.Clean(resolved)
	if resolved != root && !strings.HasPrefix(resolved, root+string(filepath.Separator)) {
		return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrNotExist}
	}
	return os.Open(resolved) // #nosec G304 -- path jails via EvalSymlinks under root
}

func serveCompressed(w http.ResponseWriter, r *http.Request, sub fs.FS, name string) bool {
	if isServiceWorkerAsset(name) {
		return false
	}
	enc := preferredEncoding(r)
	if enc.suffix == "" {
		return false
	}
	compressed := name + enc.suffix
	if _, err := fs.Stat(sub, compressed); err != nil {
		return false
	}
	f, err := sub.Open(compressed)
	if err != nil {
		return false
	}
	defer f.Close()
	setCache(w, name)
	w.Header().Set("Content-Encoding", enc.header)
	w.Header().Set("Vary", "Accept-Encoding")
	if info, err := fs.Stat(sub, compressed); err == nil {
		if rs, ok := f.(io.ReadSeeker); ok {
			http.ServeContent(w, r, path.Base(name), info.ModTime(), rs)
			return true
		}
	}
	_, _ = io.Copy(w, f)
	return true
}

type encodingChoice struct {
	suffix string
	header string
}

func preferredEncoding(r *http.Request) encodingChoice {
	ae := r.Header.Get("Accept-Encoding")
	if strings.Contains(ae, "br") {
		return encodingChoice{suffix: ".br", header: "br"}
	}
	if strings.Contains(ae, "gzip") {
		return encodingChoice{suffix: ".gz", header: "gzip"}
	}
	return encodingChoice{}
}

// setCache applies long-lived caching to hashed asset bundles and a
// conservative policy to the HTML shell so updates roll out cleanly.
func setCache(w http.ResponseWriter, name string) {
	if isServiceWorkerAsset(name) {
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		return
	}
	if strings.HasPrefix(name, "assets/") {
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		return
	}
	w.Header().Set("Cache-Control", "no-cache")
}

func isServiceWorkerAsset(name string) bool {
	switch name {
	case "sw.js", "manifest.webmanifest":
		return true
	default:
		return strings.HasPrefix(name, "workbox-") && strings.HasSuffix(name, ".js")
	}
}
