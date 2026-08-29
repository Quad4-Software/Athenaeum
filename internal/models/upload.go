package models

import "time"

// UploadSession tracks a resumable library file upload.
type UploadSession struct {
	ID        string    `json:"id"`
	LibraryID int64     `json:"libraryId"`
	UserID    int64     `json:"userId"`
	RelPath   string    `json:"relPath"`
	TotalSize int64     `json:"totalSize"`
	Offset    int64     `json:"offset"`
	Done      bool      `json:"done"`
	BookID    int64     `json:"bookId,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}
