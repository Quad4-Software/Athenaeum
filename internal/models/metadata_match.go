package models

// MetadataProvider describes an external metadata source.
type MetadataProvider struct {
	ID           string `json:"id"`
	Label        string `json:"label"`
	Description  string `json:"description,omitempty"`
	RequiresASIN bool   `json:"requiresAsin,omitempty"`
}

// MetadataSearchQuery carries user search terms for external lookup.
type MetadataSearchQuery struct {
	Title     string   `json:"title"`
	Author    string   `json:"author"`
	ISBN      string   `json:"isbn"`
	ASIN      string   `json:"asin"`
	Providers []string `json:"providers,omitempty"`
}

// MetadataMatch is one candidate result from an external metadata provider.
type MetadataMatch struct {
	Source        string  `json:"source"`
	SourceID      string  `json:"sourceId,omitempty"`
	Title         string  `json:"title"`
	Author        string  `json:"author"`
	Description   string  `json:"description,omitempty"`
	Language      string  `json:"language,omitempty"`
	Series        string  `json:"series,omitempty"`
	SeriesIndex   float64 `json:"seriesIndex,omitempty"`
	ISBN          string  `json:"isbn,omitempty"`
	ASIN          string  `json:"asin,omitempty"`
	CoverURL      string  `json:"coverUrl,omitempty"`
	PublishedYear int     `json:"publishedYear,omitempty"`
}

// MetadataSearchResult lists matches from one or more providers.
type MetadataSearchResult struct {
	Matches []MetadataMatch `json:"matches"`
}

// MetadataApplyRequest applies a chosen match to a book.
type MetadataApplyRequest struct {
	Match      MetadataMatch `json:"match"`
	ApplyCover bool          `json:"applyCover"`
}
