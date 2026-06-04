# docker-image-optimiser (`dio`)

A CLI that makes Docker images **smaller** and builds **faster** by detecting
Dockerfile anti-patterns — and, eventually, fixing them and proving the win.

## Status

| Command | What it does | State |
|---------|--------------|-------|
| `dio lint`    | Parse a Dockerfile and report size / build-speed issues | ✅ working |
| `dio fix`     | Rewrite safe issues in place, annotate the rest         | ✅ working |
| `dio bench`   | Build before/after and compare size + build time        | ✅ working |
| `dio inspect` | Show per-layer sizes for an image, largest first         | ✅ working |

## Quick start

```bash
go run ./cmd/dio lint testdata/Dockerfile.bad   # report issues
go run ./cmd/dio fix  testdata/Dockerfile.bad    # print a fixed Dockerfile
go run ./cmd/dio fix -w Dockerfile               # rewrite in place
go run ./cmd/dio bench Dockerfile                # build both, compare size + time (needs Docker)
go run ./cmd/dio inspect myimage:latest          # per-layer size breakdown (needs Docker)
# or build a binary:
go build -o dio ./cmd/dio && ./dio lint Dockerfile
```

`lint` exits non-zero when it finds issues, so it can gate a CI pipeline.
`fix` applies the safe, deterministic fixes (DIO002/003/004) directly and leaves
a `# dio[...]` comment above issues that need a human decision (DIO001 reorder,
DIO005 version pin). Re-running `fix` is idempotent.

`bench` builds the original and the rewritten Dockerfile (with `--no-cache` by
default for a fair comparison) and prints a size/time table. It needs a running
Docker daemon and cleans up its images afterwards (`--keep` to retain them).

`inspect <image>` lists the image's layers largest-first with each layer's share
of the total, so you can see where the bytes went (`-n N` to show only the top N).

## Lint rules

| ID | Checks for |
|----|------------|
| DIO001 | `COPY . .` before a dependency install (busts the layer cache) |
| DIO002 | `apt-get install` without `--no-install-recommends` |
| DIO003 | `apt-get install` that leaves `/var/lib/apt/lists` in the image |
| DIO004 | `pip install` without `--no-cache-dir` |
| DIO005 | Base image pinned to `:latest` or with no tag |

## Project layout

```
cmd/dio/            CLI entrypoint
internal/
  cli/              cobra command tree (lint, fix, bench, inspect)
  parser/           Dockerfile -> []Instruction
  rules/            one file per lint rule; register in registry.go
  rewrite/          applies fixes: rewrites safe issues, annotates the rest
  build/            shells out to Docker to build + measure images (bench)
  inspect/          per-layer size breakdown via `docker history`
  report/           terminal output
testdata/           sample Dockerfiles
```

## Adding a lint rule

1. Add a file in `internal/rules/` implementing the `Rule` interface.
2. Append it to `All()` in `internal/rules/registry.go`.
3. Add a case to `internal/rules/rules_test.go`.

## Test

```bash
go test ./...
```
