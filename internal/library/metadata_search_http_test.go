package library

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSearchGoogleBooks(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"items": []map[string]any{{
				"id": "vol1",
				"volumeInfo": map[string]any{
					"title":         "The Martian",
					"subtitle":      "A Novel",
					"authors":       []string{"Andy Weir"},
					"description":   "Mars",
					"language":      "en",
					"publishedDate": "2014-02-11",
					"industryIdentifiers": []map[string]string{
						{"type": "ISBN_13", "identifier": "9780804139021"},
					},
					"imageLinks": map[string]string{
						"thumbnail": "http://books.google.com/cover&edge=curl",
					},
				},
			}},
		})
	}))
	defer srv.Close()

	old := googleBooksAPIURL
	googleBooksAPIURL = srv.URL
	t.Cleanup(func() { googleBooksAPIURL = old })

	s := &metadataSearcher{client: srv.Client()}
	got := s.searchGoogleBooks(context.Background(), "The Martian", "Andy Weir", "")
	if len(got) != 1 {
		t.Fatalf("got=%d", len(got))
	}
	m := got[0]
	if m.Title != "The Martian: A Novel" || m.Author != "Andy Weir" {
		t.Fatalf("match=%+v", m)
	}
	if m.ISBN != "9780804139021" || m.PublishedYear != 2014 {
		t.Fatalf("ids=%+v", m)
	}
	if m.CoverURL != "https://books.google.com/cover" {
		t.Fatalf("cover=%q", m.CoverURL)
	}
	if s.searchGoogleBooks(context.Background(), "", "", "") != nil {
		t.Fatal("empty query")
	}
}

func TestSearchOpenLibrary(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"docs": []map[string]any{{
				"key":                "/works/OL123W",
				"title":              "OL Title",
				"author_name":        []string{"OL Author"},
				"first_sentence":     []string{"Once."},
				"language":           []string{"eng"},
				"cover_i":            99,
				"isbn":               []string{"9780143127741"},
				"first_publish_year": 2011,
			}},
		})
	}))
	defer srv.Close()

	old := openLibrarySearchURL
	openLibrarySearchURL = srv.URL
	t.Cleanup(func() { openLibrarySearchURL = old })

	s := &metadataSearcher{client: srv.Client()}
	got := s.searchOpenLibrary(context.Background(), "OL Title", "OL Author", "")
	if len(got) != 1 {
		t.Fatalf("got=%d", len(got))
	}
	m := got[0]
	if m.SourceID != "OL123W" || m.Title != "OL Title" || m.Author != "OL Author" {
		t.Fatalf("match=%+v", m)
	}
	if m.CoverURL == "" || m.PublishedYear != 2011 {
		t.Fatalf("match=%+v", m)
	}
}

func TestAudnexusBook(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/books/B00TEST" {
			http.NotFound(w, r)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"title":       "Audible Title",
			"description": "Desc",
			"isbn":        "978000",
			"language":    "en",
			"image":       "https://img.example/c.jpg",
			"authors":     []map[string]string{{"name": "Narrator"}},
			"series":      []map[string]string{{"name": "Series A"}},
			"releaseDate": "2020-05-01",
		})
	}))
	defer srv.Close()

	old := audnexBaseURL
	audnexBaseURL = srv.URL
	t.Cleanup(func() { audnexBaseURL = old })

	s := &metadataSearcher{client: srv.Client()}
	m, ok := s.audnexusBook(context.Background(), "B00TEST")
	if !ok {
		t.Fatal("expected match")
	}
	if m.Title != "Audible Title" || m.Author != "Narrator" || m.Series != "Series A" {
		t.Fatalf("match=%+v", m)
	}
	if m.PublishedYear != 2020 || m.ASIN != "B00TEST" {
		t.Fatalf("match=%+v", m)
	}
	if _, ok := s.audnexusBook(context.Background(), ""); ok {
		t.Fatal("empty asin")
	}
	if _, ok := s.audnexusBook(context.Background(), "MISSING"); ok {
		t.Fatal("missing asin")
	}
}
