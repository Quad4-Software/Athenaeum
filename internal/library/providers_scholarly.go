package library

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"athenaeum/internal/models"
)

const scholarlyUserAgent = "Athenaeum/1.0 (https://github.com/Quad4-Software/Athenaeum; scholarly-metadata)"
const scholarlyHTTPBodyLimit = 2 << 20

var (
	crossrefWorksURL  = "https://api.crossref.org/works"
	arxivAPIURL       = "http://export.arxiv.org/api/query"
	pubmedESearchURL  = "https://eutils.ncbi.nlm.nih.gov/entrez/eutils/esearch.fcgi"
	pubmedESummaryURL = "https://eutils.ncbi.nlm.nih.gov/entrez/eutils/esummary.fcgi"
)

func decodeJSONLimited(r io.Reader, dst any) error {
	return json.NewDecoder(io.LimitReader(r, scholarlyHTTPBodyLimit)).Decode(dst)
}

func enrichScholarlyMetadata(ctx context.Context, book *models.Book) {
	if book == nil {
		return
	}
	if book.DOI == "" && book.ArxivID == "" && book.PubmedID == "" {
		return
	}
	// Skip when citation fields already look complete.
	if book.Journal != "" && book.PublishedYear > 0 && book.Description != "" {
		return
	}
	searcher := newMetadataSearcher()
	in := MetadataSearchInput{
		Title:    book.Title,
		Author:   book.Author,
		DOI:      book.DOI,
		ArxivID:  book.ArxivID,
		PubmedID: book.PubmedID,
	}
	var matches []models.MetadataMatch
	switch {
	case book.DOI != "":
		matches = searcher.searchCrossref(ctx, in)
	case book.ArxivID != "":
		matches = searcher.searchArxiv(ctx, in)
	case book.PubmedID != "":
		matches = searcher.searchPubmed(ctx, in)
	}
	if len(matches) == 0 {
		return
	}
	m := matches[0]
	if !scholarlyMatchIDsAlign(book, m) {
		return
	}
	applyLookupMeta(book, matchToSidecar(m))
}

func scholarlyMatchIDsAlign(book *models.Book, m models.MetadataMatch) bool {
	if book.DOI != "" {
		got := NormalizeDOI(m.DOI)
		return got != "" && strings.EqualFold(got, book.DOI)
	}
	if book.ArxivID != "" {
		got := NormalizeArxivID(m.ArxivID)
		return got != "" && strings.EqualFold(got, book.ArxivID)
	}
	if book.PubmedID != "" {
		got := NormalizePubmedID(m.PubmedID)
		return got != "" && got == book.PubmedID
	}
	return false
}

func matchToSidecar(m models.MetadataMatch) sidecarFields {
	return sidecarFields{
		Title:         m.Title,
		Author:        m.Author,
		Description:   m.Description,
		Language:      m.Language,
		Series:        m.Series,
		SeriesIndex:   m.SeriesIndex,
		DOI:           m.DOI,
		ArxivID:       m.ArxivID,
		PubmedID:      m.PubmedID,
		Journal:       m.Journal,
		Volume:        m.Volume,
		Issue:         m.Issue,
		Pages:         m.Pages,
		PublishedYear: m.PublishedYear,
	}
}

func (s *metadataSearcher) searchCrossref(ctx context.Context, in MetadataSearchInput) []models.MetadataMatch {
	doi := NormalizeDOI(in.DOI)
	var reqURL string
	if doi != "" {
		reqURL = crossrefWorksURL + "/" + url.PathEscape(doi)
		if m, ok := s.fetchCrossrefWork(ctx, reqURL); ok {
			return []models.MetadataMatch{m}
		}
		return nil
	}
	title := strings.TrimSpace(in.Title)
	if title == "" {
		return nil
	}
	q := url.Values{}
	q.Set("query.bibliographic", title)
	if a := strings.TrimSpace(in.Author); a != "" {
		q.Set("query.author", a)
	}
	q.Set("rows", "8")
	reqURL = crossrefWorksURL + "?" + q.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil
	}
	req.Header.Set("User-Agent", scholarlyUserAgent)
	res, err := s.client.Do(req)
	if err != nil || res.StatusCode != http.StatusOK {
		if res != nil {
			_ = res.Body.Close()
		}
		return nil
	}
	defer res.Body.Close()
	var body struct {
		Message struct {
			Items []crossrefWork `json:"items"`
		} `json:"message"`
	}
	if err := decodeJSONLimited(res.Body, &body); err != nil {
		return nil
	}
	var out []models.MetadataMatch
	for _, item := range body.Message.Items {
		if m, ok := crossrefWorkToMatch(item); ok {
			out = append(out, m)
		}
	}
	return out
}

func (s *metadataSearcher) fetchCrossrefWork(ctx context.Context, reqURL string) (models.MetadataMatch, bool) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return models.MetadataMatch{}, false
	}
	req.Header.Set("User-Agent", scholarlyUserAgent)
	res, err := s.client.Do(req)
	if err != nil || res.StatusCode != http.StatusOK {
		if res != nil {
			_ = res.Body.Close()
		}
		return models.MetadataMatch{}, false
	}
	defer res.Body.Close()
	var body struct {
		Message crossrefWork `json:"message"`
	}
	if err := decodeJSONLimited(res.Body, &body); err != nil {
		return models.MetadataMatch{}, false
	}
	return crossrefWorkToMatch(body.Message)
}

type crossrefWork struct {
	DOI    string   `json:"DOI"`
	Title  []string `json:"title"`
	Author []struct {
		Given  string `json:"given"`
		Family string `json:"family"`
		Name   string `json:"name"`
	} `json:"author"`
	ContainerTitle []string `json:"container-title"`
	Volume         string   `json:"volume"`
	Issue          string   `json:"issue"`
	Page           string   `json:"page"`
	Abstract       string   `json:"abstract"`
	PublishedPrint *struct {
		DateParts [][]int `json:"date-parts"`
	} `json:"published-print"`
	PublishedOnline *struct {
		DateParts [][]int `json:"date-parts"`
	} `json:"published-online"`
	Issued *struct {
		DateParts [][]int `json:"date-parts"`
	} `json:"issued"`
}

func crossrefWorkToMatch(w crossrefWork) (models.MetadataMatch, bool) {
	title := ""
	if len(w.Title) > 0 {
		title = strings.TrimSpace(w.Title[0])
	}
	if title == "" {
		return models.MetadataMatch{}, false
	}
	var authors []string
	for _, a := range w.Author {
		name := strings.TrimSpace(a.Name)
		if name == "" {
			name = strings.TrimSpace(strings.TrimSpace(a.Given) + " " + strings.TrimSpace(a.Family))
		}
		if name != "" {
			authors = append(authors, name)
		}
	}
	m := models.MetadataMatch{
		Source:      "crossref",
		SourceID:    NormalizeDOI(w.DOI),
		Title:       title,
		Author:      strings.Join(authors, ", "),
		DOI:         NormalizeDOI(w.DOI),
		Volume:      strings.TrimSpace(w.Volume),
		Issue:       strings.TrimSpace(w.Issue),
		Pages:       strings.TrimSpace(w.Page),
		Description: stripXMLTags(w.Abstract),
	}
	if len(w.ContainerTitle) > 0 {
		m.Journal = strings.TrimSpace(w.ContainerTitle[0])
	}
	m.PublishedYear = crossrefYear(w)
	return m, true
}

func crossrefYear(w crossrefWork) int {
	for _, src := range []*struct {
		DateParts [][]int `json:"date-parts"`
	}{w.PublishedPrint, w.PublishedOnline, w.Issued} {
		if src == nil || len(src.DateParts) == 0 || len(src.DateParts[0]) == 0 {
			continue
		}
		if y := src.DateParts[0][0]; y > 0 {
			return y
		}
	}
	return 0
}

func stripXMLTags(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	var b strings.Builder
	inTag := false
	for _, r := range s {
		switch {
		case r == '<':
			inTag = true
		case r == '>':
			inTag = false
		case !inTag:
			b.WriteRune(r)
		}
	}
	return strings.Join(strings.Fields(b.String()), " ")
}

func (s *metadataSearcher) searchArxiv(ctx context.Context, in MetadataSearchInput) []models.MetadataMatch {
	id := NormalizeArxivID(in.ArxivID)
	q := url.Values{}
	if id != "" {
		q.Set("id_list", id)
		q.Set("max_results", "1")
	} else {
		title := strings.TrimSpace(in.Title)
		if title == "" {
			return nil
		}
		q.Set("search_query", "ti:"+quoteQueryTerm(title))
		q.Set("start", "0")
		q.Set("max_results", "8")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, arxivAPIURL+"?"+q.Encode(), nil)
	if err != nil {
		return nil
	}
	req.Header.Set("User-Agent", scholarlyUserAgent)
	res, err := s.client.Do(req)
	if err != nil || res.StatusCode != http.StatusOK {
		if res != nil {
			_ = res.Body.Close()
		}
		return nil
	}
	defer res.Body.Close()
	data, err := io.ReadAll(io.LimitReader(res.Body, scholarlyHTTPBodyLimit))
	if err != nil {
		return nil
	}
	return parseArxivAtom(data)
}

type arxivFeed struct {
	Entries []arxivEntry `xml:"entry"`
}

type arxivEntry struct {
	ID        string `xml:"id"`
	Title     string `xml:"title"`
	Summary   string `xml:"summary"`
	Published string `xml:"published"`
	Authors   []struct {
		Name string `xml:"name"`
	} `xml:"author"`
	DOI string `xml:"http://arxiv.org/schemas/atom doi"`
}

func parseArxivAtom(data []byte) []models.MetadataMatch {
	var feed arxivFeed
	if err := xml.Unmarshal(data, &feed); err != nil {
		return nil
	}
	var out []models.MetadataMatch
	for _, e := range feed.Entries {
		title := strings.Join(strings.Fields(e.Title), " ")
		if title == "" {
			continue
		}
		var authors []string
		for _, a := range e.Authors {
			if n := strings.TrimSpace(a.Name); n != "" {
				authors = append(authors, n)
			}
		}
		arxivID := NormalizeArxivID(e.ID)
		if arxivID == "" {
			arxivID = NormalizeArxivID(strings.TrimPrefix(e.ID, "http://arxiv.org/abs/"))
		}
		m := models.MetadataMatch{
			Source:        "arxiv",
			SourceID:      arxivID,
			Title:         title,
			Author:        strings.Join(authors, ", "),
			Description:   strings.Join(strings.Fields(e.Summary), " "),
			ArxivID:       arxivID,
			DOI:           NormalizeDOI(e.DOI),
			Journal:       "arXiv",
			PublishedYear: parsePublishedYear(e.Published),
		}
		out = append(out, m)
	}
	return out
}

func (s *metadataSearcher) searchPubmed(ctx context.Context, in MetadataSearchInput) []models.MetadataMatch {
	pmid := NormalizePubmedID(in.PubmedID)
	if pmid == "" {
		title := strings.TrimSpace(in.Title)
		if title == "" {
			return nil
		}
		ids := s.pubmedSearchIDs(ctx, title, in.Author)
		if len(ids) == 0 {
			return nil
		}
		return s.pubmedSummaries(ctx, ids)
	}
	return s.pubmedSummaries(ctx, []string{pmid})
}

func (s *metadataSearcher) pubmedSearchIDs(ctx context.Context, title, author string) []string {
	var term string
	if a := strings.TrimSpace(author); a != "" {
		term = fmt.Sprintf("%s[Title] AND %s[Author]", title, a)
	} else {
		term = title + "[Title]"
	}
	q := url.Values{}
	q.Set("db", "pubmed")
	q.Set("retmode", "json")
	q.Set("retmax", "8")
	q.Set("term", term)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, pubmedESearchURL+"?"+q.Encode(), nil)
	if err != nil {
		return nil
	}
	req.Header.Set("User-Agent", scholarlyUserAgent)
	res, err := s.client.Do(req)
	if err != nil || res.StatusCode != http.StatusOK {
		if res != nil {
			_ = res.Body.Close()
		}
		return nil
	}
	defer res.Body.Close()
	var body struct {
		ESearchResult struct {
			IDList []string `json:"idlist"`
		} `json:"esearchresult"`
	}
	if err := decodeJSONLimited(res.Body, &body); err != nil {
		return nil
	}
	return body.ESearchResult.IDList
}

func (s *metadataSearcher) pubmedSummaries(ctx context.Context, ids []string) []models.MetadataMatch {
	if len(ids) == 0 {
		return nil
	}
	q := url.Values{}
	q.Set("db", "pubmed")
	q.Set("retmode", "json")
	q.Set("id", strings.Join(ids, ","))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, pubmedESummaryURL+"?"+q.Encode(), nil)
	if err != nil {
		return nil
	}
	req.Header.Set("User-Agent", scholarlyUserAgent)
	res, err := s.client.Do(req)
	if err != nil || res.StatusCode != http.StatusOK {
		if res != nil {
			_ = res.Body.Close()
		}
		return nil
	}
	defer res.Body.Close()
	var body struct {
		Result map[string]json.RawMessage `json:"result"`
	}
	if err := decodeJSONLimited(res.Body, &body); err != nil {
		return nil
	}
	var out []models.MetadataMatch
	for _, id := range ids {
		raw, ok := body.Result[id]
		if !ok {
			continue
		}
		var doc struct {
			UID         string `json:"uid"`
			Title       string `json:"title"`
			FullJournal string `json:"fulljournalname"`
			Source      string `json:"source"`
			Volume      string `json:"volume"`
			Issue       string `json:"issue"`
			Pages       string `json:"pages"`
			PubDate     string `json:"pubdate"`
			Authors     []struct {
				Name string `json:"name"`
			} `json:"authors"`
			ArticleIDs []struct {
				ID   string `json:"value"`
				Type string `json:"idtype"`
			} `json:"articleids"`
		}
		if err := json.Unmarshal(raw, &doc); err != nil {
			continue
		}
		title := strings.TrimSpace(doc.Title)
		if title == "" {
			continue
		}
		var authors []string
		for _, a := range doc.Authors {
			if n := strings.TrimSpace(a.Name); n != "" {
				authors = append(authors, n)
			}
		}
		journal := strings.TrimSpace(doc.FullJournal)
		if journal == "" {
			journal = strings.TrimSpace(doc.Source)
		}
		m := models.MetadataMatch{
			Source:   "pubmed",
			SourceID: id,
			Title:    title,
			Author:   strings.Join(authors, ", "),
			PubmedID: id,
			Journal:  journal,
			Volume:   strings.TrimSpace(doc.Volume),
			Issue:    strings.TrimSpace(doc.Issue),
			Pages:    strings.TrimSpace(doc.Pages),
		}
		if y := parsePublishedYear(doc.PubDate); y > 0 {
			m.PublishedYear = y
		} else if len(doc.PubDate) >= 4 {
			if y, err := strconv.Atoi(doc.PubDate[:4]); err == nil {
				m.PublishedYear = y
			}
		}
		for _, aid := range doc.ArticleIDs {
			if strings.EqualFold(aid.Type, "doi") && m.DOI == "" {
				m.DOI = NormalizeDOI(aid.ID)
			}
		}
		out = append(out, m)
	}
	return out
}
