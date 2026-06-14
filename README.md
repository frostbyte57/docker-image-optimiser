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

`dio` is built to be a **pipeline gate**: run the check first, and only build the
image once the Dockerfile is optimised. `dio optimize --check` exits non-zero when
any optimisation or manual fix is still pending, so it fails the job before the
(slow, expensive) `docker build` runs.

```bash
dio optimize --check Dockerfile          # exit 1 if anything is pending
dio optimize --check --format github ... # same, with inline PR annotations
dio optimize --check --format json  ...  # machine-readable summary
```

### GitHub Actions

Use the bundled composite action — it installs `dio`, runs the gate, and emits
inline annotations on the pull request:

```yaml
jobs:
  dio:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: yuxiangchang/docker-image-optimiser@v1
        with:
          dockerfile: Dockerfile
          context: .
  build:
    needs: dio          # only build once the Dockerfile is optimised
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: docker/build-push-action@v6
        with: { context: ., push: false, tags: myapp:${{ github.sha }} }
```

Action inputs: `dockerfile`, `context`, `conservative`, `version`,
`fail-on-issues`. Full gate-then-build examples for GitHub Actions and GitLab CI
live in [`examples/ci/`](examples/ci/).

The `github` format maps severities to annotation levels (error/warning/notice)
so findings appear directly on the changed lines in the PR.

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
