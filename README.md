# gurl

`gurl` is a Go library for inspecting, validating, and modifying URL components. It provides
focused helpers for query and fragment parameters, paths, hosts, protocols, file extensions, and
base URLs without performing network requests.

## Install

Go 1.19 or later is supported.

```bash
go get github.com/chengchuu/gurl
```

```go
import "github.com/chengchuu/gurl"
```

## Usage

```go
value, err := gurl.GetQueryParam("https://example.com/search?q=go&lang=en", "q")
if err != nil {
	log.Fatal(err)
}
fmt.Println(value) // go
```

Run the complete example with `go run ./examples/basic`.

The public API includes helpers to:

- get, set, and delete URL query parameters;
- get, set, and delete query-like parameters in URL fragments;
- read and replace paths, hosts, hostnames, and protocols;
- remove queries and fragments to obtain a base URL;
- read a file extension from a URL path; and
- check for a scheme and host with `CheckValid`, or require an `http` or `https` scheme together
  with a host by using `CheckValidHTTPURL`.

Functions that return an error pass through failures from Go's URL parser. URL setters serialize
their result with `net/url`; query serialization can normalize ordering and escaping.

## Documentation

- [Public website](https://chengchuu.github.io/gurl/)
- [Source-derived examples](https://chengchuu.github.io/gurl/examples/)
- [Generated API reference](https://chengchuu.github.io/gurl/api/)
- [pkg.go.dev API documentation](https://pkg.go.dev/github.com/chengchuu/gurl)
- [GitHub repository](https://github.com/chengchuu/gurl)
- [Basic example](examples/basic/main.go)

The examples page renders `examples/basic/main.go` and `example_test.go` directly from checked-in Go
source. The API page extracts exported declarations and GoDoc from the module-root package.

## Development

Run package checks from the repository root:

```bash
gofmt -w .
go mod tidy
go test ./...
go test -race ./...
go vet ./...
go build ./...
go run ./examples/basic
```

Generate and validate the GitHub Pages artifact:

```bash
go run ./cmd/sitegen build
go run ./cmd/sitegen validate
```

Preview the production-like site at `http://127.0.0.1:4173/gurl/`:

```bash
go run ./cmd/sitegen preview
```

`dist/pages` is generated and ignored by Git. Update source configuration, GoDoc, templates, or
assets and rebuild instead of editing generated Pages files.

## License

The Go package and project-authored website code are available under the [MIT License](LICENSE).
Vendored browser-asset notices are in [guides/THIRD_PARTY_NOTICES.md](guides/THIRD_PARTY_NOTICES.md).
