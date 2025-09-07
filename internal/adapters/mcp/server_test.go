package mcp

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/yourusername/perplexity-mcp-golang/internal/core/domain"
)

// mockLogger implements domain.Logger for testing
type mockLogger struct {
	infoLogs  []string
	errorLogs []string
	debugLogs []string
	warnLogs  []string
}

func (m *mockLogger) Info(msg string, fields ...any) {
	m.infoLogs = append(m.infoLogs, msg)
}

func (m *mockLogger) Error(msg string, fields ...any) {
	m.errorLogs = append(m.errorLogs, msg)
}

func (m *mockLogger) Debug(msg string, fields ...any) {
	m.debugLogs = append(m.debugLogs, msg)
}

func (m *mockLogger) Warn(msg string, fields ...any) {
	m.warnLogs = append(m.warnLogs, msg)
}

// mockTool implements domain.Tool for testing
type mockTool struct {
	name        string
	description string
	schema      map[string]any
	executeFunc func(ctx context.Context, args map[string]any) (*domain.ToolResult, error)
}

func (m *mockTool) Name() string {
	return m.name
}

func (m *mockTool) Description() string {
	return m.description
}

func (m *mockTool) InputSchema() map[string]any {
	return m.schema
}

func (m *mockTool) Execute(ctx context.Context, args map[string]any) (*domain.ToolResult, error) {
	if m.executeFunc != nil {
		return m.executeFunc(ctx, args)
	}
	return &domain.ToolResult{
		Content: "mock result",
		IsError: false,
	}, nil
}

// mockSearchUseCase implements SearchUseCaseInterface for testing
type mockSearchUseCase struct {
	executeFunc  func(ctx context.Context, request domain.SearchRequest) (*domain.SearchResult, error)
	validateFunc func(request domain.SearchRequest) error
}

func (m *mockSearchUseCase) Execute(ctx context.Context, request domain.SearchRequest) (*domain.SearchResult, error) {
	if m.executeFunc != nil {
		return m.executeFunc(ctx, request)
	}
	return &domain.SearchResult{
		ID:      "test-search-result",
		Content: "test search content",
		Model:   "test-model",
	}, nil
}

func (m *mockSearchUseCase) ValidateRequest(request domain.SearchRequest) error {
	if m.validateFunc != nil {
		return m.validateFunc(request)
	}
	return nil
}


func TestNewServer(t *testing.T) {
	logger := &mockLogger{}
	searchUC := &mockSearchUseCase{}

	server := NewServer(logger, searchUC)

	if server == nil {
		t.Fatal("Expected server to be created, got nil")
	}

	if server.tools == nil {
		t.Fatal("Expected tools map to be initialized")
	}

	if server.logger != logger {
		t.Error("Expected logger to be set correctly")
	}
}

func TestServer_RegisterTool(t *testing.T) {
	logger := &mockLogger{}
	searchUC := &mockSearchUseCase{}

	server := NewServer(logger, searchUC)

	t.Run("successful registration", func(t *testing.T) {
		tool := &mockTool{
			name:        "test_tool",
			description: "A test tool",
			schema:      map[string]any{"type": "object"},
		}

		err := server.RegisterTool("test_tool", tool)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if !server.HasTool("test_tool") {
			t.Error("Expected tool to be registered")
		}
	})

	t.Run("empty name error", func(t *testing.T) {
		tool := &mockTool{name: "test_tool"}

		err := server.RegisterTool("", tool)
		if err == nil {
			t.Error("Expected error for empty tool name")
		}
	})

	t.Run("nil tool error", func(t *testing.T) {
		err := server.RegisterTool("test_tool", nil)
		if err == nil {
			t.Error("Expected error for nil tool")
		}
	})

	t.Run("name mismatch error", func(t *testing.T) {
		tool := &mockTool{name: "different_name"}

		err := server.RegisterTool("test_tool", tool)
		if err == nil {
			t.Error("Expected error for name mismatch")
		}
	})
}

func TestServer_ExecuteTool(t *testing.T) {
	logger := &mockLogger{}
	searchUC := &mockSearchUseCase{}

	server := NewServer(logger, searchUC)
	ctx := context.Background()

	t.Run("successful execution", func(t *testing.T) {
		tool := &mockTool{
			name: "test_tool",
			executeFunc: func(ctx context.Context, args map[string]any) (*domain.ToolResult, error) {
				return &domain.ToolResult{
					Content: "success",
					IsError: false,
				}, nil
			},
		}

		err := server.RegisterTool("test_tool", tool)
		if err != nil {
			t.Fatalf("Failed to register tool: %v", err)
		}

		result, err := server.ExecuteTool(ctx, "test_tool", map[string]any{})
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if result.Content != "success" {
			t.Errorf("Expected content 'success', got %s", result.Content)
		}

		if result.IsError {
			t.Error("Expected IsError to be false")
		}
	})

	t.Run("tool not found", func(t *testing.T) {
		result, err := server.ExecuteTool(ctx, "nonexistent_tool", map[string]any{})
		if err == nil {
			t.Error("Expected error for nonexistent tool")
		}

		if result == nil {
			t.Error("Expected result to be non-nil even on error")
		}

		if result != nil && !result.IsError {
			t.Error("Expected result to have IsError true")
		}
	})

	t.Run("tool execution error", func(t *testing.T) {
		tool := &mockTool{
			name: "failing_tool",
			executeFunc: func(ctx context.Context, args map[string]any) (*domain.ToolResult, error) {
				return nil, errors.New("execution failed")
			},
		}

		err := server.RegisterTool("failing_tool", tool)
		if err != nil {
			t.Fatalf("Failed to register tool: %v", err)
		}

		result, err := server.ExecuteTool(ctx, "failing_tool", map[string]any{})
		if err == nil {
			t.Error("Expected error for failing tool")
		}

		if result == nil {
			t.Error("Expected result to be non-nil even on error")
		}

		if result != nil && !result.IsError {
			t.Error("Expected result to have IsError true")
		}
	})

	t.Run("context timeout", func(t *testing.T) {
		tool := &mockTool{
			name: "slow_tool",
			executeFunc: func(ctx context.Context, args map[string]any) (*domain.ToolResult, error) {
				select {
				case <-time.After(100 * time.Millisecond):
					return &domain.ToolResult{Content: "completed"}, nil
				case <-ctx.Done():
					return nil, ctx.Err()
				}
			},
		}

		err := server.RegisterTool("slow_tool", tool)
		if err != nil {
			t.Fatalf("Failed to register tool: %v", err)
		}

		// Create context with very short timeout
		shortCtx, cancel := context.WithTimeout(ctx, 10*time.Millisecond)
		defer cancel()

		result, err := server.ExecuteTool(shortCtx, "slow_tool", map[string]any{})
		if err == nil {
			t.Error("Expected timeout error")
		}

		if result != nil && !result.IsError {
			t.Error("Expected result to have IsError true on timeout")
		}
	})
}

func TestServer_ListTools(t *testing.T) {
	logger := &mockLogger{}
	searchUC := &mockSearchUseCase{}

	server := NewServer(logger, searchUC)
	ctx := context.Background()

	t.Run("list default tools", func(t *testing.T) {
		tools, err := server.ListTools(ctx)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		// Should have the 1 default tool registered by NewServer
		expectedCount := 1
		if len(tools) != expectedCount {
			t.Errorf("Expected %d tools, got %d", expectedCount, len(tools))
		}

		// Check that required tools are present
		toolNames := make(map[string]bool)
		for _, tool := range tools {
			toolNames[tool.Name] = true
		}

		requiredTools := []string{"perplexity_search"}
		for _, requiredTool := range requiredTools {
			if !toolNames[requiredTool] {
				t.Errorf("Expected tool %s to be in the list", requiredTool)
			}
		}
	})

	t.Run("list with additional tool", func(t *testing.T) {
		tool := &mockTool{
			name:        "additional_tool",
			description: "An additional test tool",
			schema:      map[string]any{"type": "object"},
		}

		err := server.RegisterTool("additional_tool", tool)
		if err != nil {
			t.Fatalf("Failed to register additional tool: %v", err)
		}

		tools, err := server.ListTools(ctx)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		// Should now have 2 tools (1 default + 1 additional)
		expectedCount := 2
		if len(tools) != expectedCount {
			t.Errorf("Expected %d tools, got %d", expectedCount, len(tools))
		}
	})
}

func TestServer_Start(t *testing.T) {
	logger := &mockLogger{}
	searchUC := &mockSearchUseCase{}

	server := NewServer(logger, searchUC)
	ctx := context.Background()

	t.Run("successful start", func(t *testing.T) {
		err := server.Start(ctx)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		// Check that info logs were written
		if len(logger.infoLogs) == 0 {
			t.Error("Expected info logs to be written during start")
		}
	})
}

func TestServer_GetToolCount(t *testing.T) {
	logger := &mockLogger{}
	searchUC := &mockSearchUseCase{}

	server := NewServer(logger, searchUC)

	// Should have 1 default tool
	expectedCount := 1
	actualCount := server.GetToolCount()
	if actualCount != expectedCount {
		t.Errorf("Expected tool count %d, got %d", expectedCount, actualCount)
	}
}

func TestServer_HasTool(t *testing.T) {
	logger := &mockLogger{}
	searchUC := &mockSearchUseCase{}

	server := NewServer(logger, searchUC)

	t.Run("has default tools", func(t *testing.T) {
		requiredTools := []string{"perplexity_search"}
		for _, toolName := range requiredTools {
			if !server.HasTool(toolName) {
				t.Errorf("Expected server to have tool %s", toolName)
			}
		}
	})

	t.Run("does not have nonexistent tool", func(t *testing.T) {
		if server.HasTool("nonexistent_tool") {
			t.Error("Expected server to not have nonexistent tool")
		}
	})
}
