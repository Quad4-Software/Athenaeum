// Package demo seeds a public-domain demo library for local demos and screenshots.
package demo

// Entry is one demo library item. Titles are public-domain classics
// (Project Gutenberg / expired copyright). Media files are tiny stubs unless
// a remote Gutenberg EPUB URL is configured and reachable.
type Entry struct {
	Slug        string
	Title       string
	Author      string
	Series      string
	SeriesIndex float64
	Format      string
	Language    string
	Description string
	FileSize    int64
	Progress    float64
	Favorite    bool
	Shelf       string
	// CoverURL is an Open Library (or similar) cover image for legal PD works.
	CoverURL string
	// GutenbergID is the Project Gutenberg ebook id when a remote EPUB may be fetched.
	GutenbergID int
}

// Catalog is a curated set of public-domain works for demos and screenshots.
func Catalog() []Entry {
	return []Entry{
		{
			Slug: "alice-in-wonderland", Title: "Alice's Adventures in Wonderland", Author: "Lewis Carroll",
			Format: "epub", Language: "en", GutenbergID: 11,
			Description: "A girl falls down a rabbit hole into a world of peculiar creatures and impossible logic.",
			FileSize:    180000, Progress: 0.42, Favorite: true, Shelf: "Classics",
			CoverURL: "https://covers.openlibrary.org/b/id/10527843-L.jpg",
		},
		{
			Slug: "pride-and-prejudice", Title: "Pride and Prejudice", Author: "Jane Austen",
			Format: "epub", Language: "en", GutenbergID: 1342,
			Description: "Elizabeth Bennet navigates manners, marriage, and misunderstanding in Georgian England.",
			FileSize:    720000, Progress: 0.18, Shelf: "Classics",
			CoverURL: "https://covers.openlibrary.org/b/id/14348537-L.jpg",
		},
		{
			Slug: "frankenstein", Title: "Frankenstein", Author: "Mary Shelley",
			Format: "epub", Language: "en", GutenbergID: 84,
			Description: "A scientist's creation of life leads to tragedy, exile, and a chase across the ice.",
			FileSize:    450000, Favorite: true, Shelf: "Classics",
			CoverURL: "https://covers.openlibrary.org/b/id/12356249-L.jpg",
		},
		{
			Slug: "sherlock-holmes", Title: "The Adventures of Sherlock Holmes", Author: "Arthur Conan Doyle",
			Format: "pdf", Language: "en", GutenbergID: 1661,
			Description: "Twelve cases from Baker Street: scandals, red-headed leagues, and speckled bands.",
			FileSize:    1100000, Progress: 0.55,
			CoverURL: "https://covers.openlibrary.org/b/id/6717853-L.jpg",
		},
		{
			Slug: "dracula", Title: "Dracula", Author: "Bram Stoker",
			Format: "pdf", Language: "en", GutenbergID: 345,
			Description: "Letters and diaries chart Count Dracula's arrival in England and the hunt that follows.",
			FileSize:    980000,
			CoverURL:    "https://covers.openlibrary.org/b/id/12216503-L.jpg",
		},
		{
			Slug: "moby-dick", Title: "Moby-Dick; or, The Whale", Author: "Herman Melville",
			Format: "m4b", Language: "en", GutenbergID: 2701,
			Description: "Ishmael ships aboard the Pequod as Captain Ahab hunts the white whale.",
			FileSize:    64000000, Progress: 0.28, Favorite: true, Shelf: "Listen Next",
			CoverURL: "https://covers.openlibrary.org/b/id/12116552-L.jpg",
		},
		{
			Slug: "jungle-book", Title: "The Jungle Book", Author: "Rudyard Kipling",
			Format: "mp3", Language: "en", GutenbergID: 236,
			Description: "Mowgli and other animal tales from the Indian jungle, first published in 1894.",
			FileSize:    42000000, Shelf: "Listen Next",
			CoverURL: "https://covers.openlibrary.org/b/id/3344204-L.jpg",
		},
		{
			Slug: "christmas-carol", Title: "A Christmas Carol", Author: "Charles Dickens",
			Format: "audiobook", Language: "en", GutenbergID: 46,
			Description: "Ebenezer Scrooge is visited by four spirits on Christmas Eve.",
			FileSize:    28000000, Progress: 0.08, Shelf: "Listen Next",
			CoverURL: "https://covers.openlibrary.org/b/id/12875748-L.jpg",
		},
		{
			Slug: "wizard-of-oz", Title: "The Wonderful Wizard of Oz", Author: "L. Frank Baum",
			Format: "cbz", Language: "en", GutenbergID: 55,
			Description: "Dorothy is carried by a cyclone to Oz and seeks the Wizard with unusual companions.",
			FileSize:    2200000, Favorite: true, Shelf: "Classics",
			CoverURL: "https://covers.openlibrary.org/b/id/552443-L.jpg",
		},
		{
			Slug: "odyssey", Title: "The Odyssey", Author: "Homer",
			Format: "epub", Language: "en", GutenbergID: 1727,
			Description: "Odysseus wanders for ten years after Troy before returning to Ithaca.",
			FileSize:    640000, Progress: 0.73,
			CoverURL: "https://covers.openlibrary.org/b/id/10876521-L.jpg",
		},
		{
			Slug: "metamorphosis", Title: "The Metamorphosis", Author: "Franz Kafka",
			Format: "mobi", Language: "en", GutenbergID: 5200,
			Description: "Gregor Samsa wakes to find himself transformed into a monstrous insect.",
			FileSize:    120000,
			CoverURL:    "https://covers.openlibrary.org/b/id/12820198-L.jpg",
		},
		{
			Slug: "faust", Title: "Faust", Author: "Johann Wolfgang von Goethe",
			Format: "epub", Language: "de", GutenbergID: 2229,
			Description: "A scholar wagers his soul with Mephistopheles in Goethe's dramatic poem.",
			FileSize:    520000, Shelf: "Classics",
			CoverURL: "https://covers.openlibrary.org/b/id/1002485-L.jpg",
		},
	}
}
