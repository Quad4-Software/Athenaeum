package rss

import (
	"bytes"
	"encoding/xml"
	"strings"
	"testing"
	"time"
)

func TestWriteProducesValidRSS2(t *testing.T) {
	var buf bytes.Buffer
	w := Writer{Title: "My Feed", Link: "https://lib.example", Description: "Audio items"}
	pub := time.Date(2024, 3, 1, 12, 0, 0, 0, time.UTC)
	items := []Item{{
		Title:         "Chapter <One> & \"Two\"",
		GUID:          "book-7",
		PubDate:       pub,
		EnclosureURL:  "https://lib.example/feed/tok/item/7/file",
		EnclosureSize: 42,
		EnclosureMIME: "audio/mpeg",
		Author:        "A & B",
		ImageURL:      "https://lib.example/feed/tok/item/7/cover",
	}}
	if err := w.Write(&buf, items); err != nil {
		t.Fatal(err)
	}
	out := buf.String()

	var doc struct {
		XMLName xml.Name `xml:"rss"`
		Version string   `xml:"version,attr"`
		Channel struct {
			Title string `xml:"title"`
			Link  string `xml:"link"`
			Items []struct {
				Title     string `xml:"title"`
				GUID      string `xml:"guid"`
				PubDate   string `xml:"pubDate"`
				Enclosure struct {
					URL    string `xml:"url,attr"`
					Length int64  `xml:"length,attr"`
					Type   string `xml:"type,attr"`
				} `xml:"enclosure"`
				Author string `xml:"author"`
				Image  *struct {
					Href string `xml:"href,attr"`
				} `xml:"image"`
			} `xml:"item"`
		} `xml:"channel"`
	}
	if err := xml.Unmarshal(buf.Bytes(), &doc); err != nil {
		t.Fatalf("output is not well-formed XML: %v\n%s", err, out)
	}
	if doc.Version != "2.0" {
		t.Errorf("version=%q want 2.0", doc.Version)
	}
	if doc.Channel.Title != "My Feed" || doc.Channel.Link != "https://lib.example" {
		t.Errorf("channel=%+v", doc.Channel)
	}
	if len(doc.Channel.Items) != 1 {
		t.Fatalf("items=%d", len(doc.Channel.Items))
	}
	it := doc.Channel.Items[0]
	if it.Title != `Chapter <One> & "Two"` {
		t.Errorf("title=%q", it.Title)
	}
	if it.Author != "A & B" {
		t.Errorf("author=%q", it.Author)
	}
	if it.GUID != "book-7" {
		t.Errorf("guid=%q", it.GUID)
	}
	if it.PubDate != "Fri, 01 Mar 2024 12:00:00 +0000" {
		t.Errorf("pubDate=%q", it.PubDate)
	}
	if it.Enclosure.URL != items[0].EnclosureURL || it.Enclosure.Length != 42 || it.Enclosure.Type != "audio/mpeg" {
		t.Errorf("enclosure=%+v", it.Enclosure)
	}
	if it.Image == nil || it.Image.Href != items[0].ImageURL {
		t.Errorf("image=%+v", it.Image)
	}
	if !strings.Contains(out, `xmlns:itunes="http://www.itunes.com/dtds/podcast-1.0.dtd"`) {
		t.Error("missing itunes namespace declaration")
	}
}

func TestWriteOmitsOptionalFields(t *testing.T) {
	var buf bytes.Buffer
	w := Writer{Title: "T", Link: "http://x", Description: "d"}
	if err := w.Write(&buf, []Item{{
		Title: "plain", GUID: "book-1", PubDate: time.Unix(0, 0),
		EnclosureURL: "http://x/f", EnclosureSize: 1, EnclosureMIME: "audio/ogg",
	}}); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if strings.Contains(out, "itunes:image") {
		t.Error("empty ImageURL must not emit itunes:image")
	}
	if strings.Contains(out, "itunes:author") {
		t.Error("empty Author must not emit itunes:author")
	}
	if len(buf.Bytes()) == 0 {
		t.Fatal("empty output")
	}
}
