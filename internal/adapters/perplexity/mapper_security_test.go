package perplexity

import (
	"strings"
	"testing"

	"github.com/yourusername/perplexity-mcp-golang/internal/core/domain"
)

// mockLogger implements domain.Logger for testing
type mockLogger struct{}

func (m *mockLogger) Debug(msg string, keysAndValues ...interface{}) {}
func (m *mockLogger) Info(msg string, keysAndValues ...interface{})  {}
func (m *mockLogger) Warn(msg string, keysAndValues ...interface{})  {}
func (m *mockLogger) Error(msg string, keysAndValues ...interface{}) {}

// TestProcessSearchOptions_SecurityValidation tests that search options are validated securely
func TestProcessSearchOptions_SecurityValidation(t *testing.T) {
	tests := []struct {
		name     string
		options  map[string]string
		validate func(t *testing.T, apiReq APIChatRequest)
	}{
		{
			name: "valid temperature within bounds",
			options: map[string]string{
				"temperature": "1.0",
			},
			validate: func(t *testing.T, apiReq APIChatRequest) {
				if apiReq.Temperature == nil || *apiReq.Temperature != 1.0 {
					t.Error("Valid temperature should be set")
				}
			},
		},
		{
			name: "temperature below minimum ignored",
			options: map[string]string{
				"temperature": "-0.5",
			},
			validate: func(t *testing.T, apiReq APIChatRequest) {
				if apiReq.Temperature != nil {
					t.Error("Invalid temperature should be ignored")
				}
			},
		},
		{
			name: "temperature above maximum ignored",
			options: map[string]string{
				"temperature": "3.0",
			},
			validate: func(t *testing.T, apiReq APIChatRequest) {
				if apiReq.Temperature != nil {
					t.Error("Invalid temperature should be ignored")
				}
			},
		},
		{
			name: "valid top_p within bounds",
			options: map[string]string{
				"top_p": "0.9",
			},
			validate: func(t *testing.T, apiReq APIChatRequest) {
				if apiReq.TopP == nil || *apiReq.TopP != 0.9 {
					t.Error("Valid top_p should be set")
				}
			},
		},
		{
			name: "top_p above maximum ignored",
			options: map[string]string{
				"top_p": "1.5",
			},
			validate: func(t *testing.T, apiReq APIChatRequest) {
				if apiReq.TopP != nil {
					t.Error("Invalid top_p should be ignored")
				}
			},
		},
		{
			name: "oversized option key ignored",
			options: map[string]string{
				strings.Repeat("k", domain.MaxOptionKeyLength+1): "value",
			},
			validate: func(t *testing.T, apiReq APIChatRequest) {
				if apiReq.Temperature != nil || apiReq.TopP != nil {
					t.Error("Oversized key option should be ignored")
				}
			},
		},
		{
			name: "oversized option value ignored",
			options: map[string]string{
				"temperature": strings.Repeat("v", domain.MaxOptionValueLength+1),
			},
			validate: func(t *testing.T, apiReq APIChatRequest) {
				if apiReq.Temperature != nil {
					t.Error("Oversized value option should be ignored")
				}
			},
		},
		{
			name: "valid search domain filter",
			options: map[string]string{
				"search_domain_filter": "example.com,trusted.org",
			},
			validate: func(t *testing.T, apiReq APIChatRequest) {
				expected := []string{"example.com", "trusted.org"}
				if len(apiReq.SearchDomainFilter) != 2 ||
					apiReq.SearchDomainFilter[0] != expected[0] ||
					apiReq.SearchDomainFilter[1] != expected[1] {
					t.Errorf("Expected %v, got %v", expected, apiReq.SearchDomainFilter)
				}
			},
		},
		{
			name: "domain filter with invalid domains ignored",
			options: map[string]string{
				"search_domain_filter": strings.Repeat("x", 254) + ".com,valid.com", // First domain too long
			},
			validate: func(t *testing.T, apiReq APIChatRequest) {
				if len(apiReq.SearchDomainFilter) != 1 || apiReq.SearchDomainFilter[0] != "valid.com" {
					t.Errorf("Expected only valid domain, got %v", apiReq.SearchDomainFilter)
				}
			},
		},
		{
			name: "too many domains ignored",
			options: map[string]string{
				"search_domain_filter": strings.Repeat("x.com,", 12)[:len("x.com,")*11-1], // 11 domains
			},
			validate: func(t *testing.T, apiReq APIChatRequest) {
				if apiReq.SearchDomainFilter != nil {
					t.Error("Too many domains should be ignored")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			apiReq := APIChatRequest{}
			err := processSearchOptions(&apiReq, tt.options, &mockLogger{})
			if err != nil {
				t.Logf("processSearchOptions error (may be expected): %v", err)
			}
			tt.validate(t, apiReq)
		})
	}
}

// TestProcessSearchOptionsAdditional_SecurityValidation tests additional search options validation
func TestProcessSearchOptionsAdditional_SecurityValidation(t *testing.T) {
	tests := []struct {
		name     string
		options  map[string]string
		validate func(t *testing.T, apiReq APIChatRequest)
	}{
		{
			name: "valid search mode accepted",
			options: map[string]string{
				"search_mode": "academic",
			},
			validate: func(t *testing.T, apiReq APIChatRequest) {
				if apiReq.SearchMode != "academic" {
					t.Error("Valid search mode should be set")
				}
			},
		},
		{
			name: "invalid search mode ignored",
			options: map[string]string{
				"search_mode": "malicious_mode",
			},
			validate: func(t *testing.T, apiReq APIChatRequest) {
				if apiReq.SearchMode != "" {
					t.Error("Invalid search mode should be ignored")
				}
			},
		},
		{
			name: "valid top_p within bounds",
			options: map[string]string{
				"top_p": "0.8",
			},
			validate: func(t *testing.T, apiReq APIChatRequest) {
				if apiReq.TopP == nil || *apiReq.TopP != 0.8 {
					t.Error("Valid top_p should be set")
				}
			},
		},
		{
			name: "top_p below minimum ignored",
			options: map[string]string{
				"top_p": "-0.1",
			},
			validate: func(t *testing.T, apiReq APIChatRequest) {
				if apiReq.TopP != nil {
					t.Error("Invalid top_p should be ignored")
				}
			},
		},
		{
			name: "disable search valid boolean",
			options: map[string]string{
				"disable_search": "true",
			},
			validate: func(t *testing.T, apiReq APIChatRequest) {
				if apiReq.DisableSearch == nil || !*apiReq.DisableSearch {
					t.Error("Valid disable_search should be set")
				}
			},
		},
		{
			name: "disable search invalid value ignored",
			options: map[string]string{
				"disable_search": "maybe",
			},
			validate: func(t *testing.T, apiReq APIChatRequest) {
				if apiReq.DisableSearch != nil {
					t.Error("Invalid disable_search should be ignored")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			apiReq := APIChatRequest{}
			err := processSearchOptions(&apiReq, tt.options, &mockLogger{})
			if err != nil {
				t.Logf("processSearchOptions error (may be expected): %v", err)
			}
			tt.validate(t, apiReq)
		})
	}
}


// TestOptionProcessing_DoSPrevention tests that option processing prevents DoS attacks
func TestOptionProcessing_DoSPrevention(t *testing.T) {
	t.Run("excessive options handled gracefully", func(t *testing.T) {
		// Create an excessive number of options
		options := make(map[string]string)
		for i := 0; i < 1000; i++ {
			key := strings.Repeat("k", i%100+1)
			value := strings.Repeat("v", i%1000+1)
			options[key] = value
		}

		apiReq := APIChatRequest{}
		err := processSearchOptions(&apiReq, options, &mockLogger{})
		if err != nil {
			t.Logf("processSearchOptions error (may be expected): %v", err)
		}

		// Processing should complete without panic or excessive resource use
		// Invalid options should be filtered out
		if apiReq.Temperature != nil || apiReq.TopP != nil {
			t.Error("Invalid options should not be processed")
		}
	})

	t.Run("malformed values handled gracefully", func(t *testing.T) {
		options := map[string]string{
			"temperature":          "not-a-number",
			"top_p":                "infinity",
			"disable_search":       "not-a-bool",
			"search_domain_filter": strings.Repeat("malicious-domain,", 100),
		}

		apiReq := APIChatRequest{}
		err := processSearchOptions(&apiReq, options, &mockLogger{})
		if err != nil {
			t.Logf("processSearchOptions error (may be expected): %v", err)
		}

		// All malformed values should be ignored
		if apiReq.Temperature != nil || apiReq.TopP != nil ||
			apiReq.DisableSearch != nil || apiReq.SearchDomainFilter != nil {
			t.Error("Malformed options should be ignored")
		}
	})

	t.Run("memory exhaustion prevention", func(t *testing.T) {
		// Try to create very large option values
		largeValue := strings.Repeat("x", domain.MaxOptionValueLength*2)
		options := map[string]string{
			"temperature":          largeValue,
			"top_p":                largeValue,
			"search_domain_filter": largeValue,
		}

		apiReq := APIChatRequest{}
		err := processSearchOptions(&apiReq, options, &mockLogger{})
		if err != nil {
			t.Logf("processSearchOptions error (may be expected): %v", err)
		}

		// Large values should be ignored
		if apiReq.Temperature != nil || apiReq.TopP != nil || apiReq.SearchDomainFilter != nil {
			t.Error("Large values should be ignored to prevent memory exhaustion")
		}
	})
}

// TestSearchModeValidation tests that only valid search modes are accepted
func TestSearchModeValidation(t *testing.T) {
	validModes := []string{"web", "academic", "news"}
	invalidModes := []string{"", "invalid", "malicious", "admin", "debug"}

	for _, mode := range validModes {
		t.Run("valid mode: "+mode, func(t *testing.T) {
			options := map[string]string{"search_mode": mode}
			apiReq := APIChatRequest{}
			err := processSearchOptions(&apiReq, options, &mockLogger{})
		if err != nil {
			t.Logf("processSearchOptions error (may be expected): %v", err)
		}

			if apiReq.SearchMode != mode {
				t.Errorf("Valid search mode %s should be accepted", mode)
			}
		})
	}

	for _, mode := range invalidModes {
		t.Run("invalid mode: "+mode, func(t *testing.T) {
			options := map[string]string{"search_mode": mode}
			apiReq := APIChatRequest{}
			err := processSearchOptions(&apiReq, options, &mockLogger{})
		if err != nil {
			t.Logf("processSearchOptions error (may be expected): %v", err)
		}

			if apiReq.SearchMode != "" {
				t.Errorf("Invalid search mode %s should be rejected", mode)
			}
		})
	}
}

// TestDomainFilterSecurity tests that domain filtering is secure
func TestDomainFilterSecurity(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "normal domains",
			input:    "example.com,test.org",
			expected: []string{"example.com", "test.org"},
		},
		{
			name:     "domains with spaces",
			input:    " example.com , test.org ",
			expected: []string{"example.com", "test.org"},
		},
		{
			name:     "empty domains filtered out",
			input:    "example.com,,test.org, ,",
			expected: []string{"example.com", "test.org"},
		},
		{
			name:     "domains too long filtered out",
			input:    strings.Repeat("x", 254) + ".com,valid.com",
			expected: []string{"valid.com"},
		},
		{
			name:     "too many domains rejected",
			input:    strings.Join(make([]string, 15), ","), // 15 empty domains
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			options := map[string]string{"search_domain_filter": tt.input}
			apiReq := APIChatRequest{}
			err := processSearchOptions(&apiReq, options, &mockLogger{})
		if err != nil {
			t.Logf("processSearchOptions error (may be expected): %v", err)
		}

			if tt.expected == nil {
				if apiReq.SearchDomainFilter != nil {
					t.Error("Expected no domains to be set")
				}
			} else {
				if len(apiReq.SearchDomainFilter) != len(tt.expected) {
					t.Errorf("Expected %d domains, got %d", len(tt.expected), len(apiReq.SearchDomainFilter))
				}
				for i, expected := range tt.expected {
					if i >= len(apiReq.SearchDomainFilter) || apiReq.SearchDomainFilter[i] != expected {
						t.Errorf("Expected domain %s at index %d, got %v", expected, i, apiReq.SearchDomainFilter)
					}
				}
			}
		})
	}
}
