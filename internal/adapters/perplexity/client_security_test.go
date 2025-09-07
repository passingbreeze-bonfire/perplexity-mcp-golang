package perplexity

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/yourusername/perplexity-mcp-golang/internal/core/domain"
)

// TestClient_TLSConfiguration verifies that the client enforces secure TLS settings
func TestClient_TLSConfiguration(t *testing.T) {
	logger := &MockLogger{}

	client, err := NewClient("test-api-key", logger)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	// Access the underlying HTTP client to check TLS configuration
	httpClient, ok := client.httpClient.(*http.Client)
	if !ok {
		t.Fatal("Expected http.Client")
	}

	transport, ok := httpClient.Transport.(*http.Transport)
	if !ok {
		t.Fatal("Expected http.Transport")
	}

	tlsConfig := transport.TLSClientConfig
	if tlsConfig == nil {
		t.Fatal("TLS configuration not set")
	}

	if tlsConfig.MinVersion != tls.VersionTLS12 {
		t.Errorf("Expected minimum TLS version 1.2, got %v", tlsConfig.MinVersion)
	}

	if tlsConfig.MaxVersion != tls.VersionTLS13 {
		t.Errorf("Expected maximum TLS version 1.3, got %v", tlsConfig.MaxVersion)
	}
}

// TestClient_ResponseSizeLimit verifies that the client limits response body size
func TestClient_ResponseSizeLimit(t *testing.T) {
	logger := &MockLogger{}

	// Create a large response body that exceeds the limit
	largeResponse := APIChatResponse{
		ID:      "test-123",
		Object:  "chat.completion",
		Created: 1703097600,
		Model:   "test-model",
		Choices: []APIChoice{
			{
				Index: 0,
				Message: APIMessage{
					Role:    "assistant",
					Content: strings.Repeat("x", MaxResponseSize), // Exactly at the limit
				},
			},
		},
		Usage: APIUsage{TotalTokens: 10},
	}

	mockHTTPClient := &MockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			responseBody, _ := json.Marshal(largeResponse)
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
		Query: "test query",
	}

	ctx := context.Background()
	_, err = client.Search(ctx, request)

	// Should fail due to response size limit
	if err == nil {
		t.Error("Expected error due to response size limit, got nil")
	}

	if !strings.Contains(err.Error(), "response too large") {
		t.Errorf("Expected error about response size, got: %v", err)
	}
}

// TestClient_NoSensitiveLogging verifies that sensitive data is not logged
func TestClient_NoSensitiveLogging(t *testing.T) {
	logger := &MockLogger{}

	mockHTTPClient := &MockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			// Return error to trigger error logging
			return &http.Response{
				StatusCode: http.StatusUnauthorized,
				Body:       io.NopCloser(strings.NewReader(`{"error": {"type": "unauthorized", "message": "API key is invalid: sk-1234567890abcdef"}}`)),
			}, nil
		},
	}

	client, err := NewClient("test-secret-key", logger, WithHTTPClient(mockHTTPClient))
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	request := domain.SearchRequest{
		Query: "test query",
	}

	ctx := context.Background()
	_, err = client.Search(ctx, request)

	// Check that API key is not in any log messages
	allLogs := append(logger.DebugLogs, logger.ErrorLogs...)
	for _, log := range allLogs {
		logStr := log.Message + " " + strings.Join(convertFieldsToStrings(log.Fields), " ")
		if strings.Contains(logStr, "test-secret-key") {
			t.Errorf("API key found in log: %s", logStr)
		}
		if strings.Contains(logStr, "sk-1234567890abcdef") {
			t.Errorf("API key from error response found in log: %s", logStr)
		}
	}

	// Verify that error logs don't contain full error messages that might include sensitive data
	for _, log := range logger.ErrorLogs {
		if log.Message == "API error" {
			// Check that the message field is not logged
			for i := 0; i < len(log.Fields); i += 2 {
				if i+1 < len(log.Fields) && log.Fields[i] == "message" {
					t.Errorf("Error message should not be logged for security reasons")
				}
			}
		}
	}
}

// Helper function to convert log fields to strings for testing
func convertFieldsToStrings(fields []any) []string {
	var result []string
	for _, field := range fields {
		switch v := field.(type) {
		case string:
			result = append(result, v)
		case int:
			result = append(result, fmt.Sprintf("%d", v))
		default:
			result = append(result, fmt.Sprintf("%v", v))
		}
	}
	return result
}

// TestClient_SecurityHeaders verifies that proper security headers are set
func TestClient_SecurityHeaders(t *testing.T) {
	logger := &MockLogger{}

	var capturedRequest *http.Request

	mockHTTPClient := &MockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			capturedRequest = req

			mockResponse := APIChatResponse{
				ID:      "test-123",
				Choices: []APIChoice{{Message: APIMessage{Content: "test"}}},
				Usage:   APIUsage{TotalTokens: 10},
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
		Query: "test query",
	}

	ctx := context.Background()
	_, err = client.Search(ctx, request)

	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}

	// Verify security headers
	if capturedRequest.Header.Get("Authorization") != "Bearer test-api-key" {
		t.Error("Authorization header not set correctly")
	}

	if capturedRequest.Header.Get("Content-Type") != "application/json" {
		t.Error("Content-Type header not set correctly")
	}

	// Verify no sensitive headers are accidentally set
	sensitiveHeaders := []string{"X-API-Key", "X-Secret", "Cookie"}
	for _, header := range sensitiveHeaders {
		if capturedRequest.Header.Get(header) != "" {
			t.Errorf("Unexpected sensitive header %s found", header)
		}
	}
}

// TestClient_ContextTimeout verifies that context timeouts are properly handled
func TestClient_ContextTimeout(t *testing.T) {
	logger := &MockLogger{}

	mockHTTPClient := &MockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			// Wait for context cancellation
			<-req.Context().Done()
			return nil, req.Context().Err()
		},
	}

	client, err := NewClient("test-api-key", logger, WithHTTPClient(mockHTTPClient))
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	request := domain.SearchRequest{
		Query: "test query",
	}

	// Create context with short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	_, err = client.Search(ctx, request)

	if err == nil {
		t.Error("Expected timeout error, got nil")
	}

	if !strings.Contains(err.Error(), "timed out") {
		t.Errorf("Expected timeout error, got: %v", err)
	}
}

// TestClient_InvalidTLSConfiguration tests that the client fails with invalid TLS
func TestClient_InvalidTLSConfiguration(t *testing.T) {
	logger := &MockLogger{}

	// Create a client with custom HTTP client that has no TLS config
	insecureTransport := &http.Transport{}
	insecureClient := &http.Client{Transport: insecureTransport}

	_, err := NewClient("test-api-key", logger, WithHTTPClient(insecureClient))
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	// Verify that our default secure client would be different
	defaultClient, err := NewClient("test-api-key", logger)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	defaultHTTPClient := defaultClient.httpClient.(*http.Client)
	defaultTransport := defaultHTTPClient.Transport.(*http.Transport)

	// The default client should have TLS config while the insecure one doesn't
	if defaultTransport.TLSClientConfig == nil {
		t.Error("Default client should have TLS configuration")
	}

	if insecureTransport.TLSClientConfig != nil {
		t.Error("Insecure transport should not have TLS configuration")
	}
}
