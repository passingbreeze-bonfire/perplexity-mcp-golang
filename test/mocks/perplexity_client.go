package mocks

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/yourusername/perplexity-mcp-golang/internal/core/domain"
)

// MockPerplexityClient provides a mock implementation of domain.PerplexityClient
// for reliable testing without external API dependencies
type MockPerplexityClient struct {
	mu              sync.RWMutex
	searchResponses map[string]*domain.SearchResult
	callHistory     []MockCall
	errors          map[string]error
	delay           time.Duration
	callCount       int
}

// MockCall represents a recorded API call for verification
type MockCall struct {
	Method    string
	Query     string
	Timestamp time.Time
	Args      interface{}
}

// NewMockPerplexityClient creates a new mock client with sensible defaults
func NewMockPerplexityClient() *MockPerplexityClient {
	client := &MockPerplexityClient{
		searchResponses: make(map[string]*domain.SearchResult),
		callHistory:     make([]MockCall, 0),
		errors:          make(map[string]error),
		delay:           0,
	}

	// Set up default responses for common test cases
	client.setupDefaultResponses()
	return client
}

// Search implements domain.PerplexityClient interface
func (m *MockPerplexityClient) Search(ctx context.Context, request domain.SearchRequest) (*domain.SearchResult, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Record call
	m.callCount++
	m.callHistory = append(m.callHistory, MockCall{
		Method:    "Search",
		Query:     request.Query,
		Timestamp: time.Now(),
		Args:      request,
	})

	// Simulate network delay if configured
	if m.delay > 0 {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(m.delay):
		}
	}

	// Check for configured error
	if err, exists := m.errors[request.Query]; exists {
		return nil, err
	}

	// Return configured response or default
	if response, exists := m.searchResponses[request.Query]; exists {
		// Clone response to avoid data races
		clone := *response
		return &clone, nil
	}

	// Generate default response
	return &domain.SearchResult{
		ID:      fmt.Sprintf("search_%d_%d", m.callCount, time.Now().Unix()),
		Content: fmt.Sprintf("Mock search result for query: %s", request.Query),
		Model:   getModelOrDefault(request.Model),
		Usage: domain.Usage{
			PromptTokens:     len(request.Query) / 4, // Rough estimation
			CompletionTokens: 100,
			TotalTokens:      len(request.Query)/4 + 100,
		},
		Citations: []domain.Citation{
			{Number: 1, URL: "https://example.com/source1", Title: "Mock Source 1"},
		},
		Sources: []domain.Source{
			{URL: "https://example.com/source1", Title: "Mock Source 1", Snippet: "Mock snippet"},
		},
		Created: time.Now(),
	}, nil
}



// Mock configuration methods for testing

// SetSearchResponse configures a specific response for a search query
func (m *MockPerplexityClient) SetSearchResponse(query string, response *domain.SearchResult) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.searchResponses[query] = response
}


// SetError configures an error to be returned for a specific query/topic
func (m *MockPerplexityClient) SetError(queryOrTopic string, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.errors[queryOrTopic] = err
}

// SetDelay configures a delay for all API calls to simulate network latency
func (m *MockPerplexityClient) SetDelay(delay time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.delay = delay
}

// GetCallHistory returns the history of all API calls for verification
func (m *MockPerplexityClient) GetCallHistory() []MockCall {
	m.mu.RLock()
	defer m.mu.RUnlock()
	// Return a copy to avoid data races
	history := make([]MockCall, len(m.callHistory))
	copy(history, m.callHistory)
	return history
}

// GetCallCount returns the total number of API calls made
func (m *MockPerplexityClient) GetCallCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.callCount
}

// Reset clears all call history and configured responses
func (m *MockPerplexityClient) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.searchResponses = make(map[string]*domain.SearchResult)
	m.callHistory = make([]MockCall, 0)
	m.errors = make(map[string]error)
	m.callCount = 0
	m.delay = 0
}

// FindCalls returns calls matching the given method and query pattern
func (m *MockPerplexityClient) FindCalls(method string, queryPattern string) []MockCall {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	var matches []MockCall
	for _, call := range m.callHistory {
		if (method == "" || call.Method == method) &&
			(queryPattern == "" || strings.Contains(call.Query, queryPattern)) {
			matches = append(matches, call)
		}
	}
	return matches
}

// setupDefaultResponses configures common responses for typical test scenarios
func (m *MockPerplexityClient) setupDefaultResponses() {
	// Default search responses
	m.searchResponses["test query"] = &domain.SearchResult{
		ID:      "default_search_1",
		Content: "This is a default search result for testing purposes.",
		Model:   "sonar",
		Usage:   domain.Usage{PromptTokens: 10, CompletionTokens: 50, TotalTokens: 60},
		Created: time.Now(),
	}
}

// Helper functions

func getModelOrDefault(model string) string {
	if model == "" {
		return "sonar"
	}
	return model
}