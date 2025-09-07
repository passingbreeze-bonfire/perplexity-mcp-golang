package mocks

import (
	"context"
	"testing"
	"time"

	"github.com/yourusername/perplexity-mcp-golang/internal/core/domain"
)

// TestMockPerplexityClient tests basic functionality of the mock client
func TestMockPerplexityClient(t *testing.T) {
	client := NewMockPerplexityClient()
	ctx := context.Background()

	t.Run("Search", func(t *testing.T) {
		request := domain.SearchRequest{
			Query: "test query",
			Model: "test-model",
		}

		result, err := client.Search(ctx, request)
		if err != nil {
			t.Fatalf("Search failed: %v", err)
		}

		if result == nil {
			t.Fatal("Search result is nil")
		}

		if result.Content == "" {
			t.Error("Search result has empty content")
		}
	})


	t.Run("CallHistory", func(t *testing.T) {
		history := client.GetCallHistory()
		if len(history) != 1 {
			t.Errorf("Expected 1 call in history, got %d", len(history))
		}

		// Verify call method
		if len(history) > 0 && history[0].Method != "Search" {
			t.Errorf("Expected method Search, got %s", history[0].Method)
		}
	})
}

// TestMockConfig tests the mock configuration provider
func TestMockConfig(t *testing.T) {
	config := NewMockConfig()

	t.Run("DefaultValues", func(t *testing.T) {
		if config.GetPerplexityAPIKey() == "" {
			t.Error("API key should have default value")
		}

		if config.GetDefaultModel() == "" {
			t.Error("Default model should have value")
		}

		if config.GetRequestTimeout() <= 0 {
			t.Error("Request timeout should be positive")
		}

		if config.GetLogLevel() == "" {
			t.Error("Log level should have default value")
		}
	})

	t.Run("ConfigurationChanges", func(t *testing.T) {
		config.SetPerplexityAPIKey("new-key")
		config.SetDefaultModel("new-model")
		config.SetRequestTimeout(45)
		config.SetLogLevel("debug")

		if config.GetPerplexityAPIKey() != "new-key" {
			t.Error("API key not updated")
		}

		if config.GetDefaultModel() != "new-model" {
			t.Error("Default model not updated")
		}

		if config.GetRequestTimeout() != 45 {
			t.Error("Request timeout not updated")
		}

		if config.GetLogLevel() != "debug" {
			t.Error("Log level not updated")
		}
	})

	t.Run("Validation", func(t *testing.T) {
		// Test valid config
		if err := config.Validate(); err != nil {
			t.Errorf("Valid config failed validation: %v", err)
		}

		// Test invalid config
		config.SetPerplexityAPIKey("")
		if err := config.Validate(); err == nil {
			t.Error("Empty API key should fail validation")
		}
	})
}

// TestMockLogger tests the mock logger functionality
func TestMockLogger(t *testing.T) {
	logger := NewMockLogger()

	t.Run("BasicLogging", func(t *testing.T) {
		logger.Clear()
		logger.SetLevel("debug") // Ensure all levels are logged
		
		logger.Info("test info message", "key", "value")
		logger.Error("test error message", "error", "test")
		logger.Debug("test debug message")
		logger.Warn("test warn message")

		entries := logger.GetEntries()
		if len(entries) != 4 {
			t.Errorf("Expected 4 log entries, got %d", len(entries))
		}
	})

	t.Run("LogFiltering", func(t *testing.T) {
		logger.Clear()
		logger.SetLevel("warn")

		logger.Debug("debug message") // Should be filtered
		logger.Info("info message")   // Should be filtered
		logger.Warn("warn message")   // Should be logged
		logger.Error("error message") // Should be logged

		entries := logger.GetEntries()
		if len(entries) != 2 {
			t.Errorf("Expected 2 log entries with warn level, got %d", len(entries))
		}
	})

	t.Run("LogSearch", func(t *testing.T) {
		logger.Clear()
		logger.SetLevel("debug") // Ensure logs are captured
		
		logger.Info("unique test message")
		logger.Info("another message")

		entries := logger.GetEntriesWithMessage("unique")
		if len(entries) != 1 {
			t.Errorf("Expected 1 entry with 'unique', got %d", len(entries))
		}

		if !logger.HasEntry("info", "unique") {
			t.Error("Should have info entry with 'unique'")
		}
	})
}

// TestMockConfiguration tests configurable mock behavior
func TestMockConfiguration(t *testing.T) {
	client := NewMockPerplexityClient()
	ctx := context.Background()

	t.Run("ConfiguredResponse", func(t *testing.T) {
		customResponse := &domain.SearchResult{
			ID:      "custom-id",
			Content: "Custom response content",
			Model:   "custom-model",
		}

		client.SetSearchResponse("custom query", customResponse)

		request := domain.SearchRequest{Query: "custom query"}
		result, err := client.Search(ctx, request)

		if err != nil {
			t.Fatalf("Search failed: %v", err)
		}

		if result.ID != "custom-id" {
			t.Errorf("Expected custom ID, got %s", result.ID)
		}

		if result.Content != "Custom response content" {
			t.Errorf("Expected custom content, got %s", result.Content)
		}
	})

	t.Run("ConfiguredError", func(t *testing.T) {
		client.Reset()
		client.SetError("error query", domain.ErrAPIError)

		request := domain.SearchRequest{Query: "error query"}
		_, err := client.Search(ctx, request)

		if err != domain.ErrAPIError {
			t.Errorf("Expected API error, got %v", err)
		}
	})

	t.Run("ConfiguredDelay", func(t *testing.T) {
		client.Reset()
		client.SetDelay(10 * time.Millisecond)

		start := time.Now()
		request := domain.SearchRequest{Query: "delay query"}
		_, err := client.Search(ctx, request)
		duration := time.Since(start)

		if err != nil {
			t.Fatalf("Search failed: %v", err)
		}

		if duration < 10*time.Millisecond {
			t.Error("Expected delay was not applied")
		}
	})
}