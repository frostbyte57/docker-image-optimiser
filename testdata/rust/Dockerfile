# Single-stage Rust build.
FROM rust:1.77
WORKDIR /app
COPY . .
RUN cargo build --release
CMD ["./target/release/app"]
