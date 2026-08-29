package library

// Example custom metadata provider for forks.
// Copy to providers_custom.go and implement Search against your API.
//
//	func init() {
//		RegisterMetadataProvider(MetadataProviderDef{
//			Info: models.MetadataProvider{
//				ID:          "myapi",
//				Label:       "My Metadata API",
//				Description: "Internal catalog lookup",
//			},
//			Search: func(ctx context.Context, s *metadataSearcher, in MetadataSearchInput) []models.MetadataMatch {
//				// return matches from your HTTP API
//				return nil
//			},
//		})
//	}
