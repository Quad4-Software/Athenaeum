package library

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"athenaeum/internal/models"
)

const (
	comicVineProviderID = "comicvine"
	comicVineUserAgent  = "Athenaeum/1.0 (https://github.com/Quad4-Software/Athenaeum; comic-metadata)"
	comicVineFieldList  = "id,name,deck,description,image,publisher,start_year,issue_number,cover_date,volume,resource_type"
	comicVineSearchMax  = "10"
)

var (
	comicVineSearchURL = "https://comicvine.gamespot.com/api/search/"
	comicVineAPIKey    string
)

// SetComicVineAPIKey configures the Comic Vine metadata provider.
// An empty key disables it: it is hidden from provider listings and
// never contacted.
func SetComicVineAPIKey(key string) {
	comicVineAPIKey = strings.TrimSpace(key)
}

func comicVineReady() bool { return comicVineAPIKey != "" }

func (s *metadataSearcher) searchComicVine(ctx context.Context, in MetadataSearchInput) []models.MetadataMatch {
	key := comicVineAPIKey
	title := strings.TrimSpace(in.Title)
	if key == "" || title == "" {
		return nil
	}
	q := url.Values{}
	q.Set("api_key", key)
	q.Set("format", "json")
	q.Set("query", title)
	q.Set("resources", "volume,issue")
	q.Set("field_list", comicVineFieldList)
	q.Set("limit", comicVineSearchMax)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, comicVineSearchURL+"?"+q.Encode(), nil)
	if err != nil {
		return nil
	}
	req.Header.Set("User-Agent", comicVineUserAgent)
	res, err := s.client.Do(req)
	if err != nil {
		// The error text embeds the request URL, which carries api_key.
		slog.Warn("comicvine search request failed")
		return nil
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		slog.Warn("comicvine search failed", "status", res.StatusCode)
		return nil
	}

	var body struct {
		StatusCode int               `json:"status_code"`
		Error      string            `json:"error"`
		Results    []comicVineResult `json:"results"`
	}
	if err := decodeJSONLimited(res.Body, &body); err != nil {
		return nil
	}
	if body.StatusCode != 1 {
		slog.Warn("comicvine search rejected", "code", body.StatusCode, "error", body.Error)
		return nil
	}

	var out []models.MetadataMatch
	for _, r := range body.Results {
		if m, ok := comicVineResultToMatch(r); ok {
			out = append(out, m)
		}
	}
	return out
}

type comicVineImage struct {
	SuperURL  string `json:"super_url"`
	MediumURL string `json:"medium_url"`
	SmallURL  string `json:"small_url"`
	ThumbURL  string `json:"thumb_url"`
}

type comicVineResult struct {
	ResourceType string         `json:"resource_type"`
	ID           int64          `json:"id"`
	Name         string         `json:"name"`
	Deck         string         `json:"deck"`
	Description  string         `json:"description"`
	IssueNumber  string         `json:"issue_number"`
	CoverDate    string         `json:"cover_date"`
	StartYear    string         `json:"start_year"`
	Image        comicVineImage `json:"image"`
	Publisher    struct {
		Name string `json:"name"`
	} `json:"publisher"`
	Volume struct {
		ID   int64  `json:"id"`
		Name string `json:"name"`
	} `json:"volume"`
}

func comicVineResultToMatch(r comicVineResult) (models.MetadataMatch, bool) {
	m := models.MetadataMatch{
		Source:      comicVineProviderID,
		SourceID:    fmt.Sprintf("%s-%d", r.ResourceType, r.ID),
		Description: comicVineDescription(r.Deck, r.Description),
		CoverURL:    comicVineCoverURL(r.Image),
	}
	switch r.ResourceType {
	case "volume":
		m.Title = strings.TrimSpace(r.Name)
		m.Series = m.Title
		// MetadataMatch has no publisher field; Journal is the only
		// persisted slot that survives MatchToBookUpdate.
		m.Journal = strings.TrimSpace(r.Publisher.Name)
		m.PublishedYear = parsePublishedYear(r.StartYear)
	case "issue":
		m.Title = comicVineIssueTitle(r)
		m.Series = strings.TrimSpace(r.Volume.Name)
		m.Issue = strings.TrimSpace(r.IssueNumber)
		if f, err := strconv.ParseFloat(m.Issue, 64); err == nil {
			m.SeriesIndex = f
		}
		m.PublishedYear = parsePublishedYear(r.CoverDate)
	default:
		return models.MetadataMatch{}, false
	}
	if m.Title == "" {
		return models.MetadataMatch{}, false
	}
	return m, true
}

// comicVineIssueTitle builds "Volume #N: Issue name". Issue names are often
// empty upstream, so the volume name is the base.
func comicVineIssueTitle(r comicVineResult) string {
	vol := strings.TrimSpace(r.Volume.Name)
	name := strings.TrimSpace(r.Name)
	title := vol
	if title == "" {
		title = name
	}
	if title == "" {
		return ""
	}
	if num := strings.TrimSpace(r.IssueNumber); num != "" {
		title += " #" + num
	}
	if name != "" && vol != "" && !strings.EqualFold(name, vol) {
		title += ": " + name
	}
	return title
}

func comicVineDescription(deck, description string) string {
	if d := strings.TrimSpace(deck); d != "" {
		return d
	}
	return stripXMLTags(description)
}

func comicVineCoverURL(img comicVineImage) string {
	for _, u := range []string{img.SuperURL, img.MediumURL, img.SmallURL, img.ThumbURL} {
		if u = strings.TrimSpace(u); u != "" {
			return strings.Replace(u, "http://", "https://", 1)
		}
	}
	return ""
}
