package integration

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/yourusername/perplexity-mcp-golang/internal/core/domain"
)

// TestRunner provides utilities for running comprehensive integration tests
type TestRunner struct {
	env         *TestEnvironment
	testResults []TestResult
	startTime   time.Time
	config      TestConfig
}

// TestResult captures the result of an individual test
type TestResult struct {
	TestName     string
	Success      bool
	Duration     time.Duration
	Error        error
	Details      map[string]interface{}
	APICallCount int
	LogEntries   int
}

// TestConfig configures test execution parameters
type TestConfig struct {
	Timeout            time.Duration
	EnableDetailedLogs bool
	MaxConcurrentTests int
	FailFast           bool
	GenerateReport     bool
	ReportOutputPath   string
}

// NewTestRunner creates a new test runner with default configuration
func NewTestRunner(t *testing.T) *TestRunner {
	return &TestRunner{
		env:         NewTestEnvironment(t),
		testResults: make([]TestResult, 0),
		startTime:   time.Now(),
		config: TestConfig{
			Timeout:            30 * time.Second,
			EnableDetailedLogs: false,
			MaxConcurrentTests: 1, // Single-thread first policy
			FailFast:           false,
			GenerateReport:     true,
			ReportOutputPath:   "test_results.json",
		},
	}
}

// WithConfig applies custom configuration to the test runner
func (tr *TestRunner) WithConfig(config TestConfig) *TestRunner {
	tr.config = config
	return tr
}

// RunAllIntegrationTests executes a comprehensive suite of integration tests
func (tr *TestRunner) RunAllIntegrationTests(t *testing.T) {
	t.Log("Starting comprehensive integration test suite")

	// Core functionality tests
	tr.runTest(t, "ServerInitialization", tr.testServerInitialization)
	tr.runTest(t, "ToolRegistration", tr.testToolRegistration)
	tr.runTest(t, "SearchToolFlow", tr.testSearchToolFlow)
	tr.runTest(t, "ChatToolFlow", tr.testChatToolFlow)
	tr.runTest(t, "ResearchToolFlow", tr.testResearchToolFlow)

	// Validation tests
	tr.runTest(t, "InputValidation", tr.testInputValidation)
	tr.runTest(t, "ErrorHandling", tr.testErrorHandling)

	// Edge cases
	tr.runTest(t, "TimeoutHandling", tr.testTimeoutHandling)
	tr.runTest(t, "ConcurrentAccess", tr.testConcurrentAccess)

	// End-to-end workflows
	tr.runTest(t, "CompleteWorkflow", tr.testCompleteWorkflow)

	// Generate report if configured
	if tr.config.GenerateReport {
		tr.generateReport(t)
	}

	// Summary
	tr.printSummary(t)
}

// runTest executes a single test and records the result
func (tr *TestRunner) runTest(t *testing.T, testName string, testFunc func() error) {
	if tr.config.FailFast && tr.hasFailures() {
		t.Logf("Skipping test %s due to FailFast configuration", testName)
		return
	}

	t.Logf("Running test: %s", testName)
	startTime := time.Now()

	// Reset environment for clean state
	tr.env.Reset()

	// Record initial state
	initialAPICallCount := tr.env.MockClient.GetCallCount()
	initialLogEntries := len(tr.env.MockLogger.GetEntries())

	// Execute test with timeout
	ctx, cancel := context.WithTimeout(context.Background(), tr.config.Timeout)
	defer cancel()

	var testErr error
	done := make(chan error, 1)

	go func() {
		done <- testFunc()
	}()

	select {
	case testErr = <-done:
		// Test completed
	case <-ctx.Done():
		testErr = fmt.Errorf("test timeout after %v", tr.config.Timeout)
	}

	duration := time.Since(startTime)

	// Record result
	result := TestResult{
		TestName:     testName,
		Success:      testErr == nil,
		Duration:     duration,
		Error:        testErr,
		APICallCount: tr.env.MockClient.GetCallCount() - initialAPICallCount,
		LogEntries:   len(tr.env.MockLogger.GetEntries()) - initialLogEntries,
		Details:      make(map[string]interface{}),
	}

	// Add detailed information if enabled
	if tr.config.EnableDetailedLogs {
		result.Details["log_entries"] = tr.env.MockLogger.GetEntries()[initialLogEntries:]
		result.Details["api_calls"] = tr.env.MockClient.GetCallHistory()[initialAPICallCount:]
	}

	tr.testResults = append(tr.testResults, result)

	if testErr != nil {
		t.Errorf("Test %s failed: %v", testName, testErr)
	} else {
		t.Logf("Test %s passed in %v", testName, duration)
	}
}

// Test implementations

func (tr *TestRunner) testServerInitialization() error {
	// Verify server started correctly
	if tr.env.Server == nil {
		return fmt.Errorf("server is nil")
	}

	// Verify required tools are registered
	expectedTools := ExpectedToolNames()
	for _, toolName := range expectedTools {
		if !tr.env.Server.HasTool(toolName) {
			return fmt.Errorf("required tool %s not found", toolName)
		}
	}

	// Verify tool count
	actualCount := tr.env.Server.GetToolCount()
	if actualCount != len(expectedTools) {
		return fmt.Errorf("expected %d tools, got %d", len(expectedTools), actualCount)
	}

	return nil
}

func (tr *TestRunner) testToolRegistration() error {
	ctx := CreateTestContext()
	tools, err := tr.env.Server.ListTools(ctx)
	if err != nil {
		return fmt.Errorf("failed to list tools: %w", err)
	}

	if len(tools) == 0 {
		return fmt.Errorf("no tools registered")
	}

	// Verify each tool has proper structure
	for _, tool := range tools {
		if tool.Name == "" {
			return fmt.Errorf("tool has empty name")
		}
		if tool.Description == "" {
			return fmt.Errorf("tool %s has empty description", tool.Name)
		}
		if tool.InputSchema == nil {
			return fmt.Errorf("tool %s has nil input schema", tool.Name)
		}
	}

	return nil
}

func (tr *TestRunner) testSearchToolFlow() error {
	ctx := CreateTestContext()
	testData := NewTestData()

	args := map[string]any{
		"query": testData.SearchQuery,
		"model": "llama-3.1-sonar-small-128k-online",
	}

	result, err := tr.env.Server.ExecuteTool(ctx, "perplexity_search", args)
	if err != nil {
		return fmt.Errorf("search tool execution failed: %w", err)
	}

	if result == nil {
		return fmt.Errorf("search tool returned nil result")
	}

	if result.IsError {
		return fmt.Errorf("search tool returned error: %s", result.Content)
	}

	if result.Content == "" {
		return fmt.Errorf("search tool returned empty content")
	}

	return nil
}

func (tr *TestRunner) testChatToolFlow() error {
	ctx := CreateTestContext()
	testData := NewTestData()

	args := map[string]any{
		"messages": []map[string]string{
			{"role": "user", "content": testData.ChatMessages[0].Content},
		},
	}

	result, err := tr.env.Server.ExecuteTool(ctx, "perplexity_chat", args)
	if err != nil {
		return fmt.Errorf("chat tool execution failed: %w", err)
	}

	if result == nil || result.IsError {
		return fmt.Errorf("chat tool failed: %v", result)
	}

	return nil
}

func (tr *TestRunner) testResearchToolFlow() error {
	ctx := CreateTestContext()
	testData := NewTestData()

	args := map[string]any{
		"topic": testData.ResearchTopic,
	}

	result, err := tr.env.Server.ExecuteTool(ctx, "perplexity_research", args)
	if err != nil {
		return fmt.Errorf("research tool execution failed: %w", err)
	}

	if result == nil || result.IsError {
		return fmt.Errorf("research tool failed: %v", result)
	}

	return nil
}

func (tr *TestRunner) testInputValidation() error {
	ctx := CreateTestContext()

	// Test invalid inputs
	invalidCases := []struct {
		toolName string
		args     map[string]any
	}{
		{"perplexity_search", map[string]any{"query": ""}},
		{"perplexity_chat", map[string]any{"messages": []map[string]string{}}},
		{"perplexity_research", map[string]any{"topic": ""}},
	}

	for _, tc := range invalidCases {
		result, err := tr.env.Server.ExecuteTool(ctx, tc.toolName, tc.args)
		// Should either return error or result with IsError=true
		if err == nil && (result == nil || !result.IsError) {
			return fmt.Errorf("expected validation error for %s but none occurred", tc.toolName)
		}
	}

	return nil
}

func (tr *TestRunner) testErrorHandling() error {
	ctx := CreateTestContext()

	// Test non-existent tool
	result, err := tr.env.Server.ExecuteTool(ctx, "nonexistent_tool", map[string]any{})
	if err == nil {
		return fmt.Errorf("expected error for non-existent tool")
	}
	if result == nil || !result.IsError {
		return fmt.Errorf("expected error result for non-existent tool")
	}

	// Test API error
	errorQuery := "error_test_query"
	tr.env.SimulateAPIError(errorQuery, domain.ErrAPIError)

	args := map[string]any{"query": errorQuery}
	result, err = tr.env.Server.ExecuteTool(ctx, "perplexity_search", args)

	// Should handle API error gracefully
	if err == nil && (result == nil || !result.IsError) {
		return fmt.Errorf("expected error handling for API failure")
	}

	return nil
}

func (tr *TestRunner) testTimeoutHandling() error {
	// Configure delay that exceeds timeout
	tr.env.SimulateNetworkDelay(100 * time.Millisecond)

	// Create short timeout context
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	args := map[string]any{"query": "timeout test"}
	result, err := tr.env.Server.ExecuteTool(ctx, "perplexity_search", args)

	// Should handle timeout gracefully
	if err == nil && (result == nil || !result.IsError) {
		return fmt.Errorf("expected timeout handling")
	}

	return nil
}

func (tr *TestRunner) testConcurrentAccess() error {
	ctx := CreateTestContext()
	concurrency := 5
	errors := make(chan error, concurrency)

	// Execute multiple tools concurrently
	for i := 0; i < concurrency; i++ {
		go func(id int) {
			args := map[string]any{"query": fmt.Sprintf("concurrent test %d", id)}
			_, err := tr.env.Server.ExecuteTool(ctx, "perplexity_search", args)
			errors <- err
		}(i)
	}

	// Collect results
	for i := 0; i < concurrency; i++ {
		if err := <-errors; err != nil {
			return fmt.Errorf("concurrent execution %d failed: %w", i, err)
		}
	}

	return nil
}

func (tr *TestRunner) testCompleteWorkflow() error {
	ctx := CreateTestContext()
	testData := NewTestData()

	// Step 1: Search
	searchArgs := map[string]any{"query": testData.SearchQuery}
	result, err := tr.env.Server.ExecuteTool(ctx, "perplexity_search", searchArgs)
	if err != nil || result.IsError {
		return fmt.Errorf("workflow search step failed: %w", err)
	}

	// Step 2: Chat
	chatArgs := map[string]any{
		"messages": []map[string]string{
			{"role": "user", "content": testData.ChatMessages[0].Content},
		},
	}
	result, err = tr.env.Server.ExecuteTool(ctx, "perplexity_chat", chatArgs)
	if err != nil || result.IsError {
		return fmt.Errorf("workflow chat step failed: %w", err)
	}

	// Step 3: Research
	researchArgs := map[string]any{"topic": testData.ResearchTopic}
	result, err = tr.env.Server.ExecuteTool(ctx, "perplexity_research", researchArgs)
	if err != nil || result.IsError {
		return fmt.Errorf("workflow research step failed: %w", err)
	}

	return nil
}

// Utility methods

func (tr *TestRunner) hasFailures() bool {
	for _, result := range tr.testResults {
		if !result.Success {
			return true
		}
	}
	return false
}

func (tr *TestRunner) generateReport(t *testing.T) {
	// Implementation would generate JSON/HTML report
	t.Logf("Test report would be generated at %s", tr.config.ReportOutputPath)
}

func (tr *TestRunner) printSummary(t *testing.T) {
	totalDuration := time.Since(tr.startTime)
	passed := 0
	failed := 0
	totalAPIalls := 0

	for _, result := range tr.testResults {
		if result.Success {
			passed++
		} else {
			failed++
		}
		totalAPIalls += result.APICallCount
	}

	t.Logf("\n=== Test Summary ===")
	t.Logf("Total Tests: %d", len(tr.testResults))
	t.Logf("Passed: %d", passed)
	t.Logf("Failed: %d", failed)
	t.Logf("Success Rate: %.2f%%", float64(passed)/float64(len(tr.testResults))*100)
	t.Logf("Total Duration: %v", totalDuration)
	t.Logf("Total API Calls: %d", totalAPIalls)
	t.Logf("==================")

	if failed > 0 {
		t.Logf("\nFailed Tests:")
		for _, result := range tr.testResults {
			if !result.Success {
				t.Logf("- %s: %v", result.TestName, result.Error)
			}
		}
	}
}

// GetTestResults returns all test results for external analysis
func (tr *TestRunner) GetTestResults() []TestResult {
	return tr.testResults
}
