package domain

import "errors"

var (
	ErrAPIKeyMissing      = errors.New("perplexity API key not configured")
	ErrInvalidQuery       = errors.New("invalid search query")
	ErrInvalidModel       = errors.New("invalid model specified")
	ErrInvalidRequest     = errors.New("invalid request parameters")
	ErrRateLimited        = errors.New("rate limit exceeded")
	ErrAPIError           = errors.New("perplexity API error")
	ErrNetworkError       = errors.New("network connection error")
	ErrTimeout            = errors.New("request timeout")
	ErrToolNotFound       = errors.New("tool not found")
	ErrToolExecution      = errors.New("tool execution failed")
	ErrMCPProtocol        = errors.New("MCP protocol error")
	ErrConfigurationError = errors.New("configuration error")
)
