package main

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/yourusername/perplexity-mcp-golang/internal/infrastructure/config"
	"github.com/yourusername/perplexity-mcp-golang/internal/infrastructure/logger"
)

func TestWireDependencies(t *testing.T) {
	// Set up test environment
	os.Setenv("PERPLEXITY_API_KEY", "test-key")
	defer os.Unsetenv("PERPLEXITY_API_KEY")

	ctx := context.Background()
	cfg := config.NewConfig()
	log := logger.NewLogger("info", os.Stdout)

	// Test dependency wiring with timeout
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	deps, err := wireDependencies(ctx, cfg, log)
	if err != nil {
		t.Fatalf("Expected wireDependencies to succeed, got error: %v", err)
	}

	if deps == nil {
		t.Fatal("Expected dependencies to be non-nil")
	}

	// Verify all components are properly wired
	if deps.Config == nil {
		t.Error("Expected Config to be non-nil")
	}

	if deps.Logger == nil {
		t.Error("Expected Logger to be non-nil")
	}

	if deps.PerplexityClient == nil {
		t.Error("Expected PerplexityClient to be non-nil")
	}

	if deps.SearchUseCase == nil {
		t.Error("Expected SearchUseCase to be non-nil")
	}

	if deps.MCPServer == nil {
		t.Error("Expected MCPServer to be non-nil")
	}

	// Verify MCP server has required tools
	requiredTools := []string{"perplexity_search"}
	for _, toolName := range requiredTools {
		if !deps.MCPServer.HasTool(toolName) {
			t.Errorf("Expected tool '%s' to be registered", toolName)
		}
	}

	// Verify tool count
	expectedToolCount := 1
	actualToolCount := deps.MCPServer.GetToolCount()
	if actualToolCount != expectedToolCount {
		t.Errorf("Expected %d tools, got %d", expectedToolCount, actualToolCount)
	}
}

func TestWireDependenciesMissingAPIKey(t *testing.T) {
	// Ensure API key is not set
	os.Unsetenv("PERPLEXITY_API_KEY")

	ctx := context.Background()
	cfg := config.NewConfig()
	log := logger.NewLogger("info", os.Stdout)

	_, err := wireDependencies(ctx, cfg, log)
	if err == nil {
		t.Fatal("Expected wireDependencies to fail with missing API key")
	}

	// Should fail when creating Perplexity client
	if err.Error() == "" {
		t.Error("Expected error message to be non-empty")
	}
}

func TestValidateConfiguration(t *testing.T) {
	log := logger.NewLogger("info", os.Stdout)

	testCases := []struct {
		name          string
		apiKey        string
		model         string
		timeout       string
		logLevel      string
		shouldSucceed bool
	}{
		{
			name:          "Valid configuration",
			apiKey:        "test-key",
			model:         "llama-3.1-sonar-small-128k-online",
			timeout:       "30",
			logLevel:      "info",
			shouldSucceed: true,
		},
		{
			name:          "Missing API key",
			apiKey:        "",
			model:         "llama-3.1-sonar-small-128k-online",
			timeout:       "30",
			logLevel:      "info",
			shouldSucceed: false,
		},
		{
			name:          "Invalid log level",
			apiKey:        "test-key",
			model:         "llama-3.1-sonar-small-128k-online",
			timeout:       "30",
			logLevel:      "invalid",
			shouldSucceed: false,
		},
		{
			name:          "Invalid timeout (falls back to default)",
			apiKey:        "test-key",
			model:         "llama-3.1-sonar-small-128k-online",
			timeout:       "-1",
			logLevel:      "info",
			shouldSucceed: true, // Config handles invalid values by using defaults
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Set environment variables
			os.Setenv("PERPLEXITY_API_KEY", tc.apiKey)
			os.Setenv("PERPLEXITY_DEFAULT_MODEL", tc.model)
			os.Setenv("REQUEST_TIMEOUT_SECONDS", tc.timeout)
			os.Setenv("LOG_LEVEL", tc.logLevel)

			defer func() {
				os.Unsetenv("PERPLEXITY_API_KEY")
				os.Unsetenv("PERPLEXITY_DEFAULT_MODEL")
				os.Unsetenv("REQUEST_TIMEOUT_SECONDS")
				os.Unsetenv("LOG_LEVEL")
			}()

			cfg := config.NewConfig()
			err := validateConfiguration(cfg, log)

			if tc.shouldSucceed && err != nil {
				t.Errorf("Expected validation to succeed, got error: %v", err)
			}

			if !tc.shouldSucceed && err == nil {
				t.Error("Expected validation to fail, but it succeeded")
			}
		})
	}
}

func TestEnvironmentVariableDefaults(t *testing.T) {
	// Clear all environment variables
	os.Unsetenv("PERPLEXITY_API_KEY")
	os.Unsetenv("PERPLEXITY_DEFAULT_MODEL")
	os.Unsetenv("REQUEST_TIMEOUT_SECONDS")
	os.Unsetenv("LOG_LEVEL")

	cfg := config.NewConfig()

	// Test default values
	if cfg.GetDefaultModel() == "" {
		t.Error("Expected default model to be set")
	}

	if cfg.GetRequestTimeout() <= 0 {
		t.Error("Expected default timeout to be positive")
	}

	if cfg.GetLogLevel() == "" {
		t.Error("Expected default log level to be set")
	}

	// API key should be empty by default but that's valid for testing
	if cfg.GetPerplexityAPIKey() != "" {
		t.Error("Expected API key to be empty when not set")
	}
}