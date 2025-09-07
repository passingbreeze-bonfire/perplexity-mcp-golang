package tools

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
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
		ID:      "test-search-id",
		Content: "test search content",
		Model:   "sonar",
		Usage: domain.Usage{
			PromptTokens:     10,
			CompletionTokens: 20,
			TotalTokens:      30,
		},
		Citations: []domain.Citation{
			{Number: 1, URL: "https://example.com", Title: "Test Source"},
		},
		Sources: []domain.Source{
			{URL: "https://example.com", Title: "Test Source", Snippet: "Test snippet"},
		},
		Created: time.Now(),
	}, nil
}

func (m *mockSearchUseCase) ValidateRequest(request domain.SearchRequest) error {
	if m.validateFunc != nil {
		return m.validateFunc(request)
	}
	return nil
}

func TestNewSearchTool(t *testing.T) {
	useCase := &mockSearchUseCase{}
	logger := &mockLogger{}

	tool := NewSearchTool(useCase, logger)

	if tool == nil {
		t.Fatal("Expected tool to be created, got nil")
	}

	if tool.useCase != useCase {
		t.Error("Expected useCase to be set correctly")
	}

	if tool.logger != logger {
		t.Error("Expected logger to be set correctly")
	}
}

func TestSearchTool_Name(t *testing.T) {
	useCase := &mockSearchUseCase{}
	logger := &mockLogger{}

	tool := NewSearchTool(useCase, logger)

	expectedName := "perplexity_search"
	if tool.Name() != expectedName {
		t.Errorf("Expected name %s, got %s", expectedName, tool.Name())
	}
}

func TestSearchTool_Description(t *testing.T) {
	useCase := &mockSearchUseCase{}
	logger := &mockLogger{}

	tool := NewSearchTool(useCase, logger)

	description := tool.Description()
	if description == "" {
		t.Error("Expected description to be non-empty")
	}

	if !strings.Contains(description, "search") {
		t.Error("Expected description to contain 'search'")
	}
}

func TestSearchTool_InputSchema(t *testing.T) {
	useCase := &mockSearchUseCase{}
	logger := &mockLogger{}

	tool := NewSearchTool(useCase, logger)

	schema := tool.InputSchema()
	if schema == nil {
		t.Fatal("Expected schema to be non-nil")
	}

	// Check that it's an object type
	schemaType, ok := schema["type"].(string)
	if !ok || schemaType != "object" {
		t.Error("Expected schema type to be 'object'")
	}

	// Check that properties exist
	properties, ok := schema["properties"].(map[string]any)
	if !ok {
		t.Fatal("Expected properties to be a map")
	}

	// Check required query field
	queryProp, ok := properties["query"].(map[string]any)
	if !ok {
		t.Fatal("Expected query property to exist")
	}

	if queryProp["type"] != "string" {
		t.Error("Expected query type to be 'string'")
	}

	// Check required fields
	required, ok := schema["required"].([]string)
	if !ok {
		t.Fatal("Expected required to be a string slice")
	}

	if len(required) != 1 || required[0] != "query" {
		t.Error("Expected query to be the only required field")
	}
}

func TestSearchTool_Execute_Success(t *testing.T) {
	useCase := &mockSearchUseCase{}
	logger := &mockLogger{}

	tool := NewSearchTool(useCase, logger)
	ctx := context.Background()

	args := map[string]any{
		"query": "test query",
		"model": "sonar",
	}

	result, err := tool.Execute(ctx, args)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result == nil {
		t.Fatal("Expected result to be non-nil")
	}

	if result.IsError {
		t.Error("Expected result to not be an error")
	}

	if result.Content == "" {
		t.Error("Expected result content to be non-empty")
	}

	// Check that content is valid JSON
	var contentData map[string]any
	if err := json.Unmarshal([]byte(result.Content), &contentData); err != nil {
		t.Errorf("Expected content to be valid JSON, got error: %v", err)
	}

	// Check metadata
	if result.Metadata == nil {
		t.Error("Expected metadata to be non-nil")
	}

	if result.Citations == nil {
		t.Error("Expected citations to be non-nil")
	}

	if len(result.Citations) != 1 {
		t.Errorf("Expected 1 citation, got %d", len(result.Citations))
	}
}

func TestSearchTool_Execute_ValidationError(t *testing.T) {
	useCase := &mockSearchUseCase{}
	logger := &mockLogger{}

	tool := NewSearchTool(useCase, logger)
	ctx := context.Background()

	t.Run("missing query", func(t *testing.T) {
		args := map[string]any{
			"model": "sonar",
		}

		result, err := tool.Execute(ctx, args)
		if err == nil {
			t.Error("Expected error for missing query")
		}

		if result == nil || !result.IsError {
			t.Error("Expected error result")
		}
	})

	t.Run("invalid query type", func(t *testing.T) {
		args := map[string]any{
			"query": 123, // should be string
		}

		result, err := tool.Execute(ctx, args)
		if err == nil {
			t.Error("Expected error for invalid query type")
		}

		if result == nil || !result.IsError {
			t.Error("Expected error result")
		}
	})

	t.Run("invalid max_tokens type", func(t *testing.T) {
		args := map[string]any{
			"query":      "test query",
			"max_tokens": "invalid", // should be number
		}

		result, err := tool.Execute(ctx, args)
		if err == nil {
			t.Error("Expected error for invalid max_tokens type")
		}

		if result == nil || !result.IsError {
			t.Error("Expected error result")
		}
	})

	t.Run("invalid options type", func(t *testing.T) {
		args := map[string]any{
			"query":   "test query",
			"options": "invalid", // should be object
		}

		result, err := tool.Execute(ctx, args)
		if err == nil {
			t.Error("Expected error for invalid options type")
		}

		if result == nil || !result.IsError {
			t.Error("Expected error result")
		}
	})
}

func TestSearchTool_Execute_UseCaseError(t *testing.T) {
	useCase := &mockSearchUseCase{
		executeFunc: func(ctx context.Context, request domain.SearchRequest) (*domain.SearchResult, error) {
			return nil, errors.New("use case error")
		},
	}
	logger := &mockLogger{}

	tool := NewSearchTool(useCase, logger)
	ctx := context.Background()

	args := map[string]any{
		"query": "test query",
	}

	result, err := tool.Execute(ctx, args)
	if err == nil {
		t.Error("Expected error from use case")
	}

	if result == nil || !result.IsError {
		t.Error("Expected error result")
	}

	if !strings.Contains(result.Content, "Search failed") {
		t.Error("Expected error message to indicate search failure")
	}
}

func TestSearchTool_parseSearchRequest(t *testing.T) {
	useCase := &mockSearchUseCase{}
	logger := &mockLogger{}

	tool := NewSearchTool(useCase, logger)

	t.Run("valid minimal request", func(t *testing.T) {
		args := map[string]any{
			"query": "test query",
		}

		request, err := tool.parseSearchRequest(args)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if request.Query != "test query" {
			t.Errorf("Expected query 'test query', got %s", request.Query)
		}

		if request.Model != "" {
			t.Error("Expected model to be empty")
		}
	})

	t.Run("valid full request", func(t *testing.T) {
		args := map[string]any{
			"query":       "test query",
			"model":       "sonar",
			"search_mode": "web",
			"max_tokens":  float64(100),
			"options": map[string]any{
				"key1": "value1",
				"key2": "value2",
			},
		}

		request, err := tool.parseSearchRequest(args)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if request.Query != "test query" {
			t.Errorf("Expected query 'test query', got %s", request.Query)
		}

		if request.Model != "sonar" {
			t.Errorf("Expected model 'sonar', got %s", request.Model)
		}

		if request.SearchMode != "web" {
			t.Errorf("Expected search_mode 'web', got %s", request.SearchMode)
		}

		if request.MaxTokens != 100 {
			t.Errorf("Expected max_tokens 100, got %d", request.MaxTokens)
		}

		if len(request.Options) != 2 {
			t.Errorf("Expected 2 options, got %d", len(request.Options))
		}

		if request.Options["key1"] != "value1" {
			t.Error("Expected options to be parsed correctly")
		}
	})

	t.Run("domain validation error", func(t *testing.T) {
		args := map[string]any{
			"query": "", // empty query should fail domain validation
		}

		_, err := tool.parseSearchRequest(args)
		if err == nil {
			t.Error("Expected error for empty query")
		}

		if !strings.Contains(err.Error(), "validation failed") {
			t.Error("Expected validation error message")
		}
	})
}

func TestSearchTool_formatSearchResult(t *testing.T) {
	useCase := &mockSearchUseCase{}
	logger := &mockLogger{}

	tool := NewSearchTool(useCase, logger)

	result := &domain.SearchResult{
		ID:      "test-id",
		Content: "test content",
		Model:   "sonar",
		Usage: domain.Usage{
			PromptTokens:     10,
			CompletionTokens: 20,
			TotalTokens:      30,
		},
		Citations: []domain.Citation{
			{Number: 1, URL: "https://example.com", Title: "Test Source"},
		},
		Sources: []domain.Source{
			{URL: "https://example.com", Title: "Test Source", Snippet: "Test snippet"},
		},
		Created: time.Now(),
	}

	formatted, err := tool.formatSearchResult(result)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if formatted == "" {
		t.Error("Expected formatted result to be non-empty")
	}

	// Check that it's valid JSON
	var data map[string]any
	if err := json.Unmarshal([]byte(formatted), &data); err != nil {
		t.Errorf("Expected formatted result to be valid JSON, got error: %v", err)
	}

	// Check that required fields are present
	if data["id"] != result.ID {
		t.Error("Expected id to be preserved")
	}

	if data["content"] != result.Content {
		t.Error("Expected content to be preserved")
	}

	if data["model"] != result.Model {
		t.Error("Expected model to be preserved")
	}

	// Check citations and sources
	if data["citations"] == nil {
		t.Error("Expected citations to be included")
	}

	if data["sources"] == nil {
		t.Error("Expected sources to be included")
	}
}
