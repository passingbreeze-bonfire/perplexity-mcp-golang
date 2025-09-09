# Multi-stage Dockerfile for Perplexity MCP Server (dynamic linking, wolfi-base)

# Build stage
FROM golang:1.25.1-trixie AS builder

# Install build dependencies  
# Install build dependencies (cgo enabled build for dynamic linking)
RUN apt-get update && apt-get install -y --no-install-recommends \
    git \
    ca-certificates \
    tzdata \
    build-essential \
    && rm -rf /var/lib/apt/lists/*

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
  -o perplexity-mcp-server \
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
COPY --from=builder --chown=1001:1001 /app/perplexity-mcp-server /app/perplexity-mcp-server

# Switch to non-root user
USER appuser

# Set default environment variables
ENV LOG_LEVEL=info \
  REQUEST_TIMEOUT=300 \
  TZ=UTC

# Run the server
ENTRYPOINT ["./perplexity-mcp-server"]
