# Deployment & Releases

Releases are built by [GoReleaser](https://goreleaser.com/) from `v*` tags. The
release workflow ([`.github/workflows/release.yml`](../.github/workflows/release.yml))
builds cross-platform binaries, generates checksums and SBOMs, and signs
checksums with keyless [cosign](https://docs.sigstore.dev/cosign/overview/).

## Cutting a release

```bash
git tag v0.1.0
git push origin v0.1.0
```

Pushing a `v*` tag triggers the release workflow. Artifacts are attached to the
GitHub release.

## Consuming a release

Install a specific version with `go install`:

```bash
go install github.com/yuxiangchang/docker-image-optimiser/cmd/dio@v0.1.0
```

In CI, the bundled GitHub Action's `version` input pins the same ref — see
[`examples/ci/`](../examples/ci/).
