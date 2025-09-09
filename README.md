# Perplexity Search MCP Server (Go)

A Model Context Protocol (MCP) server that provides access to Perplexity AI's Sonar search models through a clean, secure Go implementation.

## Features

🔍 **Sonar Search Models**:
- `sonar` - Fast, efficient search
- `sonar-pro` - Enhanced search capabilities
- `sonar-reasoning` - Advanced reasoning with search
- `sonar-reasoning-pro` - Professional reasoning capabilities
- `sonar-deep-research` - Comprehensive deep research

🏗️ **Clean Architecture**:
- Simple, maintainable structure with clear separation
- Single-thread-first approach with context-based timeouts
- Comprehensive error handling and validation

🔒 **Security First**:
- TLS 1.2+ enforcement
- Input validation and sanitization
- Secure logging without sensitive data exposure
- Environment-based secret management

📊 **Performance & Testing**:
- Comprehensive test coverage
- Integration tests for stdio transport
- Performance benchmarks for optimization

## Quick Start

### Prerequisites

- Go 1.25.1 or later
- [Perplexity API key](https://docs.perplexity.ai/docs/getting-started)

### Installation

```bash
# Clone repository
git clone https://github.com/yourusername/perplexity-mcp-golang
cd perplexity-mcp-golang

# Set up environment
cp .env.example .env
# Edit .env to add your PERPLEXITY_API_KEY

# Build server
make build

# Run server
./perplexity-mcp-server
```

### Usage with MCP Clients

The server exposes a search tool through the MCP protocol:

#### Search Tool
```json
{
  "name": "perplexity_search",
  "arguments": {
    "query": "What is quantum computing?",
    "model": "sonar",
    "search_mode": "web",
    "max_tokens": 1000,
    "sources": ["arxiv.org", "nature.com"]
  }
}
```

**Parameters:**
- `query` (required): The search query
- `model` (optional): Sonar model to use (sonar, sonar-pro, sonar-reasoning, sonar-reasoning-pro, sonar-deep-research)
- `search_mode` (optional): Search mode (web, academic, news)
- `max_tokens` (optional): Maximum response tokens
- `sources` (optional): List of domains to search within
- `options` (optional): Additional options like temperature, top_p

## Configuration

Configure the server using environment variables:

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `PERPLEXITY_API_KEY` | ✅ | - | Your Perplexity API key |
| `PERPLEXITY_DEFAULT_MODEL` | ❌ | `sonar` | Default Sonar model |
| `REQUEST_TIMEOUT_SECONDS` | ❌ | `30` | Request timeout in seconds |
| `LOG_LEVEL` | ❌ | `info` | Log level (debug, info, warn, error) |

## Architecture

Simple, maintainable structure focused on clarity and reliability:

```
┌─────────────────────────────────────────────────────────┐
│                    cmd/server                           │
│                 (Entry Point)                          │
│             • main.go                                   │
│             • integration_test.go                      │
└─────────────────────┬───────────────────────────────────┘
                      │
┌─────────────────────▼───────────────────────────────────┐
│                   internal/                            │
│  ┌─────────────────┐  ┌─────────────────────────────────┤
│  │   client.go     │  │        tools.go                 │
│  │  (HTTP Client)  │  │    (MCP Tool Handler)           │
│  └─────────────────┘  └─────────────────────────────────┤
│  ┌─────────────────┐  ┌─────────────────────────────────┤
│  │   config.go     │  │        types.go                 │
│  │ (Configuration) │  │   (Data Structures)             │
│  └─────────────────┘  └─────────────────────────────────┤
└─────────────────────────────────────────────────────────┘
```

## Sonar Models

### sonar
Fast, efficient search for quick answers and basic queries.

### sonar-pro
Enhanced search with better understanding and more comprehensive results.

### sonar-reasoning
Advanced model that combines search with step-by-step reasoning capabilities.

### sonar-reasoning-pro
Professional-grade reasoning model for complex queries requiring logical analysis.

### sonar-deep-research
Comprehensive research model that performs thorough, multi-step research with extensive citations.

## Development

### Running Tests

```bash
# Run all tests
make test

# Run with coverage
make test-coverage

# Run integration tests
make test-integration

# Run benchmarks
make test-benchmark
```

### Building with Docker

```bash
# Build Docker image
make docker-build

# Run container
make docker-run
```

### Project Structure

```
.
├── cmd/server/          # Application entry point and integration tests
│   ├── main.go         # Server main function
│   └── integration_test.go # Integration tests
├── internal/           # Internal packages
│   ├── client.go       # Perplexity API client
│   ├── config.go       # Configuration management
│   ├── tools.go        # MCP tool implementations
│   └── types.go        # Data types and structures
├── build/              # Build artifacts directory
├── Dockerfile          # Multi-stage Docker build
├── Makefile           # Build automation
├── .mise.toml         # Development environment setup
├── .env.example       # Environment variable template
├── go.mod             # Go module definition
└── go.sum             # Go module checksums
```

## Development Tools

This project uses `mise` for development environment management:

```bash
# Install mise if not already installed
# See: https://mise.jdx.dev/getting-started.html

# Install project dependencies
mise install

# Run development tasks
mise run fmt     # Format code
mise run lint    # Lint code
mise run build   # Build project
mise run test    # Run tests
mise run dev     # Start development server
```

## Security Considerations

- **API Key Protection**: Never commit API keys. Use `.env` file for local development.
- **Input Validation**: All inputs are validated and sanitized.
- **TLS Enforcement**: TLS 1.2+ required for API communications.
- **Resource Bounds**: Maximum query lengths and response sizes enforced.
- **Secure Logging**: Sensitive information is never logged.

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Acknowledgments

- Built with the [Model Context Protocol](https://modelcontextprotocol.io/)
- Powered by [Perplexity AI](https://www.perplexity.ai/)
- Follows Go best practices and clean architecture principles