package domain

import (
	"fmt"
	"strings"
	"time"
)

const (
	// MaxQueryLength is the maximum allowed length for search queries
	MaxQueryLength = 10000
	// MaxOptionsCount is the maximum number of options allowed
	MaxOptionsCount = 20
	// MaxOptionKeyLength is the maximum length for option keys
	MaxOptionKeyLength = 100
	// MaxOptionValueLength is the maximum length for option values
	MaxOptionValueLength = 1000
	// MaxSourcesCount is the maximum number of source domains allowed
	MaxSourcesCount = 10
)

type SearchRequest struct {
	Query     string            `json:"query"`
	Model     string            `json:"model,omitempty"`
	SearchMode string           `json:"search_mode,omitempty"`
	MaxTokens int               `json:"max_tokens,omitempty"`
	DateRange string            `json:"date_range,omitempty"`
	Sources   []string          `json:"sources,omitempty"`
	Options   map[string]string `json:"options,omitempty"`
}

// Validate checks if the SearchRequest has valid parameters
func (r *SearchRequest) Validate() error {
	if strings.TrimSpace(r.Query) == "" {
		return ErrInvalidQuery
	}
	if len(r.Query) > MaxQueryLength {
		return fmt.Errorf("%w: query length %d exceeds maximum %d", ErrInvalidRequest, len(r.Query), MaxQueryLength)
	}
	if r.MaxTokens < 0 {
		return fmt.Errorf("%w: max_tokens cannot be negative", ErrInvalidRequest)
	}
	if r.MaxTokens > 128000 {
		return fmt.Errorf("%w: max_tokens %d exceeds maximum 128000", ErrInvalidRequest, r.MaxTokens)
	}
	
	// Validate Sonar model if specified
	if r.Model != "" {
		validSonarModels := map[string]bool{
			"sonar":               true,
			"sonar-pro":          true,
			"sonar-reasoning":    true,
			"sonar-reasoning-pro": true,
			"sonar-deep-research": true,
		}
		if !validSonarModels[r.Model] {
			return fmt.Errorf("%w: invalid model '%s', must be one of: sonar, sonar-pro, sonar-reasoning, sonar-reasoning-pro, sonar-deep-research", ErrInvalidRequest, r.Model)
		}
	}
	
	// Validate search mode if specified
	if r.SearchMode != "" {
		validSearchModes := map[string]bool{
			"web":      true,
			"academic": true,
			"news":     true,
		}
		if !validSearchModes[r.SearchMode] {
			return fmt.Errorf("%w: invalid search_mode '%s', must be one of: web, academic, news", ErrInvalidRequest, r.SearchMode)
		}
	}
	
	// Validate date range if specified
	if r.DateRange != "" {
		validDateRanges := map[string]bool{
			"day":   true,
			"week":  true,
			"month": true,
			"year":  true,
		}
		if !validDateRanges[r.DateRange] {
			return fmt.Errorf("%w: invalid date_range '%s', must be one of: day, week, month, year", ErrInvalidRequest, r.DateRange)
		}
	}
	
	// Validate sources count
	if len(r.Sources) > MaxSourcesCount {
		return fmt.Errorf("%w: sources count %d exceeds maximum %d", ErrInvalidRequest, len(r.Sources), MaxSourcesCount)
	}
	
	if len(r.Options) > MaxOptionsCount {
		return fmt.Errorf("%w: options count %d exceeds maximum %d", ErrInvalidRequest, len(r.Options), MaxOptionsCount)
	}
	for key, value := range r.Options {
		if len(key) > MaxOptionKeyLength {
			return fmt.Errorf("%w: option key length %d exceeds maximum %d", ErrInvalidRequest, len(key), MaxOptionKeyLength)
		}
		if len(value) > MaxOptionValueLength {
			return fmt.Errorf("%w: option value length %d exceeds maximum %d", ErrInvalidRequest, len(value), MaxOptionValueLength)
		}
	}
	return nil
}

type SearchResult struct {
	ID        string     `json:"id"`
	Content   string     `json:"content"`
	Model     string     `json:"model"`
	Usage     Usage      `json:"usage"`
	Citations []Citation `json:"citations,omitempty"`
	Sources   []Source   `json:"sources,omitempty"`
	Created   time.Time  `json:"created"`
}




type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type Citation struct {
	Number int    `json:"number"`
	URL    string `json:"url"`
	Title  string `json:"title"`
}

type Source struct {
	URL     string `json:"url"`
	Title   string `json:"title"`
	Snippet string `json:"snippet"`
}

type ToolInfo struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	InputSchema map[string]any `json:"inputSchema"`
}

type ToolResult struct {
	Content   string         `json:"content"`
	IsError   bool           `json:"isError"`
	Metadata  map[string]any `json:"metadata,omitempty"`
	Citations []Citation     `json:"citations,omitempty"`
}
