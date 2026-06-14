# dio[DIO010]: Create and switch to a non-root user, e.g. USER nonroot (or a distroless :nonroot base)
FROM debian:12-slim
RUN apt update && apt install --no-install-recommends -y curl && rm -rf /var/lib/apt/lists/*
