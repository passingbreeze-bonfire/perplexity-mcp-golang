package internal

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// Core request and response types
type SearchRequest struct {
	Query      string            `json:"query"`
	Model      string            `json:"model,omitempty"`
	SearchMode string            `json:"search_mode,omitempty"`
	MaxTokens  int               `json:"max_tokens,omitempty"`
	DateRange  string            `json:"date_range,omitempty"`
	Sources    []string          `json:"sources,omitempty"`
	Options    map[string]string `json:"options,omitempty"`
}

func (r *SearchRequest) Validate() error {
	if strings.TrimSpace(r.Query) == "" {
		return fmt.Errorf("query cannot be empty")
	}
	if len(r.Query) > 10000 {
		return fmt.Errorf("query too long: %d > 10000", len(r.Query))
	}
	if r.MaxTokens < 0 || r.MaxTokens > 128000 {
		return fmt.Errorf("invalid max_tokens: %d", r.MaxTokens)
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

// MCP-related types
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

// Perplexity API types
type APIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type APIChatRequest struct {
	Model              string       `json:"model"`
	Messages           []APIMessage `json:"messages"`
	MaxTokens          *int         `json:"max_tokens,omitempty"`
	Temperature        *float64     `json:"temperature,omitempty"`
	TopP               *float64     `json:"top_p,omitempty"`
	Stream             bool         `json:"stream"`
	SearchMode         string       `json:"search_mode,omitempty"`
	SearchDomainFilter []string     `json:"search_domain_filter,omitempty"`
	DisableSearch      *bool        `json:"disable_search,omitempty"`
	ReasoningEffort    string       `json:"reasoning_effort,omitempty"`
}

type APIChoice struct {
	Index        int        `json:"index"`
	Message      APIMessage `json:"message"`
	FinishReason string     `json:"finish_reason"`
}

type APIUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type APIChatResponse struct {
	ID        string          `json:"id"`
	Object    string          `json:"object"`
	Created   int64           `json:"created"`
	Model     string          `json:"model"`
	Choices   []APIChoice     `json:"choices"`
	Usage     APIUsage        `json:"usage"`
	Citations json.RawMessage `json:"citations,omitempty"`
	Sources   []Source        `json:"sources,omitempty"`
}

func (r *APIChatResponse) GetContent() string {
	if len(r.Choices) > 0 {
		return r.Choices[0].Message.Content
	}
	return ""
}

func (r *APIChatResponse) GetCreatedTime() time.Time {
	return time.Unix(r.Created, 0)
}

type APIError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Code    string `json:"code"`
}

type APIErrorResponse struct {
	Error struct {
		Error APIError `json:"error"`
	} `json:"error"`
}
