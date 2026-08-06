package feed

import (
	"encoding/xml"
	"os"
	"sort"
	"time"

	"github.com/freddie-northam/freddienortham.com/internal/content"
)

type rss struct {
	XMLName xml.Name `xml:"rss"`
	Version string   `xml:"version,attr"`
	Atom    string   `xml:"xmlns:atom,attr"`
	Channel channel  `xml:"channel"`
}

type channel struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	Language    string `xml:"language"`
	Updated     string `xml:"lastBuildDate,omitempty"`
	Self        selfIn `xml:"atom:link"`
	Items       []item `xml:"item"`
}

type selfIn struct {
	Href string `xml:"href,attr"`
	Rel  string `xml:"rel,attr"`
	Type string `xml:"type,attr"`
}

type item struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	GUID        guid   `xml:"guid"`
	Description string `xml:"description"`
	Published   string `xml:"pubDate,omitempty"`
}

type guid struct {
	Value string `xml:",chardata"`
	Perma bool   `xml:"isPermaLink,attr"`
}

// Write emits a single merged feed across every section marked InFeed, newest
// first. Undated pages are skipped: a feed entry without a date is noise in
// every reader.
func Write(path string, cfg content.Config, pages []content.Page) error {
	dated := make([]content.Page, 0, len(pages))
	for _, p := range pages {
		if p.Date != nil {
			dated = append(dated, p)
		}
	}
	sort.Slice(dated, func(i, j int) bool { return dated[i].Date.After(*dated[j].Date) })

	ch := channel{
		Title:       cfg.Name,
		Link:        cfg.BaseURL + "/",
		Description: cfg.Description,
		Language:    "en",
		Self:        selfIn{Href: cfg.BaseURL + "/feed.xml", Rel: "self", Type: "application/rss+xml"},
	}
	if len(dated) > 0 {
		ch.Updated = dated[0].Date.Format(time.RFC1123Z)
	}
	for _, p := range dated {
		u := cfg.BaseURL + p.URL()
		ch.Items = append(ch.Items, item{
			Title:       p.Title,
			Link:        u,
			GUID:        guid{Value: u, Perma: true},
			Description: p.Standfirst,
			Published:   p.Date.Format(time.RFC1123Z),
		})
	}

	out, err := xml.MarshalIndent(rss{
		Version: "2.0",
		Atom:    "http://www.w3.org/2005/Atom",
		Channel: ch,
	}, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append([]byte(xml.Header), out...), 0o644)
}
