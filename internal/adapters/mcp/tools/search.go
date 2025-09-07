package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/yourusername/perplexity-mcp-golang/internal/core/domain"
)

// SearchTool implements the MCP tool interface for Perplexity search operations
type SearchTool struct {
	useCase SearchUseCaseInterface
	logger  domain.Logger
}

// SearchUseCaseInterface defines the contract for search use case
type SearchUseCaseInterface interface {
	Execute(ctx context.Context, request domain.SearchRequest) (*domain.SearchResult, error)
	ValidateRequest(request domain.SearchRequest) error
}

// NewSearchTool creates a new search tool instance
func NewSearchTool(useCase SearchUseCaseInterface, logger domain.Logger) *SearchTool {
	return &SearchTool{
		useCase: useCase,
		logger:  logger,
	}
}

// Name returns the name of the tool
func (t *SearchTool) Name() string {
	return "perplexity_search"
}

// Description returns the description of the tool
func (t *SearchTool) Description() string {
	return "Search for information using Perplexity AI Sonar models. Provides real-time web search with citations and sources, supporting academic search, news search, and domain filtering."
}

// InputSchema returns the JSON schema for the tool's input parameters
func (t *SearchTool) InputSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"query": map[string]any{
				"type":        "string",
				"description": "The search query to execute",
				"minLength":   1,
				"maxLength":   10000,
			},
			"model": map[string]any{
				"type":        "string",
				"description": "The Sonar model to use for search (optional, defaults to 'sonar')",
				"enum": []string{
					"sonar",
					"sonar-pro",
					"sonar-reasoning",
					"sonar-reasoning-pro",
					"sonar-deep-research",
				},
				"default": "sonar",
			},
			"search_mode": map[string]any{
				"type":        "string",
				"description": "The search mode to use (optional, defaults to 'web')",
				"enum":        []string{"web", "academic", "news"},
				"default":     "web",
			},
			"max_tokens": map[string]any{
				"type":        "number",
				"description": "Maximum number of tokens in the response (optional)",
				"minimum":     1,
				"maximum":     128000,
			},
			"date_range": map[string]any{
				"type":        "string",
				"description": "Filter search results by date range (optional)",
				"enum":        []string{"day", "week", "month", "year"},
			},
			"sources": map[string]any{
				"type":        "array",
				"description": "Limit search to specific domains (optional, max 10)",
				"items": map[string]any{
					"type": "string",
				},
				"maxItems": 10,
			},
			"options": map[string]any{
				"type":        "object",
				"description": "Additional search options (optional)",
				"additionalProperties": map[string]any{
					"type": "string",
				},
			},
		},
		"required":             []string{"query"},
		"additionalProperties": false,
	}
}

// Execute performs the search operation
func (t *SearchTool) Execute(ctx context.Context, args map[string]any) (*domain.ToolResult, error) {
	t.logger.Debug("Executing search tool", "args_count", len(args))

	// Parse and validate arguments
	request, err := t.parseSearchRequest(args)
	if err != nil {
		t.logger.Error("Failed to parse search request", "error", err.Error())
		return &domain.ToolResult{
			Content: fmt.Sprintf("Invalid search request: %s", err.Error()),
			IsError: true,
			Metadata: map[string]any{
				"error_type": "validation_error",
			},
		}, err
	}

	// Execute the search through the use case
	result, err := t.useCase.Execute(ctx, *request)
	if err != nil {
		t.logger.Error("Search execution failed", "error", err.Error(), "query", request.Query)
		return &domain.ToolResult{
			Content: fmt.Sprintf("Search failed: %s", err.Error()),
			IsError: true,
			Metadata: map[string]any{
				"error_type": "execution_error",
				"query":      request.Query,
				"model":      request.Model,
			},
		}, err
	}

	// Format the successful result
	content, err := t.formatSearchResult(result)
	if err != nil {
		t.logger.Error("Failed to format search result", "error", err.Error())
		return &domain.ToolResult{
			Content: fmt.Sprintf("Failed to format search result: %s", err.Error()),
			IsError: true,
		}, err
	}

	t.logger.Info("Search tool execution completed successfully",
		"result_id", result.ID,
		"content_length", len(result.Content),
		"citations_count", len(result.Citations),
		"sources_count", len(result.Sources),
	)

	return &domain.ToolResult{
		Content:   content,
		IsError:   false,
		Citations: result.Citations,
		Metadata: map[string]any{
			"result_id":       result.ID,
			"model":           result.Model,
			"usage":           result.Usage,
			"created":         result.Created,
			"sources_count":   len(result.Sources),
			"citations_count": len(result.Citations),
		},
	}, nil
}

// parseSearchRequest converts the raw arguments into a domain SearchRequest
func (t *SearchTool) parseSearchRequest(args map[string]any) (*domain.SearchRequest, error) {
	// Extract query (required)
	query, ok := args["query"].(string)
	if !ok {
		return nil, fmt.Errorf("%w: query must be a string", domain.ErrInvalidRequest)
	}

	request := &domain.SearchRequest{
		Query: query,
	}

	// Extract optional model
	if model, ok := args["model"].(string); ok {
		request.Model = model
	}

	// Extract optional search mode
	if searchMode, ok := args["search_mode"].(string); ok {
		request.SearchMode = searchMode
	}

	// Extract optional max_tokens
	if maxTokensRaw, exists := args["max_tokens"]; exists {
		if maxTokens, ok := maxTokensRaw.(float64); ok {
			request.MaxTokens = int(maxTokens)
		} else {
			return nil, fmt.Errorf("%w: max_tokens must be a number", domain.ErrInvalidRequest)
		}
	}

	// Extract optional date_range
	if dateRange, ok := args["date_range"].(string); ok {
		request.DateRange = dateRange
	}

	// Extract optional sources
	if sourcesRaw, ok := args["sources"]; ok {
		if sourcesList, ok := sourcesRaw.([]any); ok {
			sources := make([]string, 0, len(sourcesList))
			for i, source := range sourcesList {
				if strSource, ok := source.(string); ok {
					sources = append(sources, strSource)
				} else {
					return nil, fmt.Errorf("%w: source[%d] must be a string", domain.ErrInvalidRequest, i)
				}
			}
			request.Sources = sources
		} else {
			return nil, fmt.Errorf("%w: sources must be an array", domain.ErrInvalidRequest)
		}
	}

	// Extract optional options
	if optionsRaw, ok := args["options"]; ok {
		if optionsMap, ok := optionsRaw.(map[string]any); ok {
			options := make(map[string]string)
			for key, value := range optionsMap {
				if strValue, ok := value.(string); ok {
					options[key] = strValue
				} else {
					return nil, fmt.Errorf("%w: option '%s' must be a string", domain.ErrInvalidRequest, key)
				}
			}
			request.Options = options
		} else {
			return nil, fmt.Errorf("%w: options must be an object", domain.ErrInvalidRequest)
		}
	}

	// Validate the request using domain validation
	if err := request.Validate(); err != nil {
		return nil, fmt.Errorf("request validation failed: %w", err)
	}

	return request, nil
}

// formatSearchResult formats the search result for MCP response
func (t *SearchTool) formatSearchResult(result *domain.SearchResult) (string, error) {
	// Create a structured response
	response := map[string]any{
		"id":      result.ID,
		"content": result.Content,
		"model":   result.Model,
		"usage":   result.Usage,
		"created": result.Created,
	}

	// Add citations if available
	if len(result.Citations) > 0 {
		response["citations"] = result.Citations
	}

	// Add sources if available
	if len(result.Sources) > 0 {
		response["sources"] = result.Sources
	}

	// Convert to JSON string
	jsonBytes, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal search result: %w", err)
	}

	return string(jsonBytes), nil
}
