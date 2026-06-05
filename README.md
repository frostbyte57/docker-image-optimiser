# docker-image-optimiser (`dio`)

A CLI that makes **any** Docker image **smaller** and its builds **faster** by
detecting Dockerfile anti-patterns, fixing them, and proving the win. It is
language-agnostic: Node, Python, Go, Rust, Java (Maven/Gradle), Ruby, PHP, .NET,
and the apt/apk/dnf system package managers are all understood.

## Commands at a glance

| Command | What it does | Needs Docker |
|---------|--------------|--------------|
| `dio lint`    | Parse a Dockerfile and report size / build-speed issues | no  |
| `dio fix`     | Rewrite safe issues, annotate the rest                  | no  |
| `dio bench`   | Build before/after and compare size + (warm) build time | yes |
| `dio inspect` | Show per-layer sizes for an image, largest first        | yes |

## Requirements

- **Go 1.24+** to build or run the CLI.
- **Docker** (with BuildKit, the default on modern Docker) only for `bench` and
  `inspect`, which shell out to the `docker` binary. `lint` and `fix` are pure
  Go and need nothing else.

## Installation

Pick whichever fits your workflow.

```bash
# 1. Install the binary onto your PATH (uses $GOBIN, or $GOPATH/bin):
go install github.com/yuxiangchang/docker-image-optimiser/cmd/dio@latest
dio lint Dockerfile

# 2. Build a local binary from a clone of this repo:
git clone https://github.com/yuxiangchang/docker-image-optimiser
cd docker-image-optimiser
go build -o dio ./cmd/dio
./dio lint Dockerfile

# 3. Run straight from source without building (handy while hacking on dio):
go run ./cmd/dio lint Dockerfile
```

The examples below use `dio`; if you went with option 3, replace it with
`go run ./cmd/dio`.

## Usage

Every command takes an optional Dockerfile path and defaults to `./Dockerfile`:

```bash
dio lint                       # lints ./Dockerfile
dio lint path/to/Dockerfile    # lints a specific file
```

### `dio lint [Dockerfile]`

Parses the Dockerfile, runs every rule, and prints what it found. It **exits
non-zero when there are findings**, so you can drop it straight into a CI step to
fail a build on regressions.

```bash
dio lint testdata/go/Dockerfile     # report issues in a sample file
dio lint                            # lint ./Dockerfile
dio lint -c ./app Dockerfile        # also run the .dockerignore check (DIO009)
```

| Flag | Default | Meaning |
|------|---------|---------|
| `-c`, `--context <dir>` | `.` | Build context directory. Enables the `.dockerignore` check (DIO009), which needs the filesystem. |

### `dio fix [Dockerfile]`

Applies the safe fixes and leaves a `# dio[...]` comment above the issues that
need a human decision (see the rules table below). By **default it prints the
result to stdout** and writes its change log to stderr, so stdout stays a clean
Dockerfile you can pipe or redirect:

```bash
dio fix Dockerfile                  # print the fixed Dockerfile to stdout
dio fix Dockerfile > Dockerfile.new # redirect the clean output, ignore the log
dio fix -w Dockerfile               # rewrite the file in place
dio fix --conservative -w Dockerfile # use --no-cache-dir cleanup, no BuildKit
```

| Flag | Default | Meaning |
|------|---------|---------|
| `-w`, `--write` | off | Write changes back to the file instead of printing to stdout. |
| `--conservative` | off | Use `--no-cache-dir`-style cleanup instead of BuildKit cache mounts, for environments without BuildKit. |

Re-running `fix` is idempotent: a second pass over an already-fixed file changes
nothing and re-stacks no annotations.

### `dio bench [Dockerfile]`

Builds the original Dockerfile and the `dio fix` rewrite, then prints a size and
build-time comparison. Requires a reachable Docker daemon. By default both builds
use `--no-cache` for a fair comparison, and the two temporary images are removed
afterwards.

```bash
dio bench Dockerfile                # compare size + cold build time
dio bench --incremental Dockerfile  # also measure the warm rebuild (builds twice)
dio bench -c ./app --keep Dockerfile # custom context, keep the built images
```

| Flag | Default | Meaning |
|------|---------|---------|
| `-c`, `--context <dir>` | `.` | Build context directory. |
| `--incremental` | off | Also measure the warm rebuild after a simulated source change (see below). |
| `--cache` | off | Allow the build cache. By default `bench` builds with `--no-cache`. |
| `--keep` | off | Keep the `dio-bench-before`/`dio-bench-after` images instead of removing them. |

### `dio inspect <image>`

Shows where an existing image's bytes went, layer by layer, largest first, by
reading `docker history`. Requires Docker and an image that is already built or
pulled. The image reference is **required**.

```bash
dio inspect myimage:latest          # full per-layer breakdown
dio inspect -n 10 myimage:latest    # only the 10 largest layers, rest summarised
```

| Flag | Default | Meaning |
|------|---------|---------|
| `-n`, `--top <N>` | `0` | Show only the N largest layers (0 = all). The remaining layers are collapsed into one summary line. |

## The three kinds of "caching" (why this tool exists)

Most Dockerfile advice conflates three different things. `dio` treats them
separately:

1. **Layer cache** — order so dependency installs come *before* `COPY . .`, so a
   code edit doesn't reinstall dependencies. Speeds up *rebuilds*. (DIO001)
2. **BuildKit cache mounts** (`RUN --mount=type=cache`) — keep the package
   manager's download cache *out of the image* while reusing it *across* builds.
   This gives **both** a small image and fast rebuilds, and is strictly better
   than `--no-cache-dir`, which only shrinks the image and forces a re-download
   every build. This is `dio`'s default fix for language installs. (DIO004)
3. **`--no-cache` at build time** — only used by `bench` for a fair comparison.

### Caching policy

| Manager type | Examples | Default fix |
|---|---|---|
| **Language** | pip, poetry, uv, npm, yarn, pnpm, go, cargo, maven, gradle, composer, dotnet | **cache mount** + auto-added `# syntax=docker/dockerfile:1` |
| **System** | apt, apk, dnf/yum | **conservative cleanup** (`--no-install-recommends`, `rm` caches, `apk --no-cache`) |

System managers stay conservative because apt cache mounts need extra, error-prone
setup. Use `dio fix --conservative` to force the `--no-cache-dir`/cleanup form
everywhere (for environments without BuildKit). Re-running `fix` is idempotent.

## Lint rules

| ID | Checks for | Fix |
|----|------------|-----|
| DIO001 | broad `COPY . .` before a language dependency install | annotate (reorder) |
| DIO002 | `apt-get install` / `apt install` without `--no-install-recommends` | auto |
| DIO003 | system install (apt/apk/dnf) leaving its cache in the image | auto |
| DIO004 | language install with no cache mount (or `--no-cache-dir` in conservative mode) | auto |
| DIO005 | base image pinned to `:latest` / no tag | annotate |
| DIO006 | fat base image with a smaller official variant | annotate |
| DIO007 | single-stage build on a build-toolchain base (go/rust/maven/gradle) | annotate (multi-stage) |
| DIO008 | hand-written cache mounts missing the `# syntax` directive | auto (prepend) |
| DIO009 | build context with no `.dockerignore` (`lint --context`) | inform |
| DIO010 | final stage runs as root (no `USER`) | annotate |

`fix` applies the **auto** fixes in place and leaves a `# dio[...]` comment above
the **annotate**/**inform** issues that need a human decision.

## `bench --incremental`

Cold builds measure **size**; the `--incremental` flag also measures the
**warm rebuild** — build once, change a source file, rebuild — which is where
cache mounts and layer ordering pay off:

```
                 before        after         change
  size         900 MB       120 MB       -780 MB (-86.7%)
  cold            58s          55s        -3s
  warm            55s          12s        -43s          <- the rebuild win
```

`bench` builds the output of `dio fix`, which adds cache mounts but only
**annotates** the layer reorder (DIO001 is a structural change left to you). So
the warm number above is what cache mounts alone buy: the dependency step still
re-runs on a source change, but its downloads come from the mount instead of the
network. Apply the DIO001 reorder by hand — copy the manifest, install, then
`COPY . .` — and the install layer is skipped entirely, taking the warm rebuild
down to a few seconds.

## Project layout

```
cmd/dio/            CLI entrypoint
internal/
  cli/              cobra command tree (lint, fix, bench, inspect)
  parser/           Dockerfile -> []Instruction
  ecosystem/        data-driven package-manager registry (the rule backbone)
  rules/            one file per lint rule; register in registry.go
  rewrite/          applies fixes: rewrites safe issues, annotates the rest
  build/            shells out to Docker to build + measure images (bench)
  inspect/          per-layer size breakdown via `docker history`
  report/           terminal output
testdata/<lang>/    sample Dockerfiles per language (go, node, java, ...)
```

## Adding a language

Most support is data, not code: add one entry to `registry` in
`internal/ecosystem/ecosystem.go` (detection substrings, manifest files, cache
mount dirs, conservative flag/cleanup). The generic rules pick it up automatically.

## Adding a lint rule

1. Add a file in `internal/rules/` implementing the `Rule` interface (a `Finding`
   with a non-nil `Rewrite` is auto-fixed; nil is annotate-only).
2. Append it to `All()` in `internal/rules/registry.go`.
3. Add a case to `internal/rules/rules_test.go`.

## Test

```bash
go test ./...
```
