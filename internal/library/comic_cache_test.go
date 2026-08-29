package library

import (
	"archive/zip"
	"os"
	"path/filepath"
	"testing"
)

func TestComicMetaCache(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.cbz")

	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	zw := zip.NewWriter(f)
	w, err := zw.Create("001.jpg")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := w.Write([]byte("fake-jpeg")); err != nil {
		t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}

	total1, pages1 := ComicManifest(path)
	if total1 != 1 || len(pages1) != 1 {
		t.Fatalf("expected 1 page, got %d pages %v", total1, pages1)
	}

	comicMetaMu.RLock()
	cached := comicMetaByPath[path]
	comicMetaMu.RUnlock()
	if cached.meta.PageCount != 1 {
		t.Fatalf("expected cached meta, got %+v", cached)
	}

	total2, _ := ComicManifest(path)
	if total2 != 1 {
		t.Fatalf("expected cached manifest total 1, got %d", total2)
	}

	data1, mime1, err := OpenComicPage(path, 0)
	if err != nil {
		t.Fatal(err)
	}
	if mime1 != "image/jpeg" || string(data1) != "fake-jpeg" {
		t.Fatalf("unexpected page payload mime=%s data=%q", mime1, data1)
	}
	data2, _, err := OpenComicPage(path, 0)
	if err != nil {
		t.Fatal(err)
	}
	if string(data2) != "fake-jpeg" {
		t.Fatalf("expected page cache hit, got %q", data2)
	}

	if err := os.WriteFile(path, []byte("not-a-zip"), 0o644); err != nil {
		t.Fatal(err)
	}
	total3, pages3 := ComicManifest(path)
	if total3 != 0 || len(pages3) != 0 {
		t.Fatalf("expected empty manifest after file change, got %d pages", total3)
	}
}
