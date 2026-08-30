package library

import (
	"path/filepath"
	"strings"
	"testing"
)

func FuzzPathAllowed(f *testing.F) {
	f.Add("/library", "/library")
	f.Add("/library", "/library/books")
	f.Add("/library", "/library2")
	f.Add("/library", "/tmp/secret")
	f.Add("/library", "/library/../etc")
	f.Add("/data/books", "/data/books/a/b")
	f.Add("/lib", "/lib/ety/..rarc")

	f.Fuzz(func(t *testing.T, root, path string) {
		root = strings.TrimSpace(root)
		path = strings.TrimSpace(path)
		if root == "" || root == "/" || path == "" {
			return
		}
		if strings.ContainsRune(root, 0) || strings.ContainsRune(path, 0) {
			return
		}

		r := filepath.Clean(root)
		p := filepath.Clean(path)
		if r == "/" || r == "." {
			return
		}

		got := pathAllowed(p, []string{r})
		lexUnder := p == r || strings.HasPrefix(p, r+string(filepath.Separator))
		_, pathErr := filepath.EvalSymlinks(p)

		// Missing paths that are lexically under the configured root must be allowed
		// even when the root itself is a symlink (for example /lib -> /usr/lib).
		if lexUnder && pathErr != nil && !got {
			t.Fatalf("pathAllowed(%q, [%q]) = false, unresolved lexical child", p, r)
		}

		if !got {
			return
		}

		// Allowed paths must sit under the root lexically or after symlink resolution.
		if lexUnder {
			return
		}
		rp, err := filepath.EvalSymlinks(p)
		if err != nil {
			t.Fatalf("pathAllowed(%q, [%q]) = true but path does not resolve", p, r)
		}
		rr := r
		if resolved, err := filepath.EvalSymlinks(r); err == nil {
			rr = filepath.Clean(resolved)
		}
		rp = filepath.Clean(rp)
		if rp != rr && !strings.HasPrefix(rp, rr+string(filepath.Separator)) {
			t.Fatalf("pathAllowed(%q, [%q]) = true for escape resolved=%q root=%q", p, r, rp, rr)
		}
	})
}
