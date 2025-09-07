# MCP Server

This is the main server entrypoint for the Perplexity MCP (Model Context Protocol) server implementation.

## Environment Variables

The following environment variables must be configured:

| Variable | Description | Default | Required |
|----------|-------------|---------|-----------|
| `PERPLEXITY_API_KEY` | Your Perplexity API key | - | Yes |
| `PERPLEXITY_DEFAULT_MODEL` | Default model to use | `llama-3.1-sonar-small-128k-online` | No |
| `REQUEST_TIMEOUT_SECONDS` | Request timeout in seconds | `30` | No |
| `LOG_LEVEL` | Log level (debug, info, warn, error) | `info` | No |

## Building

```bash
go build -o server cmd/server/main.go
```

## Running

```bash
export PERPLEXITY_API_KEY="your-api-key-here"
./server
```

Or run directly with Go:

```bash
PERPLEXITY_API_KEY="your-api-key-here" go run cmd/server/main.go
```

## Available Tools

The server exposes three MCP tools:

1. **perplexity_search** - Search for information using Perplexity AI
2. **perplexity_chat** - Chat with Perplexity AI using conversational messages  
3. **perplexity_research** - Perform comprehensive research on a topic

## Graceful Shutdown

The server handles graceful shutdown on SIGINT (Ctrl+C) or SIGTERM signals with a configurable timeout.

## Testing

Run the server tests:

```bash
go test ./cmd/server -v
```

## Architecture

The server follows clean architecture principles:

- **Domain Layer**: Core business logic and interfaces
- **Use Case Layer**: Application-specific business rules
- **Adapter Layer**: External interfaces (MCP server, Perplexity API client)
- **Infrastructure Layer**: Cross-cutting concerns (config, logging)

Dependencies are wired using dependency injection pattern for maintainable and testable code.