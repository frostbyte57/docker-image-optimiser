# dio Help

Use `dio --help` for the current command list.

```bash
dio --help
dio <command> --help
```

## Commands

### `dio lint [Dockerfile]`

Reports Dockerfile size and build-speed issues.

Flags:

```bash
-c, --context string   build context dir (default ".")
```

### `dio fix [Dockerfile]`

Prints safe Dockerfile rewrites, or writes them back to the file.

Flags:

```bash
-w, --write          write changes back to the file in place
    --conservative   use no-cache cleanup instead of BuildKit cache mounts
```

### `dio optimize [Dockerfile]`

Runs linting and safe rewriting in one command.

Alias:

```bash
dio optimise
```

Flags:

```bash
    --check            exit non-zero when optimisations or manual fixes are pending
    --conservative     use no-cache cleanup instead of BuildKit cache mounts
-c, --context string   build context dir (default ".")
    --format string    output format: text or json (default "text")
-w, --write            write optimised Dockerfile back in place
```

### `dio bench [Dockerfile]`

Builds the original Dockerfile and its optimised rewrite, then compares them.
Requires Docker.

Flags:

```bash
    --cache            allow the build cache
-c, --context string   build context directory (default ".")
    --incremental      also measure warm rebuild time after a source change
    --keep             keep the built images instead of removing them
```

### `dio inspect <image>`

Shows per-layer image sizes, largest first. Requires Docker.

Flags:

```bash
-n, --top int   show only the N largest layers (0 = all)
```

## Built-In Commands

Cobra also provides:

```bash
dio help
dio completion
```
