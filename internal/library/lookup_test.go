package library

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMetadataLookupAudnex(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/books/B00TEST" {
			http.NotFound(w, r)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"title":       "Audible Title",
			"authors":     []string{"Narrator Author"},
			"description": "Long description",
		})
	}))
	defer srv.Close()

	lookup := &metadataLookup{client: srv.Client()}
	old := audnexBaseURL
	audnexBaseURL = srv.URL
	t.Cleanup(func() { audnexBaseURL = old })

	fields := lookup.enrich(context.Background(), "", "", "", "B00TEST")
	if fields.Title != "Audible Title" || fields.Author != "Narrator Author" {
		t.Fatalf("fields = %+v", fields)
	}
}

func TestMetadataLookupOpenLibraryISBN(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"docs": []map[string]any{{
				"title":          "OL Title",
				"author_name":    []string{"OL Author"},
				"first_sentence": []string{"First line."},
				"language":       []string{"eng"},
			}},
		})
	}))
	defer srv.Close()

	lookup := &metadataLookup{client: srv.Client()}
	old := openLibrarySearchURL
	openLibrarySearchURL = srv.URL + "/search.json"
	t.Cleanup(func() { openLibrarySearchURL = old })

	fields := lookup.enrich(context.Background(), "", "", "9780143127741", "")
	if fields.Title != "OL Title" || fields.Author != "OL Author" {
		t.Fatalf("fields = %+v", fields)
	}
	if fields.ISBN != "9780143127741" {
		t.Errorf("isbn = %q", fields.ISBN)
	}
}
