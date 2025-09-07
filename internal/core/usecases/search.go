package usecases

import (
	"context"
	"fmt"
	"time"

	"github.com/yourusername/perplexity-mcp-golang/internal/core/domain"
)

// SearchUseCase orchestrates search operations using the Perplexity API
type SearchUseCase struct {
	client         domain.PerplexityClient
	config         domain.ConfigProvider
	logger         domain.Logger
	requestTimeout time.Duration
}

// NewSearchUseCase creates a new search use case instance
func NewSearchUseCase(
	client domain.PerplexityClient,
	config domain.ConfigProvider,
	logger domain.Logger,
) *SearchUseCase {
	timeout := time.Duration(config.GetRequestTimeout()) * time.Second

	return &SearchUseCase{
		client:         client,
		config:         config,
		logger:         logger,
		requestTimeout: timeout,
	}
}

// Execute performs a search operation with validation and error handling
func (uc *SearchUseCase) Execute(ctx context.Context, request domain.SearchRequest) (*domain.SearchResult, error) {
	// Add timeout to context if not already set
	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, uc.requestTimeout)
		defer cancel()
	}

	// Log the start of the operation
	uc.logger.Info("Starting search operation",
		"query", request.Query,
		"model", request.Model,
		"search_mode", request.SearchMode,
		"date_range", request.DateRange,
		"sources_count", len(request.Sources),
	)

	// Validate the request
	if err := request.Validate(); err != nil {
		uc.logger.Error("Search request validation failed",
			"error", err.Error(),
			"query", request.Query,
		)
		return nil, fmt.Errorf("search request validation failed: %w", err)
	}

	// Apply default model if not specified
	if request.Model == "" {
		request.Model = uc.config.GetDefaultModel()
		uc.logger.Debug("Applied default model to search request",
			"model", request.Model,
		)
	}

	// Execute the search through the client
	uc.logger.Debug("Executing search through Perplexity client")
	result, err := uc.client.Search(ctx, request)
	if err != nil {
		uc.logger.Error("Search operation failed",
			"error", err.Error(),
			"query", request.Query,
			"model", request.Model,
		)
		return nil, fmt.Errorf("search operation failed: %w", err)
	}

	// Log successful completion
	uc.logger.Info("Search operation completed successfully",
		"result_id", result.ID,
		"model", result.Model,
		"usage_total_tokens", result.Usage.TotalTokens,
		"citations_count", len(result.Citations),
		"sources_count", len(result.Sources),
	)

	return result, nil
}

// ValidateRequest validates a search request without executing it
func (uc *SearchUseCase) ValidateRequest(request domain.SearchRequest) error {
	uc.logger.Debug("Validating search request",
		"query", request.Query,
		"model", request.Model,
	)

	if err := request.Validate(); err != nil {
		uc.logger.Debug("Search request validation failed",
			"error", err.Error(),
		)
		return err
	}

	uc.logger.Debug("Search request validation passed")
	return nil
}
