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
- Domain-driven design with clear layer separation
- Dependency injection for testability
- Single-thread-first approach with context-based timeouts

🔒 **Security First**:
- TLS 1.2+ enforcement
- Input validation and sanitization
- Secure logging without sensitive data exposure
- Rate limiting and resource bounds

📊 **Performance & Testing**:
- Comprehensive test coverage
- Performance benchmarks for optimization
- Integration tests with mock dependencies

## Quick Start

### Prerequisites

- Go 1.22 or later
- [Perplexity API key](https://docs.perplexity.ai/docs/getting-started)

### Installation

```bash
# Clone repository
git clone https://github.com/yourusername/perplexity-mcp-golang
cd perplexity-mcp-golang

# Build server
go build -o server cmd/server/main.go

# Set environment variables
export PERPLEXITY_API_KEY="your-api-key-here"
export LOG_LEVEL="info"

# Run server
./server
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
    "date_range": "week",
    "sources": ["arxiv.org", "nature.com"],
    "max_tokens": 1000
  }
}
```

**Parameters:**
- `query` (required): The search query
- `model` (optional): Sonar model to use (sonar, sonar-pro, sonar-reasoning, sonar-reasoning-pro, sonar-deep-research)
- `search_mode` (optional): Search mode (web, academic, news)
- `date_range` (optional): Time filter (day, week, month, year)
- `sources` (optional): List of domains to search within
- `max_tokens` (optional): Maximum response tokens
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

```
┌─────────────────────────────────────────────────────────┐
│                    cmd/server                           │
│                  (Entry Point)                         │
└─────────────────────┬───────────────────────────────────┘
                      │
┌─────────────────────▼───────────────────────────────────┐
│                MCP Adapters                             │
│  ┌─────────────────┐  ┌─────────────────────────────────┤
│  │   MCP Server    │  │        MCP Tools                │
│  │   Handlers      │  │  • search.go                   │
│  └─────────────────┘  └─────────────────────────────────┤
└─────────────────────┬───────────────────────────────────┘
                      │
┌─────────────────────▼───────────────────────────────────┐
│                  Use Cases                              │
│  ┌──────────────────────────────────────────────────────┤
│  │            SearchUseCase                             │
│  └──────────────────────────────────────────────────────┤
└─────────────────────┬───────────────────────────────────┘
                      │
┌─────────────────────▼───────────────────────────────────┐
│                   Domain                                │
│  ┌─────────────────┐  ┌─────────────────────────────────┤
│  │   Interfaces    │  │         Entities                │
│  │   • Client      │  │  • SearchRequest                │
│  │   • Logger      │  │  • SearchResult                 │
│  │   • Config      │  │  • Usage, Citation, Source      │
│  └─────────────────┘  └─────────────────────────────────┤
└─────────────────────┬───────────────────────────────────┘
                      │
┌─────────────────────▼───────────────────────────────────┐
│                Infrastructure                           │
│  ┌─────────────────┐  ┌─────────────────┐  ┌───────────┤
│  │ PerplexityClient│  │     Config      │  │  Logger   │
│  │   (HTTP)        │  │   (Env Vars)    │  │ (slog)    │
│  └─────────────────┘  └─────────────────┘  └───────────┤
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

## Search Modes

- **web**: General web search across all sources
- **academic**: Focus on academic papers and scholarly sources
- **news**: Recent news articles and current events

## Date Ranges

Filter results by recency:
- **day**: Last 24 hours
- **week**: Last 7 days
- **month**: Last 30 days
- **year**: Last 12 months

## Development

### Running Tests

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific test package
go test ./internal/core/usecases

# Run benchmarks
go test -bench=. ./test/benchmark
```

### Building with Docker

```bash
# Build Docker image
docker build -t perplexity-mcp-server .

# Run container
docker run -e PERPLEXITY_API_KEY="your-key" perplexity-mcp-server
```

### Project Structure

```
.
├── cmd/server/          # Application entry point
├── internal/
│   ├── adapters/        # External adapters (MCP, Perplexity)
│   ├── core/           # Business logic
│   │   ├── domain/     # Domain entities and interfaces
│   │   └── usecases/   # Use case implementations
│   └── infrastructure/ # Infrastructure implementations
├── test/               # Test files
│   ├── integration/    # Integration tests
│   └── benchmark/      # Performance benchmarks
└── docs/              # Documentation
```

## Security Considerations

- **API Key Protection**: Never commit API keys. Use environment variables.
- **Input Validation**: All inputs are validated and sanitized.
- **TLS Enforcement**: TLS 1.2+ required for API communications.
- **Rate Limiting**: Built-in rate limiting to prevent abuse.
- **Resource Bounds**: Maximum query lengths and response sizes enforced.

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