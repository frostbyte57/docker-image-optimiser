# docker-image-optimiser (`dio`)

`dio` is a CLI for improving Dockerfiles before they are built.

It finds common image-size and build-speed problems, applies safe rewrites, and
reports changes that need human review. It understands common Node, Python, Go,
Rust, Java, Ruby, PHP, .NET, apt, apk, and dnf workflows.

## Install

Requires Go 1.24+.

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

```bash
go test ./...
go build ./cmd/dio
```
