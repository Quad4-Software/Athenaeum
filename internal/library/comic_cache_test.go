package library

import (
	"strings"
	"testing"
)

func TestComicPageCacheByteBudget(t *testing.T) {
	comicPageMu.Lock()
	comicPageCache = nil
	comicPageMu.Unlock()
	t.Cleanup(func() {
		comicPageMu.Lock()
		comicPageCache = nil
		comicPageMu.Unlock()
	})

	big := []byte(strings.Repeat("x", 20<<20)) // 20 MiB
	for i := range 8 {
		putComicPageCached(comicPageCacheKey{path: "a.cbz", index: i, mtime: 1, size: 1}, big, "image/jpeg")
	}

	comicPageMu.Lock()
	defer comicPageMu.Unlock()
	if comicPageCacheBytesLocked() > maxComicPageBytes {
		t.Fatalf("cache bytes %d exceed budget %d", comicPageCacheBytesLocked(), maxComicPageBytes)
	}
	if len(comicPageCache) > maxComicPageEntries {
		t.Fatalf("cache entries %d exceed max %d", len(comicPageCache), maxComicPageEntries)
	}
	// 20 MiB * 4 = 80 MiB would exceed 64 MiB, so at most 3 full pages fit.
	if len(comicPageCache) > 3 {
		t.Fatalf("expected byte budget to keep at most 3 pages, got %d", len(comicPageCache))
	}
}
