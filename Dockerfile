# Multi-stage Dockerfile for Perplexity MCP Server (dynamic linking, wolfi-base)

# Build stage
FROM golang:1.25-alpine AS builder

# Install build dependencies
# Install build dependencies (cgo enabled build for dynamic linking)
RUN apk add --no-cache git ca-certificates tzdata build-base

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download && go mod verify

# Copy source code
COPY . .

# Build the application with optimizations
# Build the application with optimizations (dynamic linking, cgo enabled)
# Note: Do not force GOARCH to avoid cgo cross-compile toolchain issues.
RUN CGO_ENABLED=1 GOOS=linux go build \
  -trimpath \
  -ldflags='-w -s' \
  -o server \
  ./cmd/server

# Final stage - minimal runtime image
FROM cgr.dev/chainguard/wolfi-base

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata \
  && addgroup -g 1001 -S appgroup \
  && adduser -u 1001 -S appuser -G appgroup

# Set working directory
WORKDIR /app

# Copy binary from builder stage with proper ownership
COPY --from=builder --chown=1001:1001 /app/server /app/server

# Switch to non-root user
USER appuser

# Set default environment variables
ENV LOG_LEVEL=info \
  REQUEST_TIMEOUT_SECONDS=30 \
  TZ=UTC

# Run the server
ENTRYPOINT ["./server"]
