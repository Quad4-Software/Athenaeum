package models

// Tag is a user-defined label that can be attached to books.
type Tag struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// BookRating is a per-user star rating for a book.
type BookRating struct {
	UserID    int64 `json:"userId,omitempty"`
	BookID    int64 `json:"bookId"`
	Rating    int   `json:"rating"`
	UpdatedAt int64 `json:"updatedAt"`
}

// ReaderPrefs holds a user's synced reading preferences, such as font
// and theme settings for the epub reader. Prefs are stored as opaque
// JSON so new fields can be added without a schema migration.
type ReaderPrefs struct {
	UserID    int64          `json:"userId,omitempty"`
	Prefs     map[string]any `json:"prefs"`
	UpdatedAt int64          `json:"updatedAt,omitempty"`
}
