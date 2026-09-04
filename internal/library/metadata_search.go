package library

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"

	"athenaeum/internal/models"
)

var googleBooksAPIURL = "https://www.googleapis.com/books/v1/volumes"

// MetadataProviders lists external sources available in the metadata editor.
func MetadataProviders() []models.MetadataProvider {
	defs := metadataProviderDefs()
	out := make([]models.MetadataProvider, len(defs))
	for i, def := range defs {
		out[i] = def.Info
		out[i].RequiresASIN = def.RequiresASIN
	}
	return out
}

type metadataSearcher struct {
	client *http.Client
}

var sharedMetadataSearcher = &metadataSearcher{client: sharedMetadataSearchClient}

func newMetadataSearcher() *metadataSearcher {
	return sharedMetadataSearcher
}

// SearchMetadata queries external providers and returns ranked matches.
func SearchMetadata(ctx context.Context, q models.MetadataSearchQuery) []models.MetadataMatch {
	searcher := newMetadataSearcher()
	providers := normalizeMetadataProviders(q.Providers)

	title := strings.TrimSpace(q.Title)
	author := strings.TrimSpace(q.Author)
	isbn := strings.TrimSpace(q.ISBN)
	asin := strings.TrimSpace(q.ASIN)
	doi := NormalizeDOI(q.DOI)
	arxivID := NormalizeArxivID(q.ArxivID)
	pubmedID := NormalizePubmedID(q.PubmedID)

	if title == "" && author == "" && isbn == "" && asin == "" && doi == "" && arxivID == "" && pubmedID == "" {
		return nil
	}

	in := MetadataSearchInput{
		Title: title, Author: author, ISBN: isbn, ASIN: asin,
		DOI: doi, ArxivID: arxivID, PubmedID: pubmedID,
	}

	var (
		mu      sync.Mutex
		matches []models.MetadataMatch
		wg      sync.WaitGroup
	)

	for _, p := range providers {
		wg.Add(1)
		go func(providerID string) {
			defer wg.Done()
			def, ok := metadataProviderByID(providerID)
			if !ok || def.Search == nil {
				return
			}
			found := def.Search(ctx, searcher, in)
			if len(found) == 0 {
				return
			}
			mu.Lock()
			matches = append(matches, found...)
			mu.Unlock()
		}(p)
	}
	wg.Wait()

	return dedupeMatches(matches, 20)
}

func dedupeMatches(in []models.MetadataMatch, limit int) []models.MetadataMatch {
	seen := map[string]struct{}{}
	var out []models.MetadataMatch
	for _, m := range in {
		key := strings.ToLower(m.Source) + "|" + strings.ToLower(m.SourceID)
		if m.SourceID == "" {
			key = strings.ToLower(m.Source) + "|" + strings.ToLower(m.Title) + "|" + strings.ToLower(m.Author)
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, m)
		if len(out) >= limit {
			break
		}
	}
	return out
}

func (s *metadataSearcher) searchGoogleBooks(ctx context.Context, title, author, isbn string) []models.MetadataMatch {
	q := googleBooksQuery(title, author, isbn)
	if q == "" {
		return nil
	}
	params := url.Values{}
	params.Set("q", q)
	params.Set("maxResults", "8")
	params.Set("printType", "all")

	u := googleBooksAPIURL + "?" + params.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil
	}
	res, err := s.client.Do(req)
	if err != nil || res.StatusCode != http.StatusOK {
		if res != nil {
			_ = res.Body.Close()
		}
		return nil
	}
	defer res.Body.Close()

	var body struct {
		Items []struct {
			ID         string `json:"id"`
			VolumeInfo struct {
				Title               string   `json:"title"`
				Subtitle            string   `json:"subtitle"`
				Authors             []string `json:"authors"`
				Description         string   `json:"description"`
				Language            string   `json:"language"`
				PublishedDate       string   `json:"publishedDate"`
				IndustryIdentifiers []struct {
					Type       string `json:"type"`
					Identifier string `json:"identifier"`
				} `json:"industryIdentifiers"`
				ImageLinks struct {
					Thumbnail      string `json:"thumbnail"`
					SmallThumbnail string `json:"smallThumbnail"`
				} `json:"imageLinks"`
			} `json:"volumeInfo"`
		} `json:"items"`
	}
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		return nil
	}

	var out []models.MetadataMatch
	for _, item := range body.Items {
		vi := item.VolumeInfo
		t := strings.TrimSpace(vi.Title)
		if t == "" {
			continue
		}
		if sub := strings.TrimSpace(vi.Subtitle); sub != "" {
			t = t + ": " + sub
		}
		m := models.MetadataMatch{
			Source:      "google",
			SourceID:    item.ID,
			Title:       t,
			Description: strings.TrimSpace(vi.Description),
			Language:    strings.TrimSpace(vi.Language),
			CoverURL:    bestGoogleCover(vi.ImageLinks.Thumbnail, vi.ImageLinks.SmallThumbnail),
		}
		if len(vi.Authors) > 0 {
			m.Author = strings.TrimSpace(vi.Authors[0])
		}
		for _, id := range vi.IndustryIdentifiers {
			switch id.Type {
			case "ISBN_13", "ISBN_10":
				if m.ISBN == "" {
					m.ISBN = strings.TrimSpace(id.Identifier)
				}
			}
		}
		if y := parsePublishedYear(vi.PublishedDate); y > 0 {
			m.PublishedYear = y
		}
		out = append(out, m)
	}
	return out
}

func googleBooksQuery(title, author, isbn string) string {
	if isbn != "" {
		return "isbn:" + isbn
	}
	var parts []string
	if title != "" {
		parts = append(parts, "intitle:"+quoteQueryTerm(title))
	}
	if author != "" {
		parts = append(parts, "inauthor:"+quoteQueryTerm(author))
	}
	return strings.Join(parts, "+")
}

func quoteQueryTerm(s string) string {
	s = strings.TrimSpace(s)
	if strings.Contains(s, " ") {
		return `"` + s + `"`
	}
	return s
}

func bestGoogleCover(urls ...string) string {
	for _, u := range urls {
		u = strings.TrimSpace(u)
		if u == "" {
			continue
		}
		u = strings.Replace(u, "http://", "https://", 1)
		u = strings.Replace(u, "&edge=curl", "", 1)
		return u
	}
	return ""
}

func parsePublishedYear(s string) int {
	s = strings.TrimSpace(s)
	if len(s) < 4 {
		return 0
	}
	y, err := strconv.Atoi(s[:4])
	if err != nil || y < 1000 {
		return 0
	}
	return y
}

func (s *metadataSearcher) searchOpenLibrary(ctx context.Context, title, author, isbn string) []models.MetadataMatch {
	params := url.Values{}
	if isbn != "" {
		params.Set("q", "isbn:"+isbn)
	} else {
		if title != "" {
			params.Set("title", title)
		}
		if author != "" {
			params.Set("author", author)
		}
	}
	params.Set("limit", "8")
	params.Set("fields", "key,title,author_name,first_sentence,language,cover_i,isbn,first_publish_year")

	u := openLibrarySearchURL + "?" + params.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil
	}
	res, err := s.client.Do(req)
	if err != nil || res.StatusCode != http.StatusOK {
		if res != nil {
			_ = res.Body.Close()
		}
		return nil
	}
	defer res.Body.Close()

	var body struct {
		Docs []struct {
			Key              string   `json:"key"`
			Title            string   `json:"title"`
			AuthorName       []string `json:"author_name"`
			FirstSentence    []string `json:"first_sentence"`
			Language         []string `json:"language"`
			CoverI           int64    `json:"cover_i"`
			ISBN             []string `json:"isbn"`
			FirstPublishYear int      `json:"first_publish_year"`
		} `json:"docs"`
	}
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		return nil
	}

	var out []models.MetadataMatch
	for _, doc := range body.Docs {
		t := strings.TrimSpace(doc.Title)
		if t == "" {
			continue
		}
		m := models.MetadataMatch{
			Source:   "openlibrary",
			SourceID: strings.TrimPrefix(doc.Key, "/works/"),
			Title:    t,
		}
		if len(doc.AuthorName) > 0 {
			m.Author = strings.TrimSpace(doc.AuthorName[0])
		}
		if len(doc.FirstSentence) > 0 {
			m.Description = strings.TrimSpace(doc.FirstSentence[0])
		}
		if len(doc.Language) > 0 {
			m.Language = strings.TrimSpace(doc.Language[0])
		}
		if doc.CoverI > 0 {
			m.CoverURL = fmt.Sprintf("https://covers.openlibrary.org/b/id/%d-L.jpg", doc.CoverI)
		}
		if len(doc.ISBN) > 0 {
			m.ISBN = strings.TrimSpace(doc.ISBN[0])
		}
		if doc.FirstPublishYear > 0 {
			m.PublishedYear = doc.FirstPublishYear
		}
		out = append(out, m)
	}
	return out
}

func (s *metadataSearcher) audnexusBook(ctx context.Context, asin string) (models.MetadataMatch, bool) {
	asin = strings.TrimSpace(asin)
	if asin == "" {
		return models.MetadataMatch{}, false
	}
	u := audnexBaseURL + "/books/" + url.PathEscape(asin)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return models.MetadataMatch{}, false
	}
	res, err := s.client.Do(req)
	if err != nil || res.StatusCode != http.StatusOK {
		if res != nil {
			_ = res.Body.Close()
		}
		return models.MetadataMatch{}, false
	}
	defer res.Body.Close()

	var body struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		Summary     string `json:"summary"`
		ISBN        string `json:"isbn"`
		Language    string `json:"language"`
		Image       string `json:"image"`
		Authors     []struct {
			Name string `json:"name"`
		} `json:"authors"`
		Series []struct {
			Name string `json:"name"`
		} `json:"series"`
		ReleaseDate string `json:"releaseDate"`
	}
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		return models.MetadataMatch{}, false
	}

	desc := strings.TrimSpace(body.Description)
	if desc == "" {
		desc = strings.TrimSpace(body.Summary)
	}
	m := models.MetadataMatch{
		Source:      "audnexus",
		SourceID:    asin,
		Title:       strings.TrimSpace(body.Title),
		Description: desc,
		Language:    strings.TrimSpace(body.Language),
		ISBN:        strings.TrimSpace(body.ISBN),
		ASIN:        asin,
		CoverURL:    strings.TrimSpace(body.Image),
	}
	if len(body.Authors) > 0 {
		m.Author = strings.TrimSpace(body.Authors[0].Name)
	}
	if len(body.Series) > 0 {
		m.Series = strings.TrimSpace(body.Series[0].Name)
	}
	if y := parsePublishedYear(body.ReleaseDate); y > 0 {
		m.PublishedYear = y
	}
	if m.Title == "" {
		return models.MetadataMatch{}, false
	}
	return m, true
}

// MatchToBookUpdate converts an external match into editable book fields.
func MatchToBookUpdate(m models.MetadataMatch) models.BookUpdate {
	return models.BookUpdate{
		Title:         strings.TrimSpace(m.Title),
		Author:        strings.TrimSpace(m.Author),
		Series:        CleanSeriesName(m.Series),
		SeriesIndex:   m.SeriesIndex,
		Language:      strings.TrimSpace(m.Language),
		Description:   strings.TrimSpace(m.Description),
		DOI:           NormalizeDOI(m.DOI),
		ArxivID:       NormalizeArxivID(m.ArxivID),
		PubmedID:      NormalizePubmedID(m.PubmedID),
		Journal:       strings.TrimSpace(m.Journal),
		Volume:        strings.TrimSpace(m.Volume),
		Issue:         strings.TrimSpace(m.Issue),
		Pages:         strings.TrimSpace(m.Pages),
		PublishedYear: m.PublishedYear,
	}
}

// BestMetadataMatch picks the highest-confidence match for a book.
func BestMetadataMatch(book models.Book, matches []models.MetadataMatch) (models.MetadataMatch, bool) {
	if len(matches) == 0 {
		return models.MetadataMatch{}, false
	}
	best := matches[0]
	bestScore := scoreMetadataMatch(book, matches[0])
	for _, m := range matches[1:] {
		if s := scoreMetadataMatch(book, m); s > bestScore {
			best, bestScore = m, s
		}
	}
	return best, bestScore > 0
}

func scoreMetadataMatch(book models.Book, m models.MetadataMatch) float64 {
	bt := strings.ToLower(strings.TrimSpace(book.Title))
	ba := strings.ToLower(strings.TrimSpace(book.Author))
	mt := strings.ToLower(strings.TrimSpace(m.Title))
	ma := strings.ToLower(strings.TrimSpace(m.Author))

	score := 0.0
	switch {
	case bt != "" && bt == mt:
		score += 12
	case bt != "" && (strings.Contains(mt, bt) || strings.Contains(bt, mt)):
		score += 7
	}
	switch {
	case ba != "" && ba == ma:
		score += 8
	case ba != "" && ma != "" && (strings.Contains(ma, ba) || strings.Contains(ba, ma)):
		score += 4
	}
	if m.ISBN != "" {
		score += 2
	}
	if m.CoverURL != "" {
		score += 1
	}
	if book.DOI != "" && NormalizeDOI(m.DOI) == NormalizeDOI(book.DOI) {
		score += 20
	}
	if book.ArxivID != "" && NormalizeArxivID(m.ArxivID) == NormalizeArxivID(book.ArxivID) {
		score += 20
	}
	if book.PubmedID != "" && NormalizePubmedID(m.PubmedID) == NormalizePubmedID(book.PubmedID) {
		score += 20
	}
	if m.DOI != "" || m.ArxivID != "" || m.PubmedID != "" {
		score += 3
	}
	return score
}
