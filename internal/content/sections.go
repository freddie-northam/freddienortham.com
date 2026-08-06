package content

import "strings"

// Section describes one area of the site. Adding a list-shaped section is one
// entry here plus a content folder: nav, index page, favicon tint, feed
// inclusion and the accent custom property are all derived from this table.
//
// Photos and Games are deliberately absent. Neither is list-shaped (a gallery
// and an interactive surface respectively), so both need a real Template and
// were never "one line".
type Section struct {
	Slug     string
	Title    string
	Accent   string // empty on a child means: inherit the parent's
	Parent   string // empty means top level
	Template string // empty means the generic list template
	InNav    bool
	InFeed   bool
}

// Sections is the single source of truth for site structure.
var Sections = []Section{
	// Top level.
	{Slug: "blog", Title: "Blog", Accent: "#CC3A00", InNav: true, InFeed: true},
	{Slug: "projects", Title: "Projects", Accent: "#0033FF", InNav: true},
	{Slug: "code", Title: "Code", Accent: "#7C3AED", InNav: true},
	{Slug: "not-code", Title: "Not Code", Accent: "#047857", InNav: true},

	// Code sub-pages. Tabs are real URLs, so these are sections, not JS state.
	{Slug: "components", Title: "Components", Parent: "code", Template: "components"},
	{Slug: "languages", Title: "Languages", Parent: "code"},
	{Slug: "stacks", Title: "Stacks", Parent: "code"},
	{Slug: "skills", Title: "Skills", Parent: "code", Template: "skills"},

	// Not Code sub-pages.
	{Slug: "reading", Title: "Reading", Parent: "not-code"},
	{Slug: "music", Title: "Music", Parent: "not-code"},
	{Slug: "podcasts", Title: "Podcasts", Parent: "not-code"},
	{Slug: "sports", Title: "Sports", Parent: "not-code"},
	{Slug: "cars", Title: "Cars", Parent: "not-code"},
	{Slug: "tv", Title: "TV", Parent: "not-code"},
	{Slug: "film", Title: "Film", Parent: "not-code"},
	{Slug: "gaming", Title: "Gaming", Parent: "not-code"},
}

// Reserved accents for sections not yet built, kept here so the palette stays
// coherent when they arrive. Every value clears 4.5:1 against #FFFFFF; see
// TestAccentContrast, which enforces that for Sections above.
var Reserved = map[string]string{
	"photos": "#C2185B",
	"games":  "#A16207",
}

func Find(slug string) (Section, bool) {
	for _, s := range Sections {
		if s.Slug == slug {
			return s, true
		}
	}
	return Section{}, false
}

func TopLevel() []Section {
	var out []Section
	for _, s := range Sections {
		if s.Parent == "" {
			out = append(out, s)
		}
	}
	return out
}

func Nav() []Section {
	var out []Section
	for _, s := range Sections {
		if s.Parent == "" && s.InNav {
			out = append(out, s)
		}
	}
	return out
}

// Children returns the sub-sections of a parent, in table order. These render
// as the tab bar.
func Children(parent string) []Section {
	var out []Section
	for _, s := range Sections {
		if s.Parent == parent {
			out = append(out, s)
		}
	}
	return out
}

// AccentOf resolves a section's accent, walking up to the parent when the
// child does not set one. Sub-pages of Not Code are all green by design.
func AccentOf(s Section) string {
	if s.Accent != "" {
		return s.Accent
	}
	if s.Parent != "" {
		if p, ok := Find(s.Parent); ok {
			return AccentOf(p)
		}
	}
	return "#0A0A0A"
}

// Path is the canonical, trailing-slash URL for a section.
func (s Section) Path() string {
	if s.Parent == "" {
		return "/" + s.Slug + "/"
	}
	return "/" + s.Parent + "/" + s.Slug + "/"
}

// FaviconSlug is the section whose favicon a page under this section uses.
// Children share their parent's mark so a tab switch does not flicker the tab
// icon.
func (s Section) FaviconSlug() string {
	if s.Parent != "" {
		return s.Parent
	}
	return s.Slug
}

// IsIndex reports whether this section lists pages (Blog, Projects) rather
// than holding sub-pages (Code, Not Code).
func (s Section) IsIndex() bool {
	return s.Parent == "" && len(Children(s.Slug)) == 0
}

func (s Section) ContentDir() string {
	if s.Parent == "" {
		return s.Slug
	}
	return s.Parent + "/" + s.Slug
}

// AccentVar is the value written to --accent on <body>.
func AccentVar(s Section) string { return strings.ToUpper(AccentOf(s)) }
