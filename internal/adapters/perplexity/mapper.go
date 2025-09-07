package perplexity

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/yourusername/perplexity-mcp-golang/internal/core/domain"
)

// SearchRequestToAPI converts domain.SearchRequest to APIChatRequest
func SearchRequestToAPI(req domain.SearchRequest, logger domain.Logger) (APIChatRequest, error) {
	apiReq := APIChatRequest{
		Model: req.Model,
		Messages: []APIMessage{
			{
				Role:    "user",
				Content: req.Query,
			},
		},
		SearchMode: req.SearchMode,
	}

	// Set max_tokens if specified
	if req.MaxTokens > 0 {
		apiReq.MaxTokens = &req.MaxTokens
	}

	// Handle domain filtering from Sources
	if len(req.Sources) > 0 {
		apiReq.SearchDomainFilter = req.Sources
	}

	// Process options (can override domain filtering and other settings)
	if req.Options != nil {
		if err := processSearchOptions(&apiReq, req.Options, logger); err != nil {
			return APIChatRequest{}, fmt.Errorf("failed to process search options: %w", err)
		}
	}

	return apiReq, nil
}


// APIResponseToSearchResult converts APIChatResponse to domain.SearchResult
func APIResponseToSearchResult(apiResp APIChatResponse) domain.SearchResult {
	result := domain.SearchResult{
		ID:      apiResp.ID,
		Content: apiResp.GetContent(),
		Model:   apiResp.Model,
		Usage: domain.Usage{
			PromptTokens:     apiResp.Usage.PromptTokens,
			CompletionTokens: apiResp.Usage.CompletionTokens,
			TotalTokens:      apiResp.Usage.TotalTokens,
		},
		Created: apiResp.GetCreatedTime(),
	}

	// Convert citations
	if len(apiResp.Citations) > 0 {
		result.Citations = make([]domain.Citation, len(apiResp.Citations))
		for i, citation := range apiResp.Citations {
			result.Citations[i] = domain.Citation{
				Number: citation.Number,
				URL:    citation.URL,
				Title:  citation.Title,
			}
		}
	}

	// Convert sources
	if len(apiResp.Sources) > 0 {
		result.Sources = make([]domain.Source, len(apiResp.Sources))
		for i, source := range apiResp.Sources {
			result.Sources[i] = domain.Source{
				URL:     source.URL,
				Title:   source.Title,
				Snippet: source.Snippet,
			}
		}
	}

	return result
}


// Helper functions to process options

func processSearchOptions(apiReq *APIChatRequest, options map[string]string, logger domain.Logger) error {
	for key, value := range options {
		// Validate key and value lengths with logging
		if len(key) > domain.MaxOptionKeyLength {
			logger.Warn("Option key too long, skipping", "key", key, "length", len(key), "max", domain.MaxOptionKeyLength)
			continue
		}
		if len(value) > domain.MaxOptionValueLength {
			return fmt.Errorf("option value for key '%s' is too long: %d > %d", key, len(value), domain.MaxOptionValueLength)
		}

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
		case "search_domain_filter":
			// Resource exhaustion protection - check total length before split
			const maxDomainFilterLength = 3000 // 10 domains * 253 chars + separators
			if len(value) > maxDomainFilterLength {
				return fmt.Errorf("search_domain_filter value too long: %d > %d", len(value), maxDomainFilterLength)
			}
			
			// Expecting comma-separated domains
			domains := strings.Split(value, ",")
			validDomains := make([]string, 0, len(domains))
			for _, domain := range domains {
				domain = strings.TrimSpace(domain)
				if len(domain) == 0 {
					continue
				}
				if len(domain) > 253 { // Max domain length per RFC
					logger.Warn("Domain too long, skipping", "domain", domain, "length", len(domain))
					continue
				}
				// Basic domain format validation
				if strings.Contains(domain, "/") || strings.Contains(domain, " ") {
					logger.Warn("Invalid domain format, skipping", "domain", domain)
					continue
				}
				validDomains = append(validDomains, domain)
			}
			if len(validDomains) > 10 {
				return fmt.Errorf("too many domains in search_domain_filter: %d > 10", len(validDomains))
			}
			if len(validDomains) > 0 {
				apiReq.SearchDomainFilter = validDomains
			}
		case "search_mode":
			// Validate search mode against allowed values
			validSearchModes := map[string]bool{
				"web":      true,
				"academic": true,
				"news":     true,
			}
			if validSearchModes[value] {
				apiReq.SearchMode = value
			} else {
				logger.Warn("Invalid search mode, ignoring", "mode", value, "valid_modes", []string{"web", "academic", "news"})
			}
		default:
			logger.Warn("Unknown search option key, ignoring", "key", key)
		}
	}
	return nil
}

