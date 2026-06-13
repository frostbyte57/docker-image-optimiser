# docker-image-optimiser (`dio`)

`dio` is a CLI for improving Dockerfiles before they hit CI/CD.

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

`bench` and `inspect` require Docker. Linting and rewriting do not.

## Quick Start

```bash
dio --help
dio lint Dockerfile
dio optimize --check Dockerfile
dio fix --write Dockerfile
```

Use `dio --help` for the full command list, and `dio <command> --help` for
command-specific flags and examples.

## CI

Use `optimize --check` to fail a pipeline when a Dockerfile still has pending
optimisations or manual actions.

```bash
dio optimize --check --format json Dockerfile
```

The JSON output is intended for CI logs, bots, and annotations.

## Development

```bash
go test ./...
go build ./cmd/dio
```
