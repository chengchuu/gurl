# AGENTS.md

Guidance for contributors and automated coding agents working in `gurl`.

## Package purpose and contract

`gurl` is a Go 1.19 library for inspecting, validating, and modifying URL strings. `go.mod` owns
the module path `github.com/chengchuu/gurl`. The module-root package exposes 18 functions and no
exported types, methods, interfaces, constants, variables, or sentinel errors. Runtime code uses
only the Go standard library.

Preserve function signatures, `net/url` parsing and serialization, fragment-parameter behavior,
empty-value results, and error propagation unless a public behavior change is intentional. A public
API change requires aligned implementation, GoDoc, behavior tests, executable examples,
`examples/basic`, README usage, website examples, and generated API documentation.

Website code must not be imported by the public package. Website changes must not silently change
runtime behavior.

## Identity and configuration

`go.mod` is authoritative for the module path and supported Go version. `site.config.json` is the
website and repository identity authority. Its configured owner and project derive the repository
URL, Pages URL and base path, pkg.go.dev link, theme storage key, and service-worker cache prefix.
Keep mirrored README values synchronized; build validation rejects identity drift.

The configured repository is `https://github.com/chengchuu/gurl`. Public Pages are served from
`https://chengchuu.github.io/gurl/` below `/gurl/`. Do not hard-code a different owner, repository,
module path, Pages path, theme key, or cache prefix in templates or tooling.

## Project structure

- `doc.go`, `gurl.go`: package overview, public URL helpers, and the private fragment parser.
- `gurl_test.go`, `example_test.go`: behavior tests and executable consumer documentation.
- `examples/basic`: runnable consumer-style example using the module-root import.
- `examples.txt`: manually maintained extended examples; it is not compiled by tests.
- `cmd/sitegen`: Go CLI with `build`, `validate`, and `preview` commands.
- `internal/site/config.go`: strict configuration loading, derived identity, and invariant checks.
- `internal/site/apidoc.go`: exported root-package API and GoDoc extraction; it excludes internal
  packages and test files.
- `internal/site`: deterministic generation, SEO/PWA output, artifact validation, preview, and tests.
- `site/templates`: shared layout and navigation plus crawlable home, examples, and API pages.
- `site/assets/css` and `site/assets/js`: small Bootstrap-variable styling and shared browser behavior.
- `site/assets/vendor`: website-only Bootstrap 5.3.8 assets.
- `site/assets/icons`: checked-in favicon, PWA, maskable, Apple-touch, and Open Graph source images.
- `guides/THIRD_PARTY_NOTICES.md`: license information for vendored browser assets.
- `dist/pages`: ignored generated Pages artifact, produced locally or in CI.
- `.github/workflows/ci.yml`: package, example, site generation, and artifact quality gates.
- `.github/workflows/pages.yml`: validated artifact upload and GitHub Pages deployment.

## Runtime and data flow

There is no application startup sequence. Consumers import the package and call an exported
function. Each operation parses the input with `url.Parse`, inspects or changes a `url.URL`
component, and returns either a value, a serialized URL, or a validation result.

Query functions use `url.Values`, so setters and deleters can normalize query escaping and order.
Hash-parameter functions split and rebuild query-like fragment text through `parseFragment`; do not
replace that path with `url.Values` without an approved behavior change. Path, host, hostname, and
protocol helpers read or replace parsed URL fields. `GetURLFileType` uses the parsed path, and
`GetBaseURL` clears the query and fragment. Functions returning `error` propagate parser errors;
validation helpers return `false` after parsing failures.

## Site generation flow

`cmd/sitegen` finds the module root and loads `site.config.json`. A build then extracts the public Go
API, reads `examples/basic/main.go` and `example_test.go`, renders three static pages, copies browser
assets and icons, and generates metadata, crawler files, the manifest, and a scoped service worker.
The output is written to `dist/pages`.

Validation checks configured identity, real routes, one `h1` per page, canonical and Open Graph
metadata, initial visible content, internal links, source-derived imports, sitemap entries, icon
dimensions, PWA scope, service-worker behavior, and stale identities. Fix source configuration,
templates, assets, or generator code when validation fails.

## Stable Pages contract

Keep real static pages at `index.html`, `examples/index.html`, and `api/index.html`. All links,
assets, manifest entries, and worker requests must remain repository-subpath-aware below `/gurl/`.
Do not rely on single-page application fallback routing.

The examples page must read checked-in Go source. The API page must parse actual exported
module-root declarations and GoDoc; do not manually maintain generated API declarations or expose
`internal` packages as consumer API.

Every new public page needs crawlable initial HTML, semantic landmarks, exactly one descriptive
`h1`, keyboard-accessible controls, visible focus indicators, and unique accurate canonical and
Open Graph metadata.

## Theme, SEO, and PWA behavior

The shared theme control supports exactly `system`, `light`, and `dark` through Bootstrap color
modes. Keep its storage key target-specific, preserve safe storage handling, respond to system-theme
changes, minimize theme flash, and synchronize `theme-color` on every page.

SEO metadata, JSON-LD, `robots.txt`, and `sitemap.xml` derive from central configuration and use
absolute production URLs. The generated service worker handles only same-origin GET requests within
the project base path, uses network-first HTML and cache-first local assets, bounds runtime entries,
and cleans only target-specific caches. Ordinary non-production file viewing must not register it.

## Commands and generated files

Run from the repository root:

```bash
gofmt -w .
go mod tidy
go test ./...
go test -race ./...
go vet ./...
go build ./...
go list ./...
go list ./... | xargs -n 1 go doc
go run ./examples/basic
go run ./cmd/sitegen build
go run ./cmd/sitegen validate
go run ./cmd/sitegen preview
```

`preview` rebuilds and validates before serving `http://127.0.0.1:4173/gurl/`. Stop the server after
verification.

Generated Pages files must not be edited manually. `dist/pages` should be generated in CI and not
committed. Keep generated output ignored by Git.

## Workflow and change discipline

CI verifies formatting and module tidiness, then runs tests, race tests, vet, builds, the basic
example, site generation, and artifact validation. The Pages workflow repeats package checks,
generates and validates `dist/pages`, uploads that directory, and deploys through the
`github-pages` environment with only the required permissions.

Keep all checks independent of external network services. Add deterministic tests for configuration,
identity, routes, metadata, API extraction, theme, scope, caching, or validation changes. Preserve
third-party license notices. Do not add Node.js for site generation or import Bootstrap into the Go
package. Do not publish releases or tags unless explicitly requested.
