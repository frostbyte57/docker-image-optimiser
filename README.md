# docker-image-optimiser (`dio`)

`dio` is a CLI for improving Dockerfiles before they are built.

It finds common image-size and build-speed problems, applies safe rewrites, and
reports changes that need human review. It understands common Node, Python, Go,
Rust, Java, Ruby, PHP, .NET, apt, apk, and dnf workflows.

## Install

Requires Go 1.25.8+.

```bash
go install github.com/yuxiangchang/docker-image-optimiser/cmd/dio@latest
```

From a local clone:

```bash
go build -o dio ./cmd/dio
```

For command help, see [docs/help/README.md](docs/help/README.md) or run
`dio --help`.

## Development

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

Releases are built by GoReleaser from `v*` tags. The release workflow builds
cross-platform binaries, generates checksums and SBOMs, and signs checksums with
keyless cosign signing.
