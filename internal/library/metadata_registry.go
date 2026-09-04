package library

import (
	"context"
	"strings"
	"sync"

	"athenaeum/internal/models"
)

// MetadataSearchInput carries normalized search terms for a provider.
type MetadataSearchInput struct {
	Title    string
	Author   string
	ISBN     string
	ASIN     string
	DOI      string
	ArxivID  string
	PubmedID string
}

// MetadataProviderDef registers one external metadata source.
// Forks add entries via RegisterMetadataProvider or providers_custom.go.
type MetadataProviderDef struct {
	Info         models.MetadataProvider
	RequiresASIN bool
	Search       func(ctx context.Context, s *metadataSearcher, in MetadataSearchInput) []models.MetadataMatch
}

var (
	metadataRegistry   []MetadataProviderDef
	metadataRegistryMu sync.RWMutex
)

func init() {
	RegisterMetadataProvider(MetadataProviderDef{
		Info: models.MetadataProvider{
			ID:          "google",
			Label:       "Google Books",
			Description: "Books and ebooks",
		},
		Search: func(ctx context.Context, s *metadataSearcher, in MetadataSearchInput) []models.MetadataMatch {
			return s.searchGoogleBooks(ctx, in.Title, in.Author, in.ISBN)
		},
	})
	RegisterMetadataProvider(MetadataProviderDef{
		Info: models.MetadataProvider{
			ID:          "openlibrary",
			Label:       "Open Library",
			Description: "Books and editions",
		},
		Search: func(ctx context.Context, s *metadataSearcher, in MetadataSearchInput) []models.MetadataMatch {
			return s.searchOpenLibrary(ctx, in.Title, in.Author, in.ISBN)
		},
	})
	RegisterMetadataProvider(MetadataProviderDef{
		Info: models.MetadataProvider{
			ID:          "audnexus",
			Label:       "Audnexus",
			Description: "Audiobooks by ASIN (Audible)",
		},
		RequiresASIN: true,
		Search: func(ctx context.Context, s *metadataSearcher, in MetadataSearchInput) []models.MetadataMatch {
			if in.ASIN == "" {
				return nil
			}
			if m, ok := s.audnexusBook(ctx, in.ASIN); ok {
				return []models.MetadataMatch{m}
			}
			return nil
		},
	})
	RegisterMetadataProvider(MetadataProviderDef{
		Info: models.MetadataProvider{
			ID:          "crossref",
			Label:       "Crossref",
			Description: "Journal articles and scholarly works by DOI",
		},
		Search: func(ctx context.Context, s *metadataSearcher, in MetadataSearchInput) []models.MetadataMatch {
			return s.searchCrossref(ctx, in)
		},
	})
	RegisterMetadataProvider(MetadataProviderDef{
		Info: models.MetadataProvider{
			ID:          "arxiv",
			Label:       "arXiv",
			Description: "Preprints by arXiv ID or title",
		},
		Search: func(ctx context.Context, s *metadataSearcher, in MetadataSearchInput) []models.MetadataMatch {
			return s.searchArxiv(ctx, in)
		},
	})
	RegisterMetadataProvider(MetadataProviderDef{
		Info: models.MetadataProvider{
			ID:          "pubmed",
			Label:       "PubMed",
			Description: "Medical and life-science literature by PMID or title",
		},
		Search: func(ctx context.Context, s *metadataSearcher, in MetadataSearchInput) []models.MetadataMatch {
			return s.searchPubmed(ctx, in)
		},
	})
}

// RegisterMetadataProvider adds or replaces a metadata source at runtime (e.g. from a fork).
func RegisterMetadataProvider(def MetadataProviderDef) {
	def.Info.ID = strings.ToLower(strings.TrimSpace(def.Info.ID))
	if def.Info.ID == "" || def.Search == nil {
		return
	}
	metadataRegistryMu.Lock()
	defer metadataRegistryMu.Unlock()
	for i, existing := range metadataRegistry {
		if existing.Info.ID == def.Info.ID {
			metadataRegistry[i] = def
			return
		}
	}
	metadataRegistry = append(metadataRegistry, def)
}

func metadataProviderDefs() []MetadataProviderDef {
	metadataRegistryMu.RLock()
	defer metadataRegistryMu.RUnlock()
	out := make([]MetadataProviderDef, len(metadataRegistry))
	copy(out, metadataRegistry)
	return out
}

func metadataProviderByID(id string) (MetadataProviderDef, bool) {
	id = strings.ToLower(strings.TrimSpace(id))
	for _, def := range metadataProviderDefs() {
		if def.Info.ID == id {
			return def, true
		}
	}
	return MetadataProviderDef{}, false
}

func defaultMetadataProviderIDs() []string {
	defs := metadataProviderDefs()
	out := make([]string, len(defs))
	for i, def := range defs {
		out[i] = def.Info.ID
	}
	return out
}

func normalizeMetadataProviders(in []string) []string {
	if len(in) == 0 {
		return defaultMetadataProviderIDs()
	}
	var out []string
	seen := map[string]struct{}{}
	for _, p := range in {
		p = strings.ToLower(strings.TrimSpace(p))
		if p == "all" {
			return defaultMetadataProviderIDs()
		}
		if _, ok := metadataProviderByID(p); !ok {
			continue
		}
		if _, dup := seen[p]; dup {
			continue
		}
		seen[p] = struct{}{}
		out = append(out, p)
	}
	if len(out) == 0 {
		return defaultMetadataProviderIDs()
	}
	return out
}
