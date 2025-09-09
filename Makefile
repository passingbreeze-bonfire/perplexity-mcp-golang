# Perplexity MCP Server - Makefile

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
BINARY_NAME=pplxity_mcp_server
BINARY_UNIX=$(BINARY_NAME)
MAIN_PATH=./cmd/server

# Build flags
LDFLAGS=-ldflags "-w -s"
BUILD_FLAGS=-v

.PHONY: all build clean test test-coverage test-integration test-benchmark deps fmt lint security help

# Default target
all: test build

# Build the application
build:
	@echo "Building $(BINARY_NAME)..."
	$(GOBUILD) $(BUILD_FLAGS) $(LDFLAGS) -o $(BINARY_NAME) $(MAIN_PATH)

# Build for Linux
build-linux:
	@echo "Building $(BINARY_NAME) for Linux..."
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) $(BUILD_FLAGS) $(LDFLAGS) -o $(BINARY_UNIX) $(MAIN_PATH)

# Clean build artifacts
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_UNIX)

# Run all tests
test:
	@echo "Running tests..."
	$(GOTEST) -v ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run integration tests
test-integration:
	@echo "Running integration tests..."
	$(GOTEST) -v ./test/integration/

# Run quick integration tests
test-quick:
	@echo "Running quick integration tests..."
	$(GOTEST) -v ./test/integration/ -run TestQuickIntegrationSuite

# Run benchmarks
test-benchmark:
	@echo "Running benchmarks..."
	$(GOTEST) -v ./test/benchmark/ -bench=. -benchmem

# Install dependencies
deps:
	@echo "Installing dependencies..."
	$(GOMOD) download
	$(GOMOD) verify

# Update dependencies
deps-update:
	@echo "Updating dependencies..."
	$(GOMOD) tidy
	$(GOGET) -u ./...

# Format code
fmt:
	@echo "Formatting code..."
	goimports -w .
	gofmt -s -w .

# Lint code
lint:
	@echo "Linting code..."
	golangci-lint run

# Run security scan
security:
	@echo "Running security scan..."
	gosec ./...

# Development server
dev:
	@echo "Starting development server..."
	$(GOCMD) run $(MAIN_PATH)

# Install development tools
install-tools:
	@echo "Installing development tools..."
	$(GOGET) golang.org/x/tools/cmd/goimports@latest
	$(GOGET) github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	$(GOGET) github.com/securecodewarrior/gosec/v2/cmd/gosec@latest

# Docker build
docker-build:
	@echo "Building Docker image..."
	docker build -t perplexity-mcp-server .

# Docker run
docker-run:
	@echo "Running Docker container..."
	docker run -p 8080:8080 --env-file .env perplexity-mcp-server

# Validate project setup
validate:
	@echo "Validating project setup..."
	@if [ ! -f ".env" ]; then echo "⚠️  .env file not found. Copy .env.example to .env and configure it."; fi
	@$(GOCMD) version
	@echo "✅ Go environment ready"

# Generate documentation
docs:
	@echo "Generating documentation..."
	godoc -http=:6060 &
	@echo "Documentation server started at http://localhost:6060"

# Release build (optimized)
release:
	@echo "Building release version..."
	$(GOBUILD) -a -installsuffix cgo $(LDFLAGS) -o $(BINARY_NAME) $(MAIN_PATH)
	upx --best --lzma $(BINARY_NAME) || echo "UPX not available, skipping compression"

# Help
help:
	@echo "Available commands:"
	@echo "  build          Build the application"
	@echo "  build-linux    Build for Linux"
	@echo "  clean          Clean build artifacts"
	@echo "  test           Run all tests"
	@echo "  test-coverage  Run tests with coverage report"
	@echo "  test-integration Run integration tests"
	@echo "  test-quick     Run quick integration tests"
	@echo "  test-benchmark Run performance benchmarks"
	@echo "  deps           Install dependencies"
	@echo "  deps-update    Update dependencies"
	@echo "  fmt            Format code"
	@echo "  lint           Lint code"
	@echo "  security       Run security scan"
	@echo "  dev            Start development server"
	@echo "  install-tools  Install development tools"
	@echo "  docker-build   Build Docker image"
	@echo "  docker-run     Run Docker container"
	@echo "  validate       Validate project setup"
	@echo "  docs           Generate documentation"
	@echo "  release        Build optimized release"
	@echo "  help           Show this help"

# Default environment check
check-env:
	@if [ -z "$(PERPLEXITY_API_KEY)" ]; then \
		echo "⚠️  PERPLEXITY_API_KEY environment variable not set"; \
		echo "   Please set it or create a .env file"; \
		exit 1; \
	fi