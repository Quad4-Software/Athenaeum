package library

import (
	"testing"
)

func TestNaturalLess(t *testing.T) {
	cases := []struct {
		a, b string
		want bool
	}{
		{"page1.jpg", "page2.jpg", true},
		{"page2.jpg", "page1.jpg", false},
		{"page10.jpg", "page2.jpg", true},
		{"a.jpg", "b.jpg", true},
		{"Page2.jpg", "page2.jpg", false},
		{"img001.png", "img002.png", true},
		{"same", "same", false},
	}
	for _, tc := range cases {
		if got := naturalLess(tc.a, tc.b); got != tc.want {
			t.Errorf("naturalLess(%q, %q)=%v want %v", tc.a, tc.b, got, tc.want)
		}
	}
}

func TestSortComicPages(t *testing.T) {
	pages := []comicPageEntry{
		{Name: "page10.jpg", Mime: "image/jpeg"},
		{Name: "page2.jpg", Mime: "image/jpeg"},
		{Name: "page1.jpg", Mime: "image/jpeg"},
	}
	sortComicPages(pages)
	if pages[0].Name != "page1.jpg" {
		t.Fatalf("first=%q", pages[0].Name)
	}
	if pages[1].Name != "page10.jpg" {
		t.Fatalf("second=%q (string-number order)", pages[1].Name)
	}
	if pages[2].Name != "page2.jpg" {
		t.Fatalf("third=%q", pages[2].Name)
	}
}
