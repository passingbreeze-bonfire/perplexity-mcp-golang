package integration

import (
	"context"
	"testing"
	"time"

	"github.com/yourusername/perplexity-mcp-golang/internal/adapters/mcp"
	"github.com/yourusername/perplexity-mcp-golang/internal/core/domain"
	"github.com/yourusername/perplexity-mcp-golang/internal/core/usecases"
	"github.com/yourusername/perplexity-mcp-golang/test/mocks"
)

// TestEnvironment holds all components needed for integration testing
type TestEnvironment struct {
	MockClient *mocks.MockPerplexityClient
	MockConfig *mocks.MockConfigProvider
	MockLogger *mocks.MockLogger
	Server     *mcp.Server

	// Use cases for direct testing
	SearchUseCase *usecases.SearchUseCase
}

// NewTestEnvironment creates a fully configured test environment
func NewTestEnvironment(t *testing.T) *TestEnvironment {
	t.Helper()

	// Create mocks
	mockClient := mocks.NewMockPerplexityClient()
	mockConfig := mocks.NewMockConfig()
	mockLogger := mocks.NewMockLogger()

	// Create use cases with mocks
	searchUseCase := usecases.NewSearchUseCase(mockClient, mockConfig, mockLogger)

	// Create MCP server
	server := mcp.NewServer(mockLogger, searchUseCase)

	env := &TestEnvironment{
		MockClient:    mockClient,
		MockConfig:    mockConfig,
		MockLogger:    mockLogger,
		Server:        server,
		SearchUseCase: searchUseCase,
	}

	// Start server in test mode
	ctx := context.Background()
	if err := server.Start(ctx); err != nil {
		t.Fatalf("Failed to start MCP server: %v", err)
	}

	return env
}

// Reset clears all mock state for a clean test slate
func (env *TestEnvironment) Reset() {
	env.MockClient.Reset()
	env.MockLogger.Clear()
}

// AssertNoErrors checks that no errors were logged during test execution
func (env *TestEnvironment) AssertNoErrors(t *testing.T) {
	t.Helper()
	if env.MockLogger.HasError() {
		t.Errorf("Unexpected errors logged during test:\n%s", env.MockLogger.String())
	}
}

// AssertToolExists verifies that a tool is registered in the server
func (env *TestEnvironment) AssertToolExists(t *testing.T, toolName string) {
	t.Helper()
	if !env.Server.HasTool(toolName) {
		t.Errorf("Tool %s not found in server", toolName)
	}
}

// AssertAPICallMade verifies that a specific API call was made
func (env *TestEnvironment) AssertAPICallMade(t *testing.T, method, queryPattern string) {
	t.Helper()
	calls := env.MockClient.FindCalls(method, queryPattern)
	if len(calls) == 0 {
		t.Errorf("Expected API call not found: method=%s, pattern=%s", method, queryPattern)
	}
}

// AssertLogEntryExists verifies that a log entry with specific level and message exists
func (env *TestEnvironment) AssertLogEntryExists(t *testing.T, level, messagePattern string) {
	t.Helper()
	if !env.MockLogger.HasEntry(level, messagePattern) {
		t.Errorf("Expected log entry not found: level=%s, pattern=%s", level, messagePattern)
	}
}

// CreateTestContext creates a context with a reasonable timeout for testing
func CreateTestContext() context.Context {
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	return ctx
}

// CreateTimeoutContext creates a context that will timeout quickly for testing timeout scenarios
func CreateTimeoutContext() context.Context {
	ctx, _ := context.WithTimeout(context.Background(), 1*time.Millisecond)
	time.Sleep(2 * time.Millisecond) // Ensure context is already timed out
	return ctx
}

// TestData contains common test data for consistent testing
type TestData struct {
	SearchQuery  string
	InvalidQuery string
	LongQuery    string
	EmptyQuery   string
}

// NewTestData creates common test data
func NewTestData() *TestData {
	longQuery := make([]byte, domain.MaxQueryLength+1)
	for i := range longQuery {
		longQuery[i] = 'a'
	}

	return &TestData{
		SearchQuery:  "What is artificial intelligence?",
		InvalidQuery: "", // Empty query should be invalid
		LongQuery:    string(longQuery),
		EmptyQuery:   "",
	}
}

// ExpectedToolNames returns the list of tools that should be registered
func ExpectedToolNames() []string {
	return []string{
		"perplexity_search",
		"perplexity_chat",
		"perplexity_research",
	}
}

// ValidateToolResult validates that a tool result has the expected structure
func ValidateToolResult(t *testing.T, result *domain.ToolResult, expectedContent string, shouldBeError bool) {
	t.Helper()

	if result == nil {
		t.Fatal("Tool result is nil")
	}

	if result.IsError != shouldBeError {
		t.Errorf("Expected IsError=%v, got %v", shouldBeError, result.IsError)
	}

	if expectedContent != "" && result.Content != expectedContent {
		t.Errorf("Expected content %q, got %q", expectedContent, result.Content)
	}

	if result.Content == "" {
		t.Error("Tool result content is empty")
	}
}

// SimulateNetworkDelay configures the mock client to simulate network latency
func (env *TestEnvironment) SimulateNetworkDelay(delay time.Duration) {
	env.MockClient.SetDelay(delay)
}

// SimulateAPIError configures the mock client to return an error for specific queries
func (env *TestEnvironment) SimulateAPIError(queryOrTopic string, err error) {
	env.MockClient.SetError(queryOrTopic, err)
}

// GetToolSchema retrieves and validates the input schema for a tool
func (env *TestEnvironment) GetToolSchema(t *testing.T, toolName string) map[string]any {
	t.Helper()

	ctx := CreateTestContext()
	tools, err := env.Server.ListTools(ctx)
	if err != nil {
		t.Fatalf("Failed to list tools: %v", err)
	}

	for _, tool := range tools {
		if tool.Name == toolName {
			return tool.InputSchema
		}
	}

	t.Fatalf("Tool %s not found", toolName)
	return nil
}

// ValidateInputSchema checks that an input schema has the expected structure
func ValidateInputSchema(t *testing.T, schema map[string]any, expectedFields []string) {
	t.Helper()

	if schema == nil {
		t.Fatal("Input schema is nil")
	}

	// Check for required fields
	properties, ok := schema["properties"].(map[string]any)
	if !ok {
		t.Fatal("Schema missing properties field")
	}

	for _, field := range expectedFields {
		if _, exists := properties[field]; !exists {
			t.Errorf("Schema missing required field: %s", field)
		}
	}
}

// BenchmarkHelper provides utilities for benchmark tests
type BenchmarkHelper struct {
	env *TestEnvironment
}

// NewBenchmarkHelper creates a new benchmark helper
func NewBenchmarkHelper() *BenchmarkHelper {
	// Note: We don't use testing.T here since this is for benchmarks
	mockClient := mocks.NewMockPerplexityClient()
	mockConfig := mocks.NewMockConfig()
	mockLogger := mocks.NewMockLogger()

	searchUseCase := usecases.NewSearchUseCase(mockClient, mockConfig, mockLogger)

	server := mcp.NewServer(mockLogger, searchUseCase)

	// Start server
	ctx := context.Background()
	if err := server.Start(ctx); err != nil {
		panic("Failed to start server for benchmark: " + err.Error())
	}

	return &BenchmarkHelper{
		env: &TestEnvironment{
			MockClient:    mockClient,
			MockConfig:    mockConfig,
			MockLogger:    mockLogger,
			Server:        server,
			SearchUseCase: searchUseCase,
		},
	}
}

// GetEnvironment returns the test environment for benchmarking
func (bh *BenchmarkHelper) GetEnvironment() *TestEnvironment {
	return bh.env
}
