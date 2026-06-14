# Use `dio` in CI

`dio` is built to be a **pipeline gate**: run the check first, and only build the
image once the Dockerfile is optimised. `dio optimize --check` exits non-zero when
any optimisation or manual fix is still pending, so it fails the job before the
(slow, expensive) `docker build` runs.

```bash
dio optimize --check Dockerfile          # exit 1 if anything is pending
dio optimize --check --format github ... # same, with inline PR annotations
dio optimize --check --format json  ...  # machine-readable summary
```

## Output formats

| Format   | Use for                                                        |
| -------- | -------------------------------------------------------------- |
| `text`   | Human-readable terminal output (default).                      |
| `json`   | Machine-readable summary for custom tooling.                   |
| `github` | GitHub Actions workflow commands → inline PR annotations.      |

The `github` format maps severities to annotation levels (error/warning/notice)
so findings appear directly on the changed lines in the PR.

## GitHub Actions

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

### Action inputs

| Input            | Default      | Description                                                  |
| ---------------- | ------------ | ----------------------------------------------------------- |
| `dockerfile`     | `Dockerfile` | Path to the Dockerfile to check.                            |
| `context`        | `.`          | Build context dir (enables the `.dockerignore` check).      |
| `conservative`   | `false`      | Use `--no-cache-dir` cleanup instead of BuildKit mounts.    |
| `version`        | `latest`     | `dio` version to install (a Git ref, e.g. `v0.1.0`).        |
| `fail-on-issues` | `true`       | Fail the job when optimisations/manual fixes are pending.   |

The action is defined in [`action.yml`](../../action.yml).

## Other CI systems

`dio` is a single static binary, so any runner can use it. Install with
`go install`, then run the gate before the build stage.

## Example workflows

- [`github-actions.yml`](github-actions.yml) — gate-then-build on GitHub Actions.
- [`gitlab-ci.yml`](gitlab-ci.yml) — the same pattern on GitLab CI.
