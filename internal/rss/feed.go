// Package rss renders RSS 2.0 podcast feeds so podcast clients can play
// a user's audio books through a token-scoped URL.
package rss

import (
	"encoding/xml"
	"io"
	"time"
)

const nsITunes = "http://www.itunes.com/dtds/podcast-1.0.dtd"

// Writer renders one RSS channel of playable items.
type Writer struct {
	Title       string
	Link        string
	Description string
}

// Item is one playable episode in the feed.
type Item struct {
	Title         string
	GUID          string
	PubDate       time.Time
	EnclosureURL  string
	EnclosureSize int64
	EnclosureMIME string
	Author        string
	ImageURL      string
}

type document struct {
	XMLName xml.Name `xml:"rss"`
	Version string   `xml:"version,attr"`
	ITunes  string   `xml:"xmlns:itunes,attr"`
	Channel channel  `xml:"channel"`
}

type channel struct {
	Title       string    `xml:"title"`
	Link        string    `xml:"link"`
	Description string    `xml:"description"`
	Items       []rssItem `xml:"item"`
}

type rssItem struct {
	Title     string       `xml:"title"`
	GUID      guid         `xml:"guid"`
	PubDate   string       `xml:"pubDate"`
	Enclosure enclosure    `xml:"enclosure"`
	Author    string       `xml:"itunes:author,omitempty"`
	Image     *itunesImage `xml:"itunes:image,omitempty"`
}

type guid struct {
	IsPermaLink bool   `xml:"isPermaLink,attr"`
	Value       string `xml:",chardata"`
}

type enclosure struct {
	URL    string `xml:"url,attr"`
	Length int64  `xml:"length,attr"`
	Type   string `xml:"type,attr"`
}

type itunesImage struct {
	Href string `xml:"href,attr"`
}

// Write encodes the channel and its items as RSS 2.0.
func (w Writer) Write(out io.Writer, items []Item) error {
	doc := document{
		Version: "2.0",
		ITunes:  nsITunes,
		Channel: channel{
			Title:       w.Title,
			Link:        w.Link,
			Description: w.Description,
		},
	}
	for _, it := range items {
		ri := rssItem{
			Title:     it.Title,
			GUID:      guid{IsPermaLink: false, Value: it.GUID},
			PubDate:   it.PubDate.UTC().Format(time.RFC1123Z),
			Enclosure: enclosure{URL: it.EnclosureURL, Length: it.EnclosureSize, Type: it.EnclosureMIME},
			Author:    it.Author,
		}
		if it.ImageURL != "" {
			ri.Image = &itunesImage{Href: it.ImageURL}
		}
		doc.Channel.Items = append(doc.Channel.Items, ri)
	}
	if _, err := out.Write([]byte(xml.Header)); err != nil {
		return err
	}
	enc := xml.NewEncoder(out)
	enc.Indent("", "  ")
	return enc.Encode(doc)
}
