package server

import (
	"path/filepath"
	"strings"
	"testing"
	"testing/quick"
)

func TestPropertySanitizeUploadRelPathNoDotDot(t *testing.T) {
	fn := func(rel string) bool {
		out, err := sanitizeUploadRelPath(rel)
		if err != nil {
			return true
		}
		if strings.Contains(out, "..") {
			return false
		}
		clean := filepath.ToSlash(filepath.Clean(out))
		return clean != ".." && !strings.HasPrefix(clean, "../")
	}
	if err := quick.Check(fn, &quick.Config{MaxCount: 1000}); err != nil {
		t.Fatal(err)
	}
}

func TestAdversarialSanitizeUploadRelPath(t *testing.T) {
	cases := []string{
		"../etc/passwd",
		"foo/../../etc/passwd",
		"/abs/path.epub",
		"..",
		"....//....//x.epub",
		`foo\..\bar.epub`,
	}
	for _, c := range cases {
		if out, err := sanitizeUploadRelPath(c); err == nil {
			if strings.Contains(out, "..") || strings.HasPrefix(out, "/") {
				t.Fatalf("accepted hostile relPath %q -> %q", c, out)
			}
		}
	}
}
