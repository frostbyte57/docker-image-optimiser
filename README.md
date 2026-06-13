# docker-image-optimiser (`dio`)

`dio` is a command-line tool that makes any Dockerfile build smaller images and
faster. It scans a Dockerfile for common size and build-speed anti-patterns,
rewrites the safe ones for you, annotates the rest, and can build before/after to
prove the win. It is language-agnostic: Node, Python, Go, Rust, Java, Ruby, PHP,
.NET, and apt/apk/dnf are all understood.

## Install

Requires Go 1.24+. `bench` and `inspect` also need Docker; `lint` and `fix` do
not.

```bash
# install onto your PATH
go install github.com/yuxiangchang/docker-image-optimiser/cmd/dio@latest

# or build from a clone
git clone https://github.com/yuxiangchang/docker-image-optimiser
cd docker-image-optimiser && go build -o dio ./cmd/dio
```

## Usage

Every command takes an optional Dockerfile path and defaults to `./Dockerfile`.

```bash
dio lint [Dockerfile]      # report size / build-speed issues
dio fix  [Dockerfile]      # rewrite safe issues, annotate the rest
dio optimize [Dockerfile]  # CI-friendly optimise/check workflow
dio bench [Dockerfile]     # build before/after and compare (needs Docker)
dio inspect <image>        # per-layer size breakdown of an image (needs Docker)
```

### lint

Reports issues and exits non-zero when it finds any, so it can gate CI.

```bash
dio lint Dockerfile
dio lint -c ./app Dockerfile   # -c/--context: also run the .dockerignore check
```

### fix

Prints a fixed Dockerfile to stdout (the change log goes to stderr), or rewrites
in place with `-w`. Running it twice changes nothing.

```bash
dio fix Dockerfile                     # print the result
dio fix -w Dockerfile                  # rewrite the file in place
dio fix --conservative -w Dockerfile   # --no-cache-dir cleanup instead of BuildKit cache mounts
```

### optimize

Runs the linter and safe rewriter in one CI-oriented command. Use `--check` to
fail a pipeline when a Dockerfile still has automatic optimisations or manual
actions pending, or `--write` to update it in place.

```bash
dio optimize Dockerfile
dio optimize --check Dockerfile
dio optimize --check --format json Dockerfile
dio optimise -w Dockerfile             # British spelling alias
```

### bench

Builds the original and the fixed Dockerfile and prints a size and build-time
comparison.

```bash
dio bench Dockerfile                # size + cold build time
dio bench --incremental Dockerfile  # also measure the warm rebuild after a source change
dio bench --keep Dockerfile         # keep the built images instead of removing them
```

### inspect

Shows where an existing image's bytes went, largest layer first.

```bash
dio inspect myimage:latest
dio inspect -n 10 myimage:latest    # -n/--top: only the 10 largest layers
```
