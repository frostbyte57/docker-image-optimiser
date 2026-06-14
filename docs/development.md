# Development

`dio` is a Go CLI built with Cobra. Dockerfile parsing is delegated to
BuildKit's Dockerfile parser, then adapted into dio's small internal rule API.
Docker is required only for `dio bench`, `dio inspect`, and integration tests.

```bash
go test ./...
go test -race ./...
go build ./cmd/dio
```

Optional hardening checks mirror CI:

```bash
golangci-lint run
govulncheck ./...
go test ./internal/parser -run '^$' -fuzz=FuzzParse -fuzztime=20s
go test -tags=integration ./internal/build ./internal/inspect
```

Install local pre-commit hooks with:

```bash
pre-commit install
```
