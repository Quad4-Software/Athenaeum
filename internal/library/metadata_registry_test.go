package library

import (
	"context"
	"slices"
	"testing"

	"athenaeum/internal/models"
)

func TestMetadataProviderRegistry(t *testing.T) {
	RegisterMetadataProvider(MetadataProviderDef{
		Info: models.MetadataProvider{ID: "testfork", Label: "Test Fork"},
		Search: func(ctx context.Context, s *metadataSearcher, in MetadataSearchInput) []models.MetadataMatch {
			return []models.MetadataMatch{{Source: "testfork", Title: in.Title}}
		},
	})
	t.Cleanup(func() {
		metadataRegistryMu.Lock()
		metadataRegistry = metadataRegistry[:len(metadataRegistry)-1]
		metadataRegistryMu.Unlock()
	})

	ids := defaultMetadataProviderIDs()
	found := slices.Contains(ids, "testfork")
	if !found {
		t.Fatalf("expected testfork in provider ids: %v", ids)
	}

	matches := SearchMetadata(context.Background(), models.MetadataSearchQuery{
		Title:     "Example",
		Providers: []string{"testfork"},
	})
	if len(matches) != 1 || matches[0].Title != "Example" {
		t.Fatalf("unexpected matches: %+v", matches)
	}
}
