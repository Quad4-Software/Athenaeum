// Package models defines the core domain types shared across the
// Athenaeum application.
package models

import "time"

// Format enumerates supported library item formats.
const (
	FormatEPUB      = "epub"
	FormatPDF       = "pdf"
	FormatMOBI      = "mobi"
	FormatAZW3      = "azw3"
	FormatAZW       = "azw"
	FormatKFX       = "kfx"
	FormatCBZ       = "cbz"
	FormatCBR       = "cbr"
	FormatMP3       = "mp3"
	FormatM4B       = "m4b"
	FormatM4A       = "m4a"
	FormatOGG       = "ogg"
	FormatFLAC      = "flac"
	FormatAudiobook = "audiobook" // multi-file folder aggregate
	FormatAudio     = "audio"     // query filter alias, not stored on disk
	FormatComic     = "comic"     // query filter alias for cbz/cbr
	FormatKindle    = "kindle"    // query filter alias for mobi/azw/azw3
	FormatPapers    = "papers"    // query filter: items with DOI/arXiv/PMID
)

// AudioFormats lists every stored audio extension.
var AudioFormats = []string{FormatMP3, FormatM4B, FormatM4A, FormatOGG, FormatFLAC}

// EbookFormats lists text-oriented formats readable in the browser.
var EbookFormats = []string{FormatEPUB, FormatPDF, FormatMOBI, FormatAZW3, FormatAZW}

// ComicFormats lists comic archive extensions.
var ComicFormats = []string{FormatCBZ, FormatCBR}

// IsAudio reports whether format is an audiobook extension or multi-file set.
func IsAudio(format string) bool {
	switch format {
	case FormatMP3, FormatM4B, FormatM4A, FormatOGG, FormatFLAC, FormatAudiobook:
		return true
	default:
		return false
	}
}

// IsComic reports whether format is a comic archive.
func IsComic(format string) bool {
	switch format {
	case FormatCBZ, FormatCBR:
		return true
	default:
		return false
	}
}

// IsMobiFamily reports Kindle/Mobipocket formats (KFX is download-only).
func IsMobiFamily(format string) bool {
	switch format {
	case FormatMOBI, FormatAZW3, FormatAZW:
		return true
	default:
		return false
	}
}

// IsPaper reports whether the book has a scholarly identifier.
func IsPaper(b Book) bool {
	return b.DOI != "" || b.ArxivID != "" || b.PubmedID != ""
}

// AnonymousUserID is used for progress and collections when auth is disabled.
const AnonymousUserID int64 = 0

// Book represents a single item in the library.
type Book struct {
	ID            int64     `json:"id"`
	LibraryID     int64     `json:"libraryId,omitempty"`
	Title         string    `json:"title"`
	Author        string    `json:"author"`
	Series        string    `json:"series,omitempty"`
	SeriesIndex   float64   `json:"seriesIndex,omitempty"`
	Format        string    `json:"format"`
	RelPath       string    `json:"relPath"`
	FileSize      int64     `json:"fileSize"`
	HasCover      bool      `json:"hasCover"`
	Language      string    `json:"language,omitempty"`
	Description   string    `json:"description,omitempty"`
	DOI           string    `json:"doi,omitempty"`
	ArxivID       string    `json:"arxivId,omitempty"`
	PubmedID      string    `json:"pubmedId,omitempty"`
	Journal       string    `json:"journal,omitempty"`
	Volume        string    `json:"volume,omitempty"`
	Issue         string    `json:"issue,omitempty"`
	Pages         string    `json:"pages,omitempty"`
	PublishedYear int       `json:"publishedYear,omitempty"`
	AddedAt       time.Time `json:"addedAt"`
	ModifiedAt    time.Time `json:"modifiedAt"`
	MetaEdited    bool      `json:"metaEdited,omitempty"`
	CoverEdited   bool      `json:"coverEdited,omitempty"`
	ContentHash   string    `json:"contentHash,omitempty"`
	DuplicateOf   int64     `json:"duplicateOf,omitempty"`

	// ProgressPercent is the reader's completion for this book (0–1), when listing with a user context.
	ProgressPercent float64 `json:"progressPercent,omitempty"`

	// Tags lists the names attached to this book.
	Tags []string `json:"tags,omitempty"`

	// UserRating is the current user's 1-5 star rating for this book, when available.
	UserRating int `json:"userRating,omitempty"`

	// AbsPath is the absolute path on disk and is never serialised to clients.
	AbsPath string `json:"-"`

	// Hidden marks per-file tracks merged into a multi-file audiobook set.
	Hidden bool `json:"-"`
}

// BookUpdate carries user-edited metadata fields.
type BookUpdate struct {
	Title         string  `json:"title"`
	Author        string  `json:"author"`
	Series        string  `json:"series,omitempty"`
	SeriesIndex   float64 `json:"seriesIndex,omitempty"`
	Language      string  `json:"language,omitempty"`
	Description   string  `json:"description,omitempty"`
	DOI           string  `json:"doi,omitempty"`
	ArxivID       string  `json:"arxivId,omitempty"`
	PubmedID      string  `json:"pubmedId,omitempty"`
	Journal       string  `json:"journal,omitempty"`
	Volume        string  `json:"volume,omitempty"`
	Issue         string  `json:"issue,omitempty"`
	Pages         string  `json:"pages,omitempty"`
	PublishedYear int     `json:"publishedYear,omitempty"`
}

// Progress captures a reader's position within a book so reading can be
// resumed across devices.
type Progress struct {
	BookID      int64     `json:"bookId"`
	UserID      int64     `json:"userId,omitempty"`
	Location    string    `json:"location"`
	Percent     float64   `json:"percent"`
	ReadSeconds int64     `json:"readSeconds,omitempty"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// LibraryStats summarises the library for dashboard widgets.
type LibraryStats struct {
	TotalBooks        int64  `json:"totalBooks"`
	EPUBCount         int64  `json:"epubCount"`
	PDFCount          int64  `json:"pdfCount"`
	AudioCount        int64  `json:"audioCount"`
	TotalSizeBytes    int64  `json:"totalSizeBytes"`
	AuthorCount       int64  `json:"authorCount"`
	SeriesCount       int64  `json:"seriesCount"`
	LibraryCount      int64  `json:"libraryCount"`
	AddedLast7Days    int64  `json:"addedLast7Days"`
	CollectionCount   int64  `json:"collectionCount"`
	ReadingInProgress int64  `json:"readingInProgress,omitempty"`
	ReadingCompleted  int64  `json:"readingCompleted,omitempty"`
	FavoriteCount     int64  `json:"favoriteCount,omitempty"`
	UserCount         int64  `json:"userCount,omitempty"`
	LastScanAt        string `json:"lastScanAt,omitempty"`
	Scanning          bool   `json:"scanning"`
	AuthEnabled       bool   `json:"authEnabled"`
}

// BookQuery describes the filtering, sorting and pagination options used
// when listing books.
type BookQuery struct {
	Search       string
	Sort         string
	Format       string
	Author       string
	Series       string
	LibraryID    int64
	CollectionID int64
	UserID       int64
	AddedAfter   int64
	Favorites    bool
	InProgress   bool
	Tag          string
	Limit        int
	Offset       int
	LibraryIDs   []int64
}

// BookPage is a paginated slice of books plus the total match count.
type BookPage struct {
	Items  []Book `json:"items"`
	Total  int64  `json:"total"`
	Limit  int    `json:"limit"`
	Offset int    `json:"offset"`
}

// SeriesInfo summarises one series name in the library.
type SeriesInfo struct {
	Name  string `json:"name"`
	Count int64  `json:"count"`
}
