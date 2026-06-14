# syntax=docker/dockerfile:1
# dio[DIO006]: Consider python:<ver>-slim
# dio[DIO010]: Create and switch to a non-root user, e.g. USER nonroot (or a distroless :nonroot base)
FROM python:3.12
# dio[DIO001]: Copy requirements.txt first, run the install, then COPY the rest
COPY . .
RUN --mount=type=cache,target=/root/.cache/pip pip install -r requirements.txt
