package internal

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	BaseURL                 = "https://api.perplexity.ai"
	ChatCompletionsEndpoint = "/chat/completions"
	DefaultTimeout          = 30 * time.Second
	DefaultModel            = "sonar"
	MaxResponseSize         = 10 * 1024 * 1024
)

type PerplexityClient struct {
	httpClient *http.Client
	apiKey     string
	baseURL    string
	logger     *log.Logger
}

func NewPerplexityClient(apiKey string) (*PerplexityClient, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("API key is required")
	}

	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
		MaxVersion: tls.VersionTLS13,
	}

	transport := &http.Transport{
		TLSClientConfig: tlsConfig,
	}

	return &PerplexityClient{
		httpClient: &http.Client{
			Transport: transport,
			Timeout:   DefaultTimeout,
		},
		apiKey:  apiKey,
		baseURL: BaseURL,
		logger:  log.New(os.Stdout, "[PERPLEXITY] ", log.LstdFlags|log.Lshortfile),
	}, nil
}

func (c *PerplexityClient) Search(ctx context.Context, req SearchRequest) (*SearchResult, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	if req.Model == "" {
		req.Model = DefaultModel
	}

	apiReq := c.searchToAPIRequest(req)
	apiResp, err := c.makeRequest(ctx, apiReq)
	if err != nil {
		return nil, err
	}

	result := c.apiToSearchResult(*apiResp)
	return &result, nil
}

func (c *PerplexityClient) searchToAPIRequest(req SearchRequest) APIChatRequest {
	apiReq := APIChatRequest{
		Model: req.Model,
		Messages: []APIMessage{
			{
				Role:    "user",
				Content: req.Query,
			},
		},
		Stream:     false,
		SearchMode: req.SearchMode,
	}

	if req.MaxTokens > 0 {
		apiReq.MaxTokens = &req.MaxTokens
	}

	if len(req.Sources) > 0 {
		apiReq.SearchDomainFilter = req.Sources
	}

	// Process options
	for key, value := range req.Options {
		switch strings.ToLower(key) {
		case "temperature":
			if temp, err := strconv.ParseFloat(value, 64); err == nil && temp >= 0 && temp <= 2.0 {
				apiReq.Temperature = &temp
			}
		case "top_p":
			if topP, err := strconv.ParseFloat(value, 64); err == nil && topP >= 0 && topP <= 1.0 {
				apiReq.TopP = &topP
			}
		case "disable_search":
			if disable, err := strconv.ParseBool(value); err == nil {
				apiReq.DisableSearch = &disable
			}
		}
	}

	return apiReq
}

func (c *PerplexityClient) apiToSearchResult(apiResp APIChatResponse) SearchResult {
	result := SearchResult{
		ID:      apiResp.ID,
		Content: apiResp.GetContent(),
		Model:   apiResp.Model,
		Usage: Usage{
			PromptTokens:     apiResp.Usage.PromptTokens,
			CompletionTokens: apiResp.Usage.CompletionTokens,
			TotalTokens:      apiResp.Usage.TotalTokens,
		},
		Created: apiResp.GetCreatedTime(),
	}

	if len(apiResp.Citations) > 0 {
		result.Citations = make([]Citation, len(apiResp.Citations))
		copy(result.Citations, apiResp.Citations)
	}

	if len(apiResp.Sources) > 0 {
		result.Sources = make([]Source, len(apiResp.Sources))
		copy(result.Sources, apiResp.Sources)
	}

	return result
}

func (c *PerplexityClient) makeRequest(ctx context.Context, apiReq APIChatRequest) (*APIChatResponse, error) {
	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, DefaultTimeout)
		defer cancel()
	}

	reqBody, err := json.Marshal(apiReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := c.baseURL + ChatCompletionsEndpoint
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)

	c.logger.Printf("Making API request to %s with model %s", url, apiReq.Model)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("request timed out")
		}
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	limitedReader := io.LimitReader(resp.Body, MaxResponseSize)
	respBody, err := io.ReadAll(limitedReader)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, c.handleErrorResponse(resp.StatusCode, respBody)
	}

	var apiResp APIChatResponse
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &apiResp, nil
}

func (c *PerplexityClient) handleErrorResponse(statusCode int, body []byte) error {
	var apiError APIErrorResponse
	if err := json.Unmarshal(body, &apiError); err == nil {
		return c.mapAPIError(statusCode, apiError.Error.Error.Message)
	}

	return c.mapStatusCodeError(statusCode, string(body))
}

func (c *PerplexityClient) mapAPIError(statusCode int, message string) error {
	c.logger.Printf("API error: status=%d, message=%s", statusCode, message)

	switch statusCode {
	case http.StatusBadRequest:
		return fmt.Errorf("bad request: %s", message)
	case http.StatusUnauthorized:
		return fmt.Errorf("unauthorized: %s", message)
	case http.StatusTooManyRequests:
		return fmt.Errorf("rate limited: %s", message)
	case http.StatusInternalServerError, http.StatusBadGateway, http.StatusServiceUnavailable:
		return fmt.Errorf("server error: %s", message)
	default:
		return fmt.Errorf("API error (HTTP %d): %s", statusCode, message)
	}
}

func (c *PerplexityClient) mapStatusCodeError(statusCode int, body string) error {
	c.logger.Printf("HTTP error: status=%d, body_length=%d", statusCode, len(body))

	switch statusCode {
	case http.StatusBadRequest:
		return fmt.Errorf("bad request")
	case http.StatusUnauthorized:
		return fmt.Errorf("unauthorized")
	case http.StatusTooManyRequests:
		return fmt.Errorf("rate limited")
	case http.StatusInternalServerError, http.StatusBadGateway, http.StatusServiceUnavailable:
		return fmt.Errorf("server error (HTTP %d)", statusCode)
	default:
		return fmt.Errorf("HTTP error %d", statusCode)
	}
}
