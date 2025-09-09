package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/mark3labs/mcp-go/mcp"
)

// CreatePerplexitySearchTool creates the perplexity_search tool for use with mcp-go
func CreatePerplexitySearchTool(client *PerplexityClient) mcp.Tool {
	return mcp.Tool{
		Name:        "perplexity_search",
		Description: "Search for information using Perplexity AI Sonar models. Provides real-time web search with citations and sources, supporting academic search, news search, and domain filtering.",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
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
			Required: []string{"query"},
		},
	}
}

// PerplexitySearchHandler creates the handler function for the perplexity_search tool
func PerplexitySearchHandler(client *PerplexityClient) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Parse the search request
		req, err := parseSearchRequestFromMCP(request)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					mcp.TextContent{
						Type: "text",
						Text: fmt.Sprintf("Invalid search request: %s", err.Error()),
					},
				},
				IsError: true,
			}, err
		}

		// Execute search using the Perplexity client
		result, err := client.Search(ctx, *req)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					mcp.TextContent{
						Type: "text",
						Text: fmt.Sprintf("Search failed: %s", err.Error()),
					},
				},
				IsError: true,
			}, err
		}

		// Format the result
		content, err := formatSearchResultForMCP(result)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					mcp.TextContent{
						Type: "text",
						Text: fmt.Sprintf("Failed to format result: %s", err.Error()),
					},
				},
				IsError: true,
			}, err
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.TextContent{
					Type: "text",
					Text: content,
				},
			},
			IsError: false,
		}, nil
	}
}

// parseSearchRequestFromMCP converts mcp.CallToolRequest to internal SearchRequest
func parseSearchRequestFromMCP(request mcp.CallToolRequest) (*SearchRequest, error) {
	query, err := request.RequireString("query")
	if err != nil {
		return nil, fmt.Errorf("query must be a string")
	}

	req := &SearchRequest{
		Query: query,
	}

	// Optional model parameter
	if model := request.GetString("model", ""); model != "" {
		req.Model = model
	}

	// Optional search_mode parameter
	if searchMode := request.GetString("search_mode", ""); searchMode != "" {
		req.SearchMode = searchMode
	}

	// Optional max_tokens parameter
	if maxTokensStr := request.GetString("max_tokens", ""); maxTokensStr != "" {
		if maxTokens, err := strconv.Atoi(maxTokensStr); err == nil {
			req.MaxTokens = maxTokens
		} else {
			return nil, fmt.Errorf("max_tokens must be a valid number")
		}
	}

	// Optional date_range parameter
	if dateRange := request.GetString("date_range", ""); dateRange != "" {
		req.DateRange = dateRange
	}

	// Optional sources parameter
	if sources := request.GetStringSlice("sources", nil); sources != nil {
		req.Sources = sources
	}

	// Optional options parameter - use BindArguments for complex objects
	var optionsMap map[string]string
	if args := request.GetArguments(); args != nil {
		if optionsRaw, exists := args["options"]; exists {
			if optionsData, ok := optionsRaw.(map[string]any); ok {
				options := make(map[string]string)
				for key, value := range optionsData {
					if strValue, ok := value.(string); ok {
						options[key] = strValue
					} else {
						return nil, fmt.Errorf("option '%s' must be a string", key)
					}
				}
				optionsMap = options
			} else {
				return nil, fmt.Errorf("options must be an object")
			}
		}
	}
	if optionsMap != nil {
		req.Options = optionsMap
	}

	return req, nil
}

// formatSearchResultForMCP formats SearchResult as JSON string for MCP response
func formatSearchResultForMCP(result *SearchResult) (string, error) {
	response := map[string]any{
		"id":      result.ID,
		"content": result.Content,
		"model":   result.Model,
		"usage":   result.Usage,
		"created": result.Created,
	}

	if len(result.Citations) > 0 {
		response["citations"] = result.Citations
	}

	if len(result.Sources) > 0 {
		response["sources"] = result.Sources
	}

	jsonBytes, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal search result: %w", err)
	}

	return string(jsonBytes), nil
}