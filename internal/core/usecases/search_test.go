package usecases

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/yourusername/perplexity-mcp-golang/internal/core/domain"
)

// MockPerplexityClient is a mock implementation of domain.PerplexityClient
type MockPerplexityClient struct {
	SearchFunc func(ctx context.Context, request domain.SearchRequest) (*domain.SearchResult, error)
}

func (m *MockPerplexityClient) Search(ctx context.Context, request domain.SearchRequest) (*domain.SearchResult, error) {
	if m.SearchFunc != nil {
		return m.SearchFunc(ctx, request)
	}
	return nil, errors.New("search not implemented")
}

// MockConfigProvider is a mock implementation of domain.ConfigProvider
type MockConfigProvider struct {
	APIKey         string
	DefaultModel   string
	RequestTimeout int
	LogLevel       string
}

func (m *MockConfigProvider) GetPerplexityAPIKey() string {
	return m.APIKey
}

func (m *MockConfigProvider) GetDefaultModel() string {
	return m.DefaultModel
}

func (m *MockConfigProvider) GetRequestTimeout() int {
	return m.RequestTimeout
}

func (m *MockConfigProvider) GetLogLevel() string {
	return m.LogLevel
}

// MockLogger is a mock implementation of domain.Logger
type MockLogger struct {
	InfoCalls  []LogCall
	ErrorCalls []LogCall
	DebugCalls []LogCall
	WarnCalls  []LogCall
}

type LogCall struct {
	Message string
	Fields  []any
}

func (m *MockLogger) Info(msg string, fields ...any) {
	m.InfoCalls = append(m.InfoCalls, LogCall{Message: msg, Fields: fields})
}

func (m *MockLogger) Error(msg string, fields ...any) {
	m.ErrorCalls = append(m.ErrorCalls, LogCall{Message: msg, Fields: fields})
}

func (m *MockLogger) Debug(msg string, fields ...any) {
	m.DebugCalls = append(m.DebugCalls, LogCall{Message: msg, Fields: fields})
}

func (m *MockLogger) Warn(msg string, fields ...any) {
	m.WarnCalls = append(m.WarnCalls, LogCall{Message: msg, Fields: fields})
}

func TestNewSearchUseCase(t *testing.T) {
	mockClient := &MockPerplexityClient{}
	mockConfig := &MockConfigProvider{
		DefaultModel:   "test-model",
		RequestTimeout: 30,
	}
	mockLogger := &MockLogger{}

	useCase := NewSearchUseCase(mockClient, mockConfig, mockLogger)

	if useCase == nil {
		t.Fatal("Expected non-nil use case")
	}

	if useCase.client != mockClient {
		t.Error("Expected client to be set correctly")
	}

	if useCase.config != mockConfig {
		t.Error("Expected config to be set correctly")
	}

	if useCase.logger != mockLogger {
		t.Error("Expected logger to be set correctly")
	}

	expectedTimeout := time.Duration(30) * time.Second
	if useCase.requestTimeout != expectedTimeout {
		t.Errorf("Expected timeout %v, got %v", expectedTimeout, useCase.requestTimeout)
	}
}

func TestSearchUseCase_Execute_Success(t *testing.T) {
	expectedResult := &domain.SearchResult{
		ID:      "test-id",
		Content: "test content",
		Model:   "test-model",
		Usage: domain.Usage{
			TotalTokens: 100,
		},
		Created: time.Now(),
	}

	mockClient := &MockPerplexityClient{
		SearchFunc: func(ctx context.Context, request domain.SearchRequest) (*domain.SearchResult, error) {
			return expectedResult, nil
		},
	}

	mockConfig := &MockConfigProvider{
		DefaultModel:   "sonar",
		RequestTimeout: 30,
	}

	mockLogger := &MockLogger{}

	useCase := NewSearchUseCase(mockClient, mockConfig, mockLogger)

	request := domain.SearchRequest{
		Query: "test query",
		Model: "sonar",
	}

	ctx := context.Background()
	result, err := useCase.Execute(ctx, request)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if result != expectedResult {
		t.Errorf("Expected result %+v, got %+v", expectedResult, result)
	}

	// Check that info was logged
	if len(mockLogger.InfoCalls) == 0 {
		t.Error("Expected info logs to be called")
	}
}

func TestSearchUseCase_Execute_WithDefaultModel(t *testing.T) {
	expectedResult := &domain.SearchResult{
		ID:      "test-id",
		Content: "test content",
		Model:   "sonar",
		Usage: domain.Usage{
			TotalTokens: 100,
		},
		Created: time.Now(),
	}

	var capturedRequest domain.SearchRequest

	mockClient := &MockPerplexityClient{
		SearchFunc: func(ctx context.Context, request domain.SearchRequest) (*domain.SearchResult, error) {
			capturedRequest = request
			return expectedResult, nil
		},
	}

	mockConfig := &MockConfigProvider{
		DefaultModel:   "sonar",
		RequestTimeout: 30,
	}

	mockLogger := &MockLogger{}

	useCase := NewSearchUseCase(mockClient, mockConfig, mockLogger)

	request := domain.SearchRequest{
		Query: "test query",
		// Model intentionally left empty to test default behavior
	}

	ctx := context.Background()
	_, err := useCase.Execute(ctx, request)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if capturedRequest.Model != "sonar" {
		t.Errorf("Expected default model to be applied, got: %s", capturedRequest.Model)
	}
}

func TestSearchUseCase_Execute_ValidationError(t *testing.T) {
	mockClient := &MockPerplexityClient{}
	mockConfig := &MockConfigProvider{
		DefaultModel:   "default-model",
		RequestTimeout: 30,
	}
	mockLogger := &MockLogger{}

	useCase := NewSearchUseCase(mockClient, mockConfig, mockLogger)

	// Invalid request with empty query
	request := domain.SearchRequest{
		Query: "", // Invalid empty query
	}

	ctx := context.Background()
	result, err := useCase.Execute(ctx, request)

	if err == nil {
		t.Fatal("Expected validation error, got nil")
	}

	if result != nil {
		t.Error("Expected nil result on validation error")
	}

	// Check that error was logged
	if len(mockLogger.ErrorCalls) == 0 {
		t.Error("Expected error logs to be called")
	}
}

func TestSearchUseCase_Execute_ClientError(t *testing.T) {
	mockClient := &MockPerplexityClient{
		SearchFunc: func(ctx context.Context, request domain.SearchRequest) (*domain.SearchResult, error) {
			return nil, domain.ErrAPIError
		},
	}

	mockConfig := &MockConfigProvider{
		DefaultModel:   "default-model",
		RequestTimeout: 30,
	}

	mockLogger := &MockLogger{}

	useCase := NewSearchUseCase(mockClient, mockConfig, mockLogger)

	request := domain.SearchRequest{
		Query: "test query",
	}

	ctx := context.Background()
	result, err := useCase.Execute(ctx, request)

	if err == nil {
		t.Fatal("Expected client error, got nil")
	}

	if result != nil {
		t.Error("Expected nil result on client error")
	}

	if !errors.Is(err, domain.ErrAPIError) {
		t.Errorf("Expected wrapped API error, got: %v", err)
	}

	// Check that error was logged
	if len(mockLogger.ErrorCalls) == 0 {
		t.Error("Expected error logs to be called")
	}
}

func TestSearchUseCase_Execute_ContextTimeout(t *testing.T) {
	mockClient := &MockPerplexityClient{
		SearchFunc: func(ctx context.Context, request domain.SearchRequest) (*domain.SearchResult, error) {
			// Check that context has a deadline
			if _, hasDeadline := ctx.Deadline(); !hasDeadline {
				t.Error("Expected context to have a deadline")
			}
			return &domain.SearchResult{}, nil
		},
	}

	mockConfig := &MockConfigProvider{
		DefaultModel:   "default-model",
		RequestTimeout: 30,
	}

	mockLogger := &MockLogger{}

	useCase := NewSearchUseCase(mockClient, mockConfig, mockLogger)

	request := domain.SearchRequest{
		Query: "test query",
	}

	// Context without deadline - should be enhanced by use case
	ctx := context.Background()
	_, err := useCase.Execute(ctx, request)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
}

func TestSearchUseCase_ValidateRequest_Success(t *testing.T) {
	mockClient := &MockPerplexityClient{}
	mockConfig := &MockConfigProvider{}
	mockLogger := &MockLogger{}

	useCase := NewSearchUseCase(mockClient, mockConfig, mockLogger)

	request := domain.SearchRequest{
		Query: "valid query",
	}

	err := useCase.ValidateRequest(request)

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Check that debug was logged
	if len(mockLogger.DebugCalls) == 0 {
		t.Error("Expected debug logs to be called")
	}
}

func TestSearchUseCase_ValidateRequest_Error(t *testing.T) {
	mockClient := &MockPerplexityClient{}
	mockConfig := &MockConfigProvider{}
	mockLogger := &MockLogger{}

	useCase := NewSearchUseCase(mockClient, mockConfig, mockLogger)

	request := domain.SearchRequest{
		Query: "", // Invalid empty query
	}

	err := useCase.ValidateRequest(request)

	if err == nil {
		t.Error("Expected validation error, got nil")
	}

	if !errors.Is(err, domain.ErrInvalidQuery) {
		t.Errorf("Expected invalid query error, got: %v", err)
	}
}
