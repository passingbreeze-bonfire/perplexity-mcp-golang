package perplexity

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/yourusername/perplexity-mcp-golang/internal/core/domain"
)

const (
	// BaseURL is the base URL for the Perplexity API
	BaseURL = "https://api.perplexity.ai"
	// ChatCompletionsEndpoint is the endpoint for chat completions
	ChatCompletionsEndpoint = "/chat/completions"
	// DefaultTimeout is the default request timeout
	DefaultTimeout = 30 * time.Second
	// DefaultModel is the default Sonar model
	DefaultModel = "sonar"
	// MaxResponseSize is the maximum allowed response body size (10MB)
	MaxResponseSize = 10 * 1024 * 1024
)

// HTTPClient defines the interface for HTTP operations
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// Client implements the domain.PerplexityClient interface
type Client struct {
	httpClient HTTPClient
	baseURL    string
	apiKey     string
	timeout    time.Duration
	logger     domain.Logger
}

// NewClient creates a new Perplexity API client
func NewClient(apiKey string, logger domain.Logger, opts ...ClientOption) (*Client, error) {
	if apiKey == "" {
		return nil, domain.ErrAPIKeyMissing
	}

	// Create secure HTTP client with TLS 1.2+ enforcement
	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
		MaxVersion: tls.VersionTLS13,
	}

	transport := &http.Transport{
		TLSClientConfig: tlsConfig,
	}

	client := &Client{
		httpClient: &http.Client{
			Transport: transport,
			Timeout:   DefaultTimeout,
		},
		baseURL: BaseURL,
		apiKey:  apiKey,
		timeout: DefaultTimeout,
		logger:  logger,
	}

	// Apply options
	for _, opt := range opts {
		opt(client)
	}

	return client, nil
}

// ClientOption defines options for configuring the client
type ClientOption func(*Client)

// WithHTTPClient sets a custom HTTP client
func WithHTTPClient(httpClient HTTPClient) ClientOption {
	return func(c *Client) {
		c.httpClient = httpClient
	}
}

// WithTimeout sets the request timeout
func WithTimeout(timeout time.Duration) ClientOption {
	return func(c *Client) {
		c.timeout = timeout
	}
}

// WithBaseURL sets a custom base URL
func WithBaseURL(baseURL string) ClientOption {
	return func(c *Client) {
		c.baseURL = baseURL
	}
}

// Search implements the PerplexityClient interface for search operations
func (c *Client) Search(ctx context.Context, request domain.SearchRequest) (*domain.SearchResult, error) {
	// Validate request
	if err := request.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrInvalidRequest, err)
	}

	// Set default model if not provided
	if request.Model == "" {
		request.Model = DefaultModel
	}

	// Convert to API request
	apiReq, err := SearchRequestToAPI(request, c.logger)
	if err != nil {
		return nil, fmt.Errorf("failed to convert search request: %w", err)
	}

	// Make API call
	apiResp, err := c.makeRequest(ctx, apiReq)
	if err != nil {
		return nil, err
	}

	// Convert response
	result := APIResponseToSearchResult(*apiResp)

	c.logger.Debug("Search completed", "request_id", result.ID, "tokens_used", result.Usage.TotalTokens)

	return &result, nil
}


// makeRequest handles the actual HTTP request to the Perplexity API
func (c *Client) makeRequest(ctx context.Context, apiReq APIChatRequest) (*APIChatResponse, error) {
	// Add timeout to context if not already set
	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, c.timeout)
		defer cancel()
	}

	// Marshal request body
	reqBody, err := json.Marshal(apiReq)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to marshal request: %v", domain.ErrInvalidRequest, err)
	}

	// Create HTTP request
	url := c.baseURL + ChatCompletionsEndpoint
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("%w: failed to create request: %v", domain.ErrNetworkError, err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)

	// Log request details (without sensitive data)
	c.logger.Debug("Making API request", "url", url, "model", apiReq.Model, "message_count", len(apiReq.Messages))

	// Make the request
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		// Check if it's a timeout error
		if ctx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("%w: request timed out", domain.ErrTimeout)
		}
		return nil, fmt.Errorf("%w: request failed: %v", domain.ErrNetworkError, err)
	}
	defer resp.Body.Close()

	// Read response body with size limit to prevent memory exhaustion
	limitedReader := io.LimitReader(resp.Body, MaxResponseSize)
	respBody, err := io.ReadAll(limitedReader)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to read response: %v", domain.ErrNetworkError, err)
	}

	// Check if response was truncated due to size limit
	if len(respBody) == MaxResponseSize {
		c.logger.Warn("Response body truncated due to size limit", "max_size", MaxResponseSize)
		return nil, fmt.Errorf("%w: response too large (exceeded %d bytes)", domain.ErrAPIError, MaxResponseSize)
	}

	// Handle non-200 status codes
	if resp.StatusCode != http.StatusOK {
		return nil, c.handleErrorResponse(resp.StatusCode, respBody)
	}

	// Parse successful response
	var apiResp APIChatResponse
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return nil, fmt.Errorf("%w: failed to parse response: %v", domain.ErrAPIError, err)
	}

	return &apiResp, nil
}

// handleErrorResponse processes error responses from the API
func (c *Client) handleErrorResponse(statusCode int, body []byte) error {
	// Try to parse as structured API error
	var apiError APIErrorResponse
	if err := json.Unmarshal(body, &apiError); err == nil {
		return c.mapAPIError(statusCode, apiError.Error.Error.Message)
	}

	// Fallback to status code mapping
	return c.mapStatusCodeError(statusCode, string(body))
}

// mapAPIError maps structured API errors to domain errors
func (c *Client) mapAPIError(statusCode int, message string) error {
	// Log error without exposing sensitive information
	c.logger.Error("API error", "status_code", statusCode, "error_type", "structured")

	switch statusCode {
	case http.StatusBadRequest:
		return fmt.Errorf("%w: %s", domain.ErrInvalidRequest, message)
	case http.StatusUnauthorized:
		return fmt.Errorf("%w: %s", domain.ErrAPIKeyMissing, message)
	case http.StatusTooManyRequests:
		return fmt.Errorf("%w: %s", domain.ErrRateLimited, message)
	case http.StatusInternalServerError, http.StatusBadGateway, http.StatusServiceUnavailable:
		return fmt.Errorf("%w: %s", domain.ErrAPIError, message)
	default:
		return fmt.Errorf("%w: HTTP %d: %s", domain.ErrAPIError, statusCode, message)
	}
}

// mapStatusCodeError maps HTTP status codes to domain errors when no structured error is available
func (c *Client) mapStatusCodeError(statusCode int, body string) error {
	// Log error without exposing response body which might contain sensitive data
	c.logger.Error("HTTP error", "status_code", statusCode, "body_length", len(body))

	switch statusCode {
	case http.StatusBadRequest:
		return fmt.Errorf("%w: bad request", domain.ErrInvalidRequest)
	case http.StatusUnauthorized:
		return fmt.Errorf("%w: unauthorized", domain.ErrAPIKeyMissing)
	case http.StatusTooManyRequests:
		return fmt.Errorf("%w: rate limited", domain.ErrRateLimited)
	case http.StatusInternalServerError, http.StatusBadGateway, http.StatusServiceUnavailable:
		return fmt.Errorf("%w: server error (HTTP %d)", domain.ErrAPIError, statusCode)
	default:
		return fmt.Errorf("%w: HTTP %d", domain.ErrAPIError, statusCode)
	}
}
