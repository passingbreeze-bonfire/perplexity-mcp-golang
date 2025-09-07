package perplexity

import "time"

// APIMessage represents a message in the Perplexity API format
type APIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// APIChatRequest represents the request payload for Perplexity chat completions API
type APIChatRequest struct {
	Model              string       `json:"model"`
	Messages           []APIMessage `json:"messages"`
	MaxTokens          *int         `json:"max_tokens,omitempty"`
	Temperature        *float64     `json:"temperature,omitempty"`
	TopP               *float64     `json:"top_p,omitempty"`
	Stream             bool         `json:"stream,omitempty"`
	SearchMode         string       `json:"search_mode,omitempty"`
	DisableSearch      *bool        `json:"disable_search,omitempty"`
	SearchDomainFilter []string     `json:"search_domain_filter,omitempty"`
	ReasoningEffort    string       `json:"reasoning_effort,omitempty"`
}

// APIUsage represents usage statistics from the API response
type APIUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// APICitation represents a citation in the API response
type APICitation struct {
	Number int    `json:"number"`
	URL    string `json:"url"`
	Title  string `json:"title"`
}

// APISource represents a search source in the API response
type APISource struct {
	URL     string `json:"url"`
	Title   string `json:"title"`
	Snippet string `json:"snippet"`
}

// APIChoice represents a completion choice in the API response
type APIChoice struct {
	Index        int        `json:"index"`
	Message      APIMessage `json:"message"`
	FinishReason string     `json:"finish_reason"`
}

// APIChatResponse represents the response from Perplexity chat completions API
type APIChatResponse struct {
	ID      string      `json:"id"`
	Object  string      `json:"object"`
	Created int64       `json:"created"`
	Model   string      `json:"model"`
	Choices []APIChoice `json:"choices"`
	Usage   APIUsage    `json:"usage"`
	// Optional fields for search results
	Citations []APICitation `json:"citations,omitempty"`
	Sources   []APISource   `json:"sources,omitempty"`
}

// APIError represents an error response from the Perplexity API
type APIError struct {
	Error struct {
		Type    string `json:"type"`
		Message string `json:"message"`
		Code    string `json:"code,omitempty"`
	} `json:"error"`
}

// APIErrorResponse represents the full error response structure
type APIErrorResponse struct {
	Error APIError `json:"error"`
}

// GetCreatedTime returns the created timestamp as a time.Time
func (r *APIChatResponse) GetCreatedTime() time.Time {
	return time.Unix(r.Created, 0)
}

// GetContent returns the content from the first choice message
func (r *APIChatResponse) GetContent() string {
	if len(r.Choices) > 0 {
		return r.Choices[0].Message.Content
	}
	return ""
}
