package build

import (
	"bufio"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/freddie-northam/freddienortham.com/internal/content"
	"github.com/freddie-northam/freddienortham.com/internal/ui"
)

// favicons draws the site mark per section, tinted with that section's accent.
//
// PNG is emitted alongside SVG because Safari does not support SVG favicons:
// an SVG-only build shows a blank default there, which on a site this fussy is
// a visible defect. The mark is a square with one corner cut, chosen because
// it is legible at 16px and drawable directly with image/draw, avoiding an SVG
// rasteriser dependency entirely.
func (s *Site) favicons() error {
	marks := map[string]string{"home": "#0A0A0A"}
	for _, sec := range content.Sections {
		if sec.Parent == "" {
			marks[sec.Slug] = content.AccentOf(sec)
		}
	}
	for slug, hex := range marks {
		c, err := parseHex(hex)
		if err != nil {
			return err
		}
		for _, size := range []int{32, 180} {
			p := filepath.Join(s.Opts.OutDir, fmt.Sprintf("favicon-%s-%d.png", slug, size))
			if err := writePNG(p, size, c); err != nil {
				return err
			}
		}
		svg := fmt.Sprintf(
			`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 32 32">`+
				`<path d="M0 0h32v20L20 32H0z" fill="%s"/></svg>`, hex)
		p := filepath.Join(s.Opts.OutDir, fmt.Sprintf("favicon-%s.svg", slug))
		if err := os.WriteFile(p, []byte(svg), 0o644); err != nil {
			return err
		}
	}
	return nil
}

// writePNG draws the square-with-a-diagonal-cut mark: a filled square whose
// bottom-right corner is sliced away along a 45 degree line.
func writePNG(path string, size int, c color.RGBA) error {
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	cut := size * 3 / 8 // where the diagonal starts, from each edge
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			// Outside the cut when x+y exceeds the diagonal threshold.
			if x+y > (size-1)+(size-1-cut) {
				continue
			}
			img.Set(x, y, c)
		}
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}

func parseHex(h string) (color.RGBA, error) {
	h = strings.TrimPrefix(h, "#")
	if len(h) != 6 {
		return color.RGBA{}, fmt.Errorf("bad hex colour %q", h)
	}
	v, err := strconv.ParseUint(h, 16, 32)
	if err != nil {
		return color.RGBA{}, err
	}
	return color.RGBA{R: uint8(v >> 16), G: uint8(v >> 8), B: uint8(v), A: 255}, nil
}

// redirects writes the Cloudflare Pages _redirects file from every page alias,
// so renaming a slug never breaks a published URL.
func (s *Site) redirects() error {
	if len(s.Aliases) == 0 {
		return os.WriteFile(filepath.Join(s.Opts.OutDir, "_redirects"), []byte("# no aliases\n"), 0o644)
	}
	var b strings.Builder
	b.WriteString("# generated from page aliases; do not edit\n")
	for from, to := range s.Aliases {
		fmt.Fprintf(&b, "%s %s 301\n", from, to)
	}
	return os.WriteFile(filepath.Join(s.Opts.OutDir, "_redirects"), []byte(b.String()), 0o644)
}

// llms writes a plain-text index of the site for agents and crawlers.
func (s *Site) llms() error {
	var b strings.Builder
	fmt.Fprintf(&b, "# %s\n\n> %s\n\n", content.Site.Name, content.Site.Description)
	for _, sec := range content.Sections {
		pages := s.Pages[sec.Slug]
		if len(pages) == 0 && sec.Parent != "" {
			continue
		}
		fmt.Fprintf(&b, "## %s\n%s%s\n", sec.Title, content.Site.BaseURL, sec.Path())
		for _, p := range pages {
			fmt.Fprintf(&b, "- [%s](%s%s)", p.Title, content.Site.BaseURL, p.URL())
			if p.Standfirst != "" {
				fmt.Fprintf(&b, ": %s", p.Standfirst)
			}
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}
	return os.WriteFile(filepath.Join(s.Opts.OutDir, "llms.txt"), []byte(b.String()), 0o644)
}

// loadSkills reads Agent Skills from SKILL.md frontmatter so the Skills page
// is generated from source rather than transcribed from it.
func loadSkills(dir string) ([]ui.Skill, error) {
	if dir == "" {
		return nil, nil
	}
	entries, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var out []ui.Skill
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		path := filepath.Join(dir, e.Name(), "SKILL.md")
		name, desc, err := skillMeta(path)
		if err != nil {
			continue // a directory without a SKILL.md is simply not a skill
		}
		if name == "" {
			name = e.Name()
		}
		out = append(out, ui.Skill{
			Name:        name,
			Description: desc,
			Install:     "/" + e.Name(),
		})
	}
	return out, nil
}

// skillMeta pulls name and description out of SKILL.md YAML frontmatter
// without a full YAML dependency: the format is two flat scalar keys.
func skillMeta(path string) (name, desc string, err error) {
	f, err := os.Open(path)
	if err != nil {
		return "", "", err
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	in := false
	for sc.Scan() {
		line := sc.Text()
		if strings.TrimSpace(line) == "---" {
			if in {
				break
			}
			in = true
			continue
		}
		if !in {
			continue
		}
		k, v, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		v = strings.TrimSpace(v)
		switch strings.TrimSpace(k) {
		case "name":
			name = v
		case "description":
			desc = v
		}
	}
	return name, desc, sc.Err()
}
