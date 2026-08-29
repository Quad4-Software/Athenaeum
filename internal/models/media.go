package models

// ComicPage describes one page in a CBZ/CBR archive.
type ComicPage struct {
	Index    int    `json:"index"`
	Name     string `json:"name"`
	MimeType string `json:"mimeType"`
}

// ComicManifest is the page list for a comic book.
type ComicManifest struct {
	Total int         `json:"total"`
	Pages []ComicPage `json:"pages"`
}

// MobiSection is one readable HTML chunk from a MOBI/AZW file.
type MobiSection struct {
	Index int    `json:"index"`
	Title string `json:"title"`
	HTML  string `json:"html"`
}

// AudiobookTrack is one file in a multi-part audiobook folder.
type AudiobookTrack struct {
	Index    int    `json:"index"`
	Title    string `json:"title"`
	RelPath  string `json:"relPath"`
	Format   string `json:"format"`
	FileSize int64  `json:"fileSize"`
}

// ConvertResult reports the outcome of an e-book conversion job.
type ConvertResult struct {
	TargetFormat string `json:"targetFormat"`
	OutputPath   string `json:"outputPath"`
	BookID       int64  `json:"bookId,omitempty"`
	Message      string `json:"message,omitempty"`
}
