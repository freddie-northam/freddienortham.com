# Personal website: design spec

Created: 2026-08-05
Revised: 2026-08-06 (architecture review findings ARCH-001 through ARCH-019 applied)
Status: awaiting review

## 1. Purpose

A personal site whose primary job is to be a craft artefact. The design, typography,
and motion are the portfolio; the content is what fills it. A visitor should form the
impression "this person builds beautiful things" before reading a full sentence.

Secondary job: a durable home for writing, projects, and personal interests that grows
over years without a rewrite.

## 2. Constraints

These drive every decision below.

1. **Minimal LOC.** The owner maintains this alone. Target is roughly 450 lines of Go
   and templ, 400 lines of CSS and JS. No file over ~80 lines.
2. **Modular.** Adding a list-shaped section costs one table entry and a content folder.
   Never a code change. See §4.1 for the honest limits of this claim.
3. **Not TypeScript.** The owner writes enough of it elsewhere.
4. **No Node.** Tailwind v4 runs from its standalone binary; config lives in CSS.
5. **No runtime.** Static output only. No server, no ops, no bill.

## 3. Decisions

| Decision | Choice | Rejected |
|---|---|---|
| Language | Go | TypeScript (owner fatigue), Rust (slow iteration on a design-led site), Elixir (needs a running server) |
| Templating | templ | `html/template` (no type safety, more call-site boilerplate) |
| Output | Static generator to `dist/` | Server-rendered with htmx (adds ops cost for mostly-static prose) |
| Hosting | Cloudflare Pages, building from a connected repo | Direct `wrangler` upload from CI (more reproducible, loses push-to-deploy) |
| CSS | Tailwind v4, standalone CLI | Plain CSS with custom properties (fewer lines, but loses the component vocabulary) |
| Typeface | Geist and Geist Mono, self-hosted, SIL OFL 1.1 | Roboto Flex, Instrument Serif, Bricolage Grotesque |
| Icons | Six hand-rolled inline SVGs as templ components | Lucide (1500 icons for a site needing 6), font glyph arrows (fallback risk) |
| Visual direction | Restrained single column, one loud display title per page | Full Cloudflare-style blueprint (high maintenance, unforgiving of thin content); pure Nico-style restraint (does not demonstrate craft) |
| Colour | Monochrome plus one accent per section | Single sitewide accent, zero colour, illustrated characters |
| Echo stack technique | Layered `text-shadow` on one element | Duplicate DOM spans (find-in-page and selection pollution) |
| Dark mode | Not in v1 | |

**Note on Geist.** Geist is variable on weight only. It has no width axis, so the
"font glyph warp" idea cannot be a width warp. It is reinterpreted as a weight wave
(§8.2), which has the incidental benefit of causing no layout shift.

**Note on Tailwind and component ports.** Tailwind was kept partly to preserve access
to templUI and the shadcn vocabulary. Be aware that templUI is substantially younger
and thinner than shadcn itself. Tailwind is defensible on its own merits; do not count
on the port ecosystem being there when the Components section arrives.

**Accepted compromise.** `Section.Accent` (§4.1) puts a presentation concern in the
content domain. This is deliberate: separating a theme map from the section table costs
more lines and more indirection than it saves at this scale. The LOC constraint wins.

## 4. Architecture

A build binary walks `content/`, renders templ components, and writes static HTML to
`dist/`. There is no runtime component.

```
cmd/site/main.go             flag parsing only                       ~15
internal/build/build.go      orchestration: walk, render, write      ~70
internal/build/assets.go     favicon generation, static copy         ~50
internal/content/parse.go    frontmatter + markdown -> Page          ~70
internal/content/sections.go section table, index building           ~50
internal/ui/layout.templ     shell, nav, head metadata, footer       ~70
internal/ui/list.templ       generic section index, all sections     ~40
internal/ui/page.templ       generic article page, all sections      ~35
internal/ui/title.templ      echo stack display title                ~15
internal/ui/home.templ       homepage                                ~50
internal/ui/icon.templ       six inline SVG icons                    ~40
internal/feed/rss.go         build-time RSS                          ~40
assets/site.js               four motion modules                    ~170
assets/site.css              design system + Tailwind directives    ~220
content/                     markdown, one folder per section
dist/                        build output, gitignored
```

`main.go` is deliberately reduced to argument parsing. Orchestration moved to
`internal/build` so the file most likely to accrete responsibilities is the file least
able to hold them.

**LOC caveat.** These estimates assume Tailwind utility classes inline in markup.
Expect `internal/ui/*.templ` to run 20 to 40 percent longer than the equivalent would
with plain CSS. That cost was accepted knowingly when Tailwind was chosen.

### 4.1 Sections are data

The central modularity mechanism. One slice defines every section. Nav, index pages,
favicon tint, RSS inclusion, and the accent custom property all derive from it.

```go
type Section struct {
    Slug     string
    Title    string
    Accent   string
    Template string // "" means the generic list template
    InNav    bool
    InFeed   bool   // contributes to /feed.xml
}

var Sections = []Section{
    // Top level.
    {Slug: "blog",     Title: "Blog",     Accent: "#CC3A00", InNav: true, InFeed: true},
    {Slug: "projects", Title: "Projects", Accent: "#0033FF", InNav: true},
    {Slug: "code",     Title: "Code",     Accent: "#7C3AED", InNav: true},
    {Slug: "not-code", Title: "Not Code", Accent: "#047857", InNav: true},

    // Sub-pages. Tabs are real URLs, so these are sections, not JS state,
    // and each one is shareable and works without JavaScript.
    {Slug: "components", Title: "Components", Parent: "code", Template: "components"},
    {Slug: "cars",       Title: "Cars",       Parent: "not-code"},
    // ...twelve children in total
}
```

Children inherit their parent's accent, so every page under Not Code is the same green,
and share the parent's favicon so switching tabs does not flicker the browser tab icon.

**What this claim actually covers.** Blog and Projects are the same shape: a title, a
subtitle, and one metadata value on a right rail. Only the metadata differs (a date
versus a status). So v1 ships **one** list template, not two, and there is no `Kind`
enum. Code and Not Code will fit the same template when they arrive.

**Where the claim stops.** Photos is a gallery: masonry or justified rows, a lightbox,
no titles, no standfirst. Games is an interactive application surface. Neither is a
list of anything. Both will need `Template: "gallery"` and `Template: "game"`, which
means a real template file, not a config line.

Stated plainly so it is not discovered mid-build: **adding a list-shaped section is one
line. Photos and Games are custom templates and always were.** Four of six deferred
sections are free; two are not.

### 4.2 Content model

One struct. Everything else is derived.

```go
type Page struct {
    Section    string     // owning section slug
    Slug       string     // from frontmatter, NOT the filename
    Aliases    []string   // previous slugs, generate redirects
    Title      string
    Standfirst string     // one line under the title, also the RSS description
    Date       *time.Time // nil when undated; never a zero-value sentinel
    Draft      bool
    Body       template.HTML
}
```

Markdown frontmatter:

```yaml
---
title: Billions of tokens later
slug: billions-of-tokens-later
standfirst: What a year inside the model actually taught me.
date: 2026-01-14
aliases: []
draft: false
---
```

**Slug is explicit, not derived from the filename.** Renaming a file must never break a
public URL. `aliases` lists previously-published slugs and generates a Cloudflare Pages
`_redirects` entry per alias at build time. Both cost a handful of lines now and are
impossible to retrofit once links are in the wild.

`draft: true` is excluded from production output, indexes, and RSS, but **included by
`make dev`** (§6.2) so drafts are previewable. This matters because most posts are
unwritten.

### 4.3 Routes and URL policy

```
/                        home
/blog/                   section index
/blog/<slug>/            article
/projects/               section index
/projects/<slug>/        project, only where a file exists
/feed.xml                RSS: every section with InFeed, merged, newest first
/_redirects              generated from all Page.Aliases
/404.html
```

Routes generate from `Sections` and the content tree. There is no route table.

**Trailing slash policy: directory-style with a trailing slash.** Every page is written
as `<path>/index.html`. This is pinned rather than left to Cloudflare's defaults so that
relative asset URLs resolve consistently and the same page is never reachable at two
URLs.

`/feed.xml` is a single merged feed across all `InFeed` sections, newest first. It is
declared in `<head>` via `<link rel="alternate" type="application/rss+xml">`.

### 4.4 Page metadata

Every page emits, from `Page` and `Section` with no per-page authoring:

- `<title>` and `<meta name="description">` from Title and Standfirst
- `<link rel="canonical">` at the absolute trailing-slash URL on `https://freddienortham.com`
- OpenGraph and Twitter card tags: title, description, url, type, site name
- The section's favicon (§6.3)

Build-time OG **image** generation is deferred, but the tags ship in v1 so that a shared
link is never bare. This is a design surface: the first impression of a site built to
make impressions is often the link preview, not the page.

## 5. Markdown pipeline

goldmark, with:

- YAML frontmatter extension.
- chroma for code blocks, monochrome theme, accent on the active line only.
- A small custom renderer applying the shadcn typeset rules to prose elements
  (`https://ui.shadcn.com/docs/typeset`), so article body styling is defined once.
- Typographer extension for correct quotes and dashes.
- **An image renderer that stamps `width`, `height`, `loading="lazy"` and
  `decoding="async"`.** Dimensions are read from the file at build time. Un-dimensioned
  images cause layout shift, which is the most visible way a craft-forward site looks
  amateur.

### 5.1 Images

Images live in `content/<section>/<slug>/` beside the markdown that references them, and
are copied to the matching output path. Referenced relatively (`![](photo.jpg)`), so a
post is a self-contained folder that can be moved or deleted as a unit.

v1 does no resizing or format conversion. Responsive `srcset` generation is deferred to
whenever Photos is built, since that is the section that will force the issue.

## 6. Build and development

### 6.1 Production build

```makefile
build:
	go run github.com/a-h/templ/cmd/templ@$(TEMPL_VERSION) generate
	./bin/tailwindcss -i assets/site.css -o dist/site.css --minify
	go run ./cmd/site -out dist

check:
	go vet ./...
	go test ./...
	go run ./cmd/site -check-links -out dist
```

templ runs via `go run`, so its version is verified through `go.sum` for free. The
Tailwind standalone binary is a raw GitHub release download and gets **no such
guarantee**: the Makefile target that fetches it must verify a pinned sha256 before the
binary is executed. An unverified third-party executable in a deploy pipeline is not
acceptable, and the fix is one line.

Cloudflare Pages runs `make build && make check`. The link checker lives in `check`
rather than `build` so a broken link fails the deploy but does not block local preview
while drafting. `GO_VERSION` is pinned in the Pages environment.

### 6.2 Local development

The loop where 95 percent of the time is spent, and the thing the first draft of this
spec forgot entirely.

```makefile
dev:
	templ generate --watch &
	./bin/tailwindcss -i assets/site.css -o dist/site.css --watch &
	go run ./cmd/site -out dist -drafts -serve :8080 -watch
```

`-serve` runs `http.FileServer` over `dist/` with a rebuild on content change. `-drafts`
includes `draft: true` pages. Roughly 30 lines, and it makes the difference between
tuning easing curves pleasantly and doing it through `make build`.

### 6.3 Favicons

One geometric mark, drawn per section in the section's accent.

**Safari does not support SVG favicons.** Emitting SVG only means Safari shows a blank
default, which on a site this fussy is a visible defect. So the mark is drawn directly
with Go's `image` and `image/png` (no SVG rasterizer dependency) and emitted as both
`favicon-<slug>.png` at 32 and 180 pixels and `favicon-<slug>.svg` for browsers that
prefer it. Keeping the mark geometric enough to draw in ~20 lines of `image/draw` is a
design constraint, not an afterthought.

## 7. Design system

### 7.1 Type

Geist and Geist Mono, self-hosted woff2, `font-display: swap`. No Google CDN: faster,
and no third party observing readers. Both are SIL OFL 1.1; the licence file ships in
`assets/fonts/OFL.txt` as the licence requires.

| Role | Size | Weight | Leading | Tracking |
|---|---|---|---|---|
| Home display name | 96px | 800 | 0.85 | -0.05em |
| Page title (echo) | 76px | 800 | 0.85 | -0.05em |
| Article h1 | 40px | 700 | 1.1 | -0.03em |
| Article h2 | 24px | 500 | 1.3 | -0.02em |
| Article body | 17px | 400 | 1.65 | 0 |
| UI and list items | 19px | 500 | 1.4 | -0.01em |
| Secondary and standfirst | 15px | 400 | 1.5 | 0 |
| Metadata (mono) | 11px | 400 | 1.4 | 0.06em, uppercase |

Display sizes step down at 768px: 96 to 56, 76 to 44.

### 7.2 Colour

```
--paper    #FFFFFF
--ink      #0A0A0A
--grey-700 #4A4A4A   secondary prose
--grey-500 #8A8A8A   metadata
--grey-300 #C4C4C4   echo layer, disabled
--grey-100 #EDEDED   rules, hairlines
```

Section accents, every value verified against `--paper`:

| Section | Accent | Contrast | Status |
|---|---|---|---|
| blog | `#CC3A00` | 5.03:1 | live in v1 |
| projects | `#0033FF` | 7.20:1 | live in v1 |
| code | `#7C3AED` | 5.70:1 | reserved |
| not-code | `#047857` | 5.48:1 | reserved |
| photos | `#C2185B` | 5.87:1 | reserved |
| games | `#A16207` | 4.92:1 | reserved |

**Contrast rule, binding on all future sections.** Accent is applied to link hover on
19px/500 list titles, which is normal text under WCAG and requires 4.5:1. Every accent
above clears it with margin, and they sit in a 4.9 to 7.2 band so they read as one
family. The earlier palette failed this on three of six values, including the one
shipping in v1. §9 encodes this as a test so it cannot regress.

The active nav underline is a 2px border, not text, and needs only 3:1. Nav label text
stays `--ink`; only the underline carries accent. Accent appears in exactly three
places: that underline, link hover, and the favicon. Never a fill, never on more than
one element at a time.

### 7.3 Layout

Single column, 660px measure, centred. Page gutter 72px desktop, 24px below 768px.
Metadata sits on a right-aligned rail against the column edge. Section rhythm is a 56px
stack; article prose is a 28px stack.

### 7.4 Icons

**No icon library.** The site needs six icons; Lucide ships 1500. Font glyph arrows
(`↗`, `→`) were rejected because a missing glyph falls back to another family and
renders at the wrong weight and size, which is exactly the detail that betrays a craft
site.

Six inline SVGs as templ components in `icon.templ`, roughly 40 lines total:

| Icon | Used for |
|---|---|
| `arrow-up-right` | external links, in list rows and the footer |
| `arrow-right` | "view all", next article |
| `arrow-left` | previous article |
| `rss` | feed link in the footer |
| `copy` | code block copy button |
| `link` | heading anchors in articles |

All are 16px on a 24px grid, `stroke-width: 1.5` to sit correctly beside Geist at 15 to
19px, and `stroke="currentColor"` so accent inheritance on hover is free with no extra
rules. Social links are plain text labels with `arrow-up-right`, not brand logos: brand
marks would import six colour identities into a monochrome design and need licence
attention.

## 8. Motion

Four moments, one per idea, no more. Each is a self-registering module in `site.js` keyed
off a data attribute, so deleting one means deleting one function and one attribute.
Nothing else breaks.

### 8.1 Echo stack (`data-echo`)

Page titles render four ghost copies above the solid word via **layered `text-shadow` on
a single element**, at `--grey-300` through `--grey-100`, offset 12, 24, 36 and 48px
upward.

Duplicate DOM spans were rejected. Five copies of the word in the document means
find-in-page reports five phantom matches, text selection yields
`writingwritingwritingwritingwriting`, and every copy needs `aria-hidden` and
`user-select: none` to behave. `text-shadow` has none of these problems, adds zero
nodes, and is invisible to assistive technology by construction.

The cost is that per-ghost stagger is no longer possible: browsers interpolate a shadow
stack layer by layer, so the whole stack animates as one. An IntersectionObserver
transitions the stack from collapsed (all offsets 0, transparent) to rest over 420ms,
`cubic-bezier(0.22, 1, 0.36, 1)`. The ghosts rise together rather than in sequence,
which reads nearly identically at this scale and costs four fewer DOM nodes per page.

### 8.2 Weight wave (`data-wave`)

The homepage name splits into per-letter spans. Each letter's weight is interpolated by
cursor distance with a gaussian falloff (sigma 90px) across 200 to 900, lerped at 0.15
per frame from a single rAF loop. Because Geist varies on weight and not width, advance
widths are stable and the word does not reflow.

**Mobile has no cursor, so the site's signature moment would simply not exist for the
majority of visitors.** On coarse pointers (`matchMedia('(pointer: coarse)')`) the wave
runs autonomously instead: the gaussian centre travels across the name on a 3.2s loop,
easing at each end, and pauses when the element scrolls out of view. Touching the name
snaps the centre to the touch point.

The rAF loop only runs while the element is visible and the tab is focused. Weight
changes on 16 spans trigger layout, so an unconditioned loop is a real battery cost for
an offscreen decoration.

### 8.3 Hover grid (`data-grid`)

Hovering a list row fades in 1px hairlines at 24px intervals behind it, opacity 0 to 0.5
over 180ms. This is the Cloudflare blueprint showing through, on demand only. Rendered
as a `repeating-linear-gradient` background, so it costs no nodes and no JS beyond the
hover class. Disabled entirely on coarse pointers, where hover does not exist and the
sticky `:hover` state on tap would leave a row stuck highlighted.

### 8.4 Days on earth (`data-days`)

A monotonic counter in the footer from the owner's birth date, in `tabular-nums` so
digits do not jitter, showing enough decimal places to visibly flicker.

It is below the fold on every page. It updates from a rAF loop gated on an
IntersectionObserver **and** `document.visibilitychange`, so it costs nothing while
offscreen or backgrounded. A naive `setInterval` would write to the DOM ten times a
second, forever, on every page, for an element nobody is looking at.

### 8.5 Reduced motion

Under `prefers-reduced-motion: reduce`, every module renders its static end state and
registers no listeners. Echo shows fully settled, the wave rests at 800, the grid stays
hidden, the counter shows whole days and does not tick.

## 9. Testing

- Golden-file tests over `content/parse.go`: frontmatter, drafts, aliases, missing
  fields, malformed dates, nil versus present date.
- A golden-file test over one rendered article, guarding template regressions.
- **A contrast test asserting every `Section.Accent` clears 4.5:1 against `--paper`.**
  About 20 lines, and it makes the palette defect structurally unrepeatable rather than
  relying on anyone remembering.
- A slug uniqueness test: duplicate slugs within a section fail the build.
- A redirect integrity test: every alias resolves to a slug that exists.
- A link checker across `dist/`, run in `make check`.
- `go vet` and `templ fmt --check` in CI.

## 10. Scope

**Built and passing `make check`:** Home, Blog, Projects, Code (Components, Languages,
Stacks, Skills), Not Code (eight sub-pages), RSS, `/llms.txt`, 404, canonical and
OpenGraph tags, the four motion moments, six icons, per-section favicons with PNG
fallback, redirects, a local dev server with draft preview, responsive down to 375px.

Code and Not Code were pulled into v1 during implementation. Not Code is the only
section that could be filled immediately, since its content already existed in the
original brief, whereas the Blog cannot be filled until the writing is done. Deferring
the full section and keeping the empty one was the wrong way round.

**Out, deferred:** Photos, Games, dark mode, illustrated characters, the widget ideas,
build-time OG image generation, responsive `srcset`. Photos and Games both need custom
templates rather than the generic list, which is precisely why they are not "one line".

**Explicitly not doing:** a CMS, comments, analytics beyond Cloudflare's defaults, a
newsletter.

**Prerequisite, not yet met:** Go is not installed on the development machine, and this
directory is not a git repository. Cloudflare Pages builds from a connected repo, so
both are blocking before any code is written.

## 11. Migration from the current site

The existing site at `freddienortham.com` is a single-page SPA with a catch-all route:
`/blog`, `/about`, `/projects`, `/writing` and `/work` all return 200 serving the same
index page. There is no real content at any deep URL and exactly one outbound link
(XALT, `https://wearexalt.com/`).

Consequences, all favourable:

- **No redirect burden.** Nothing published lives at a URL that must be preserved. The
  `aliases` mechanism in §4.2 exists for future renames, not for this cutover.
- **Already served by Cloudflare.** Pointing the domain at Pages is a custom-domain
  configuration, not a nameserver migration.
- Paths that currently return 200 with placeholder content will become either real pages
  (`/blog/`, `/projects/`) or a proper 404. Both are an improvement on a soft 200.

Cutover is therefore: build, deploy to a Pages preview URL, verify, then attach the
custom domain.

## 12. Open questions

These block content, not construction. The site can be built and reviewed with
placeholder copy while they are resolved.

1. Bio. Partially answered by the current site: Head of Product and Innovation at XALT,
   2024 to present, leading product vision and platform strategy across brand-fan
   engagement, data intelligence, and community growth. **The facts are captured; the
   copy needs rewriting.** The existing wording ("shaping digital ecosystems at the
   intersection of media, AI, and data") is the register this site should avoid.
2. The five projects (Record, Halero, XALT, Kanzoro, Spaniel): one line each covering
   what it is, the owner's role, whether it is live, and a public link if any. Only XALT
   is known so far (`https://wearexalt.com/`).
3. Birth date, for the days-on-earth counter.
4. "J-space": what this refers to. Not guessing at an intellectual reference.
5. Which social links belong in the footer.
6. Should the Blog section be renamed? It will hold a funeral poem and a birthday speech
   alongside essays, and "Blog" sits oddly with those. Raised once, declined once,
   recorded here because the content makes it worth a second look.
