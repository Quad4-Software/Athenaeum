package models

import "time"

// ShareLink is a public, tokenized download link for one book.
type ShareLink struct {
	ID            int64      `json:"id"`
	Token         string     `json:"token"`
	BookID        int64      `json:"bookId"`
	CreatedBy     int64      `json:"createdBy"`
	ExpiresAt     *time.Time `json:"expiresAt,omitempty"`
	CreatedAt     time.Time  `json:"createdAt"`
	DownloadCount int64      `json:"downloadCount"`
	MaxDownloads  int64      `json:"maxDownloads,omitempty"`
}

// ShareLinkMeta is the public metadata returned for a token before download.
type ShareLinkMeta struct {
	Token     string     `json:"token"`
	BookTitle string     `json:"bookTitle"`
	Author    string     `json:"author"`
	Format    string     `json:"format"`
	FileSize  int64      `json:"fileSize"`
	ExpiresAt *time.Time `json:"expiresAt,omitempty"`
}
