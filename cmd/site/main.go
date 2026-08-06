// Command site builds freddienortham.com to static HTML.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/freddie-northam/freddienortham.com/internal/build"
	"github.com/freddie-northam/freddienortham.com/internal/content"
)

func main() {
	var (
		out       = flag.String("out", "dist", "output directory")
		contentIn = flag.String("content", "content", "content directory")
		assets    = flag.String("assets", "assets", "static assets directory")
		skills    = flag.String("skills", "vendor/skills", "Agent Skills directory")
		drafts    = flag.Bool("drafts", false, "include pages marked draft")
		serve     = flag.String("serve", "", "serve the output on this address, e.g. :8080")
		watch     = flag.Bool("watch", false, "rebuild when content or assets change")
		check     = flag.Bool("check-links", false, "fail on dead internal links")
		only      = flag.Bool("links-only", false, "check links in an existing build without rebuilding")
	)
	flag.Parse()

	// Checking must not rebuild: the build clears dist/ first, which would
	// delete the stylesheet Tailwind wrote after it and then report every page
	// as having a dead link to it.
	if *only {
		problems := build.CheckLinks(*out)
		for _, p := range problems {
			fmt.Fprintln(os.Stderr, "  "+p)
		}
		if len(problems) > 0 {
			fail(fmt.Errorf("%d dead link(s)", len(problems)))
		}
		fmt.Println("links ok")
		return
	}

	opts := build.Options{
		ContentDir: *contentIn,
		AssetsDir:  *assets,
		OutDir:     *out,
		SkillsDir:  *skills,
		Drafts:     *drafts,
	}

	site, err := build.Run(opts)
	if err != nil {
		fail(err)
	}
	report(site)

	if *check {
		if problems := build.CheckLinks(*out); len(problems) > 0 {
			for _, p := range problems {
				fmt.Fprintln(os.Stderr, "  "+p)
			}
			fail(fmt.Errorf("%d dead link(s)", len(problems)))
		}
		fmt.Println("links ok")
	}

	if *watch {
		go watchLoop(opts)
	}
	if *serve != "" {
		if err := build.Serve(*serve, *out); err != nil {
			fail(err)
		}
	}
}

func report(s *build.Site) {
	fmt.Printf("built %d pages\n", len(s.Rendered))
	for _, p := range content.Site.Placeholders() {
		fmt.Fprintf(os.Stderr, "warning: %s is still a placeholder\n", p)
	}
}

// watchLoop polls for changes. Polling rather than fsnotify keeps the
// dependency list at zero for a loop that only ever watches two directories.
func watchLoop(o build.Options) {
	last := stamp(o)
	for {
		time.Sleep(400 * time.Millisecond)
		if s := stamp(o); s != last {
			last = s
			if site, err := build.Run(o); err != nil {
				fmt.Fprintln(os.Stderr, "rebuild failed:", err)
			} else {
				fmt.Printf("rebuilt %d pages\n", len(site.Rendered))
			}
		}
	}
}

func stamp(o build.Options) string {
	var newest time.Time
	var count int
	for _, dir := range []string{o.ContentDir, o.AssetsDir} {
		_ = filepath.Walk(dir, func(_ string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() {
				return nil
			}
			count++
			if info.ModTime().After(newest) {
				newest = info.ModTime()
			}
			return nil
		})
	}
	return fmt.Sprintf("%d-%d", count, newest.UnixNano())
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "error:", err)
	os.Exit(1)
}
