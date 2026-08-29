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
		under := p == r || strings.HasPrefix(p, r+string(filepath.Separator))
		if under != got {
			t.Fatalf("pathAllowed(%q, [%q]) = %v, under=%v", p, r, got, under)
		}
	})
}
