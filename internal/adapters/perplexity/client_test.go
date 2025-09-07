package perplexity

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/yourusername/perplexity-mcp-golang/internal/core/domain"
)

// MockHTTPClient implements HTTPClient for testing
type MockHTTPClient struct {
	DoFunc func(req *http.Request) (*http.Response, error)
}

func (m *MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return m.DoFunc(req)
}

// MockLogger implements domain.Logger for testing
type MockLogger struct {
	InfoLogs  []LogEntry
	ErrorLogs []LogEntry
	DebugLogs []LogEntry
	WarnLogs  []LogEntry
}

type LogEntry struct {
	Message string
	Fields  []any
}

func (m *MockLogger) Info(msg string, fields ...any) {
	m.InfoLogs = append(m.InfoLogs, LogEntry{Message: msg, Fields: fields})
}

func (m *MockLogger) Error(msg string, fields ...any) {
	m.ErrorLogs = append(m.ErrorLogs, LogEntry{Message: msg, Fields: fields})
}

func (m *MockLogger) Debug(msg string, fields ...any) {
	m.DebugLogs = append(m.DebugLogs, LogEntry{Message: msg, Fields: fields})
}

func (m *MockLogger) Warn(msg string, fields ...any) {
	m.WarnLogs = append(m.WarnLogs, LogEntry{Message: msg, Fields: fields})
}

func TestNewClient(t *testing.T) {
	logger := &MockLogger{}

	tests := []struct {
		name      string
		apiKey    string
		wantError error
	}{
		{
			name:      "valid API key",
			apiKey:    "test-api-key",
			wantError: nil,
		},
		{
			name:      "empty API key",
			apiKey:    "",
			wantError: domain.ErrAPIKeyMissing,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(tt.apiKey, logger)
			if tt.wantError != nil {
				if err == nil {
					t.Errorf("NewClient() expected error %v, got nil", tt.wantError)
					return
				}
				if !errors.Is(err, tt.wantError) {
					t.Errorf("NewClient() error = %v, want %v", err, tt.wantError)
				}
				return
			}

			if err != nil {
				t.Errorf("NewClient() unexpected error = %v", err)
				return
			}

			if client == nil {
				t.Error("NewClient() returned nil client")
			}
		})
	}
}

func TestClient_Search_Success(t *testing.T) {
	logger := &MockLogger{}

	// Mock successful API response
	mockResponse := APIChatResponse{
		ID:      "search-test-123",
		Object:  "chat.completion",
		Created: 1703097600,
		Model:   "sonar",
		Choices: []APIChoice{
			{
				Index: 0,
				Message: APIMessage{
					Role:    "assistant",
					Content: "Go is a programming language developed by Google.",
				},
				FinishReason: "stop",
			},
		},
		Usage: APIUsage{
			PromptTokens:     10,
			CompletionTokens: 15,
			TotalTokens:      25,
		},
		Citations: []APICitation{
			{
				Number: 1,
				URL:    "https://golang.org",
				Title:  "The Go Programming Language",
			},
		},
	}

	mockHTTPClient := &MockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			// Verify request details
			if req.Method != http.MethodPost {
				t.Errorf("Expected POST request, got %s", req.Method)
			}
			if req.Header.Get("Authorization") != "Bearer test-api-key" {
				t.Errorf("Expected Authorization header with Bearer token")
			}
			if req.Header.Get("Content-Type") != "application/json" {
				t.Errorf("Expected Content-Type application/json")
			}

			// Return mock response
			responseBody, _ := json.Marshal(mockResponse)
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewReader(responseBody)),
			}, nil
		},
	}

	client, err := NewClient("test-api-key", logger, WithHTTPClient(mockHTTPClient))
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	request := domain.SearchRequest{
		Query: "What is Go programming?",
		Model: "sonar",
	}

	ctx := context.Background()
	result, err := client.Search(ctx, request)

	if err != nil {
		t.Errorf("Search() error = %v", err)
		return
	}

	if result == nil {
		t.Fatal("Search() returned nil result")
	}

	if result.ID != "search-test-123" {
		t.Errorf("Expected ID %s, got %s", "search-test-123", result.ID)
	}

	if result.Content != "Go is a programming language developed by Google." {
		t.Errorf("Expected content %s, got %s", "Go is a programming language developed by Google.", result.Content)
	}

	if len(result.Citations) != 1 {
		t.Errorf("Expected 1 citation, got %d", len(result.Citations))
	}
}

func TestClient_Search_ValidationError(t *testing.T) {
	logger := &MockLogger{}
	mockHTTPClient := &MockHTTPClient{}

	client, err := NewClient("test-api-key", logger, WithHTTPClient(mockHTTPClient))
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	request := domain.SearchRequest{
		Query: "", // Invalid: empty query
	}

	ctx := context.Background()
	_, err = client.Search(ctx, request)

	if err == nil {
		t.Error("Search() expected validation error, got nil")
		return
	}

	if !errors.Is(err, domain.ErrInvalidRequest) {
		t.Errorf("Search() error = %v, want %v", err, domain.ErrInvalidRequest)
	}
}

func TestClient_Search_WithSearchMode(t *testing.T) {
	logger := &MockLogger{}

	mockResponse := APIChatResponse{
		ID:      "search-academic-123",
		Object:  "chat.completion",
		Created: 1703097600,
		Model:   "sonar-pro",
		Choices: []APIChoice{
			{
				Message: APIMessage{
					Content: "Academic search result",
				},
			},
		},
		Usage: APIUsage{
			TotalTokens: 30,
		},
	}

	mockHTTPClient := &MockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			// Verify request includes search mode
			var apiReq APIChatRequest
			body, _ := io.ReadAll(req.Body)
			json.Unmarshal(body, &apiReq)

			if apiReq.SearchMode != "academic" {
				t.Errorf("Expected search_mode 'academic', got %s", apiReq.SearchMode)
			}

			responseBody, _ := json.Marshal(mockResponse)
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewReader(responseBody)),
			}, nil
		},
	}

	client, err := NewClient("test-api-key", logger, WithHTTPClient(mockHTTPClient))
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	request := domain.SearchRequest{
		Query:      "quantum computing research",
		Model:      "sonar-pro",
		SearchMode: "academic",
	}

	ctx := context.Background()
	result, err := client.Search(ctx, request)

	if err != nil {
		t.Errorf("Search() error = %v", err)
	}

	if result == nil {
		t.Fatal("Search() returned nil result")
	}
}

func TestClient_Search_WithDateRange(t *testing.T) {
	logger := &MockLogger{}

	mockResponse := APIChatResponse{
		ID:      "search-date-123",
		Model:   "sonar",
		Choices: []APIChoice{{Message: APIMessage{Content: "Recent results"}}},
		Usage:   APIUsage{TotalTokens: 20},
	}

	mockHTTPClient := &MockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			var apiReq APIChatRequest
			body, _ := io.ReadAll(req.Body)
			json.Unmarshal(body, &apiReq)

			// Date range is handled by the API differently
			// We can't check it directly in the request structure

			responseBody, _ := json.Marshal(mockResponse)
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewReader(responseBody)),
			}, nil
		},
	}

	client, err := NewClient("test-api-key", logger, WithHTTPClient(mockHTTPClient))
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	request := domain.SearchRequest{
		Query:     "recent news",
		DateRange: "week",
	}

	ctx := context.Background()
	_, err = client.Search(ctx, request)

	if err != nil {
		t.Errorf("Search() error = %v", err)
	}
}

func TestClient_Search_WithSources(t *testing.T) {
	logger := &MockLogger{}

	mockResponse := APIChatResponse{
		ID:      "search-sources-123",
		Model:   "sonar",
		Choices: []APIChoice{{Message: APIMessage{Content: "Filtered results"}}},
		Usage:   APIUsage{TotalTokens: 25},
		Sources: []APISource{
			{
				URL:     "https://example.com/article",
				Title:   "Example Article",
				Snippet: "This is an example snippet",
			},
		},
	}

	mockHTTPClient := &MockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			var apiReq APIChatRequest
			body, _ := io.ReadAll(req.Body)
			json.Unmarshal(body, &apiReq)

			if len(apiReq.SearchDomainFilter) != 2 {
				t.Errorf("Expected 2 search domain filters, got %d", len(apiReq.SearchDomainFilter))
			}

			responseBody, _ := json.Marshal(mockResponse)
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewReader(responseBody)),
			}, nil
		},
	}

	client, err := NewClient("test-api-key", logger, WithHTTPClient(mockHTTPClient))
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	request := domain.SearchRequest{
		Query:   "specific topic",
		Sources: []string{"example.com", "test.org"},
	}

	ctx := context.Background()
	result, err := client.Search(ctx, request)

	if err != nil {
		t.Errorf("Search() error = %v", err)
	}

	if result != nil && len(result.Sources) != 1 {
		t.Errorf("Expected 1 source in result, got %d", len(result.Sources))
	}
}

func TestClient_ErrorHandling(t *testing.T) {
	logger := &MockLogger{}

	tests := []struct {
		name          string
		statusCode    int
		responseBody  string
		expectedError error
		setupResponse func() string
	}{
		{
			name:          "unauthorized",
			statusCode:    http.StatusUnauthorized,
			expectedError: domain.ErrAPIKeyMissing,
			setupResponse: func() string {
				return `{"error": {"type": "unauthorized", "message": "Invalid API key"}}`
			},
		},
		{
			name:          "bad request",
			statusCode:    http.StatusBadRequest,
			expectedError: domain.ErrInvalidRequest,
			setupResponse: func() string {
				return `{"error": {"type": "invalid_request", "message": "Invalid model specified"}}`
			},
		},
		{
			name:          "rate limited",
			statusCode:    http.StatusTooManyRequests,
			expectedError: domain.ErrRateLimited,
			setupResponse: func() string {
				return `{"error": {"type": "rate_limit_exceeded", "message": "Too many requests"}}`
			},
		},
		{
			name:          "server error",
			statusCode:    http.StatusInternalServerError,
			expectedError: domain.ErrAPIError,
			setupResponse: func() string {
				return `{"error": {"type": "server_error", "message": "Internal server error"}}`
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockHTTPClient := &MockHTTPClient{
				DoFunc: func(req *http.Request) (*http.Response, error) {
					return &http.Response{
						StatusCode: tt.statusCode,
						Body:       io.NopCloser(strings.NewReader(tt.setupResponse())),
					}, nil
				},
			}

			client, err := NewClient("test-api-key", logger, WithHTTPClient(mockHTTPClient))
			if err != nil {
				t.Fatalf("NewClient() error = %v", err)
			}

			request := domain.SearchRequest{
				Query: "test query",
				Model: "sonar",
			}

			ctx := context.Background()
			_, err = client.Search(ctx, request)

			if err == nil {
				t.Error("Expected error, got nil")
				return
			}

			if !errors.Is(err, tt.expectedError) {
				t.Errorf("Expected error %v, got %v", tt.expectedError, err)
			}
		})
	}
}

func TestClient_NetworkError(t *testing.T) {
	logger := &MockLogger{}

	mockHTTPClient := &MockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return nil, errors.New("network connection failed")
		},
	}

	client, err := NewClient("test-api-key", logger, WithHTTPClient(mockHTTPClient))
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	request := domain.SearchRequest{
		Query: "test query",
		Model: "sonar",
	}

	ctx := context.Background()
	_, err = client.Search(ctx, request)

	if err == nil {
		t.Error("Expected network error, got nil")
		return
	}

	if !errors.Is(err, domain.ErrNetworkError) {
		t.Errorf("Expected network error, got %v", err)
	}
}

func TestClient_Timeout(t *testing.T) {
	logger := &MockLogger{}

	mockHTTPClient := &MockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			// Simulate timeout by checking context
			<-req.Context().Done()
			return nil, req.Context().Err()
		},
	}

	client, err := NewClient("test-api-key", logger, WithHTTPClient(mockHTTPClient), WithTimeout(10*time.Millisecond))
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	request := domain.SearchRequest{
		Query: "test query",
		Model: "sonar",
	}

	ctx := context.Background()
	_, err = client.Search(ctx, request)

	if err == nil {
		t.Error("Expected timeout error, got nil")
		return
	}

	if !errors.Is(err, domain.ErrTimeout) {
		t.Errorf("Expected timeout error, got %v", err)
	}
}

func TestClient_DefaultModel(t *testing.T) {
	logger := &MockLogger{}

	mockResponse := APIChatResponse{
		ID:      "test-123",
		Created: 1703097600,
		Model:   DefaultModel,
		Choices: []APIChoice{{Message: APIMessage{Content: "response"}}},
		Usage:   APIUsage{TotalTokens: 10},
	}

	mockHTTPClient := &MockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			// Verify that default model was set
			var apiReq APIChatRequest
			body, _ := io.ReadAll(req.Body)
			json.Unmarshal(body, &apiReq)

			if apiReq.Model != DefaultModel {
				t.Errorf("Expected model %s, got %s", DefaultModel, apiReq.Model)
			}

			responseBody, _ := json.Marshal(mockResponse)
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewReader(responseBody)),
			}, nil
		},
	}

	client, err := NewClient("test-api-key", logger, WithHTTPClient(mockHTTPClient))
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	// Test with empty model - should use default
	request := domain.SearchRequest{
		Query: "test query",
		Model: "", // Empty model
	}

	ctx := context.Background()
	_, err = client.Search(ctx, request)

	if err != nil {
		t.Errorf("Search() error = %v", err)
	}
}

func TestClient_ReasoningModels(t *testing.T) {
	logger := &MockLogger{}

	tests := []struct {
		name  string
		model string
	}{
		{
			name:  "sonar-reasoning model",
			model: "sonar-reasoning",
		},
		{
			name:  "sonar-reasoning-pro model",
			model: "sonar-reasoning-pro",
		},
		{
			name:  "sonar-deep-research model",
			model: "sonar-deep-research",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockResponse := APIChatResponse{
				ID:      "reasoning-test-" + tt.model,
				Model:   tt.model,
				Choices: []APIChoice{{Message: APIMessage{Content: "Reasoning response"}}},
				Usage:   APIUsage{TotalTokens: 50},
			}

			mockHTTPClient := &MockHTTPClient{
				DoFunc: func(req *http.Request) (*http.Response, error) {
					var apiReq APIChatRequest
					body, _ := io.ReadAll(req.Body)
					json.Unmarshal(body, &apiReq)

					if apiReq.Model != tt.model {
						t.Errorf("Expected model %s, got %s", tt.model, apiReq.Model)
					}

					responseBody, _ := json.Marshal(mockResponse)
					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       io.NopCloser(bytes.NewReader(responseBody)),
					}, nil
				},
			}

			client, err := NewClient("test-api-key", logger, WithHTTPClient(mockHTTPClient))
			if err != nil {
				t.Fatalf("NewClient() error = %v", err)
			}

			request := domain.SearchRequest{
				Query: "complex reasoning query",
				Model: tt.model,
			}

			ctx := context.Background()
			result, err := client.Search(ctx, request)

			if err != nil {
				t.Errorf("Search() with %s error = %v", tt.model, err)
			}

			if result != nil && result.Model != tt.model {
				t.Errorf("Expected model %s in result, got %s", tt.model, result.Model)
			}
		})
	}
}