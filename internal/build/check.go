package build

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

var hrefRe = regexp.MustCompile(`(?:href|src)="([^"]+)"`)

// CheckLinks walks the built output and reports internal links that resolve to
// nothing. A renamed slug should fail the build rather than ship a dead link.
//
// This runs in `make check`, not `make build`, so a broken link blocks the
// deploy without blocking local preview while drafting.
func CheckLinks(outDir string) []string {
	var problems []string
	seen := map[string]bool{}

	_ = filepath.Walk(outDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, ".html") {
			return nil
		}
		body, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		src, _ := filepath.Rel(outDir, path)
		for _, m := range hrefRe.FindAllStringSubmatch(string(body), -1) {
			target := m[1]
			if skipLink(target) {
				continue
			}
			key := src + " -> " + target
			if seen[key] {
				continue
			}
			seen[key] = true
			if !resolves(outDir, target) {
				problems = append(problems, fmt.Sprintf("%s: dead link %s", src, target))
			}
		}
		return nil
	})
	sort.Strings(problems)
	return problems
}

func skipLink(t string) bool {
	switch {
	case t == "", strings.HasPrefix(t, "#"),
		strings.HasPrefix(t, "http://"), strings.HasPrefix(t, "https://"),
		strings.HasPrefix(t, "mailto:"), strings.HasPrefix(t, "data:"):
		return true
	}
	return false
}

func resolves(outDir, target string) bool {
	if i := strings.IndexAny(target, "#?"); i >= 0 {
		target = target[:i]
	}
	if target == "" {
		return true
	}
	p := filepath.Join(outDir, filepath.FromSlash(strings.TrimPrefix(target, "/")))
	if st, err := os.Stat(p); err == nil {
		if !st.IsDir() {
			return true
		}
		_, err := os.Stat(filepath.Join(p, "index.html"))
		return err == nil
	}
	_, err := os.Stat(p + "index.html")
	return err == nil
}

// Serve runs the built site locally. Paired with -watch it is the loop where
// almost all of the design work happens.
func Serve(addr, dir string) error {
	fs := http.FileServer(http.Dir(dir))
	mux := http.NewServeMux()
	mux.Handle("/", noCache(fs))
	fmt.Printf("serving %s on http://localhost%s\n", dir, addr)
	return http.ListenAndServe(addr, mux)
}

func noCache(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		h.ServeHTTP(w, r)
	})
}
