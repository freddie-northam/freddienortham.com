package content

import (
	"strings"
	"testing"
)

func TestParseFrontmatter(t *testing.T) {
	src := []byte(`---
title: "Billions of tokens later"
slug: billions-of-tokens-later
standfirst: "What a year taught me."
date: 2026-01-14
aliases: ["/blog/tokens/"]
---

Body text here with a [link](/blog/).
`)
	p, err := Parse(src, "blog", "whatever-the-file-is-called")
	if err != nil {
		t.Fatal(err)
	}
	if p.Title != "Billions of tokens later" {
		t.Errorf("title = %q", p.Title)
	}
	// Slug comes from frontmatter, never the filename: renaming a file must
	// not be able to break a published URL.
	if p.Slug != "billions-of-tokens-later" {
		t.Errorf("slug = %q, want the frontmatter value not the filename", p.Slug)
	}
	if p.Date == nil || p.Date.Year() != 2026 || p.Date.Month() != 1 {
		t.Errorf("date = %v", p.Date)
	}
	if len(p.Aliases) != 1 {
		t.Errorf("aliases = %v", p.Aliases)
	}
	if p.URL() != "/blog/billions-of-tokens-later/" {
		t.Errorf("url = %q", p.URL())
	}
	if !strings.Contains(string(p.Body), "<p>") {
		t.Errorf("body not rendered: %q", p.Body)
	}
}

// An undated page must carry nil, not a zero time, or every sort places it in
// the year 1.
func TestUndatedIsNil(t *testing.T) {
	p, err := Parse([]byte("---\ntitle: \"X\"\nslug: x\nmeta: Live\n---\n\nhi\n"), "projects", "x")
	if err != nil {
		t.Fatal(err)
	}
	if p.Date != nil {
		t.Errorf("date = %v, want nil", p.Date)
	}
	if p.Rail() != "Live" {
		t.Errorf("rail = %q, want the meta value", p.Rail())
	}
}

func TestMissingSlugIsAnError(t *testing.T) {
	_, err := Parse([]byte("---\ntitle: \"X\"\n---\n\nhi\n"), "blog", "some-file")
	if err == nil {
		t.Fatal("want an error when slug is absent")
	}
	if !strings.Contains(err.Error(), "slug") {
		t.Errorf("error should name the missing field, got %v", err)
	}
}

func TestMissingTitleIsAnError(t *testing.T) {
	if _, err := Parse([]byte("---\nslug: x\n---\n\nhi\n"), "blog", "f"); err == nil {
		t.Fatal("want an error when title is absent")
	}
}

func TestMalformedDateIsAnError(t *testing.T) {
	_, err := Parse([]byte("---\ntitle: \"X\"\nslug: x\ndate: 14/01/2026\n---\n\nhi\n"), "blog", "f")
	if err == nil {
		t.Fatal("want an error on a non ISO date")
	}
	if !strings.Contains(err.Error(), "YYYY-MM-DD") {
		t.Errorf("error should state the expected format, got %v", err)
	}
}

// A sub-section's page is the sub-section itself, not a child of it.
func TestSubSectionURL(t *testing.T) {
	p, err := Parse([]byte("---\ntitle: \"Cars\"\nslug: _index\n---\n\nhi\n"), "cars", "cars")
	if err != nil {
		t.Fatal(err)
	}
	if got := p.URL(); got != "/not-code/cars/" {
		t.Errorf("url = %q, want /not-code/cars/", got)
	}
}

func TestFurnitureThresholds(t *testing.T) {
	short := Page{WordCount: 300}
	mid := Page{WordCount: 1200}
	long := Page{WordCount: 2500}

	if short.NeedsAnchors() || short.NeedsTOC() {
		t.Error("a short page should carry no furniture")
	}
	if !mid.NeedsAnchors() || mid.NeedsTOC() {
		t.Error("a mid-length page wants anchors but no table of contents")
	}
	if !long.NeedsTOC() {
		t.Error("a long page wants a table of contents")
	}
}

func TestImagesAreStamped(t *testing.T) {
	p, err := Parse([]byte("---\ntitle: \"X\"\nslug: x\n---\n\n![alt](photo.jpg)\n"), "blog", "f")
	if err != nil {
		t.Fatal(err)
	}
	body := string(p.Body)
	if !strings.Contains(body, `loading="lazy"`) || !strings.Contains(body, `decoding="async"`) {
		t.Errorf("image not stamped: %q", body)
	}
}
