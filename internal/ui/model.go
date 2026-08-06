package ui

import (
	"fmt"
	"strings"

	"github.com/freddie-northam/freddienortham.com/internal/content"
)

// Link is one nav item or tab.
type Link struct {
	Title  string
	URL    string
	Active bool
}

// RowData is one item in any list. Every section builds these, so Blog,
// Projects, Work and Skills cannot drift into four different row treatments.
type RowData struct {
	Title string
	Sub   string
	URL   string
	Date  string // leading mono column
	Meta  string // trailing mono column
	Thumb string
	Mark  bool // draw the site mark where there is no thumbnail
}

// View is everything a template needs. Building it in Go keeps the templ
// files small and free of logic.
type View struct {
	Site        content.Config
	Title       string // <title>, without the site name
	Description string // meta description; may be a generated fallback
	Lede        string // visible standfirst; only ever authored content
	Path        string // canonical, trailing slash
	Accent      string
	Favicon     string
	Nav         []Link
	Tabs        []Link
	Section     content.Section
	Page        content.Page
	Rows        []RowData
	Recent      []RowData
	Projects    []RowData
	Books       []content.Book
	Skills      []Skill
	Work        []Role
	ShowAllWork bool
	DaysZero    int64
}

// Skill is one authored Agent Skill, read from SKILL.md frontmatter so the
// page cannot drift from the source.
type Skill struct {
	Name        string
	Description string
	Install     string
	URL         string
}

// Role is a work history entry.
type Role struct {
	Title   string
	Org     string
	URL     string
	Period  string
	Summary string
	Current bool
}

func (v View) CanonicalURL() string {
	return strings.TrimSuffix(v.Site.BaseURL, "/") + v.Path
}

func (v View) FullTitle() string {
	if v.Title == "" {
		return v.Site.Name + " — " + v.Site.Role
	}
	return v.Title + " — " + v.Site.Name
}

// NavFor builds the top-level nav, marking the section the reader is in.
func NavFor(active string) []Link {
	out := []Link{{Title: "Home", URL: "/", Active: active == ""}}
	for _, s := range content.Nav() {
		out = append(out, Link{Title: s.Title, URL: s.Path(), Active: s.Slug == active})
	}
	return out
}

// TabsFor builds the sub-section tab bar. Each tab is a real URL, so tabs work
// without JavaScript and every one of them is shareable.
func TabsFor(parent, active string) []Link {
	kids := content.Children(parent)
	if len(kids) == 0 {
		return nil
	}
	out := make([]Link, 0, len(kids))
	for _, s := range kids {
		out = append(out, Link{Title: s.Title, URL: s.Path(), Active: s.Slug == active})
	}
	return out
}

// RowsFrom converts pages into rows. Dated pages lead with the date; undated
// ones lead with the site mark, so a list never has a ragged empty column.
func RowsFrom(pages []content.Page) []RowData {
	out := make([]RowData, 0, len(pages))
	for _, p := range pages {
		r := RowData{Title: p.Title, Sub: p.Standfirst, URL: p.URL(), Thumb: p.Thumb}
		if p.Date != nil {
			r.Date = p.Date.Format("01/06")
		} else {
			r.Mark = true
			r.Meta = p.Meta
		}
		out = append(out, r)
	}
	return out
}

func FirstN(r []RowData, n int) []RowData {
	if len(r) < n {
		return r
	}
	return r[:n]
}

func colClass(span int) string { return fmt.Sprintf("col-%d", span) }

func countStr(n int) string { return fmt.Sprint(n) }
