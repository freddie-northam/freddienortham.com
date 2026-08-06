TEMPL_VERSION := v0.3.1020
TAILWIND_VERSION := v4.3.3
TAILWIND_SHA256 := cdf646702987a743464dff4d9c60fd4480d1c1e73dd819a9a67f1078815dce9d
TAILWIND := bin/tailwindcss

.PHONY: build check dev clean tools

# Order matters: the site build clears dist/ first, so Tailwind must write
# its stylesheet afterwards or the compiled CSS is deleted before it is served.
build: tools
	go run github.com/a-h/templ/cmd/templ@$(TEMPL_VERSION) generate
	go run ./cmd/site -out dist
	$(TAILWIND) -i assets/site.css -o dist/site.css --minify

# The link checker lives here rather than in `build` so a dead link fails the
# deploy without blocking local preview while drafting.
check: build scale
	go vet ./...
	go test ./...
	go run ./cmd/site -links-only -out dist

dev: tools
	go run github.com/a-h/templ/cmd/templ@$(TEMPL_VERSION) generate
	$(TAILWIND) -i assets/site.css -o dist/site.css --watch=always & \
	go run ./cmd/site -out dist -drafts -serve :8080 -watch

tools: $(TAILWIND)

# Tailwind ships as a raw GitHub release binary with no module checksum behind
# it, so verify the hash before executing it. templ needs no equivalent: it
# runs via `go run`, which go.sum already covers.
$(TAILWIND):
	@mkdir -p bin
	@echo "fetching tailwindcss $(TAILWIND_VERSION)"
	@curl -sL "https://github.com/tailwindlabs/tailwindcss/releases/download/$(TAILWIND_VERSION)/tailwindcss-macos-arm64" -o $@
	@echo "$(TAILWIND_SHA256)  $@" | shasum -a 256 -c - || (rm -f $@; echo "checksum mismatch, refusing to run"; exit 1)
	@chmod +x $@

clean:
	rm -rf dist bin

# Fitness function for the type scale. Every font-size must reference a token
# in :root; a design system drifts one hardcoded 14px at a time.
scale:
	@if grep -nE '^\s*font-size:\s*[0-9]+px' assets/site.css; then \
		echo "raw px font-size above: use a --fs-* token"; exit 1; \
	else echo "type scale ok"; fi
