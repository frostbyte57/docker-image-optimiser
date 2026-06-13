package cli

const rootLong = `docker-image-optimiser (dio) is a CLI for making Dockerfiles better
for CI/CD builds. It lints size and build-speed anti-patterns, applies safe
rewrites, reports manual fixes, and can benchmark the result.

Available workflows:
  dio lint       Report Dockerfile optimisation issues
  dio fix        Print or write safe Dockerfile rewrites
  dio optimize   CI-friendly optimise/check command
  dio bench      Compare original vs optimised image builds
  dio inspect    Show image layer sizes`

const rootExamples = `  dio --help
  dio lint Dockerfile
  dio fix --write Dockerfile
  dio optimize --check --format json Dockerfile
  dio bench --incremental Dockerfile
  dio inspect myimage:latest --top 10`
