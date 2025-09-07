package integration

import (
	"context"
	"testing"
	"time"

	"github.com/yourusername/perplexity-mcp-golang/internal/core/domain"
)

// TestMCPServerInitialization tests that the MCP server starts correctly with all required tools
func TestMCPServerInitialization(t *testing.T) {
	env := NewTestEnvironment(t)
	defer env.Reset()

	// Verify all required tools are registered
	expectedTools := ExpectedToolNames()
	for _, toolName := range expectedTools {
		env.AssertToolExists(t, toolName)
	}

	// Verify tool count
	actualCount := env.Server.GetToolCount()
	expectedCount := len(expectedTools)
	if actualCount != expectedCount {
		t.Errorf("Expected %d tools, got %d", expectedCount, actualCount)
	}

	// Verify no errors during startup
	env.AssertNoErrors(t)
}

// TestListTools tests the MCP ListTools functionality
func TestListTools(t *testing.T) {
	env := NewTestEnvironment(t)
	defer env.Reset()

	ctx := CreateTestContext()
	tools, err := env.Server.ListTools(ctx)
	if err != nil {
		t.Fatalf("ListTools failed: %v", err)
	}

	// Verify we got the expected tools
	expectedTools := ExpectedToolNames()
	if len(tools) != len(expectedTools) {
		t.Errorf("Expected %d tools, got %d", len(expectedTools), len(tools))
	}

	// Verify each tool has proper structure
	toolMap := make(map[string]domain.ToolInfo)
	for _, tool := range tools {
		toolMap[tool.Name] = tool
	}

	for _, expectedName := range expectedTools {
		tool, exists := toolMap[expectedName]
		if !exists {
			t.Errorf("Expected tool %s not found", expectedName)
			continue
		}

		// Validate tool structure
		if tool.Description == "" {
			t.Errorf("Tool %s has empty description", expectedName)
		}
		if tool.InputSchema == nil {
			t.Errorf("Tool %s has nil input schema", expectedName)
		}
	}
}

// TestSearchTool tests the complete search tool flow
func TestSearchTool(t *testing.T) {
	env := NewTestEnvironment(t)
	defer env.Reset()

	testData := NewTestData()
	ctx := CreateTestContext()

	// Execute search tool
	args := map[string]any{
		"query": testData.SearchQuery,
		"model": "llama-3.1-sonar-small-128k-online",
	}

	result, err := env.Server.ExecuteTool(ctx, "perplexity_search", args)
	if err != nil {
		t.Fatalf("Search tool execution failed: %v", err)
	}

	// Validate result
	ValidateToolResult(t, result, "", false)

	// Verify API call was made
	env.AssertAPICallMade(t, "Search", testData.SearchQuery)

	// Verify logging
	env.AssertLogEntryExists(t, "info", "Executing tool")
	env.AssertLogEntryExists(t, "info", "Tool execution completed successfully")

	// Verify no errors
	env.AssertNoErrors(t)
}

// TestChatTool tests the complete chat tool flow
func TestChatTool(t *testing.T) {
	env := NewTestEnvironment(t)
	defer env.Reset()

	testData := NewTestData()
	ctx := CreateTestContext()

	// Execute chat tool
	args := map[string]any{
		"messages": []map[string]string{
			{"role": "user", "content": testData.ChatMessages[0].Content},
		},
		"model": "llama-3.1-sonar-small-128k-chat",
	}

	result, err := env.Server.ExecuteTool(ctx, "perplexity_chat", args)
	if err != nil {
		t.Fatalf("Chat tool execution failed: %v", err)
	}

	// Validate result
	ValidateToolResult(t, result, "", false)

	// Verify API call was made
	env.AssertAPICallMade(t, "Chat", testData.ChatMessages[0].Content)

	// Verify logging
	env.AssertLogEntryExists(t, "info", "Executing tool")
	env.AssertLogEntryExists(t, "info", "Tool execution completed successfully")

	// Verify no errors
	env.AssertNoErrors(t)
}

// TestResearchTool tests the complete research tool flow
func TestResearchTool(t *testing.T) {
	env := NewTestEnvironment(t)
	defer env.Reset()

	testData := NewTestData()
	ctx := CreateTestContext()

	// Execute research tool
	args := map[string]any{
		"topic":            testData.ResearchTopic,
		"reasoning_effort": "thorough",
		"model":            "llama-3.1-sonar-large-128k-online",
	}

	result, err := env.Server.ExecuteTool(ctx, "perplexity_research", args)
	if err != nil {
		t.Fatalf("Research tool execution failed: %v", err)
	}

	// Validate result
	ValidateToolResult(t, result, "", false)

	// Verify API call was made
	env.AssertAPICallMade(t, "Research", testData.ResearchTopic)

	// Verify logging
	env.AssertLogEntryExists(t, "info", "Executing tool")
	env.AssertLogEntryExists(t, "info", "Tool execution completed successfully")

	// Verify no errors
	env.AssertNoErrors(t)
}

// TestToolInputValidation tests validation of tool inputs
func TestToolInputValidation(t *testing.T) {
	env := NewTestEnvironment(t)
	defer env.Reset()

	ctx := CreateTestContext()

	testCases := []struct {
		name     string
		toolName string
		args     map[string]any
		wantErr  bool
	}{
		{
			name:     "search with empty query",
			toolName: "perplexity_search",
			args:     map[string]any{"query": ""},
			wantErr:  true,
		},
		{
			name:     "chat with empty messages",
			toolName: "perplexity_chat",
			args:     map[string]any{"messages": []map[string]string{}},
			wantErr:  true,
		},
		{
			name:     "research with empty topic",
			toolName: "perplexity_research",
			args:     map[string]any{"topic": ""},
			wantErr:  true,
		},
		{
			name:     "search with valid query",
			toolName: "perplexity_search",
			args:     map[string]any{"query": "valid query"},
			wantErr:  false,
		},
		{
			name:     "chat with valid messages",
			toolName: "perplexity_chat",
			args:     map[string]any{"messages": []map[string]string{{"role": "user", "content": "hello"}}},
			wantErr:  false,
		},
		{
			name:     "research with valid topic",
			toolName: "perplexity_research",
			args:     map[string]any{"topic": "valid topic"},
			wantErr:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := env.Server.ExecuteTool(ctx, tc.toolName, tc.args)

			if tc.wantErr {
				// Should either return error or result with IsError=true
				if err == nil && (result == nil || !result.IsError) {
					t.Error("Expected error but none occurred")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result != nil && result.IsError {
					t.Errorf("Result marked as error: %s", result.Content)
				}
			}
		})
	}
}

// TestToolNotFound tests handling of non-existent tools
func TestToolNotFound(t *testing.T) {
	env := NewTestEnvironment(t)
	defer env.Reset()

	ctx := CreateTestContext()
	result, err := env.Server.ExecuteTool(ctx, "nonexistent_tool", map[string]any{})

	// Should return both error and error result
	if err == nil {
		t.Error("Expected error for non-existent tool")
	}

	if result == nil {
		t.Fatal("Expected error result for non-existent tool")
	}

	if !result.IsError {
		t.Error("Result should be marked as error")
	}

	// Verify error logging
	env.AssertLogEntryExists(t, "error", "Tool not found")
}

// TestContextTimeout tests handling of context timeouts
func TestContextTimeout(t *testing.T) {
	env := NewTestEnvironment(t)
	defer env.Reset()

	// Configure delay that exceeds timeout
	env.SimulateNetworkDelay(100 * time.Millisecond)

	// Create context that times out quickly
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	args := map[string]any{
		"query": "timeout test",
	}

	result, err := env.Server.ExecuteTool(ctx, "perplexity_search", args)

	// Should handle timeout gracefully
	if err == nil && (result == nil || !result.IsError) {
		t.Error("Expected timeout error")
	}
}

// TestConcurrentToolExecution tests that tools can be executed concurrently safely
func TestConcurrentToolExecution(t *testing.T) {
	env := NewTestEnvironment(t)
	defer env.Reset()

	ctx := CreateTestContext()
	concurrency := 10
	results := make(chan error, concurrency)

	// Execute multiple tools concurrently
	for i := 0; i < concurrency; i++ {
		go func(id int) {
			args := map[string]any{
				"query": "concurrent test query",
			}
			_, err := env.Server.ExecuteTool(ctx, "perplexity_search", args)
			results <- err
		}(i)
	}

	// Collect results
	for i := 0; i < concurrency; i++ {
		if err := <-results; err != nil {
			t.Errorf("Concurrent execution %d failed: %v", i, err)
		}
	}

	// Verify all API calls were made
	calls := env.MockClient.FindCalls("Search", "concurrent test query")
	if len(calls) != concurrency {
		t.Errorf("Expected %d API calls, got %d", concurrency, len(calls))
	}
}

// TestAPIError tests handling of API errors
func TestAPIError(t *testing.T) {
	env := NewTestEnvironment(t)
	defer env.Reset()

	// Configure mock to return error
	testQuery := "error test query"
	env.SimulateAPIError(testQuery, domain.ErrAPIError)

	ctx := CreateTestContext()
	args := map[string]any{
		"query": testQuery,
	}

	result, err := env.Server.ExecuteTool(ctx, "perplexity_search", args)

	// Should handle API error gracefully
	if err == nil {
		t.Error("Expected error from API failure")
	}

	if result == nil || !result.IsError {
		t.Error("Expected error result from API failure")
	}

	// Verify error logging
	env.AssertLogEntryExists(t, "error", "Tool execution failed")
}

// TestToolSchemas tests that all tools have proper input schemas
func TestToolSchemas(t *testing.T) {
	env := NewTestEnvironment(t)
	defer env.Reset()

	expectedSchemas := map[string][]string{
		"perplexity_search":   {"query"},
		"perplexity_chat":     {"messages"},
		"perplexity_research": {"topic"},
	}

	for toolName, expectedFields := range expectedSchemas {
		t.Run(toolName, func(t *testing.T) {
			schema := env.GetToolSchema(t, toolName)
			ValidateInputSchema(t, schema, expectedFields)
		})
	}
}

// TestEndToEndFlow tests a complete end-to-end flow using all tools
func TestEndToEndFlow(t *testing.T) {
	env := NewTestEnvironment(t)
	defer env.Reset()

	ctx := CreateTestContext()
	testData := NewTestData()

	// Step 1: Search for information
	searchArgs := map[string]any{
		"query": testData.SearchQuery,
	}
	searchResult, err := env.Server.ExecuteTool(ctx, "perplexity_search", searchArgs)
	if err != nil {
		t.Fatalf("Search step failed: %v", err)
	}
	ValidateToolResult(t, searchResult, "", false)

	// Step 2: Have a chat conversation
	chatArgs := map[string]any{
		"messages": []map[string]string{
			{"role": "user", "content": testData.ChatMessages[0].Content},
		},
	}
	chatResult, err := env.Server.ExecuteTool(ctx, "perplexity_chat", chatArgs)
	if err != nil {
		t.Fatalf("Chat step failed: %v", err)
	}
	ValidateToolResult(t, chatResult, "", false)

	// Step 3: Research a topic
	researchArgs := map[string]any{
		"topic": testData.ResearchTopic,
	}
	researchResult, err := env.Server.ExecuteTool(ctx, "perplexity_research", researchArgs)
	if err != nil {
		t.Fatalf("Research step failed: %v", err)
	}
	ValidateToolResult(t, researchResult, "", false)

	// Verify all API calls were made
	env.AssertAPICallMade(t, "Search", testData.SearchQuery)
	env.AssertAPICallMade(t, "Chat", testData.ChatMessages[0].Content)
	env.AssertAPICallMade(t, "Research", testData.ResearchTopic)

	// Verify proper logging throughout
	env.AssertLogEntryExists(t, "info", "MCP server started successfully")

	// Should have exactly 3 tool executions
	executionEntries := env.MockLogger.GetEntriesWithMessage("Tool execution completed successfully")
	if len(executionEntries) != 3 {
		t.Errorf("Expected 3 tool executions, got %d", len(executionEntries))
	}

	// Verify no errors
	env.AssertNoErrors(t)
}
