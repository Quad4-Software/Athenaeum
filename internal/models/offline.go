package models

// OfflineGrants lists the book ids a user has approved for offline access.
type OfflineGrants struct {
	BookIDs []int64 `json:"bookIds"`
}
