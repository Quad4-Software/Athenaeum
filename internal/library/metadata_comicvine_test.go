package library

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

// swapComicVine points the provider at a test server and sets the API key.
func swapComicVine(t *testing.T, baseURL, key string) {
	t.Helper()
	oldURL, oldKey := comicVineSearchURL, comicVineAPIKey
	comicVineSearchURL, comicVineAPIKey = baseURL, key
	t.Cleanup(func() { comicVineSearchURL, comicVineAPIKey = oldURL, oldKey })
}

func TestSearchComicVine(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("api_key"); got != "test-key" {
			t.Errorf("api_key param=%q", got)
		}
		if got := r.URL.Query().Get("query"); got != "Saga" {
			t.Errorf("query param=%q", got)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status_code":             1,
			"error":                   "OK",
			"number_of_total_results": 2,
			"results": []map[string]any{
				{
					"resource_type": "volume",
					"id":            4050,
					"name":          "Saga",
					"deck":          "An epic space opera.",
					"description":   "<p>Longer <b>HTML</b> text.</p>",
					"start_year":    "2012",
					"publisher":     map[string]any{"name": "Image Comics"},
					"image": map[string]any{
						"medium_url": "http://img.example/m.jpg",
						"super_url":  "http://img.example/s.jpg",
					},
				},
				{
					"resource_type": "issue",
					"id":            400123,
					"name":          "Chapter One",
					"issue_number":  "1",
					"cover_date":    "2012-03-14",
					"description":   "<p>First <i>issue</i>.</p>",
					"image":         map[string]any{"small_url": "https://img.example/i1.jpg"},
					"volume":        map[string]any{"id": 4050, "name": "Saga"},
				},
			},
		})
	}))
	defer srv.Close()
	swapComicVine(t, srv.URL+"/api/search/", "test-key")

	s := &metadataSearcher{client: srv.Client()}
	got := s.searchComicVine(context.Background(), MetadataSearchInput{Title: "Saga"})
	if len(got) != 2 {
		t.Fatalf("got=%d", len(got))
	}

	vol, iss := got[0], got[1]
	if vol.Source != "comicvine" || vol.SourceID != "volume-4050" {
		t.Fatalf("volume ids=%+v", vol)
	}
	if vol.Title != "Saga" || vol.Series != "Saga" || vol.Journal != "Image Comics" {
		t.Fatalf("volume=%+v", vol)
	}
	if vol.PublishedYear != 2012 || vol.Description != "An epic space opera." {
		t.Fatalf("volume=%+v", vol)
	}
	if vol.CoverURL != "https://img.example/s.jpg" {
		t.Fatalf("cover=%q", vol.CoverURL)
	}

	if iss.SourceID != "issue-400123" || iss.Title != "Saga #1: Chapter One" {
		t.Fatalf("issue=%+v", iss)
	}
	if iss.Series != "Saga" || iss.SeriesIndex != 1 || iss.Issue != "1" {
		t.Fatalf("issue=%+v", iss)
	}
	if iss.PublishedYear != 2012 || iss.Description != "First issue." {
		t.Fatalf("issue=%+v", iss)
	}
	if iss.CoverURL != "https://img.example/i1.jpg" {
		t.Fatalf("cover=%q", iss.CoverURL)
	}

	if s.searchComicVine(context.Background(), MetadataSearchInput{}) != nil {
		t.Fatal("empty input")
	}
}

func TestSearchComicVineFailures(t *testing.T) {
	cases := []struct {
		name   string
		status int
		body   string
	}{
		{"api error envelope", http.StatusOK, `{"status_code":100,"error":"Invalid API Key","results":[]}`},
		{"malformed json", http.StatusOK, `not json at all`},
		{"http error", http.StatusBadGateway, `{"status_code":1,"results":[{"resource_type":"volume","id":1,"name":"X"}]}`},
		{"empty results", http.StatusOK, `{"status_code":1,"error":"OK","results":[]}`},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(c.status)
				_, _ = w.Write([]byte(c.body))
			}))
			defer srv.Close()
			swapComicVine(t, srv.URL, "test-key")

			s := &metadataSearcher{client: srv.Client()}
			if got := s.searchComicVine(context.Background(), MetadataSearchInput{Title: "Saga"}); got != nil {
				t.Fatalf("got=%v", got)
			}
		})
	}
}

func TestSearchComicVineNoKey(t *testing.T) {
	var hits int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&hits, 1)
	}))
	defer srv.Close()
	swapComicVine(t, srv.URL, "")

	s := &metadataSearcher{client: srv.Client()}
	if got := s.searchComicVine(context.Background(), MetadataSearchInput{Title: "Saga"}); got != nil {
		t.Fatalf("got=%v", got)
	}
	if atomic.LoadInt32(&hits) != 0 {
		t.Fatal("server contacted without an API key")
	}
}

func TestComicVineProviderGating(t *testing.T) {
	old := comicVineAPIKey
	t.Cleanup(func() { comicVineAPIKey = old })

	comicVineAPIKey = ""
	for _, p := range MetadataProviders() {
		if p.ID == "comicvine" {
			t.Fatal("comicvine listed without an API key")
		}
	}

	comicVineAPIKey = "k"
	found := false
	for _, p := range MetadataProviders() {
		if p.ID == "comicvine" {
			found = true
		}
	}
	if !found {
		t.Fatal("comicvine missing from providers with an API key")
	}
}
