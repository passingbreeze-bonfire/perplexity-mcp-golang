package domain

import (
	"strings"
	"testing"
)

// TestSearchRequest_InputLengthValidation tests that search request validation handles large inputs
func TestSearchRequest_InputLengthValidation(t *testing.T) {
	tests := []struct {
		name      string
		request   SearchRequest
		wantError bool
		errorMsg  string
	}{
		{
			name: "valid query within limits",
			request: SearchRequest{
				Query:     "What is Go programming?",
				MaxTokens: 1000,
			},
			wantError: false,
		},
		{
			name: "query exceeds maximum length",
			request: SearchRequest{
				Query:     strings.Repeat("x", MaxQueryLength+1),
				MaxTokens: 1000,
			},
			wantError: true,
			errorMsg:  "query length",
		},
		{
			name: "max_tokens exceeds limit",
			request: SearchRequest{
				Query:     "valid query",
				MaxTokens: 130000,
			},
			wantError: true,
			errorMsg:  "max_tokens",
		},
		{
			name: "too many options",
			request: SearchRequest{
				Query:   "valid query",
				Options: createLargeOptionsMap(MaxOptionsCount + 1),
			},
			wantError: true,
			errorMsg:  "options count",
		},
		{
			name: "option key too long",
			request: SearchRequest{
				Query: "valid query",
				Options: map[string]string{
					strings.Repeat("k", MaxOptionKeyLength+1): "value",
				},
			},
			wantError: true,
			errorMsg:  "option key length",
		},
		{
			name: "option value too long",
			request: SearchRequest{
				Query: "valid query",
				Options: map[string]string{
					"key": strings.Repeat("v", MaxOptionValueLength+1),
				},
			},
			wantError: true,
			errorMsg:  "option value length",
		},
		{
			name: "too many sources",
			request: SearchRequest{
				Query:     "valid query",
				Sources:   createLargeSourcesSlice(MaxSourcesCount + 1),
				MaxTokens: 100,
			},
			wantError: true,
			errorMsg:  "sources count",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.request.Validate()
			if tt.wantError {
				if err == nil {
					t.Errorf("Expected error containing '%s', got nil", tt.errorMsg)
				} else if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error containing '%s', got '%v'", tt.errorMsg, err)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}
			}
		})
	}
}

// TestSearchRequest_InputSanitization tests that search inputs are properly sanitized
func TestSearchRequest_InputSanitization(t *testing.T) {
	tests := []struct {
		name      string
		request   SearchRequest
		wantError bool
	}{
		{
			name: "query with SQL injection attempt",
			request: SearchRequest{
				Query: "'; DROP TABLE users; --",
			},
			wantError: false, // Should be allowed but treated as plain text
		},
		{
			name: "query with script tags",
			request: SearchRequest{
				Query: "<script>alert('xss')</script>",
			},
			wantError: false, // Should be allowed but treated as plain text
		},
		{
			name: "query with null bytes",
			request: SearchRequest{
				Query: "test\x00query",
			},
			wantError: false, // Null bytes should be handled gracefully
		},
		{
			name: "query with control characters",
			request: SearchRequest{
				Query: "test\r\nquery\t\b",
			},
			wantError: false, // Control characters should be allowed
		},
		{
			name: "model with invalid characters",
			request: SearchRequest{
				Query: "test query",
				Model: "sonar<script>",
			},
			wantError: true, // Invalid model name
		},
		{
			name: "search mode with injection",
			request: SearchRequest{
				Query:      "test query",
				SearchMode: "web'; DROP TABLE--",
			},
			wantError: true, // Invalid search mode
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.request.Validate()
			if tt.wantError {
				if err == nil {
					t.Error("Expected validation error for malicious input, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error for sanitized input, got %v", err)
				}
			}
		})
	}
}

// TestSearchRequest_ModelValidation tests strict model validation
func TestSearchRequest_ModelValidation(t *testing.T) {
	tests := []struct {
		name      string
		model     string
		wantError bool
	}{
		{"valid sonar model", "sonar", false},
		{"valid sonar-pro model", "sonar-pro", false},
		{"valid sonar-reasoning model", "sonar-reasoning", false},
		{"valid sonar-reasoning-pro model", "sonar-reasoning-pro", false},
		{"valid sonar-deep-research model", "sonar-deep-research", false},
		{"empty model allowed", "", false},
		{"invalid model", "gpt-4", true},
		{"model with spaces", "sonar pro", true},
		{"model with special chars", "sonar@pro", true},
		{"model with uppercase", "SONAR", true},
		{"model with path traversal", "../sonar", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := SearchRequest{
				Query: "test query",
				Model: tt.model,
			}
			err := request.Validate()
			if tt.wantError {
				if err == nil {
					t.Errorf("Expected error for model '%s', got nil", tt.model)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error for model '%s', got %v", tt.model, err)
				}
			}
		})
	}
}

// TestSearchRequest_SearchModeValidation tests strict search mode validation
func TestSearchRequest_SearchModeValidation(t *testing.T) {
	tests := []struct {
		name       string
		searchMode string
		wantError  bool
	}{
		{"valid web mode", "web", false},
		{"valid academic mode", "academic", false},
		{"valid news mode", "news", false},
		{"empty mode allowed", "", false},
		{"invalid mode", "custom", true},
		{"mode with spaces", "web search", true},
		{"mode with special chars", "web@search", true},
		{"mode with uppercase", "WEB", true},
		{"mode with injection", "web'; DROP--", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := SearchRequest{
				Query:      "test query",
				SearchMode: tt.searchMode,
			}
			err := request.Validate()
			if tt.wantError {
				if err == nil {
					t.Errorf("Expected error for search mode '%s', got nil", tt.searchMode)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error for search mode '%s', got %v", tt.searchMode, err)
				}
			}
		})
	}
}

// TestSearchRequest_DateRangeValidation tests strict date range validation
func TestSearchRequest_DateRangeValidation(t *testing.T) {
	tests := []struct {
		name      string
		dateRange string
		wantError bool
	}{
		{"valid day range", "day", false},
		{"valid week range", "week", false},
		{"valid month range", "month", false},
		{"valid year range", "year", false},
		{"empty range allowed", "", false},
		{"invalid range", "decade", true},
		{"range with spaces", "last week", true},
		{"range with numbers", "7days", true},
		{"range with uppercase", "WEEK", true},
		{"range with special chars", "week@", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := SearchRequest{
				Query:     "test query",
				DateRange: tt.dateRange,
			}
			err := request.Validate()
			if tt.wantError {
				if err == nil {
					t.Errorf("Expected error for date range '%s', got nil", tt.dateRange)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error for date range '%s', got %v", tt.dateRange, err)
				}
			}
		})
	}
}

// TestSearchRequest_ConcurrentValidation tests thread safety of validation
func TestSearchRequest_ConcurrentValidation(t *testing.T) {
	request := SearchRequest{
		Query:     "concurrent test query",
		Model:     "sonar",
		MaxTokens: 1000,
		Options: map[string]string{
			"temperature": "0.7",
		},
	}

	// Run validation concurrently
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			err := request.Validate()
			if err != nil {
				t.Errorf("Concurrent validation failed: %v", err)
			}
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}

// Helper functions

func createLargeOptionsMap(size int) map[string]string {
	options := make(map[string]string, size)
	for i := 0; i < size; i++ {
		key := string(rune('a' + i%26))
		if i >= 26 {
			key = key + string(rune('0' + (i-26)%10))
		}
		options[key] = "value"
	}
	return options
}

func createLargeSourcesSlice(size int) []string {
	sources := make([]string, size)
	for i := 0; i < size; i++ {
		sources[i] = "domain" + string(rune('0'+i%10)) + ".com"
	}
	return sources
}