package models

import "time"

// FeedToken is a token-scoped RSS feed that exposes a user's audio books
// to podcast clients. The token in the URL is the only credential, like a
// share link. LibraryID and CollectionID of 0 mean the feed covers every
// audio book the owner can access.
type FeedToken struct {
	ID           int64     `json:"id"`
	Token        string    `json:"token"`
	UserID       int64     `json:"-"`
	Name         string    `json:"name"`
	LibraryID    int64     `json:"libraryId"`
	CollectionID int64     `json:"collectionId"`
	URL          string    `json:"url,omitempty"`
	CreatedAt    time.Time `json:"createdAt"`
}
