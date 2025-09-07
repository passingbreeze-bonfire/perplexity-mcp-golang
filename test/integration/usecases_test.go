package integration

import (
	"strings"
	"testing"

	"github.com/yourusername/perplexity-mcp-golang/internal/core/domain"
)

// TestSearchUseCaseIntegration tests the search use case end-to-end
func TestSearchUseCaseIntegration(t *testing.T) {
	env := NewTestEnvironment(t)
	defer env.Reset()

	ctx := CreateTestContext()
	testData := NewTestData()

	// Test valid search request
	request := domain.SearchRequest{
		Query:      testData.SearchQuery,
		Model:      "sonar",
		SearchMode: "web",
		MaxTokens:  1000,
		Options: map[string]string{
			"temperature": "0.7",
		},
	}

	result, err := env.SearchUseCase.Execute(ctx, request)
	if err != nil {
		t.Fatalf("Search use case failed: %v", err)
	}

	// Validate result structure
	if result == nil {
		t.Fatal("Search result is nil")
	}
	if result.ID == "" {
		t.Error("Search result missing ID")
	}
	if result.Content == "" {
		t.Error("Search result missing content")
	}
	if result.Model == "" {
		t.Error("Search result missing model")
	}
	if result.Usage.TotalTokens == 0 {
		t.Error("Search result missing token usage")
	}
	if result.Created.IsZero() {
		t.Error("Search result missing creation timestamp")
	}

	// Verify API call was made
	env.AssertAPICallMade(t, "Search", testData.SearchQuery)

	// Verify no errors
	env.AssertNoErrors(t)
}

// TestSearchUseCaseValidation tests input validation in search use case
func TestSearchUseCaseValidation(t *testing.T) {
	env := NewTestEnvironment(t)
	defer env.Reset()

	ctx := CreateTestContext()
	testData := NewTestData()

	testCases := []struct {
		name    string
		request domain.SearchRequest
		wantErr bool
		errType string
	}{
		{
			name: "empty query",
			request: domain.SearchRequest{
				Query: testData.EmptyQuery,
			},
			wantErr: true,
			errType: "invalid query",
		},
		{
			name: "query too long",
			request: domain.SearchRequest{
				Query: testData.LongQuery,
			},
			wantErr: true,
			errType: "exceeds maximum",
		},
		{
			name: "negative max tokens",
			request: domain.SearchRequest{
				Query:     testData.SearchQuery,
				MaxTokens: -1,
			},
			wantErr: true,
			errType: "cannot be negative",
		},
		{
			name: "max tokens too high",
			request: domain.SearchRequest{
				Query:     testData.SearchQuery,
				MaxTokens: 150000,
			},
			wantErr: true,
			errType: "exceeds maximum",
		},
		{
			name: "valid request",
			request: domain.SearchRequest{
				Query:     testData.SearchQuery,
				MaxTokens: 1000,
			},
			wantErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := env.SearchUseCase.Execute(ctx, tc.request)

			if tc.wantErr {
				if err == nil {
					t.Error("Expected error but none occurred")
				} else if !strings.Contains(err.Error(), tc.errType) {
					t.Errorf("Expected error containing %q, got %q", tc.errType, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}