package content

import (
	"math"
	"strconv"
	"strings"
	"testing"
)

// TestAccentContrast is the fitness function for the palette.
//
// Accent is applied to link hover on 19px/500 list titles. Under WCAG that is
// normal text and needs 4.5:1, not the 3:1 that large text gets. An earlier
// palette failed this on three of six values including the one shipping first,
// which is exactly the kind of defect that should be caught by a build rather
// than by a reader.
func TestAccentContrast(t *testing.T) {
	const min = 4.5

	check := func(name, hex string) {
		t.Helper()
		got, err := contrastOnWhite(hex)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if got < min {
			t.Errorf("%s accent %s has contrast %.2f:1 on white, want >= %.1f:1", name, hex, got, min)
		}
	}

	for _, s := range Sections {
		if s.Accent == "" {
			continue // inherits its parent's, checked there
		}
		check(s.Slug, s.Accent)
	}
	// Reserved accents are checked too: a future section must not be able to
	// introduce a failing colour just because it was picked in advance.
	for slug, hex := range Reserved {
		check("reserved/"+slug, hex)
	}
}

func contrastOnWhite(hex string) (float64, error) {
	h := strings.TrimPrefix(hex, "#")
	v, err := strconv.ParseUint(h, 16, 32)
	if err != nil {
		return 0, err
	}
	lin := func(c float64) float64 {
		c /= 255
		if c <= 0.04045 {
			return c / 12.92
		}
		return math.Pow((c+0.055)/1.055, 2.4)
	}
	r := lin(float64(uint8(v >> 16)))
	g := lin(float64(uint8(v >> 8)))
	b := lin(float64(uint8(v)))
	l := 0.2126*r + 0.7152*g + 0.0722*b
	return 1.05 / (l + 0.05), nil
}

// TestAccentInheritance pins the rule that sub-sections take their parent's
// colour, so every page under Not Code is the same green.
func TestAccentInheritance(t *testing.T) {
	cars, ok := Find("cars")
	if !ok {
		t.Fatal("cars section missing")
	}
	notCode, _ := Find("not-code")
	if got, want := AccentOf(cars), AccentOf(notCode); got != want {
		t.Errorf("cars accent = %s, want inherited %s", got, want)
	}
}

func TestPaths(t *testing.T) {
	for _, tc := range []struct{ slug, want string }{
		{"blog", "/blog/"},
		{"cars", "/not-code/cars/"},
		{"skills", "/code/skills/"},
	} {
		s, ok := Find(tc.slug)
		if !ok {
			t.Fatalf("%s missing", tc.slug)
		}
		if got := s.Path(); got != tc.want {
			t.Errorf("%s path = %s, want %s", tc.slug, got, tc.want)
		}
	}
}

// TestFaviconInheritance: a tab switch must not flicker the browser tab icon.
func TestFaviconInheritance(t *testing.T) {
	cars, _ := Find("cars")
	if got := cars.FaviconSlug(); got != "not-code" {
		t.Errorf("cars favicon = %s, want not-code", got)
	}
}

// TestNoOrphans guards the section table: every child must name a parent that
// actually exists, or its URL resolves to nothing.
func TestNoOrphans(t *testing.T) {
	for _, s := range Sections {
		if s.Parent == "" {
			continue
		}
		if _, ok := Find(s.Parent); !ok {
			t.Errorf("section %s names missing parent %s", s.Slug, s.Parent)
		}
	}
}
