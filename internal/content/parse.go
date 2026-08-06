package content

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	gmhtml "github.com/yuin/goldmark/renderer/html"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	"go.abhg.dev/goldmark/frontmatter"
)

// Page is the only content type. Everything on the site is derived from it.
type Page struct {
	Section    string
	Slug       string
	Aliases    []string
	Title      string
	Standfirst string
	Date       *time.Time // nil when undated; never a zero-value sentinel
	Meta       string     // right-rail text when there is no date, e.g. "Live"
	Thumb      string
	Draft      bool
	Body       template.HTML
	WordCount  int
	Headings   []Heading
}

type Heading struct {
	Level int
	Text  string
	ID    string
}

type matter struct {
	Title      string   `yaml:"title"`
	Slug       string   `yaml:"slug"`
	Standfirst string   `yaml:"standfirst"`
	Date       string   `yaml:"date"`
	Meta       string   `yaml:"meta"`
	Thumb      string   `yaml:"thumb"`
	Aliases    []string `yaml:"aliases"`
	Draft      bool     `yaml:"draft"`
}

// URL is the canonical trailing-slash path.
func (p Page) URL() string {
	// A sub-section's page is the sub-section: /not-code/cars/, not /cars/.
	if sec, ok := Find(p.Section); ok && sec.Parent != "" {
		return sec.Path()
	}
	if p.Slug == "_index" {
		return "/" + p.Section + "/"
	}
	return "/" + p.Section + "/" + p.Slug + "/"
}

// Rail is what shows in the right-hand metadata column.
func (p Page) Rail() string {
	if p.Date != nil {
		return p.Date.Format("2006 — 01")
	}
	return p.Meta
}

// NeedsAnchors and NeedsTOC make article furniture conditional on length, so
// short posts stay clean and long ones stay navigable.
func (p Page) NeedsAnchors() bool { return p.WordCount > 800 }
func (p Page) NeedsTOC() bool     { return p.WordCount > 2000 }

func (p Page) ReadingMinutes() int {
	m := p.WordCount / 220
	if m < 1 {
		return 1
	}
	return m
}

var md = goldmark.New(
	goldmark.WithExtensions(
		extension.GFM,
		extension.Typographer,
		&frontmatter.Extender{},
		highlighting.NewHighlighting(
			highlighting.WithStyle("bw"),
			highlighting.WithFormatOptions(html.WithClasses(true)),
		),
	),
	goldmark.WithParserOptions(parser.WithAutoHeadingID()),
	goldmark.WithRendererOptions(
		gmhtml.WithUnsafe(),
		renderer.WithNodeRenderers(),
	),
)

var (
	headingRe = regexp.MustCompile(`(?m)^(#{2,3})\s+(.+)$`)
	imgRe     = regexp.MustCompile(`<img([^>]*?)src="([^"]+)"([^>]*?)>`)
	tagRe     = regexp.MustCompile(`<[^>]+>`)
)

// Load reads the markdown for a section. Drafts are excluded unless
// includeDrafts is set, which is what `make dev` passes so unpublished writing
// is previewable without being published.
//
// Three shapes, decided by the section's place in the table:
//
//	blog, projects      content/<slug>/*.md      a list of pages
//	code, not-code      content/<slug>/_index.md its own page only
//	cars, skills, ...   content/<parent>/<slug>.md  one page, the tab's content
func Load(root, section string, includeDrafts bool) ([]Page, error) {
	sec, ok := Find(section)
	if !ok {
		return nil, fmt.Errorf("unknown section %q", section)
	}

	// A sub-section is a single file sitting in its parent's directory.
	if sec.Parent != "" {
		return loadOne(filepath.Join(root, sec.Parent, sec.Slug+".md"), section, includeDrafts)
	}
	// A parent section owns only its own intro; the sibling files in its
	// directory belong to its children.
	if len(Children(sec.Slug)) > 0 {
		return loadOne(filepath.Join(root, sec.Slug, "_index.md"), section, includeDrafts)
	}

	dir := filepath.Join(root, sec.Slug)
	entries, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var pages []Page
	seen := map[string]string{}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		path := filepath.Join(dir, e.Name())
		p, err := parseFile(path, section)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", path, err)
		}
		if p.Draft && !includeDrafts {
			continue
		}
		if prev, dup := seen[p.Slug]; dup {
			return nil, fmt.Errorf("duplicate slug %q in %s (%s and %s)", p.Slug, section, prev, path)
		}
		seen[p.Slug] = path
		pages = append(pages, p)
	}

	// Dated newest first; undated keep table order after them.
	sort.SliceStable(pages, func(i, j int) bool {
		a, b := pages[i].Date, pages[j].Date
		switch {
		case a != nil && b != nil:
			return a.After(*b)
		case a != nil:
			return true
		default:
			return false
		}
	})
	return pages, nil
}

// loadOne reads a single-page section, returning nothing when the file is
// absent or drafted so an unwritten tab still renders rather than 404s.
func loadOne(path, section string, includeDrafts bool) ([]Page, error) {
	p, err := parseFile(path, section)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	p.Slug = "_index"
	if p.Draft && !includeDrafts {
		return nil, nil
	}
	return []Page{p}, nil
}

func parseFile(path, section string) (Page, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return Page{}, err
	}
	return Parse(raw, section, strings.TrimSuffix(filepath.Base(path), ".md"))
}

// Parse turns a markdown file into a Page. Exported for the golden tests.
func Parse(raw []byte, section, filename string) (Page, error) {
	var buf bytes.Buffer
	ctx := parser.NewContext()
	if err := md.Convert(raw, &buf, parser.WithContext(ctx)); err != nil {
		return Page{}, err
	}

	var fm matter
	if d := frontmatter.Get(ctx); d != nil {
		if err := d.Decode(&fm); err != nil {
			return Page{}, fmt.Errorf("frontmatter: %w", err)
		}
	}
	if fm.Title == "" {
		return Page{}, fmt.Errorf("missing title")
	}

	// Slug is explicit, never derived from the filename, so renaming a file
	// can never break a published URL. _index is the exception: it names the
	// section's own page.
	slug := fm.Slug
	if slug == "" {
		if filename == "_index" {
			slug = "_index"
		} else {
			return Page{}, fmt.Errorf("missing slug (required: renaming a file must not break a URL)")
		}
	}

	var date *time.Time
	if fm.Date != "" {
		t, err := time.Parse("2006-01-02", fm.Date)
		if err != nil {
			return Page{}, fmt.Errorf("bad date %q, want YYYY-MM-DD: %w", fm.Date, err)
		}
		date = &t
	}

	body := stampImages(buf.String())
	plain := tagRe.ReplaceAllString(body, " ")

	return Page{
		Section:    section,
		Slug:       slug,
		Aliases:    fm.Aliases,
		Title:      fm.Title,
		Standfirst: fm.Standfirst,
		Date:       date,
		Meta:       fm.Meta,
		Thumb:      fm.Thumb,
		Draft:      fm.Draft,
		Body:       template.HTML(body),
		WordCount:  len(strings.Fields(plain)),
		Headings:   headings(raw),
	}, nil
}

func headings(raw []byte) []Heading {
	var out []Heading
	for _, m := range headingRe.FindAllStringSubmatch(string(raw), -1) {
		text := strings.TrimSpace(m[2])
		out = append(out, Heading{Level: len(m[1]), Text: text, ID: slugify(text)})
	}
	return out
}

var nonWord = regexp.MustCompile(`[^a-z0-9]+`)

func slugify(s string) string {
	return strings.Trim(nonWord.ReplaceAllString(strings.ToLower(s), "-"), "-")
}

// stampImages adds lazy loading and async decoding. Dimensions come from the
// fixed thumbnail aspect ratio in CSS rather than from reading every file, so
// there is no layout shift and no image decoding in the build.
func stampImages(body string) string {
	return imgRe.ReplaceAllString(body, `<img$1src="$2"$3 loading="lazy" decoding="async">`)
}
