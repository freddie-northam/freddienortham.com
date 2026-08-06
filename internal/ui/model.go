package ui

import (
	"strings"

	"github.com/freddie-northam/freddienortham.com/internal/content"
)

// Link is one nav item or tab.
type Link struct {
	Title  string
	URL    string
	Active bool
}

// View is everything a template needs. Building it in Go keeps the templ
// files small and free of logic.
type View struct {
	Site        content.Config
	Title       string // <title>, without the site name
	Description string
	Path        string // canonical, trailing slash
	Accent      string
	Favicon     string
	Nav         []Link
	Tabs        []Link // sub-sections, rendered as the tab bar
	Section     content.Section
	Page        content.Page
	Pages       []content.Page
	Recent      []content.Page
	Projects    []content.Page
	Skills      []Skill
	Components  []Component
	Work        []Role
	ShowAllWork bool
	DaysZero    int64 // unix seconds of the birth date, for the counter
}

// Skill is one authored Agent Skill, read from SKILL.md frontmatter so this
// page can never drift from the source.
type Skill struct {
	Name        string
	Description string
	Install     string
	URL         string
}

// Component is one entry in the live component gallery: the rendered thing
// beside the source that produced it.
type Component struct {
	Name    string
	Note    string
	Snippet string
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
		return v.Site.Name
	}
	return v.Title + " — " + v.Site.Name
}

// NavFor builds the top-level nav, marking the section the reader is in.
func NavFor(active string) []Link {
	var out []Link
	out = append(out, Link{Title: "Home", URL: "/", Active: active == ""})
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
