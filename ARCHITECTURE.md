# Architecture Documentation

## Overview

The Perplexity MCP Server follows **Clean Architecture** principles with clear separation of concerns, dependency inversion, and testability as core design goals.

## Architecture Principles

### 1. Single-Thread-First Policy
- All operations start with a correct, single-threaded baseline
- Context-based timeout management throughout the system
- Concurrency only after profiling proves bottlenecks
- Deterministic and predictable behavior

### 2. Clean Architecture Layers
```
External → Adapters → Use Cases → Domain ← Infrastructure
```

### 3. Dependency Direction
- Dependencies point inward toward the domain
- Outer layers depend on inner layers, never vice versa
- Interfaces define contracts in the domain layer

## Layer Details

### Domain Layer (`internal/core/domain/`)

**Purpose**: Pure business logic and entities

**Components**:
- **Interfaces** (`interfaces.go`): Contracts for external dependencies
- **Entities** (`entities.go`): Core business objects with validation
- **Errors** (`errors.go`): Domain-specific error types

**Key Interfaces**:
```go
type PerplexityClient interface {
    Search(ctx context.Context, request SearchRequest) (*SearchResult, error)
    Chat(ctx context.Context, request ChatRequest) (*ChatResult, error)  
    Research(ctx context.Context, request ResearchRequest) (*ResearchResult, error)
}

type MCPServer interface {
    RegisterTool(name string, tool Tool) error
    ExecuteTool(ctx context.Context, name string, args map[string]any) (*ToolResult, error)
    ListTools(ctx context.Context) ([]ToolInfo, error)
    Start(ctx context.Context) error
}

type Tool interface {
    Name() string
    Description() string
    InputSchema() map[string]any
    Execute(ctx context.Context, args map[string]any) (*ToolResult, error)
}
```

**Design Decisions**:
- All entities have `Validate()` methods for input sanitization
- Context is passed through all operations for timeout/cancellation
- Errors are typed and wrapped with context

### Use Cases Layer (`internal/core/usecases/`)

**Purpose**: Application-specific business rules and workflows

**Components**:
- **SearchUseCase** (`search.go`): Orchestrates search operations
- **ChatUseCase** (`chat.go`): Manages conversational workflows  
- **ResearchUseCase** (`research.go`): Handles research requests

**Key Responsibilities**:
```go
type SearchUseCase struct {
    client domain.PerplexityClient
    config domain.ConfigProvider
    logger domain.Logger
}

func (uc *SearchUseCase) Execute(ctx context.Context, request domain.SearchRequest) (*domain.SearchResult, error) {
    // 1. Validate request
    // 2. Apply defaults from config
    // 3. Call external client
    // 4. Log operation
    // 5. Return result
}
```

**Design Patterns**:
- Constructor injection for dependencies
- Single responsibility per use case
- Structured logging without sensitive data
- Context propagation for timeouts

### Adapter Layer

#### MCP Adapter (`internal/adapters/mcp/`)

**Purpose**: Translates between MCP protocol and domain

**Components**:
- **Server** (`server.go`): Main MCP server implementation
- **Handlers** (`handlers.go`): JSON-RPC 2.0 protocol handling
- **Tools** (`tools/`): Individual MCP tool implementations

**Tool Implementation Pattern**:
```go
type SearchTool struct {
    useCase *usecases.SearchUseCase
    logger  domain.Logger
}

func (t *SearchTool) Execute(ctx context.Context, args map[string]any) (*domain.ToolResult, error) {
    // 1. Parse and validate MCP arguments
    // 2. Convert to domain request
    // 3. Call use case
    // 4. Convert domain result to MCP result
    // 5. Return formatted result
}
```

**Error Handling**:
- MCP protocol compliance (JSON-RPC 2.0)
- Proper error codes and messages
- Input validation and sanitization
- Timeout handling with context

#### Perplexity Adapter (`internal/adapters/perplexity/`)

**Purpose**: Integrates with Perplexity API

**Components**:
- **Client** (`client.go`): HTTP client implementation
- **Models** (`models.go`): API request/response models
- **Mapper** (`mapper.go`): Domain ↔ API mapping

**HTTP Client Design**:
```go
type Client struct {
    httpClient HTTPClient
    baseURL    string
    apiKey     string
    timeout    time.Duration
    logger     domain.Logger
}

func (c *Client) Search(ctx context.Context, request domain.SearchRequest) (*domain.SearchResult, error) {
    // 1. Add safety timeout if not set
    // 2. Map domain request to API request
    // 3. Make HTTP call with context
    // 4. Handle HTTP errors → domain errors
    // 5. Map API response to domain result
}
```

**Security Features**:
- TLS 1.2+ enforcement
- Response size limiting (10MB max)
- Secure logging (no API keys/secrets)
- Input validation and sanitization

### Infrastructure Layer (`internal/infrastructure/`)

**Purpose**: External concerns (config, logging, HTTP)

#### Configuration (`config/`)
```go
type Config struct {
    perplexityAPIKey    string
    defaultModel        string
    requestTimeout      time.Duration
    logLevel            string
}

// Environment variable mapping:
// PERPLEXITY_API_KEY → perplexityAPIKey
// PERPLEXITY_DEFAULT_MODEL → defaultModel  
// REQUEST_TIMEOUT_SECONDS → requestTimeout
// LOG_LEVEL → logLevel
```

**Validation**:
- Required fields checked on startup
- Default values for optional settings
- Type conversion with fallbacks
- Detailed error messages

#### Logging (`logger/`)
```go
type Logger struct {
    slogger *slog.Logger
}

// Features:
// - JSON structured logging
// - Automatic sensitive data redaction
// - Context-aware logging
// - Level filtering (debug, info, warn, error)
```

**Security**:
- Redacts sensitive keys (password, api_key, token, etc.)
- No secrets in log output
- Configurable log levels
- Context propagation for distributed tracing

## Communication Patterns

### Request Flow
```
MCP Client → MCP Handler → Tool → UseCase → Client → Perplexity API
                ↓
           Domain Validation
                ↓  
           Context Timeout
                ↓
           Structured Logging
```

### Error Flow
```
API Error → Client Mapping → Domain Error → UseCase Wrapping → Tool Result → MCP Response
```

### Context Flow
```
HTTP Request Context → Tool Context → UseCase Context → Client Context → HTTP Client Context
```

## Testing Strategy

### Unit Tests
- Mock all external dependencies  
- Test each layer in isolation
- Focus on business logic and validation
- Use table-driven tests for edge cases

### Integration Tests
- Test complete request flows
- Mock external HTTP calls
- Verify MCP protocol compliance
- Test error handling and timeouts

### Benchmarks
- Single-threaded performance baselines
- Memory allocation patterns
- Different input sizes
- Error path performance

## Security Architecture

### Defense in Depth

1. **Input Layer**: MCP argument validation
2. **Domain Layer**: Business rule validation  
3. **Client Layer**: HTTP input sanitization
4. **Network Layer**: TLS enforcement

### Resource Protection

```go
// Example limits
const (
    MaxQueryLength        = 10000    // DoS prevention
    MaxMessageContent     = 50000    // Memory protection
    MaxResponseSize       = 10MB     // Response bounds
    MaxOptionsCount       = 20       // Resource limits
    MaxMessagesCount      = 100      // Processing limits
)
```

### Secure Logging

```go
// Sensitive keys automatically redacted
sensitiveKeys := []string{
    "password", "api_key", "token", "secret", 
    "private_key", "authorization", "session", 
    "cookie", "certificate", "cert",
}
```

## Performance Characteristics

### Benchmarks (Single-Thread Baseline)

| Operation | Latency | Memory | Allocations |
|-----------|---------|---------|-------------|
| Search Tool | 7.1μs | 9.2KB | 73 |
| Research Tool | 9.0μs | 10.5KB | 79 |  
| Input Validation | 181ns | 96B | 3 |
| Tool Listing | 2.6μs | 10.2KB | 78 |

### Memory Management
- Bounded input sizes prevent memory exhaustion
- Response streaming for large results
- Context-based cancellation prevents leaks
- Efficient JSON parsing with pools

## Deployment Architecture

### Binary Deployment
```bash
# Single binary with all dependencies
go build -ldflags="-w -s" -o server cmd/server/main.go

# Environment configuration
export PERPLEXITY_API_KEY="..."
export LOG_LEVEL="info"
export REQUEST_TIMEOUT_SECONDS="30"

# Run server
./server
```

### Container Deployment
```dockerfile
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -ldflags="-w -s" -o server cmd/server/main.go

FROM alpine:latest
RUN apk add --no-cache ca-certificates
COPY --from=builder /app/server /server
EXPOSE 8080
CMD ["/server"]
```

### Configuration Management
- Environment variables for runtime config
- No secrets in container images
- Health checks and monitoring hooks
- Graceful shutdown handling

## Extension Points

### Adding New Tools
1. Create tool implementation in `internal/adapters/mcp/tools/`
2. Implement the `Tool` interface
3. Add to server registration in `server.go`
4. Write tests and documentation

### Adding New Use Cases
1. Create use case in `internal/core/usecases/`
2. Define domain interfaces if needed
3. Wire dependencies in `cmd/server/main.go`
4. Add comprehensive tests

### Adding New Clients
1. Create adapter in `internal/adapters/`
2. Implement domain interfaces
3. Add configuration support
4. Create integration tests

## Future Architecture Considerations

### Concurrency
- Current: Single-threaded baseline established
- Future: Profile-driven concurrency behind interfaces
- Pattern: Worker pools for high throughput scenarios
- Constraint: Maintain deterministic behavior

### Scaling
- Current: Single process, stateless design
- Future: Horizontal scaling with load balancer
- Pattern: Stateless design enables simple scaling
- Consideration: Rate limiting and circuit breakers

### Monitoring
- Current: Structured logging with metrics
- Future: Distributed tracing and metrics
- Tools: OpenTelemetry integration
- Focus: Performance and error tracking

This architecture provides a solid foundation that's easy to understand, test, and extend while maintaining security and performance requirements.