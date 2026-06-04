# Single-stage Go build: ships the whole toolchain, no cache mounts.
FROM golang:1.22
WORKDIR /app
COPY . .
RUN go build -o /app/server ./cmd/server
CMD ["/app/server"]
