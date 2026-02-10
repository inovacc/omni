# Multi-stage golden master testing for omni CLI
# Stage 1: Build omni binary
# Stage 2: Python runner with golden testing framework

# Stage 1: Go builder
FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git gcc musl-dev

WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o /omni .

# Stage 2: Python + golden test runner
FROM python:3.13-slim AS runner

RUN apt-get update && apt-get install -y --no-install-recommends git && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Copy omni binary from builder
COPY --from=builder /omni /usr/local/bin/omni

# Copy golden testing framework
COPY tools/golden/src/ /app/src/
COPY tools/golden/golden_tests.yaml /app/golden_tests.yaml
COPY tools/golden/pyproject.toml /app/pyproject.toml

# Install Python dependencies
RUN pip install --no-cache-dir pyyaml

ENV PYTHONPATH=/app/src
ENV OMNI_BIN=/usr/local/bin/omni

ENTRYPOINT ["python3", "-m", "golden"]
