package library

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
)

var (
	audnexBaseURL        = "https://api.audnex.us"
	openLibrarySearchURL = "https://openlibrary.org/search.json"
)

type metadataLookup struct {
	client *http.Client
}

var sharedMetadataLookup = &metadataLookup{client: sharedMetadataLookupClient}

func newMetadataLookup() *metadataLookup {
	return sharedMetadataLookup
}

func (l *metadataLookup) enrich(ctx context.Context, bookTitle, bookAuthor, isbn, asin string) sidecarFields {
	if asin != "" {
		if meta, ok := l.audnex(ctx, asin); ok {
			return meta
		}
	}
	if isbn != "" {
		if meta, ok := l.openLibraryISBN(ctx, isbn); ok {
			return meta
		}
	}
	if strings.TrimSpace(bookTitle) != "" {
		if meta, ok := l.openLibrarySearch(ctx, bookTitle, bookAuthor); ok {
			return meta
		}
	}
	return sidecarFields{}
}

func (l *metadataLookup) audnex(ctx context.Context, asin string) (sidecarFields, bool) {
	asin = strings.TrimSpace(asin)
	if asin == "" {
		return sidecarFields{}, false
	}
	u := audnexBaseURL + "/books/" + url.PathEscape(asin)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return sidecarFields{}, false
	}
	res, err := l.client.Do(req)
	if err != nil || res.StatusCode != http.StatusOK {
		if res != nil {
			_ = res.Body.Close()
		}
		return sidecarFields{}, false
	}
	defer res.Body.Close()

	var body struct {
		Title       string   `json:"title"`
		Authors     []string `json:"authors"`
		Description string   `json:"description"`
		Series      []struct {
			Name string `json:"name"`
		} `json:"series"`
	}
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		return sidecarFields{}, false
	}
	meta := sidecarFields{
		Title:       strings.TrimSpace(body.Title),
		Description: strings.TrimSpace(body.Description),
		ASIN:        asin,
	}
	if len(body.Authors) > 0 {
		meta.Author = strings.TrimSpace(body.Authors[0])
	}
	if len(body.Series) > 0 {
		meta.Series = strings.TrimSpace(body.Series[0].Name)
	}
	if meta.Title == "" && meta.Author == "" && meta.Description == "" {
		return sidecarFields{}, false
	}
	return meta, true
}

func (l *metadataLookup) openLibraryISBN(ctx context.Context, isbn string) (sidecarFields, bool) {
	isbn = strings.TrimSpace(isbn)
	if isbn == "" {
		return sidecarFields{}, false
	}
	q := url.Values{}
	q.Set("q", "isbn:"+isbn)
	q.Set("limit", "1")
	q.Set("fields", "title,author_name,first_sentence,language")
	u := openLibrarySearchURL + "?" + q.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return sidecarFields{}, false
	}
	res, err := l.client.Do(req)
	if err != nil || res.StatusCode != http.StatusOK {
		if res != nil {
			_ = res.Body.Close()
		}
		return sidecarFields{}, false
	}
	defer res.Body.Close()

	var body struct {
		Docs []struct {
			Title         string   `json:"title"`
			AuthorName    []string `json:"author_name"`
			FirstSentence []string `json:"first_sentence"`
			Language      []string `json:"language"`
		} `json:"docs"`
	}
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil || len(body.Docs) == 0 {
		return sidecarFields{}, false
	}
	doc := body.Docs[0]
	meta := sidecarFields{Title: strings.TrimSpace(doc.Title), ISBN: isbn}
	if len(doc.AuthorName) > 0 {
		meta.Author = strings.TrimSpace(doc.AuthorName[0])
	}
	if len(doc.FirstSentence) > 0 {
		meta.Description = strings.TrimSpace(doc.FirstSentence[0])
	}
	if len(doc.Language) > 0 {
		meta.Language = strings.TrimSpace(doc.Language[0])
	}
	if meta.Title == "" {
		return sidecarFields{}, false
	}
	return meta, true
}

func (l *metadataLookup) openLibrarySearch(ctx context.Context, title, author string) (sidecarFields, bool) {
	q := url.Values{}
	q.Set("title", strings.TrimSpace(title))
	if strings.TrimSpace(author) != "" {
		q.Set("author", strings.TrimSpace(author))
	}
	q.Set("limit", "1")
	q.Set("fields", "title,author_name,first_sentence,language")

	u := openLibrarySearchURL + "?" + q.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return sidecarFields{}, false
	}
	res, err := l.client.Do(req)
	if err != nil || res.StatusCode != http.StatusOK {
		if res != nil {
			_ = res.Body.Close()
		}
		return sidecarFields{}, false
	}
	defer res.Body.Close()

	var body struct {
		Docs []struct {
			Title         string   `json:"title"`
			AuthorName    []string `json:"author_name"`
			FirstSentence []string `json:"first_sentence"`
			Language      []string `json:"language"`
		} `json:"docs"`
	}
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil || len(body.Docs) == 0 {
		return sidecarFields{}, false
	}
	doc := body.Docs[0]
	meta := sidecarFields{Title: strings.TrimSpace(doc.Title)}
	if len(doc.AuthorName) > 0 {
		meta.Author = strings.TrimSpace(doc.AuthorName[0])
	}
	if len(doc.FirstSentence) > 0 {
		meta.Description = strings.TrimSpace(doc.FirstSentence[0])
	}
	if len(doc.Language) > 0 {
		meta.Language = strings.TrimSpace(doc.Language[0])
	}
	if meta.Title == "" {
		return sidecarFields{}, false
	}
	return meta, true
}
