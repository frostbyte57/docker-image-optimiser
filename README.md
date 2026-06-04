# docker-image-optimiser (`dio`)

A CLI that makes Docker images **smaller** and builds **faster** by detecting
Dockerfile anti-patterns — and, eventually, fixing them and proving the win.

## Status

| Command | What it does | State |
|---------|--------------|-------|
| `dio lint`    | Parse a Dockerfile and report size / build-speed issues | ✅ working |
| `dio fix`     | Rewrite the Dockerfile applying the fixes               | 🚧 planned |
| `dio bench`   | Build before/after and compare size + build time        | 🚧 planned |
| `dio inspect` | Show per-layer sizes and wasted space for an image      | 🚧 planned |

## Quick start

```bash
go run ./cmd/dio lint testdata/Dockerfile.bad
# or build a binary:
go build -o dio ./cmd/dio && ./dio lint Dockerfile
```

`lint` exits non-zero when it finds issues, so it can gate a CI pipeline.

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
  cli/              cobra command tree (lint + stubs for fix/bench/inspect)
  parser/           Dockerfile -> []Instruction
  rules/            one file per lint rule; register in registry.go
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
