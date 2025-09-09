# Integration Tests for Stdio Transport MCP Server

This directory contains comprehensive integration tests for the stdio transport MCP server implementation.

## Test Coverage

The integration tests cover the following scenarios:

### Core Functionality Tests
- **Basic Communication** (`TestStdioTransportBasic`): Verifies JSON-RPC protocol initialization
- **Tools List** (`TestStdioTransportToolsList`): Tests listing available tools
- **Valid API Requests** (`TestStdioTransportValidRequest`): Tests actual Perplexity API calls (requires real API key)

### Error Handling Tests
- **Malformed Input** (`TestStdioTransportMalformedInput`): Tests handling of invalid JSON
- **Empty Input** (`TestStdioTransportEmptyInput`): Tests handling of empty requests
- **Large Payloads** (`TestStdioTransportLargePayload`): Tests handling of oversized requests
- **Invalid Tool Names** (`TestStdioTransportInvalidToolName`): Tests error responses for unknown tools
- **Invalid Method Names** (`TestStdioTransportInvalidMethodName`): Tests error responses for invalid JSON-RPC methods

### Protocol Compliance Tests
- **Stderr Logging** (`TestStdioTransportStderrLogsOnly`): Verifies all logs go to stderr, not stdout
- **Resource Cleanup** (`TestStdioTransportResourceCleanup`): Tests proper process termination
- **Sequential Requests** (`TestStdioTransportSequentialRequests`): Tests handling multiple sequential requests

### Performance Tests
- **Benchmark** (`BenchmarkStdioTransport`): Measures stdio transport performance

## Running the Tests

### Run All Integration Tests
```bash
go test -v ./cmd/server -run "TestStdio"
```

### Run Individual Test
```bash
go test -v ./cmd/server -run TestStdioTransportBasic
```

### Run with Real API Calls
To test actual Perplexity API integration, set your API key:
```bash
export PERPLEXITY_API_KEY="your-real-api-key"
go test -v ./cmd/server -run TestStdioTransportValidRequest
```

### Run Performance Benchmark
```bash
go test -v ./cmd/server -bench BenchmarkStdioTransport -benchtime=5s
```

## Test Architecture

### TestHelper
The tests use a `TestHelper` struct that manages subprocess communication:
- Spawns the server as a child process using `os/exec`
- Provides stdin/stdout/stderr pipes for communication
- Handles JSON-RPC request/response serialization
- Manages process lifecycle and cleanup

### Test Pattern
Each test follows this pattern:
1. Create and start TestHelper
2. Send initialize request to establish MCP session
3. Send test-specific requests
4. Verify responses match expected format
5. Clean up subprocess

### Key Features Tested

#### Stdio Transport Protocol
- Server communicates via stdin/stdout only
- All logging redirected to stderr
- JSON-RPC 2.0 protocol compliance
- Proper request/response correlation

#### Error Handling
- Graceful handling of malformed input
- Appropriate error responses for invalid requests
- Server remains responsive after errors
- Proper exit codes and cleanup

#### Resource Management
- Clean process startup and shutdown
- No resource leaks
- Proper signal handling
- Timeout management

## Dependencies

The tests require:
- `github.com/stretchr/testify` for assertions
- Go 1.24+ for subprocess management
- Environment variable `PERPLEXITY_API_KEY` for API tests (optional)

## Test Environment

Tests create a controlled environment:
- Set test API key in environment
- Spawn fresh server process for each test
- Use timeouts to prevent hanging
- Capture and verify stderr output

## Troubleshooting

### Common Issues

1. **Test Timeouts**: Server may not be responding to requests
   - Check stderr logs for server startup errors
   - Verify API key is set correctly
   - Ensure no other process is using required resources

2. **JSON Parsing Errors**: Usually indicates stdout contamination
   - Verify all logging goes to stderr
   - Check for printf/print statements in code
   - Confirm proper JSON-RPC formatting

3. **Process Management**: Tests failing to start/stop server
   - Check system resources and process limits
   - Verify Go build environment is working
   - Ensure test permissions are sufficient

### Debug Output

To see detailed test output including stderr logs:
```bash
go test -v ./cmd/server -run TestStdioTransportBasic -args -test.v
```

## Test Coverage Statistics

The integration tests provide comprehensive coverage of:
- ✅ JSON-RPC protocol implementation
- ✅ MCP tool registration and execution
- ✅ Error handling and edge cases
- ✅ Process lifecycle management
- ✅ Stdio transport compliance
- ✅ Resource cleanup and timeouts

These tests ensure the server works correctly as a subprocess in MCP client environments.