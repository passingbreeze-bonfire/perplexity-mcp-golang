# Perplexity MCP Golang - Test Suite

This directory contains comprehensive tests for the Perplexity MCP server, including unit tests, integration tests, benchmarks, and mocks.

## Structure

```
test/
├── benchmark/           # Performance benchmarks
├── integration/         # End-to-end integration tests  
├── mocks/              # Mock implementations for testing
├── unit/               # Unit tests (organized by component)
└── README.md           # This file
```

## Test Categories

### Integration Tests (`integration/`)

Comprehensive end-to-end tests that verify the complete MCP protocol flow and system integration.

**Key Features:**
- Full MCP server lifecycle testing
- Tool registration and execution testing
- Use case integration testing
- Error handling and edge case testing
- Context timeout and cancellation testing
- Concurrent access testing

**Files:**
- `testutil.go` - Test utilities and helpers
- `mcp_server_test.go` - MCP server integration tests
- `usecases_test.go` - Use case integration tests
- `test_runner.go` - Comprehensive test runner with reporting
- `main_integration_test.go` - Main integration test entry point

### Performance Benchmarks (`benchmark/`)

Single-thread-first performance benchmarks following Go best practices for establishing performance baselines.

**Key Features:**
- Tool execution performance benchmarks
- Use case performance benchmarks  
- Memory allocation and GC pressure analysis
- Different input size performance testing
- Error handling performance testing
- Sequential workflow benchmarks

**Files:**
- `mcp_server_bench_test.go` - MCP server performance benchmarks
- `usecases_bench_test.go` - Use case performance benchmarks

### Mock Implementations (`mocks/`)

Reliable mock implementations for external dependencies to ensure consistent and fast testing.

**Key Features:**
- Complete PerplexityClient mock with configurable responses
- Mock configuration provider
- Mock logger with verification capabilities
- Thread-safe implementations
- Configurable delays and errors for testing edge cases

**Files:**
- `perplexity_client.go` - Mock Perplexity API client
- `config.go` - Mock configuration provider  
- `logger.go` - Mock logger with capture and verification

## Running Tests

### Quick Integration Tests
```bash
# Run quick integration tests for fast feedback
go test ./test/integration -run TestQuickIntegrationSuite -v
```

### Comprehensive Integration Tests
```bash
# Run full integration test suite (skipped in short mode)
go test ./test/integration -run TestComprehensiveIntegrationSuite -v
```

### Performance Benchmarks
```bash
# Run all benchmarks
go test ./test/benchmark -bench=. -benchmem -v

# Run specific benchmark categories
go test ./test/benchmark -bench=BenchmarkSearchTool -benchmem -v
go test ./test/benchmark -bench=BenchmarkUseCases -benchmem -v
```

### All Tests with Coverage
```bash
# Run all tests with coverage report
go test ./... -cover -v

# Generate detailed coverage report
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html
```

### Short Test Mode
```bash
# Run only fast tests (skips comprehensive integration tests)
go test ./... -short -v
```

## Test Configuration

### Environment Variables
- `PERPLEXITY_API_KEY` - Set to "test" or any value for integration tests (uses mocks)
- `TEST_TIMEOUT` - Override default test timeout (default: 5s for integration tests)
- `TEST_LOG_LEVEL` - Set log level for test output (debug, info, warn, error)

### Test Tags
```bash
# Run only integration tests
go test -tags=integration ./test/integration -v

# Run only benchmark tests  
go test -tags=benchmark ./test/benchmark -bench=. -v
```

## Design Principles

### Single-Thread-First Policy
All benchmarks establish single-threaded performance baselines before any concurrency optimization. This follows the project's architectural principle of:

1. **Correctness First** - Ensure single-threaded implementation is correct
2. **Performance Baseline** - Establish baseline metrics for optimization decisions  
3. **Deterministic Behavior** - Prefer predictable, deterministic execution
4. **Safe Scaling** - Only add concurrency when bottlenecks are proven via profiling

### Reliable Testing
- **Mock External Dependencies** - All external API calls are mocked for reliability
- **Deterministic Results** - Tests produce consistent results across environments
- **Fast Execution** - Integration tests complete quickly using mocks
- **Comprehensive Coverage** - Test both happy paths and error scenarios

### Clean Architecture Testing
Tests are organized by architectural layer:
- **Integration Tests** - Test complete system integration
- **Use Case Tests** - Test business logic in isolation
- **Mock Layer** - Isolate external dependencies
- **Benchmark Tests** - Measure performance at each layer

## Performance Expectations

### Baseline Performance Targets (Single-Thread)
Based on mock implementations, real performance will vary:

- **Tool Execution**: < 1ms per operation (mocked)
- **Input Validation**: < 100μs per request  
- **Memory Allocation**: Minimal per request (< 1KB)
- **Concurrent Safety**: No data races under concurrent load

### Benchmark Interpretation
- Use benchmark results to compare relative performance between changes
- Mock-based benchmarks establish overhead baselines
- Real API performance will be dominated by network latency
- Focus on allocation patterns and algorithmic complexity

## Troubleshooting

### Common Issues

**Tests failing with timeout:**
```bash
# Increase test timeout
go test ./test/integration -timeout=30s -v
```

**Mock not returning expected data:**
```go
// Configure mock responses in test setup
env.MockClient.SetSearchResponse("query", &domain.SearchResult{...})
```

**Memory allocation warnings:**
```bash
# Run with memory profiling
go test ./test/benchmark -bench=BenchmarkMemory -benchmem -memprofile=mem.prof
go tool pprof mem.prof
```

**Coverage reports missing:**
```bash
# Ensure all packages are included
go test ./... -coverprofile=coverage.out -coverpkg=./...
```

### Debug Mode
Enable detailed logging in tests:
```go
env := NewTestEnvironment(t)
env.MockLogger.SetLevel("debug")
```

## Contributing

When adding new tests:

1. **Follow naming conventions** - `Test*` for tests, `Benchmark*` for benchmarks
2. **Use test utilities** - Leverage existing test helpers and utilities
3. **Update mocks as needed** - Add new mock responses for new test scenarios  
4. **Document performance expectations** - Include comments about expected performance characteristics
5. **Test error cases** - Ensure error paths are tested, not just happy paths
6. **Single-thread first** - Establish single-threaded baselines before adding concurrency tests

## Future Enhancements

- **Load Testing** - Add load tests for production deployment validation
- **Chaos Testing** - Add failure injection for resilience testing  
- **Property-Based Testing** - Add generative testing for edge cases
- **Contract Testing** - Add API contract validation
- **Performance Regression Testing** - Automated performance regression detection