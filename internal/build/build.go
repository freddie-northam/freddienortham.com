package build

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/a-h/templ"

	"github.com/freddie-northam/freddienortham.com/internal/content"
	"github.com/freddie-northam/freddienortham.com/internal/feed"
	"github.com/freddie-northam/freddienortham.com/internal/ui"
)

type Options struct {
	ContentDir string
	AssetsDir  string
	OutDir     string
	SkillsDir  string
	Drafts     bool
}

// Site is everything loaded from disk, built once and reused by every render.
type Site struct {
	Opts     Options
	Pages    map[string][]content.Page // by section slug
	Index    map[string]content.Page   // section slug -> its _index page
	Aliases  map[string]string         // alias path -> canonical path
	Skills   []ui.Skill
	Rendered []string // every path written, for the link checker
}

func Run(o Options) (*Site, error) {
	s := &Site{
		Opts:    o,
		Pages:   map[string][]content.Page{},
		Index:   map[string]content.Page{},
		Aliases: map[string]string{},
	}

	for _, sec := range content.Sections {
		pages, err := content.Load(o.ContentDir, sec.Slug, o.Drafts)
		if err != nil {
			return nil, err
		}
		var listed []content.Page
		for _, p := range pages {
			if p.Slug == "_index" {
				s.Index[sec.Slug] = p
				continue
			}
			listed = append(listed, p)
			for _, a := range p.Aliases {
				s.Aliases[normalise(a)] = p.URL()
			}
		}
		s.Pages[sec.Slug] = listed
	}

	var err error
	if s.Skills, err = loadSkills(o.SkillsDir); err != nil {
		return nil, err
	}

	// dist/site.css is written by Tailwind, not by this build, so a clean must
	// preserve it. Without this, any rebuild (every content change under
	// `make dev`) deletes the stylesheet and it stays gone until Tailwind
	// happens to notice a change to its own source.
	css, hadCSS := os.ReadFile(filepath.Join(o.OutDir, "site.css"))
	if err := os.RemoveAll(o.OutDir); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(o.OutDir, 0o755); err != nil {
		return nil, err
	}
	if hadCSS == nil {
		if err := os.WriteFile(filepath.Join(o.OutDir, "site.css"), css, 0o644); err != nil {
			return nil, err
		}
	}

	if err := s.renderAll(); err != nil {
		return nil, err
	}
	if err := s.writeAssets(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Site) renderAll() error {
	if err := s.home(); err != nil {
		return err
	}
	for _, sec := range content.Sections {
		if err := s.section(sec); err != nil {
			return fmt.Errorf("section %s: %w", sec.Slug, err)
		}
	}
	if err := s.projectsAll(); err != nil {
		return err
	}
	return s.notFound()
}

func (s *Site) base(sec content.Section, activeTop string) ui.View {
	v := ui.View{
		Site:     content.Site,
		Section:  sec,
		Accent:   content.AccentVar(sec),
		Favicon:  sec.FaviconSlug(),
		Nav:      ui.NavFor(activeTop),
		DaysZero: content.Site.BirthDate.Unix(),
	}
	if sec.Parent != "" {
		v.Tabs = ui.TabsFor(sec.Parent, sec.Slug)
	} else {
		v.Tabs = ui.TabsFor(sec.Slug, "")
	}
	return v
}

func (s *Site) home() error {
	v := ui.View{
		Site:     content.Site,
		Accent:   "#0A0A0A",
		Favicon:  "home",
		Nav:      ui.NavFor(""),
		Path:     "/",
		Title:    "",
		DaysZero: content.Site.BirthDate.Unix(),
	}
	v.Description = content.Site.Description
	if p, ok := s.Index["home"]; ok {
		v.Page = p
	} else if raw, err := os.ReadFile(filepath.Join(s.Opts.ContentDir, "home.md")); err == nil {
		p, err := content.Parse(raw, "home", "_index")
		if err != nil {
			return fmt.Errorf("home.md: %w", err)
		}
		v.Page = p
	}
	v.Recent = firstN(s.Pages["blog"], 3)
	v.Projects = firstN(s.Pages["projects"], 3)
	return s.write("/", ui.Home(v))
}

func (s *Site) section(sec content.Section) error {
	activeTop := sec.Slug
	if sec.Parent != "" {
		activeTop = sec.Parent
	}
	v := s.base(sec, activeTop)
	v.Path = sec.Path()
	v.Title = sec.Title
	v.Pages = s.Pages[sec.Slug]
	v.Skills = s.Skills

	idx, hasIdx := s.Index[sec.Slug]
	v.Page = idx
	v.Description = idx.Standfirst
	if v.Description == "" {
		v.Description = fmt.Sprintf("%s — %s", sec.Title, content.Site.Name)
	}

	// Sub-pages of a parent render as articles from their _index content.
	if sec.Parent != "" {
		switch sec.Template {
		case "skills":
			if err := s.write(sec.Path(), ui.Skills(v)); err != nil {
				return err
			}
		case "components":
			v.Components = make([]ui.Component, 7)
			if err := s.write(sec.Path(), ui.Components(v)); err != nil {
				return err
			}
		default:
			if !hasIdx {
				// A tab with no content yet still needs a real page, or the
				// tab bar links to a 404.
				v.Page = content.Page{Title: sec.Title, Section: sec.Slug, Slug: "_index"}
			}
			if err := s.write(sec.Path(), ui.Article(v)); err != nil {
				return err
			}
		}
		return nil
	}

	// Top-level sections.
	switch {
	case sec.Slug == "projects":
		v.Work = roles(content.Roles[:min(content.VisibleRoles, len(content.Roles))])
		if err := s.write(sec.Path(), ui.Projects(v)); err != nil {
			return err
		}
	case len(content.Children(sec.Slug)) > 0:
		if err := s.write(sec.Path(), ui.SectionPage(v)); err != nil {
			return err
		}
	default:
		if err := s.write(sec.Path(), ui.List(v)); err != nil {
			return err
		}
	}

	// Individual pages within an index section.
	for _, p := range v.Pages {
		pv := s.base(sec, activeTop)
		pv.Tabs = nil
		pv.Path = p.URL()
		pv.Title = p.Title
		pv.Description = p.Standfirst
		pv.Page = p
		if err := s.write(p.URL(), ui.Article(pv)); err != nil {
			return err
		}
	}
	return nil
}

func (s *Site) projectsAll() error {
	sec, ok := content.Find("projects")
	if !ok {
		return nil
	}
	v := s.base(sec, "projects")
	v.Path = "/projects/all/"
	v.Title = "Work"
	v.Description = "Every role, in full."
	v.Work = roles(content.Roles)
	v.ShowAllWork = true
	return s.write("/projects/all/", ui.Projects(v))
}

func (s *Site) notFound() error {
	v := ui.View{
		Site: content.Site, Accent: "#0A0A0A", Favicon: "home",
		Nav: ui.NavFor(""), Path: "/404.html", Title: "Not found",
		Description: "That page has no preimage.",
		DaysZero:    content.Site.BirthDate.Unix(),
	}
	return s.writeFile("404.html", ui.NotFound(v))
}

func (s *Site) write(path string, c templ.Component) error {
	rel := strings.TrimPrefix(path, "/")
	if rel == "" || strings.HasSuffix(rel, "/") {
		rel += "index.html"
	}
	s.Rendered = append(s.Rendered, path)
	return s.writeFile(rel, c)
}

func (s *Site) writeFile(rel string, c templ.Component) error {
	full := filepath.Join(s.Opts.OutDir, rel)
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		return err
	}
	f, err := os.Create(full)
	if err != nil {
		return err
	}
	defer f.Close()
	return c.Render(context.Background(), f)
}

func (s *Site) writeAssets() error {
	if err := copyTree(s.Opts.AssetsDir, s.Opts.OutDir); err != nil {
		return err
	}
	if err := s.favicons(); err != nil {
		return err
	}
	if err := s.redirects(); err != nil {
		return err
	}
	if err := s.llms(); err != nil {
		return err
	}
	return feed.Write(filepath.Join(s.Opts.OutDir, "feed.xml"), content.Site, s.feedPages())
}

func (s *Site) feedPages() []content.Page {
	var out []content.Page
	for _, sec := range content.Sections {
		if !sec.InFeed {
			continue
		}
		out = append(out, s.Pages[sec.Slug]...)
	}
	return out
}

func roles(in []content.Role) []ui.Role {
	out := make([]ui.Role, 0, len(in))
	for _, r := range in {
		out = append(out, ui.Role(r))
	}
	return out
}

func firstN(p []content.Page, n int) []content.Page {
	if len(p) < n {
		return p
	}
	return p[:n]
}

func normalise(p string) string {
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	if !strings.HasSuffix(p, "/") {
		p += "/"
	}
	return p
}

func copyTree(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		if info.IsDir() {
			return os.MkdirAll(filepath.Join(dst, rel), 0o755)
		}
		// site.css is a Tailwind *source* file, not an asset. Copying it would
		// overwrite the compiled stylesheet with one that still contains
		// `@import "tailwindcss"`, which the browser then requests as a URL
		// and refuses for its MIME type, silently dropping preflight.
		if rel == "site.css" {
			return nil
		}
		in, err := os.Open(path)
		if err != nil {
			return err
		}
		defer in.Close()
		out, err := os.Create(filepath.Join(dst, rel))
		if err != nil {
			return err
		}
		defer out.Close()
		_, err = io.Copy(out, in)
		return err
	})
}
