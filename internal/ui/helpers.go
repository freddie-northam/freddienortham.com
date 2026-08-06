package ui

import (
	"fmt"
	"strings"

	"github.com/freddie-northam/freddienortham.com/internal/content"
)

func countLabel(n int) string {
	if n == 1 {
		return "1 entry"
	}
	return fmt.Sprintf("%d entries", n)
}

// spanLabel reports the year range covered by a list, or an empty string when
// nothing in it is dated.
func spanLabel(pages []content.Page) string {
	lo, hi := 0, 0
	for _, p := range pages {
		if p.Date == nil {
			continue
		}
		y := p.Date.Year()
		if lo == 0 || y < lo {
			lo = y
		}
		if y > hi {
			hi = y
		}
	}
	switch {
	case lo == 0:
		return ""
	case lo == hi:
		return fmt.Sprint(lo)
	default:
		return fmt.Sprintf("%d — %d", lo, hi)
	}
}

// hasThumbSlot reports whether a section's rows reserve space for a leading
// image. Projects do; prose sections do not.
func hasThumbSlot(p content.Page) bool {
	return p.Section == "projects"
}

// Words splits a string into words, then each word into characters, for the
// weight wave. Grouping by word matters: every letter is an inline-block, so
// without a word wrapper the browser will happily break "Northam" across two
// lines mid-word.
func Words(s string) [][]string {
	var out [][]string
	for _, w := range strings.Fields(s) {
		letters := make([]string, 0, len(w))
		for _, r := range w {
			letters = append(letters, string(r))
		}
		out = append(out, letters)
	}
	return out
}
