package models

import "time"

// ScanStatus reports live library scan progress.
type ScanStatus struct {
	Scanning    bool       `json:"scanning"`
	Indexed     int64      `json:"indexed"`
	Skipped     int64      `json:"skipped"`
	CurrentPath string     `json:"currentPath,omitempty"`
	LibraryName string     `json:"libraryName,omitempty"`
	StartedAt   *time.Time `json:"startedAt,omitempty"`
	FinishedAt  *time.Time `json:"finishedAt,omitempty"`
}

// AuthorInfo summarises one author name in the library.
type AuthorInfo struct {
	Name  string `json:"name"`
	Count int64  `json:"count"`
}
