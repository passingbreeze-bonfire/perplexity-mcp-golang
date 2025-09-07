package perplexity

import (
	"reflect"
	"testing"
	"time"

	"github.com/yourusername/perplexity-mcp-golang/internal/core/domain"
)


func TestSearchRequestToAPI(t *testing.T) {
	tests := []struct {
		name     string
		request  domain.SearchRequest
		expected APIChatRequest
	}{
		{
			name: "basic search request",
			request: domain.SearchRequest{
				Query: "What is Go programming?",
				Model: "sonar",
			},
			expected: APIChatRequest{
				Model: "sonar",
				Messages: []APIMessage{
					{
						Role:    "user",
						Content: "What is Go programming?",
					},
				},
			},
		},
		{
			name: "search request with search mode",
			request: domain.SearchRequest{
				Query:      "AI research papers",
				Model:      "sonar-pro",
				SearchMode: "academic",
			},
			expected: APIChatRequest{
				Model:      "sonar-pro",
				SearchMode: "academic",
				Messages: []APIMessage{
					{
						Role:    "user",
						Content: "AI research papers",
					},
				},
			},
		},
		{
			name: "search request with max tokens",
			request: domain.SearchRequest{
				Query:     "quantum computing",
				Model:     "sonar-reasoning",
				MaxTokens: 1000,
			},
			expected: APIChatRequest{
				Model:     "sonar-reasoning",
				MaxTokens: intPtr(1000),
				Messages: []APIMessage{
					{
						Role:    "user",
						Content: "quantum computing",
					},
				},
			},
		},
		{
			name: "search request with sources",
			request: domain.SearchRequest{
				Query:   "latest news",
				Model:   "sonar",
				Sources: []string{"bbc.com", "cnn.com"},
			},
			expected: APIChatRequest{
				Model: "sonar",
				Messages: []APIMessage{
					{
						Role:    "user",
						Content: "latest news",
					},
				},
				SearchDomainFilter: []string{"bbc.com", "cnn.com"},
			},
		},
		{
			name: "search request with options",
			request: domain.SearchRequest{
				Query:     "AI research",
				Model:     "sonar-pro",
				MaxTokens: 500,
				Options: map[string]string{
					"temperature":          "0.7",
					"top_p":                "0.9",
					"disable_search":       "false",
					"search_domain_filter": "arxiv.org,nature.com",
				},
			},
			expected: APIChatRequest{
				Model:     "sonar-pro",
				MaxTokens: intPtr(500),
				Messages: []APIMessage{
					{
						Role:    "user",
						Content: "AI research",
					},
				},
				Temperature:        float64Ptr(0.7),
				TopP:               float64Ptr(0.9),
				DisableSearch:      boolPtr(false),
				SearchDomainFilter: []string{"arxiv.org", "nature.com"},
			},
		},
		{
			name: "search request with invalid temperature option",
			request: domain.SearchRequest{
				Query: "test query",
				Model: "sonar",
				Options: map[string]string{
					"temperature": "3.0", // Invalid: too high
				},
			},
			expected: APIChatRequest{
				Model: "sonar",
				Messages: []APIMessage{
					{
						Role:    "user",
						Content: "test query",
					},
				},
				// Temperature should not be set due to invalid value
			},
		},
		{
			name: "search request with invalid top_p option",
			request: domain.SearchRequest{
				Query: "test query",
				Model: "sonar",
				Options: map[string]string{
					"top_p": "1.5", // Invalid: too high
				},
			},
			expected: APIChatRequest{
				Model: "sonar",
				Messages: []APIMessage{
					{
						Role:    "user",
						Content: "test query",
					},
				},
				// TopP should not be set due to invalid value
			},
		},
		{
			name: "search request with reasoning models",
			request: domain.SearchRequest{
				Query:     "complex problem solving",
				Model:     "sonar-reasoning-pro",
				MaxTokens: 2000,
			},
			expected: APIChatRequest{
				Model:     "sonar-reasoning-pro",
				MaxTokens: intPtr(2000),
				Messages: []APIMessage{
					{
						Role:    "user",
						Content: "complex problem solving",
					},
				},
			},
		},
		{
			name: "search request with deep research model",
			request: domain.SearchRequest{
				Query:     "comprehensive analysis of climate change",
				Model:     "sonar-deep-research",
				MaxTokens: 5000,
			},
			expected: APIChatRequest{
				Model:     "sonar-deep-research",
				MaxTokens: intPtr(5000),
				Messages: []APIMessage{
					{
						Role:    "user",
						Content: "comprehensive analysis of climate change",
					},
				},
			},
		},
	}

	logger := &mockLogger{}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SearchRequestToAPI(tt.request, logger)
			if err != nil {
				t.Fatalf("SearchRequestToAPI() error = %v", err)
			}
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("SearchRequestToAPI() = %+v, want %+v", got, tt.expected)
			}
		})
	}
}

func TestAPIResponseToSearchResult(t *testing.T) {
	created := time.Unix(1703097600, 0)

	tests := []struct {
		name     string
		apiResp  APIChatResponse
		expected domain.SearchResult
	}{
		{
			name: "basic search response",
			apiResp: APIChatResponse{
				ID:      "search-123",
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
					CompletionTokens: 20,
					TotalTokens:      30,
				},
			},
			expected: domain.SearchResult{
				ID:      "search-123",
				Content: "Go is a programming language developed by Google.",
				Model:   "sonar",
				Usage: domain.Usage{
					PromptTokens:     10,
					CompletionTokens: 20,
					TotalTokens:      30,
				},
				Created: created,
			},
		},
		{
			name: "search response with citations",
			apiResp: APIChatResponse{
				ID:      "search-456",
				Created: 1703097600,
				Model:   "sonar-pro",
				Choices: []APIChoice{
					{
						Message: APIMessage{
							Content: "According to recent research...",
						},
					},
				},
				Usage: APIUsage{
					TotalTokens: 50,
				},
				Citations: []APICitation{
					{
						Number: 1,
						URL:    "https://example.com/article1",
						Title:  "Research Article 1",
					},
					{
						Number: 2,
						URL:    "https://example.com/article2",
						Title:  "Research Article 2",
					},
				},
			},
			expected: domain.SearchResult{
				ID:      "search-456",
				Content: "According to recent research...",
				Model:   "sonar-pro",
				Usage: domain.Usage{
					TotalTokens: 50,
				},
				Citations: []domain.Citation{
					{
						Number: 1,
						URL:    "https://example.com/article1",
						Title:  "Research Article 1",
					},
					{
						Number: 2,
						URL:    "https://example.com/article2",
						Title:  "Research Article 2",
					},
				},
				Created: created,
			},
		},
		{
			name: "search response with sources",
			apiResp: APIChatResponse{
				ID:      "search-789",
				Created: 1703097600,
				Model:   "sonar-reasoning",
				Choices: []APIChoice{
					{
						Message: APIMessage{
							Content: "Based on multiple sources...",
						},
					},
				},
				Usage: APIUsage{
					PromptTokens:     25,
					CompletionTokens: 75,
					TotalTokens:      100,
				},
				Sources: []APISource{
					{
						URL:     "https://source1.com",
						Title:   "Source 1",
						Snippet: "This is a snippet from source 1",
					},
					{
						URL:     "https://source2.com",
						Title:   "Source 2",
						Snippet: "This is a snippet from source 2",
					},
				},
			},
			expected: domain.SearchResult{
				ID:      "search-789",
				Content: "Based on multiple sources...",
				Model:   "sonar-reasoning",
				Usage: domain.Usage{
					PromptTokens:     25,
					CompletionTokens: 75,
					TotalTokens:      100,
				},
				Sources: []domain.Source{
					{
						URL:     "https://source1.com",
						Title:   "Source 1",
						Snippet: "This is a snippet from source 1",
					},
					{
						URL:     "https://source2.com",
						Title:   "Source 2",
						Snippet: "This is a snippet from source 2",
					},
				},
				Created: created,
			},
		},
		{
			name: "search response with citations and sources",
			apiResp: APIChatResponse{
				ID:      "search-complex",
				Created: 1703097600,
				Model:   "sonar-deep-research",
				Choices: []APIChoice{
					{
						Message: APIMessage{
							Content: "Comprehensive research findings...",
						},
					},
				},
				Usage: APIUsage{
					PromptTokens:     100,
					CompletionTokens: 500,
					TotalTokens:      600,
				},
				Citations: []APICitation{
					{
						Number: 1,
						URL:    "https://research.com/paper1",
						Title:  "Research Paper 1",
					},
				},
				Sources: []APISource{
					{
						URL:     "https://research.com/data",
						Title:   "Research Data",
						Snippet: "Supporting data for the research",
					},
				},
			},
			expected: domain.SearchResult{
				ID:      "search-complex",
				Content: "Comprehensive research findings...",
				Model:   "sonar-deep-research",
				Usage: domain.Usage{
					PromptTokens:     100,
					CompletionTokens: 500,
					TotalTokens:      600,
				},
				Citations: []domain.Citation{
					{
						Number: 1,
						URL:    "https://research.com/paper1",
						Title:  "Research Paper 1",
					},
				},
				Sources: []domain.Source{
					{
						URL:     "https://research.com/data",
						Title:   "Research Data",
						Snippet: "Supporting data for the research",
					},
				},
				Created: created,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := APIResponseToSearchResult(tt.apiResp)
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("APIResponseToSearchResult() = %+v, want %+v", got, tt.expected)
			}
		})
	}
}

func TestProcessSearchOptions(t *testing.T) {
	tests := []struct {
		name     string
		options  map[string]string
		expected APIChatRequest
	}{
		{
			name: "valid temperature and top_p",
			options: map[string]string{
				"temperature": "0.5",
				"top_p":       "0.8",
			},
			expected: APIChatRequest{
				Temperature: float64Ptr(0.5),
				TopP:        float64Ptr(0.8),
			},
		},
		{
			name: "invalid temperature ignored",
			options: map[string]string{
				"temperature": "invalid",
				"top_p":       "0.9",
			},
			expected: APIChatRequest{
				TopP: float64Ptr(0.9),
			},
		},
		{
			name: "disable_search option",
			options: map[string]string{
				"disable_search": "true",
			},
			expected: APIChatRequest{
				DisableSearch: boolPtr(true),
			},
		},
		{
			name: "search_domain_filter with valid domains",
			options: map[string]string{
				"search_domain_filter": "example.com, test.org, demo.net",
			},
			expected: APIChatRequest{
				SearchDomainFilter: []string{"example.com", "test.org", "demo.net"},
			},
		},
		{
			name: "search_domain_filter with too many domains",
			options: map[string]string{
				"search_domain_filter": "d1.com,d2.com,d3.com,d4.com,d5.com,d6.com,d7.com,d8.com,d9.com,d10.com,d11.com",
			},
			expected: APIChatRequest{
				// Should be ignored as it exceeds 10 domains
			},
		},
		{
			name: "mixed valid and invalid options",
			options: map[string]string{
				"temperature":    "0.7",
				"invalid_option": "value",
				"top_p":          "0.95",
			},
			expected: APIChatRequest{
				Temperature: float64Ptr(0.7),
				TopP:        float64Ptr(0.95),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			apiReq := APIChatRequest{}
			err := processSearchOptions(&apiReq, tt.options, &mockLogger{})
			if err != nil {
				t.Logf("processSearchOptions error: %v", err)
			}
			
			// Compare only the fields that should be set
			if tt.expected.Temperature != nil {
				if apiReq.Temperature == nil || *apiReq.Temperature != *tt.expected.Temperature {
					t.Errorf("Temperature = %v, want %v", apiReq.Temperature, tt.expected.Temperature)
				}
			}
			if tt.expected.TopP != nil {
				if apiReq.TopP == nil || *apiReq.TopP != *tt.expected.TopP {
					t.Errorf("TopP = %v, want %v", apiReq.TopP, tt.expected.TopP)
				}
			}
			if tt.expected.DisableSearch != nil {
				if apiReq.DisableSearch == nil || *apiReq.DisableSearch != *tt.expected.DisableSearch {
					t.Errorf("DisableSearch = %v, want %v", apiReq.DisableSearch, tt.expected.DisableSearch)
				}
			}
			if len(tt.expected.SearchDomainFilter) > 0 {
				if !reflect.DeepEqual(apiReq.SearchDomainFilter, tt.expected.SearchDomainFilter) {
					t.Errorf("SearchDomainFilter = %v, want %v", apiReq.SearchDomainFilter, tt.expected.SearchDomainFilter)
				}
			}
		})
	}
}

// Helper functions for pointer values in tests
func intPtr(v int) *int {
	return &v
}

func float64Ptr(v float64) *float64 {
	return &v
}

func boolPtr(v bool) *bool {
	return &v
}