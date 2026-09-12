package models

import "time"

// Bookmark is a saved reading position for a user.
type Bookmark struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"userId,omitempty"`
	BookID    int64     `json:"bookId"`
	Location  string    `json:"location"`
	Label     string    `json:"label,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
}

// Highlight is a user annotation at a location in a book.
type Highlight struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"userId,omitempty"`
	BookID    int64     `json:"bookId"`
	Location  string    `json:"location"`
	Excerpt   string    `json:"excerpt,omitempty"`
	Note      string    `json:"note,omitempty"`
	Color     string    `json:"color,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
}

// ReadingSession is a contiguous stretch of reading on one book, grouped
// server-side from heartbeat writes.
type ReadingSession struct {
	ID        int64     `json:"id"`
	BookID    int64     `json:"bookId"`
	Title     string    `json:"title"`
	StartedAt time.Time `json:"startedAt"`
	EndedAt   time.Time `json:"endedAt"`
	Seconds   int64     `json:"seconds"`
}

// ReadingStats summarizes a user's reading activity.
type ReadingStats struct {
	TotalReadSeconds int64 `json:"totalReadSeconds"`
	BooksInProgress  int64 `json:"booksInProgress"`
	BooksCompleted   int64 `json:"booksCompleted"`
	CurrentStreak    int64 `json:"currentStreakDays"`
	Sessions7d       int64 `json:"sessions7d"`
	Seconds7d        int64 `json:"seconds7d"`
	Seconds30d       int64 `json:"seconds30d"`
}
