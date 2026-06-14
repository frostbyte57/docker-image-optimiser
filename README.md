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

## Use in CI

`dio` is built to be a **pipeline gate**: run `dio optimize --check` first, and
only build the image once the Dockerfile is optimised. It exits non-zero when any
fix is still pending, failing the job before the slow `docker build` runs.

```bash
dio optimize --check --format github Dockerfile
```

See [examples/ci/](examples/ci/) for the bundled GitHub Action and full
gate-then-build examples for GitHub Actions and GitLab CI.

## Documentation

- [Use in CI](examples/ci/README.md) — pipeline gate, GitHub Action, output formats.
- [Command help](docs/help/README.md) — every subcommand and flag.
- [Development](docs/development.md) — building, testing, hardening checks.
- [Deployment & Releases](docs/deployment.md) — tagging, GoReleaser, signing.
