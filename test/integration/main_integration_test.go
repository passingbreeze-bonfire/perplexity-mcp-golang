package integration

import (
	"context"
	"os"
	"testing"
	"time"
)

// TestMain is the main entry point for integration tests
func TestMain(m *testing.M) {
	// Setup any global test resources here if needed

	// Run tests
	code := m.Run()

	// Cleanup any global resources here if needed

	os.Exit(code)
}

// TestComprehensiveIntegrationSuite runs the complete integration test suite
func TestComprehensiveIntegrationSuite(t *testing.T) {
	// Skip if running short tests
	if testing.Short() {
		t.Skip("Skipping comprehensive integration tests in short mode")
	}

	runner := NewTestRunner(t)

	// Configure for comprehensive testing
	config := TestConfig{
		EnableDetailedLogs: true,
		FailFast:           false,
		GenerateReport:     true,
		ReportOutputPath:   "integration_test_report.json",
	}

	runner.WithConfig(config).RunAllIntegrationTests(t)
}

// TestQuickIntegrationSuite runs a subset of critical integration tests for quick feedback
func TestQuickIntegrationSuite(t *testing.T) {
	env := NewTestEnvironment(t)
	defer env.Reset()

	t.Run("ServerStartup", func(t *testing.T) {
		// Server should already be started by NewTestEnvironment
		expectedTools := ExpectedToolNames()
		for _, toolName := range expectedTools {
			env.AssertToolExists(t, toolName)
		}
		env.AssertNoErrors(t)
	})

	t.Run("BasicToolExecution", func(t *testing.T) {
		ctx := CreateTestContext()
		testData := NewTestData()

		// Test search tool
		args := map[string]any{"query": testData.SearchQuery}
		result, err := env.Server.ExecuteTool(ctx, "perplexity_search", args)
		if err != nil {
			t.Fatalf("Search tool failed: %v", err)
		}
		ValidateToolResult(t, result, "", false)

		// Test chat tool
		chatArgs := map[string]any{
			"messages": []map[string]string{
				{"role": "user", "content": testData.ChatMessages[0].Content},
			},
		}
		result, err = env.Server.ExecuteTool(ctx, "perplexity_chat", chatArgs)
		if err != nil {
			t.Fatalf("Chat tool failed: %v", err)
		}
		ValidateToolResult(t, result, "", false)

		// Test research tool
		researchArgs := map[string]any{"topic": testData.ResearchTopic}
		result, err = env.Server.ExecuteTool(ctx, "perplexity_research", researchArgs)
		if err != nil {
			t.Fatalf("Research tool failed: %v", err)
		}
		ValidateToolResult(t, result, "", false)

		env.AssertNoErrors(t)
	})
}

// TestIntegrationWithRealTimeouts tests behavior with realistic timeout scenarios
func TestIntegrationWithRealTimeouts(t *testing.T) {
	env := NewTestEnvironment(t)
	defer env.Reset()

	t.Run("ReasonableTimeout", func(t *testing.T) {
		// Configure reasonable delay
		env.SimulateNetworkDelay(50 * time.Millisecond)

		ctx := CreateTestContext() // 5 second timeout
		args := map[string]any{"query": "timeout test with reasonable delay"}

		result, err := env.Server.ExecuteTool(ctx, "perplexity_search", args)
		if err != nil {
			t.Fatalf("Tool should succeed with reasonable timeout: %v", err)
		}

		ValidateToolResult(t, result, "", false)
	})

	t.Run("ExcessiveDelay", func(t *testing.T) {
		// Configure excessive delay
		env.SimulateNetworkDelay(10 * time.Second)

		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		args := map[string]any{"query": "timeout test with excessive delay"}

		result, err := env.Server.ExecuteTool(ctx, "perplexity_search", args)
		// Should handle timeout gracefully
		if err == nil && (result == nil || !result.IsError) {
			t.Error("Expected timeout handling for excessive delay")
		}
	})
}

// TestIntegrationEdgeCases tests various edge cases in integration scenarios
func TestIntegrationEdgeCases(t *testing.T) {
	env := NewTestEnvironment(t)
	defer env.Reset()

	testData := NewTestData()

	t.Run("EmptyInputs", func(t *testing.T) {
		ctx := CreateTestContext()

		// Test with empty query
		args := map[string]any{"query": ""}
		result, err := env.Server.ExecuteTool(ctx, "perplexity_search", args)

		// Should handle empty input gracefully
		if err == nil && (result == nil || !result.IsError) {
			t.Error("Expected error for empty query")
		}
	})

	t.Run("VeryLongInputs", func(t *testing.T) {
		ctx := CreateTestContext()

		// Test with very long query
		args := map[string]any{"query": testData.LongQuery}
		result, err := env.Server.ExecuteTool(ctx, "perplexity_search", args)

		// Should handle long input according to validation rules
		if err == nil && (result == nil || !result.IsError) {
			t.Error("Expected validation error for overly long query")
		}
	})

	t.Run("InvalidToolArguments", func(t *testing.T) {
		ctx := CreateTestContext()

		// Test with missing required arguments
		args := map[string]any{} // No query
		result, err := env.Server.ExecuteTool(ctx, "perplexity_search", args)

		// Should handle missing arguments gracefully
		if err == nil && (result == nil || !result.IsError) {
			t.Error("Expected error for missing required arguments")
		}
	})

	t.Run("InvalidToolName", func(t *testing.T) {
		ctx := CreateTestContext()

		args := map[string]any{"query": testData.SearchQuery}
		result, err := env.Server.ExecuteTool(ctx, "invalid_tool_name", args)

		// Should handle invalid tool name gracefully
		if err == nil {
			t.Error("Expected error for invalid tool name")
		}
		if result == nil || !result.IsError {
			t.Error("Expected error result for invalid tool name")
		}
	})
}
